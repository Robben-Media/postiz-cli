package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/builtbyrobben/postiz-cli/internal/outfmt"
	"github.com/builtbyrobben/postiz-cli/internal/postiz"
)

type PostsCmd struct {
	List   PostsListCmd   `cmd:"" help:"List posts"`
	Create PostsCreateCmd `cmd:"" help:"Create a post"`
	Delete PostsDeleteCmd `cmd:"" help:"Delete a post"`
}

type PostsListCmd struct {
	From string `help:"Filter posts from this date (YYYY-MM-DD)"`
	To   string `help:"Filter posts to this date (YYYY-MM-DD)"`
}

func (cmd *PostsListCmd) Run(ctx context.Context) error {
	client, err := getPostizClient()
	if err != nil {
		return err
	}

	posts, err := client.ListPosts(ctx, cmd.From, cmd.To)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, posts)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "CONTENT", "TYPE", "STATUS", "PUBLISH_DATE"}
		var rows [][]string

		for _, p := range posts {
			content := truncate(p.Content, 60)
			rows = append(rows, []string{p.ID, content, p.Type, p.Status, p.PublishDate})
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(posts) == 0 {
		fmt.Fprintln(os.Stderr, "No posts found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d posts\n\n", len(posts))

	for _, p := range posts {
		printPost(&p)
	}

	return nil
}

type PostsCreateCmd struct {
	Type        string   `help:"Post type: now, schedule, draft" default:"now"`
	Date        string   `help:"Schedule date (ISO 8601, for type=schedule)"`
	Content     string   `help:"Post content text"`
	Integration []string `help:"Integration ID (repeatable)" name:"integration"`
	JSONInput   string   `help:"JSON file path or - for stdin" name:"json-input"`
}

func (cmd *PostsCreateCmd) Run(ctx context.Context) error {
	client, err := getPostizClient()
	if err != nil {
		return err
	}

	var post *postiz.Post

	if cmd.JSONInput != "" {
		var data []byte

		if cmd.JSONInput == "-" {
			data, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("read stdin: %w", err)
			}
		} else {
			data, err = os.ReadFile(cmd.JSONInput)
			if err != nil {
				return fmt.Errorf("read file %s: %w", cmd.JSONInput, err)
			}
		}

		var raw any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parse JSON input: %w", err)
		}

		post, err = client.CreatePostRaw(ctx, raw)
		if err != nil {
			return err
		}
	} else {
		if cmd.Content == "" {
			return fmt.Errorf("--content is required (or use --json-input)")
		}

		input := postiz.CreatePostInput{
			Type:        cmd.Type,
			Date:        cmd.Date,
			Content:     cmd.Content,
			Integration: cmd.Integration,
		}

		post, err = client.CreatePost(ctx, input)
		if err != nil {
			return err
		}
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, post)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "STATUS", "TYPE", "PUBLISH_DATE"}
		rows := [][]string{{post.ID, post.Status, post.Type, post.PublishDate}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Post created\n")
	printPost(post)

	return nil
}

type PostsDeleteCmd struct {
	ID string `arg:"" required:"" help:"Post ID"`
}

func (cmd *PostsDeleteCmd) Run(ctx context.Context) error {
	client, err := getPostizClient()
	if err != nil {
		return err
	}

	if err := client.DeletePost(ctx, cmd.ID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "Post deleted",
			"id":      cmd.ID,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "ID"}
		rows := [][]string{{"success", cmd.ID}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Post %s deleted\n", cmd.ID)

	return nil
}

func printPost(p *postiz.Post) {
	fmt.Printf("%s  %s\n", p.ID, truncate(p.Content, 60))
	fmt.Printf("  Type: %s\n", p.Type)
	fmt.Printf("  Status: %s\n", p.Status)

	if p.PublishDate != "" {
		fmt.Printf("  Publish Date: %s\n", p.PublishDate)
	}

	if len(p.Integration) > 0 {
		fmt.Printf("  Integrations: %s\n", strings.Join(p.Integration, ", "))
	}

	fmt.Println()
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")

	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen-3] + "..."
}
