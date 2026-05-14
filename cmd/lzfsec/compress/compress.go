// Package compress implements the lzfsec compress sub-command.
package compress

import (
	"fmt"

	"github.com/go-compressions/lzfse"
	"github.com/go-compressions/lzfsec/cmd/lzfsec/internal/cmdio"
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
			data, err := cmdio.ReadInput(inputPath)
			if err != nil {
				return err
			}
			// lzfse.Compress's err return is reserved for future use;
			// the current implementation has no failure mode.
			compressed, _ := lzfse.Compress(data)
			if err := cmdio.WriteOutput(outputPath, compressed); err != nil {
				return err
			}
			if outputPath != "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "compressed %d → %d bytes (%.1f%%)\n",
					len(data), len(compressed), Ratio(len(data), len(compressed)))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputPath, "input", "i", "", "Input file (default: stdin)")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file (default: stdout)")
	return cmd
}

// Ratio returns 100 × compressed / original as a percentage; 0 when
// original is zero so the caller doesn't print NaN. Exported so the
// package's own tests can pin its behaviour.
func Ratio(original, compressed int) float64 {
	if original == 0 {
		return 0
	}
	return float64(compressed) / float64(original) * 100
}
