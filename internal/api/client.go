package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
	authFn     func(*http.Request)
}

type ClientOption func(*Client)

func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

func WithUserAgent(ua string) ClientOption {
	return func(c *Client) {
		c.userAgent = ua
	}
}

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

func WithAuthFn(fn func(*http.Request)) ClientOption {
	return func(c *Client) {
		c.authFn = fn
	}
}

func NewClient(baseURL string, opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   baseURL,
		userAgent: "postiz-cli/1.0",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) Get(ctx context.Context, path string, query url.Values, result any) error {
	fullURL := c.baseURL + path
	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}

	return c.do(ctx, http.MethodGet, fullURL, nil, result)
}

func (c *Client) Post(ctx context.Context, path string, body any, result any) error {
	return c.doJSON(ctx, http.MethodPost, c.baseURL+path, body, result)
}

func (c *Client) Put(ctx context.Context, path string, body any, result any) error {
	return c.doJSON(ctx, http.MethodPut, c.baseURL+path, body, result)
}

func (c *Client) Patch(ctx context.Context, path string, body any, result any) error {
	return c.doJSON(ctx, http.MethodPatch, c.baseURL+path, body, result)
}

func (c *Client) Delete(ctx context.Context, path string, result any) error {
	return c.do(ctx, http.MethodDelete, c.baseURL+path, nil, result)
}

func (c *Client) PostMultipart(ctx context.Context, path string, fieldName string, fileName string, fileReader io.Reader, result any) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}

	if _, err := io.Copy(part, fileReader); err != nil {
		return fmt.Errorf("copy file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("close multipart writer: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("User-Agent", c.userAgent)

	if c.authFn != nil {
		c.authFn(httpReq)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return parseHTTPError(resp)
	}

	if result != nil {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read response body: %w", err)
		}

		if len(respBody) > 0 {
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("decode response: %w", err)
			}
		}
	}

	return nil
}

func (c *Client) doJSON(ctx context.Context, method string, fullURL string, body any, result any) error {
	var reader io.Reader

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}

		reader = bytes.NewReader(bodyBytes)
	}

	return c.doWithContentType(ctx, method, fullURL, "application/json", reader, result)
}

func (c *Client) do(ctx context.Context, method string, fullURL string, body io.Reader, result any) error {
	return c.doWithContentType(ctx, method, fullURL, "", body, result)
}

func (c *Client) doWithContentType(ctx context.Context, method string, fullURL string, contentType string, body io.Reader, result any) error {
	httpReq, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}

	httpReq.Header.Set("User-Agent", c.userAgent)

	if c.authFn != nil {
		c.authFn(httpReq)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return parseHTTPError(resp)
	}

	if result != nil {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read response body: %w", err)
		}

		if len(respBody) > 0 {
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("decode response: %w", err)
			}
		}
	}

	return nil
}

type APIError struct {
	StatusCode int
	Message    string
	Code       string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("API error (%s): %s", e.Code, e.Message)
	}

	return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Message)
}

func parseHTTPError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var apiErr struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}

	if json.Unmarshal(body, &apiErr) == nil {
		msg := apiErr.Message
		if msg == "" {
			msg = apiErr.Error
		}

		if msg != "" {
			return &APIError{StatusCode: resp.StatusCode, Message: msg}
		}
	}

	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    http.StatusText(resp.StatusCode),
	}
}
