package main

import (
	"encoding/json"
	"github.com/alilxxey/dnn.vk_worker_pool/internal/workerpool"
	"log"
	"net/http"
	"time"
)

type RequestData struct {
	A int `json:"a"` // first number
	B int `json:"b"` // second number
}

type ResponseData struct {
	Result int   `json:"result"` // the sum of A and B
	Error  error `json:"error"`
}

func main() {
	pool := workerpool.NewWorkerPool(5, 10, 100)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		var reqData RequestData
		err := json.NewDecoder(r.Body).Decode(&reqData)
		if err != nil {
			log.Println("ERROR: ", err)
			http.Error(w, "Bad Request: invalid JSON", http.StatusBadRequest)
			return
		}
		log.Printf("Received task: %d + %d\n", reqData.A, reqData.B)
		var result int
		err = pool.SubmitTask(func() error {
			time.Sleep(2 * time.Second)
			result = reqData.A + reqData.B

			return nil
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ResponseData{Result: result, Error: err})
	})

	log.Println("Server is running on http://localhost:8080/")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server error:", err)
	}

}
