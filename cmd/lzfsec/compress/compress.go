// Package compress implements the lzfsec compress sub-command.
package compress

import (
	"fmt"
	"io"
	"os"

	"github.com/go-compressions/lzfse"
	"github.com/spf13/cobra"
)

// Command returns the compress cobra sub-command.
func Command() *cobra.Command {
	var inputPath, outputPath string

	cmd := &cobra.Command{
		Use:   "compress",
		Short: "Compress data using LZFSE",
		Long: `compress reads raw bytes from a file (or stdin) and writes
LZFSE-compressed output to a file (or stdout).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := readInput(inputPath)
			if err != nil {
				return err
			}
			compressed, err := lzfse.Compress(data)
			if err != nil {
				return fmt.Errorf("compress: %w", err)
			}
			if err := writeOutput(outputPath, compressed); err != nil {
				return err
			}
			if outputPath != "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "compressed %d → %d bytes (%.1f%%)\n",
					len(data), len(compressed), ratio(len(data), len(compressed)))
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

func ratio(original, compressed int) float64 {
	if original == 0 {
		return 0
	}
	return float64(compressed) / float64(original) * 100
}
