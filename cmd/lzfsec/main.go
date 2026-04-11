package main

import (
	"fmt"
	"os"

	"github.com/go-compressions/lzfsec/cmd/lzfsec/compress"
	"github.com/go-compressions/lzfsec/cmd/lzfsec/decompress"
	"github.com/spf13/cobra"
)

// osExit allows tests to override os.Exit.
var osExit = os.Exit

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		osExit(1)
	}
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lzfsec",
		Short: "Compress and decompress files using Apple's LZFSE format",
		Long: `lzfsec compresses and decompresses files using Apple's LZFSE/LZVN
compression format with a pure-Go implementation.`,
	}
	cmd.AddCommand(compress.Command())
	cmd.AddCommand(decompress.Command())
	return cmd
}
