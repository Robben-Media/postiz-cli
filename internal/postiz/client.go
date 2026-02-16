package postiz

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/builtbyrobben/postiz-cli/internal/api"
)

type Client struct {
	api *api.Client
}

func NewClient(apiKey string, opts ...api.ClientOption) *Client {
	baseURL := os.Getenv("POSTIZ_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.postiz.com/public/v1"
	}

	allOpts := append([]api.ClientOption{
		api.WithAuthFn(func(r *http.Request) {
			r.Header.Set("Authorization", "Bearer "+apiKey)
		}),
		api.WithUserAgent("postiz-cli/1.0"),
	}, opts...)

	return &Client{
		api: api.NewClient(baseURL, allOpts...),
	}
}

func (c *Client) ListIntegrations(ctx context.Context) ([]Integration, error) {
	var result []Integration

	if err := c.api.Get(ctx, "/integrations", nil, &result); err != nil {
		return nil, fmt.Errorf("list integrations: %w", err)
	}

	return result, nil
}

func (c *Client) CheckIntegration(ctx context.Context, id string) (*IntegrationCheck, error) {
	var result IntegrationCheck

	if err := c.api.Get(ctx, "/integrations/"+id+"/check", nil, &result); err != nil {
		return nil, fmt.Errorf("check integration: %w", err)
	}

	return &result, nil
}

func (c *Client) FindSlot(ctx context.Context, id string) (*Slot, error) {
	var result Slot

	if err := c.api.Get(ctx, "/integrations/"+id+"/find-slot", nil, &result); err != nil {
		return nil, fmt.Errorf("find slot: %w", err)
	}

	return &result, nil
}

func (c *Client) ListPosts(ctx context.Context, from, to string) ([]Post, error) {
	query := url.Values{}

	if from != "" {
		query.Set("from", from)
	}

	if to != "" {
		query.Set("to", to)
	}

	var result []Post

	if err := c.api.Get(ctx, "/posts", query, &result); err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}

	return result, nil
}

func (c *Client) CreatePost(ctx context.Context, input CreatePostInput) (*Post, error) {
	var result Post

	if err := c.api.Post(ctx, "/posts", input, &result); err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}

	return &result, nil
}

func (c *Client) CreatePostRaw(ctx context.Context, input any) (*Post, error) {
	var result Post

	if err := c.api.Post(ctx, "/posts", input, &result); err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}

	return &result, nil
}

func (c *Client) DeletePost(ctx context.Context, id string) error {
	if err := c.api.Delete(ctx, "/posts/"+id, nil); err != nil {
		return fmt.Errorf("delete post: %w", err)
	}

	return nil
}

func (c *Client) UploadFile(ctx context.Context, filePath string) (*Upload, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	var result Upload

	if err := c.api.PostMultipart(ctx, "/uploads", "file", filepath.Base(filePath), f, &result); err != nil {
		return nil, fmt.Errorf("upload file: %w", err)
	}

	return &result, nil
}

func (c *Client) UploadURL(ctx context.Context, rawURL string) (*Upload, error) {
	var result Upload

	body := map[string]string{"url": rawURL}

	if err := c.api.Post(ctx, "/uploads/url", body, &result); err != nil {
		return nil, fmt.Errorf("upload url: %w", err)
	}

	return &result, nil
}
