# AWS Lambda Histogram Counter with Go
This Lambda function processes a string input to calculate the frequency of alphabets using concurrency for efficient execution and interacts with AWS MemoryDB for storing and retrieving the results.

## Overview
1. **Concurrency**: Input strings are segmented into parts of 100 characters each. Each segment is then processed by an individual goroutine (lightweight thread in Go). This ensures faster execution as each segment is processed concurrently.

2. **Redis Optimized Calls**: Instead of making multiple calls to fetch or store individual characters, the function leverages batch commands to make fewer calls to the Redis backend (AWS MemoryDB). This optimizes the overall performance and reduces the latency associated with multiple database operations.

## How It Works
## Input Handling
The function is designed to accept a JSON payload with the following structure:

```json
{
    "word": "your_input_string_here"
}
```

The word key should have a string value that you want to get the frequency for.

## Concurrency Implementation
On receiving the input string:

The input string is divided into segments of 100 characters each.
Each segment is processed by an individual goroutine.
The character frequencies calculated by each goroutine are combined to form the overall character frequency for the input string.
Redis Integration
To ensure efficient read and write operations with Redis:

**For reading**: The MGet command is used to fetch the frequencies of all alphabets in a single call.

**For writing**: The MSet command is used, which allows setting multiple keys and their corresponding values in one go.

**The Redis URL is configured to use rediss:// (Redis over SSL) for a secure connection to the AWS MemoryDB cluster.**

## Future Improvements
**Limit on Goroutines**: The current system spawns a goroutine for every 100 characters without an upper limit. An enhancement would be to introduce an upper limit on the number of goroutines that can be spawned to ensure system stability.

**Error Handling**: The current implementation could be further enhanced with more robust error handling, especially for Redis operations.
