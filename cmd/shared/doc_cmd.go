package shared

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docDirectory string

// kvCmd represents the kv command
var docCmd = &cobra.Command{
	Use:   "gendoc",
	Short: "Generate command documentation",
	Run:   genDoc,
}

func InitDocCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(docCmd)
	docCmd.Flags().StringVarP(&docDirectory, "directory", "d", defaultDirectory(), "Generate command documentation to given directory")
}

func genDoc(cmd *cobra.Command, _ []string) {

	err := func() error {
		ensureDirectoryExists := func(directory string) error {
			info, err := os.Stat(directory)
			if err != nil {
				if os.IsNotExist(err) {
					// Directory does not exist, create it
					return os.MkdirAll(directory, os.ModePerm)
				}
				return err // Unexpected error
			}

			if !info.IsDir() {
				return &os.PathError{Op: "ensureDirectoryExists", Path: directory, Err: os.ErrInvalid}
			}

			return nil // Directory exists
		}

		if err := ensureDirectoryExists(docDirectory); err != nil {
			return err
		}

		return doc.GenMarkdownTree(cmd.Parent(), docDirectory)
	}()

	if err != nil {
		fmt.Printf("error generating documentation: %v\n", err)
	}
}

func defaultDirectory() string {

	return filepath.Join("./docs", filepath.Base(os.Args[0]))
}
