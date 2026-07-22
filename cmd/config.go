package cmd

import (
	"fmt"

	"github.com/kentem-re-hiramatsu/gh-image-fetch/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage gh-image-fetch settings",
}

var configSetCmd = &cobra.Command{
	Use:   "set dir <path>",
	Short: "Set the default download directory",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if args[0] != "dir" {
			return fmt.Errorf("unknown config key %q: only \"dir\" is supported", args[0])
		}
		abs, err := config.SetDir(args[1])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Default download directory set to %s\n", abs)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get dir",
	Short: "Show the effective default download directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if args[0] != "dir" {
			return fmt.Errorf("unknown config key %q: only \"dir\" is supported", args[0])
		}
		dir, err := config.DefaultDir()
		if err != nil {
			return err
		}
		if dir == "" {
			return fmt.Errorf("default download directory is not set: run `gh image-fetch config set dir <path>` or set %s", config.EnvDir)
		}
		fmt.Fprintln(cmd.OutOrStdout(), dir)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd, configGetCmd)
	rootCmd.AddCommand(configCmd)
}
