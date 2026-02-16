package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/postiz-cli/internal/outfmt"
)

type IntegrationsCmd struct {
	List     IntegrationsListCmd     `cmd:"" help:"List integrations"`
	Check    IntegrationsCheckCmd    `cmd:"" help:"Check integration status"`
	FindSlot IntegrationsFindSlotCmd `cmd:"" name:"find-slot" help:"Find next available slot for an integration"`
}

type IntegrationsListCmd struct{}

func (cmd *IntegrationsListCmd) Run(ctx context.Context) error {
	client, err := getPostizClient()
	if err != nil {
		return err
	}

	integrations, err := client.ListIntegrations(ctx)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, integrations)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "PROVIDER", "IDENTIFIER", "DISABLED"}
		var rows [][]string

		for _, i := range integrations {
			rows = append(rows, []string{i.ID, i.Name, i.Provider, i.Identifier, fmt.Sprintf("%t", i.Disabled)})
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(integrations) == 0 {
		fmt.Fprintln(os.Stderr, "No integrations found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d integrations\n\n", len(integrations))

	for _, i := range integrations {
		fmt.Printf("%s  %s\n", i.ID, i.Name)
		fmt.Printf("  Provider: %s\n", i.Provider)
		fmt.Printf("  Identifier: %s\n", i.Identifier)
		fmt.Printf("  Disabled: %t\n", i.Disabled)
		fmt.Println()
	}

	return nil
}

type IntegrationsCheckCmd struct {
	ID string `arg:"" required:"" help:"Integration ID"`
}

func (cmd *IntegrationsCheckCmd) Run(ctx context.Context) error {
	client, err := getPostizClient()
	if err != nil {
		return err
	}

	check, err := client.CheckIntegration(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, check)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"CONNECTED", "ERROR"}
		rows := [][]string{{fmt.Sprintf("%t", check.Connected), check.Error}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Printf("Connected: %t\n", check.Connected)

	if check.Error != "" {
		fmt.Printf("Error: %s\n", check.Error)
	}

	return nil
}

type IntegrationsFindSlotCmd struct {
	ID string `arg:"" required:"" help:"Integration ID"`
}

func (cmd *IntegrationsFindSlotCmd) Run(ctx context.Context) error {
	client, err := getPostizClient()
	if err != nil {
		return err
	}

	slot, err := client.FindSlot(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, slot)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"TIME"}
		rows := [][]string{{slot.Time}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Printf("Next available slot: %s\n", slot.Time)

	return nil
}
