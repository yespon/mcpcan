package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := "9999"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		headers := make(map[string]string)
		for k, v := range r.Header {
			headers[k] = v[0]
		}

		resp := map[string]interface{}{
			"headers": headers,
			"method":  r.Method,
			"url":     r.URL.String(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		
		fmt.Printf("Received request: %s %s\n", r.Method, r.URL.Path)
		for k, v := range headers {
			fmt.Printf("  %s: %s\n", k, v)
		}
	})

	fmt.Printf("Mock upstream service listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
