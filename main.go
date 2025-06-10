package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alilxxey/dnn.vk_worker_pool/internal/workerpool"
)

type RequestData struct {
	A int `json:"a"`
	B int `json:"b"`
}

type ResponseData struct {
	Result any    `json:"result"`
	Error  string `json:"error,omitempty"`
}

func main() {
	initial := flag.Int("initial", 5, "начальное количество воркеров")
	maxW := flag.Int("max", 10, "максимальное количество воркеров")
	buf := flag.Int("buffer", 100, "размер буфера задач")
	addr := flag.String("addr", ":8080", "адрес HTTP-сервера")
	flag.Parse()

	pool := workerpool.NewWorkerPool[int](*initial, *maxW, *buf)
	defer pool.Shutdown()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RequestData
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Println("ERROR decoding JSON:", err)
			http.Error(w, "Bad Request: invalid JSON", http.StatusBadRequest)
			return
		}
		log.Printf("Received task: %d + %d", req.A, req.B)

		res, err := pool.SubmitTask(func() (int, error) {
			time.Sleep(3 * time.Second) // имитируем полезную нагрузку
			return req.A + req.B, nil
		})

		var resp ResponseData
		if err != nil {
			resp.Error = err.Error()
		} else {
			resp.Result = res
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	srv := &http.Server{
		Addr:         *addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	log.Printf("Server listening on %s", *addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server error:", err)
	}
}
