package main

import (
	"fmt"
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var rdb *redis.Client
var characterFrequency = make(map[string]int)


// Initialising the redis client : used rediss for extra layer of security
func init() {
	conf, err := redis.ParseURL("rediss://clustercfg.counter.n5gdlu.memorydb.us-west-2.amazonaws.com:6379")
	if err != nil {
	   panic(err)
	}

	rdb = redis.NewClient(conf)
}

// Request payload 
// {"word":"kognitos"}
type Request struct {
	Word interface{} `json:"word"`
}

func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req Request
	err := json.Unmarshal([]byte(request.Body), &req)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Invalid JSON input",
			StatusCode: 400,
		}, nil
	}

	// validation to only accept words as input
	switch word := req.Word.(type) {
	case string:
		processWord(word)
	default:
		return events.APIGatewayProxyResponse{
			Body:       "Invalid word format",
			StatusCode: 400,
		}, nil
	}

	responseBody, err := json.Marshal(characterFrequency)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Error creating response",
			StatusCode: 500,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		Body:       string(responseBody),
		StatusCode: 200,
	}, nil
}

func processWord(word string) {
	word = strings.ToLower(word)
	var wg sync.WaitGroup
	ch := make(chan map[string]int)

	// splitting the input word in segment of 100 characters each 
	// each segment is being handled by a separate goroutine(thread) to scale the system 
	// have not kept any upperbound on the number of threads
	// ToDo: To keep an upperbound of number of threads
	for i := 0; i < len(word); i += 100 {
		end := i + 100
		if end > len(word) {
			end = len(word)
		}

		subsegment := word[i:end]

		wg.Add(1)
		go func(subsegment string) {
			defer wg.Done()

			localFreq := make(map[string]int)
			for _, char := range subsegment {
				lowerChar := strings.ToLower(string(char))
				if 'a' <= char && char <= 'z' {
					localFreq[lowerChar]++
				}
			}

			ch <- localFreq
		}(subsegment)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	// Combine the results from the goroutines
	for freqMap := range ch {
		for char, freq := range freqMap {
			characterFrequency[char] += freq
		}
	}

	// Fetch current counts from MemoryDB
	frequencies, err := rdb.MGet(ctx, "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z").Result() 
	if err != nil {
		fmt.Println("error getting data from redis", err)
	}

	// Modify counts based on fetched values
	for i, char := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"} {
	    if i < len(frequencies) && frequencies[i] != nil {
	        count, err := frequencies[i].(int64)
	        if !err  {
	            characterFrequency[char] += int(count)
	        }
	    }
	}

	// Store updated counts in MemoryDB
	pairs := make([]string, 0)
	for char, freq := range characterFrequency {
		pairs = append(pairs, char, fmt.Sprint(freq))  
	}

	
	ifacePairs := make([]interface{}, len(pairs))
	for i, v := range pairs {
		ifacePairs[i] = v
	}

	err = rdb.MSet(ctx, ifacePairs...).Err()
	if err != nil {
		fmt.Println("error updating data to redis")
	}
}

func main() {
	lambda.Start(HandleRequest)
}
