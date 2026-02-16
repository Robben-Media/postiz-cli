package cmd

import (
	"fmt"
	"os"

	"github.com/builtbyrobben/postiz-cli/internal/postiz"
	"github.com/builtbyrobben/postiz-cli/internal/secrets"
)

func getPostizClient() (*postiz.Client, error) {
	if key := os.Getenv("POSTIZ_API_KEY"); key != "" {
		return postiz.NewClient(key), nil
	}

	store, err := secrets.OpenDefault()
	if err != nil {
		return nil, fmt.Errorf("open credential store: %w", err)
	}

	key, err := store.GetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("no credentials found; run: postiz-cli auth set-key --stdin")
	}

	return postiz.NewClient(key), nil
}
