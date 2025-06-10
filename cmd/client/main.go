package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

type RequestData struct {
	A int `json:"a"`
	B int `json:"b"`
}

type ResponseData struct {
	Result int   `json:"result"`
	Error  error `json:"error"`
}

func main() {
	totalRequests := 20

	var wg sync.WaitGroup
	wg.Add(totalRequests)

	start := time.Now()

	for i := 1; i <= totalRequests; i++ {
		go func(i int) {
			defer wg.Done()
			reqBody := RequestData{A: i, B: 2 * i}

			data, err := json.Marshal(reqBody)

			if err != nil {
				log.Println("ERROR: Failed to marshal request:", err)
				return
			}
			resp, err := http.Post("http://localhost:8080/", "application/json", bytes.NewBuffer(data))
			if err != nil {
				log.Println("ERROR: Request failed:", err)
				return
			}
			defer resp.Body.Close()

			var resBody ResponseData
			if err := json.NewDecoder(resp.Body).Decode(&resBody); err != nil {
				log.Println("ERROR: Failed to parse response:", err)
				return
			}
			log.Printf("Request %d: %d + %d = %d (response received)", i, reqBody.A, reqBody.B, resBody.Result)
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	log.Printf("All %d requests completed in %s\n", totalRequests, elapsed)
}
