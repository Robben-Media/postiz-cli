package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGet_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Authorization header 'Bearer test-key', got %s", r.Header.Get("Authorization"))
		}

		if r.URL.Query().Get("from") != "2024-01-01" {
			t.Errorf("expected query param from=2024-01-01, got %s", r.URL.Query().Get("from"))
		}

		resp := []map[string]any{
			{"id": "post-1", "content": "Hello"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, WithAuthFn(func(r *http.Request) {
		r.Header.Set("Authorization", "Bearer test-key")
	}))

	var result []struct {
		ID      string `json:"id"`
		Content string `json:"content"`
	}

	query := url.Values{"from": []string{"2024-01-01"}}

	err := client.Get(context.Background(), "/posts", query, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}

	if result[0].ID != "post-1" {
		t.Errorf("expected ID post-1, got %s", result[0].ID)
	}
}

func TestPost_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		body, _ := io.ReadAll(r.Body)

		var req map[string]any
		json.Unmarshal(body, &req)

		if req["content"] != "Hello world" {
			t.Errorf("expected content 'Hello world', got %v", req["content"])
		}

		resp := map[string]any{"id": "new-post", "content": "Hello world"}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, WithAuthFn(func(r *http.Request) {
		r.Header.Set("Authorization", "Bearer test-key")
	}))

	input := map[string]string{"content": "Hello world", "type": "now"}

	var result struct {
		ID      string `json:"id"`
		Content string `json:"content"`
	}

	err := client.Post(context.Background(), "/posts", input, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "new-post" {
		t.Errorf("expected ID new-post, got %s", result.ID)
	}
}

func TestDelete_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, WithAuthFn(func(r *http.Request) {
		r.Header.Set("Authorization", "Bearer test-key")
	}))

	err := client.Delete(context.Background(), "/posts/abc", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGet_HTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid API key"})
	}))
	defer server.Close()

	client := NewClient(server.URL, WithAuthFn(func(r *http.Request) {
		r.Header.Set("Authorization", "Bearer bad-key")
	}))

	var result struct{}

	err := client.Get(context.Background(), "/integrations", nil, &result)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if apiErr.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
}

func TestPostMultipart_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Errorf("expected multipart/form-data content type, got %s", r.Header.Get("Content-Type"))
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("expected form file: %v", err)
		}
		defer file.Close()

		if header.Filename != "test.png" {
			t.Errorf("expected filename test.png, got %s", header.Filename)
		}

		resp := map[string]string{"id": "upload-1", "url": "https://example.com/test.png"}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, WithAuthFn(func(r *http.Request) {
		r.Header.Set("Authorization", "Bearer test-key")
	}))

	var result struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}

	reader := strings.NewReader("fake image data")

	err := client.PostMultipart(context.Background(), "/uploads", "file", "test.png", reader, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "upload-1" {
		t.Errorf("expected ID upload-1, got %s", result.ID)
	}
}

func TestPostMultipart_EmptyBodyWithResult(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	var result struct {
		ID string `json:"id"`
	}

	err := client.PostMultipart(
		context.Background(),
		"/uploads",
		"file",
		"empty.txt",
		strings.NewReader("content"),
		&result,
	)
	if err != nil {
		t.Fatalf("expected no error for empty multipart response body, got %v", err)
	}
}
