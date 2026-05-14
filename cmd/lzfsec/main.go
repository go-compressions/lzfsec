// lzfsec is a small CLI wrapper around the pure-Go LZFSE
// implementation in github.com/go-compressions/lzfse.
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
	if err := RootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		osExit(1)
	}
}

// RootCmd returns the top-level cobra command. Exported so tests in
// the same package can exercise the wiring without spawning a child
// process.
func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lzfsec",
		Short: "Compress and decompress files using Apple's LZFSE format",
		Long: `lzfsec compresses and decompresses files using Apple's LZFSE/LZVN
compression format with a pure-Go implementation.`,
		SilenceUsage: true,
	}
	cmd.AddCommand(compress.Command())
	cmd.AddCommand(decompress.Command())
	return cmd
}
