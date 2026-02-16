package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/postiz-cli/internal/outfmt"
)

type UploadsCmd struct {
	File UploadsFileCmd `cmd:"" help:"Upload a file"`
	URL  UploadsURLCmd  `cmd:"" name:"url" help:"Upload from a URL"`
}

type UploadsFileCmd struct {
	Path string `arg:"" required:"" help:"Path to the file to upload"`
}

func (cmd *UploadsFileCmd) Run(ctx context.Context) error {
	client, err := getPostizClient()
	if err != nil {
		return err
	}

	upload, err := client.UploadFile(ctx, cmd.Path)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, upload)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "URL"}
		rows := [][]string{{upload.ID, upload.URL}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "File uploaded\n")
	fmt.Printf("ID: %s\n", upload.ID)
	fmt.Printf("URL: %s\n", upload.URL)

	return nil
}

type UploadsURLCmd struct {
	URL string `arg:"" required:"" help:"URL to upload from"`
}

func (cmd *UploadsURLCmd) Run(ctx context.Context) error {
	client, err := getPostizClient()
	if err != nil {
		return err
	}

	upload, err := client.UploadURL(ctx, cmd.URL)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, upload)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "URL"}
		rows := [][]string{{upload.ID, upload.URL}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "URL uploaded\n")
	fmt.Printf("ID: %s\n", upload.ID)
	fmt.Printf("URL: %s\n", upload.URL)

	return nil
}
