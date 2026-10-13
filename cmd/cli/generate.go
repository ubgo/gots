// generate.go defines the "generate" subcommand.
//
// This is the main workhorse of the CLI. It:
//  1. Loads the YAML config (gots.yaml by default)
//  2. Parses each Go source file listed in config.sources using go/ast
//  3. Passes the parsed result to the generator
//  4. Generator writes .ts files to the configured output paths
//
// Usage:
//
//	gots generate                    # uses ./gots.yaml
//	gots generate -c path/to.yaml   # custom config path
package main

import (
	"fmt"

	"github.com/ubgo/gots/pkg/config"
	"github.com/ubgo/gots/pkg/generator"
	"github.com/ubgo/gots/pkg/parser"
	"github.com/spf13/cobra"
)

// configPath holds the --config flag value, bound in init().
var (
	configPath string
)

// generateCmd is the "gots generate" subcommand.
//
// The pipeline is: config → parser → generator → .ts files
//
// Each source entry in the config points to a Go file (absolute or relative path).
// The parser reads all sources into a single ParseResult, then the generator
// iterates over each source and writes the corresponding .ts output file.
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate TypeScript from Go source files",
	Long:  `Reads Go source files and generates TypeScript types, enums, constants, and interfaces based on the gots.yaml config.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Step 1: Load and validate config
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Step 2: Parse all Go source files into a single result.
		// The parser accumulates types, constants, and structs across all files.
		p := parser.New()

		for _, source := range cfg.Sources {
			if err := p.ParseFile(source.File); err != nil {
				return fmt.Errorf("failed to parse %s: %w", source.File, err)
			}
		}

		// Step 3: Generate TypeScript output.
		// The generator uses the config (styles, overrides, type map) to decide
		// how each Go declaration maps to TypeScript.
		gen := generator.New(cfg, p.Result())

		if err := gen.Generate(); err != nil {
			return fmt.Errorf("failed to generate: %w", err)
		}

		fmt.Println("TypeScript generated successfully")
		return nil
	},
}

func init() {
	generateCmd.Flags().StringVarP(&configPath, "config", "c", "gots.yaml", "Path to gots.yaml config file")
}
