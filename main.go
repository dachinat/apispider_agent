package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"time"
	"strings"
	"github.com/common-nighthawk/go-figure"
)

// RequestPayload matches the frontend request structure
type RequestPayload struct {
	Method   string                   `json:"method"`
	URL      string                   `json:"url"`
	Headers  map[string]string        `json:"headers"`
	Body     string                   `json:"body"`
	AuthType string                   `json:"authType"`
	AuthData map[string]interface{}   `json:"authData"`
	FormData []map[string]interface{} `json:"formData"`
}

// ResponsePayload matches the backend response structure
type ResponsePayload struct {
	Status     int               `json:"status"`
	StatusText string            `json:"status_text"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Duration   int64             `json:"duration"`
}

func enableCORS(w http.ResponseWriter) {
	// Allow requests from the frontend (adjust origin as needed)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	// Required for Private Network Access (browser security feature)
	// when accessing localhost from a public origin
	w.Header().Set("Access-Control-Allow-Private-Network", "true")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"agent":   "ApiSpider Local Agent",
		"version": "1.0.0",
	})
}

func calculateAuthHeaders(reqPayload RequestPayload) map[string]string {
	headers := make(map[string]string)

	// Start with existing headers
	for k, v := range reqPayload.Headers {
		headers[k] = v
	}

	switch reqPayload.AuthType {
	case "basic":
		if username, ok := reqPayload.AuthData["username"].(string); ok {
			if password, ok := reqPayload.AuthData["password"].(string); ok {
				auth := username + ":" + password
				headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
			}
		}
	case "bearer":
		if token, ok := reqPayload.AuthData["token"].(string); ok {
			headers["Authorization"] = "Bearer " + token
		}
	case "oauth2":
		if accessToken, ok := reqPayload.AuthData["accessToken"].(string); ok {
			tokenType, _ := reqPayload.AuthData["tokenType"].(string)
			if tokenType == "" {
				tokenType = "Bearer"
			}
			headers["Authorization"] = tokenType + " " + accessToken
		}
	case "api-key":
		if key, ok := reqPayload.AuthData["key"].(string); ok {
			if value, ok := reqPayload.AuthData["value"].(string); ok {
				addTo, _ := reqPayload.AuthData["addTo"].(string)
				if addTo == "header" || addTo == "" { // Default to header
					headers[key] = value
				}
			}
		}
	}

	return headers
}

func executeHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request payload
	var reqPayload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Printf("Executing request: %s %s (Auth: %s)", reqPayload.Method, reqPayload.URL, reqPayload.AuthType)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	executeRequest := func(headers map[string]string) (*http.Response, int64, error) {
		startTime := time.Now()

		var requestBody io.Reader
		contentType := headers["Content-Type"]

		if len(reqPayload.FormData) > 0 {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			for _, item := range reqPayload.FormData {
				key, _ := item["key"].(string)
				value, _ := item["value"].(string)
				typeStr, _ := item["type"].(string)

				if key != "" {
					if typeStr == "file" {
						fileName, _ := item["fileName"].(string)
						if fileName == "" {
							fileName = "file"
						}
						part, err := writer.CreateFormFile(key, fileName)
						if err == nil {
							if value != "" {
								// Value is base64 encoded content
								decoded, decodeErr := base64.StdEncoding.DecodeString(value)
								if decodeErr == nil {
									part.Write(decoded)
								} else {
									part.Write([]byte(value))
								}
							}
						}
					} else {
						writer.WriteField(key, value)
					}
				}
			}
			writer.Close()
			requestBody = body
			contentType = writer.FormDataContentType()
		} else if reqPayload.Body != "" && strings.Contains(headers["Content-Type"], "application/octet-stream") {
			// Treat as binary if Content-Type is application/octet-stream
			decoded, err := base64.StdEncoding.DecodeString(reqPayload.Body)
			if err == nil {
				requestBody = bytes.NewBuffer(decoded)
			} else {
				requestBody = bytes.NewBufferString(reqPayload.Body)
			}
		} else {
			requestBody = bytes.NewBufferString(reqPayload.Body)
		}

		req, err := http.NewRequest(reqPayload.Method, reqPayload.URL, requestBody)
		if err != nil {
			return nil, 0, err
		}

		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}

		for key, value := range headers {
			if key != "Accept-Encoding" {
				req.Header.Set(key, value)
			}
		}

		// Handle API Key in query if needed
		if reqPayload.AuthType == "api-key" {
			if addTo, ok := reqPayload.AuthData["addTo"].(string); ok && addTo == "query" {
				q := req.URL.Query()
				if key, ok := reqPayload.AuthData["key"].(string); ok {
					if value, ok := reqPayload.AuthData["value"].(string); ok {
						q.Add(key, value)
						req.URL.RawQuery = q.Encode()
					}
				}
			}
		}

		resp, err := client.Do(req)
		duration := time.Since(startTime).Milliseconds()
		return resp, duration, err
	}

	// Calculate initial headers
	finalHeaders := calculateAuthHeaders(reqPayload)

	// First execution
	resp, duration, err := executeRequest(finalHeaders)
	if err != nil {
		respondWithError(w, "Request failed", err)
		return
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		respondWithError(w, "Failed to read response", err)
		return
	}

	// Build response headers map
	respHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			respHeaders[key] = values[0]
		}
	}

	// Create response payload
	response := ResponsePayload{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Headers:    respHeaders,
		Body:       string(bodyBytes),
		Duration:   duration,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("Request completed: %d %s (%dms)", resp.StatusCode, resp.Status, duration)
}

func respondWithError(w http.ResponseWriter, message string, err error) {
	log.Printf("Error: %s - %v", message, err)

	response := ResponsePayload{
		Status:     0,
		StatusText: "Error",
		Headers:    make(map[string]string),
		Body:       message + ": " + err.Error(),
		Duration:   0,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Return 200 but with error in body
	json.NewEncoder(w).Encode(response)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	text := `ApiSpider Local Agent - ONLINE

Version: 1.0.0
Status:  Running on http://localhost:8889

This agent allows you to make requests to localhost and private network APIs
from apispider.com, bypassing browser CORS restrictions.

Available Endpoints:
  GET  /health  - Health check
  POST /execute - Execute HTTP request

You can now use https://apispider.com to query your local APIs!
`
	w.Write([]byte(text))
}

func main() {
	// Root landing page
	http.HandleFunc("/", rootHandler)

	// Health check endpoint
	http.HandleFunc("/health", healthHandler)

	// Execute request endpoint
	http.HandleFunc("/execute", executeHandler)

	port := ":8889"

	myFigure := figure.NewFigure("ApiSpider", "", true)
	myFigure.Print()

	log.Printf("ApiSpider Local Agent starting on http://localhost%s", port)
	log.Printf("Ready to proxy requests to localhost and private APIs")
	log.Printf("CORS enabled for browser access")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Failed to start agent:", err)
	}
}
