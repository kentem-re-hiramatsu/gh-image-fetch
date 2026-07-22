package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kentem-re-hiramatsu/gh-image-fetch/internal/attachment"
	"github.com/kentem-re-hiramatsu/gh-image-fetch/internal/config"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download <asset-url-or-id> [dest]",
	Short: "Download a user-attachments asset to a local path",
	Long: `Download a GitHub user-attachments asset using gh CLI credentials.

<asset-url-or-id> accepts either a full URL
(https://github.com/user-attachments/assets/<uuid>) or the bare UUID.

[dest] may be a file path, or an existing directory (the file is then
named <uuid> plus an extension guessed from the Content-Type).
When [dest] is omitted, the file is saved into the default download
directory (GH_IMAGE_FETCH_DIR or ` + "`gh image-fetch config set dir`" + `)
as <yyyymmdd-hhmmss>-<uuid prefix><ext>.
Existing files are overwritten.`,
	Args: cobra.RangeArgs(1, 2),
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

		var path string
		if len(args) == 2 {
			path, err = attachment.ResolveDest(args[1], asset.ID, result.ContentType)
		} else {
			path, err = defaultDestPath(asset.ID, result.ContentType)
		}
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

// defaultDestPath resolves the save path when [dest] is omitted, creating
// the default download directory if necessary.
func defaultDestPath(assetID, contentType string) (string, error) {
	dir, err := config.DefaultDir()
	if err != nil {
		return "", err
	}
	if dir == "" {
		return "", fmt.Errorf("no destination given and no default download directory is configured: run `gh image-fetch config set dir <path>` or set %s", config.EnvDir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("cannot create default download directory %q: %w", dir, err)
	}
	return filepath.Join(dir, attachment.DefaultFileName(time.Now(), assetID, contentType)), nil
}

func init() {
	rootCmd.AddCommand(downloadCmd)
}
