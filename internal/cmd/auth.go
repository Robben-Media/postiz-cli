package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/builtbyrobben/postiz-cli/internal/outfmt"
	"github.com/builtbyrobben/postiz-cli/internal/secrets"
)

type AuthCmd struct {
	SetKey AuthSetKeyCmd `cmd:"" help:"Set API key (uses --stdin by default)"`
	Status AuthStatusCmd `cmd:"" help:"Show authentication status"`
	Remove AuthRemoveCmd `cmd:"" help:"Remove stored credentials"`
}

type AuthSetKeyCmd struct {
	Stdin bool   `help:"Read API key from stdin (default: true)" default:"true"`
	Key   string `arg:"" optional:"" help:"API key (discouraged; exposes in shell history)"`
}

func (cmd *AuthSetKeyCmd) Run(ctx context.Context) error {
	var apiKey string

	switch {
	case cmd.Key != "":
		fmt.Fprintln(os.Stderr, "Warning: passing keys as arguments exposes them in shell history. Use --stdin instead.")
		apiKey = strings.TrimSpace(cmd.Key)
	case term.IsTerminal(int(os.Stdin.Fd())):
		fmt.Fprint(os.Stderr, "Enter API key: ")

		byteKey, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)

		if err != nil {
			return fmt.Errorf("read API key: %w", err)
		}

		apiKey = strings.TrimSpace(string(byteKey))
	default:
		byteKey, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read API key from stdin: %w", err)
		}

		apiKey = strings.TrimSpace(string(byteKey))
	}

	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	store, err := secrets.OpenDefault()
	if err != nil {
		return fmt.Errorf("open credential store: %w", err)
	}

	if err := store.SetAPIKey(apiKey); err != nil {
		return fmt.Errorf("store API key: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "API key stored in keyring",
		})
	}

	if outfmt.IsPlain(ctx) {
		return outfmt.WritePlain(os.Stdout, []string{"STATUS", "MESSAGE"}, [][]string{{"success", "API key stored in keyring"}})
	}

	fmt.Fprintln(os.Stderr, "API key stored in keyring")

	return nil
}

type AuthStatusCmd struct{}

func (cmd *AuthStatusCmd) Run(ctx context.Context) error {
	store, err := secrets.OpenDefault()
	if err != nil {
		return fmt.Errorf("open credential store: %w", err)
	}

	hasKey, err := store.HasKey()
	if err != nil {
		return fmt.Errorf("check API key: %w", err)
	}

	envKey := os.Getenv("POSTIZ_API_KEY")
	envOverride := envKey != ""

	status := map[string]any{
		"has_key":         hasKey,
		"env_override":    envOverride,
		"storage_backend": "keyring",
	}

	if hasKey && !envOverride {
		key, err := store.GetAPIKey()
		if err == nil && len(key) > 8 {
			status["key_redacted"] = key[:4] + "..." + key[len(key)-4:]
		}
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, status)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"HAS_KEY", "ENV_OVERRIDE", "STORAGE"}
		rows := [][]string{{
			fmt.Sprintf("%t", hasKey),
			fmt.Sprintf("%t", envOverride),
			"keyring",
		}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stdout, "Storage: %s\n", status["storage_backend"])

	switch {
	case envOverride:
		fmt.Fprintln(os.Stdout, "API Key: Using POSTIZ_API_KEY environment variable")
	case hasKey:
		fmt.Fprintln(os.Stdout, "API Key: Authenticated")

		if redacted, ok := status["key_redacted"].(string); ok {
			fmt.Fprintf(os.Stdout, "Key: %s\n", redacted)
		}
	default:
		fmt.Fprintln(os.Stdout, "API Key: Not authenticated")
		fmt.Fprintln(os.Stderr, "Run: postiz-cli auth set-key --stdin")
	}

	return nil
}

type AuthRemoveCmd struct{}

func (cmd *AuthRemoveCmd) Run(ctx context.Context) error {
	store, err := secrets.OpenDefault()
	if err != nil {
		return fmt.Errorf("open credential store: %w", err)
	}

	if err := store.DeleteAPIKey(); err != nil {
		return fmt.Errorf("remove API key: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "API key removed",
		})
	}

	if outfmt.IsPlain(ctx) {
		return outfmt.WritePlain(os.Stdout, []string{"STATUS", "MESSAGE"}, [][]string{{"success", "API key removed"}})
	}

	fmt.Fprintln(os.Stderr, "API key removed")

	return nil
}
