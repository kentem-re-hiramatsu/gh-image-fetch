package cmd

import (
	"fmt"

	"github.com/re-hiramatsu/gh-image-fetch/internal/attachment"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download <asset-url-or-id> <dest>",
	Short: "Download a user-attachments asset to a local path",
	Long: `Download a GitHub user-attachments asset using gh CLI credentials.

<asset-url-or-id> accepts either a full URL
(https://github.com/user-attachments/assets/<uuid>) or the bare UUID.
<dest> may be a file path, or an existing directory (the file is then
named <uuid> plus an extension guessed from the Content-Type).
Existing files are overwritten.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		asset, err := attachment.Parse(args[0])
		if err != nil {
			return err
		}
		result, err := attachment.Fetch(asset)
		if err != nil {
			return err
		}
		defer result.Body.Close()

		path, err := attachment.ResolveDest(args[1], asset.ID, result.ContentType)
		if err != nil {
			return err
		}
		n, err := attachment.Save(path, result.Body)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Saved %s (%d bytes)\n", path, n)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
}
