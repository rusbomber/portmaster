package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/safing/portbase/log"
	"github.com/safing/portmaster/updates/helper"
)

var (
	reset     bool
	intelOnly bool
)

func init() {
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(purgeCmd)

	flags := updateCmd.Flags()
	flags.BoolVar(&reset, "reset", false, "Delete all resources and re-download the basic set")
	flags.BoolVar(&intelOnly, "intel-only", false, "Only make downloading intel updates mandatory")
}

var (
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Run a manual update process",
		RunE: func(cmd *cobra.Command, args []string) error {
			return downloadUpdates()
		},
	}

	purgeCmd = &cobra.Command{
		Use:   "purge",
		Short: "Remove old resource versions that are superseded by at least three versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return purge()
		},
	}
)

func indexRequired(cmd *cobra.Command) bool {
	switch cmd {
	case updateCmd, purgeCmd:
		return true
	default:
		return false
	}
}

func downloadUpdates() error {
	// Check if only intel data is mandatory.
	if intelOnly {
		helper.IntelOnly()
	}

	// Set required updates.
	registry.MandatoryUpdates = helper.MandatoryUpdates()
	registry.AutoUnpack = helper.AutoUnpackUpdates()

	// logging is configured as a persistent pre-run method inherited from
	// the root command but since we don't use run.Run() we need to start
	// logging ourself.
	log.SetLogLevel(log.TraceLevel)
	err := log.Start()
	if err != nil {
		fmt.Printf("failed to start logging: %s\n", err)
	}
	defer log.Shutdown()

	if reset {
		// Delete storage.
		err = os.RemoveAll(registry.StorageDir().Path)
		if err != nil {
			return fmt.Errorf("failed to reset update dir: %w", err)
		}
		err = registry.StorageDir().Ensure()
		if err != nil {
			return fmt.Errorf("failed to create update dir: %w", err)
		}

		// Reset registry resources.
		registry.ResetResources()
	}

	// Update all indexes.
	err = registry.UpdateIndexes(context.TODO())
	if err != nil {
		return err
	}

	// Download all required updates.
	err = registry.DownloadUpdates(context.TODO())
	if err != nil {
		return err
	}

	// Select versions and unpack the selected.
	registry.SelectVersions()
	err = registry.UnpackResources()
	if err != nil {
		return fmt.Errorf("failed to unpack resources: %w", err)
	}

	if !intelOnly {
		// Fix chrome-sandbox permissions
		if err := helper.EnsureChromeSandboxPermissions(registry); err != nil {
			return fmt.Errorf("failed to fix electron permissions: %w", err)
		}
	}

	return nil
}

func purge() error {
	log.SetLogLevel(log.TraceLevel)

	// logging is configured as a persistent pre-run method inherited from
	// the root command but since we don't use run.Run() we need to start
	// logging ourself.
	err := log.Start()
	if err != nil {
		fmt.Printf("failed to start logging: %s\n", err)
	}
	defer log.Shutdown()

	registry.Purge(3)
	return nil
}
