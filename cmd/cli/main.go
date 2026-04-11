// root.go is the entry point for the gots CLI application.
//
// It defines the root cobra command and registers all subcommands.
// The binary name is "gots" and it's built from this package (cmd/cli).
//
// Subcommands:
//   - generate: Parse Go files and produce TypeScript output
//   - init:     Scaffold a default gots.yaml config file
//   - version:  Print the current binary version
//
// The version variable is set at build time via ldflags:
//
//	go build -ldflags "-X main.version=1.0.0" ./cmd/cli
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is injected at build time by goreleaser or manual ldflags.
// Defaults to "dev" for local development builds.
var version = "dev"

// rootCmd is the base command. Running "gots" with no subcommand prints help.
var rootCmd = &cobra.Command{
	Use:   "gots",
	Short: "Go to TypeScript code generator",
	Long:  `A CLI tool that generates TypeScript types, enums, and constants from Go source files.`,
}

// main is the application entry point. Cobra handles argument parsing
// and dispatches to the appropriate subcommand's RunE function.
func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// init registers all subcommands on the root command.
// Cobra calls this automatically before main().
func init() {
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(initCmd)

	// version is defined inline since it's trivial — no separate file needed.
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version and exit",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("gots", version)
		},
	})
}
