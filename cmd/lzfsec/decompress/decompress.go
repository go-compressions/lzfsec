// Package decompress implements the lzfsec decompress sub-command.
package decompress

import (
	"fmt"
	"io"
	"os"

	"github.com/go-compressions/lzfse"
	"github.com/spf13/cobra"
)

// Command returns the decompress cobra sub-command.
func Command() *cobra.Command {
	var inputPath, outputPath string

	cmd := &cobra.Command{
		Use:   "decompress",
		Short: "Decompress LZFSE-compressed data",
		Long: `decompress reads LZFSE-compressed bytes from a file (or stdin) and
writes the original raw data to a file (or stdout).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := readInput(inputPath)
			if err != nil {
				return err
			}
			decompressed, err := lzfse.Decompress(data)
			if err != nil {
				return fmt.Errorf("decompress: %w", err)
			}
			if err := writeOutput(outputPath, decompressed); err != nil {
				return err
			}
			if outputPath != "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "decompressed %d → %d bytes\n",
					len(data), len(decompressed))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputPath, "input", "i", "", "Input file (default: stdin)")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file (default: stdout)")
	return cmd
}

func readInput(path string) ([]byte, error) {
	if path == "" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func writeOutput(path string, data []byte) error {
	if path == "" {
		_, err := os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
