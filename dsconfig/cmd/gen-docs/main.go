// Command gen-docs renders consumer-facing Markdown configuration documents
// from dsconfig.json files.
//
// Usage:
//
//	# Render docs for a single plugin (writes CONFIGURATION.md next to the
//	# input file):
//	go run ./cmd/gen-docs -in path/to/dsconfig.json
//
//	# Render docs for every plugin under a registry directory
//	# (each subdirectory that contains a dsconfig.json is processed):
//	go run ./cmd/gen-docs -dir ../registry
//
//	# Write to a specific file (single-input mode only):
//	go run ./cmd/gen-docs -in path/to/dsconfig.json -out path/to/CONFIG.md
//
//	# Print to stdout instead of writing a file (single-input mode only):
//	go run ./cmd/gen-docs -in path/to/dsconfig.json -stdout
//
// The generated document is intended for operators configuring the plugin in
// Grafana — it deliberately hides internal storage keys, roles, effects,
// storage mappings and LLM-tagged instructions.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/grafana/dsconfig/dsconfig"
	// Ensure the built-in field packs are registered so that plugins using
	// `baseFields` resolve correctly.
	_ "github.com/grafana/dsconfig/dsconfig/packs"
)

const defaultOutputName = "CONFIGURATION.md"

func main() {
	var (
		inFile   = flag.String("in", "", "path to a single dsconfig.json")
		outFile  = flag.String("out", "", "output Markdown file (defaults to CONFIGURATION.md next to -in)")
		dir      = flag.String("dir", "", "directory containing plugin subdirectories with dsconfig.json")
		toStdout = flag.Bool("stdout", false, "write to stdout instead of a file (single-input mode only)")
	)
	flag.Parse()

	if (*inFile == "") == (*dir == "") {
		fmt.Fprintln(os.Stderr, "gen-docs: exactly one of -in or -dir must be provided")
		flag.Usage()
		os.Exit(2)
	}

	if *inFile != "" {
		if err := renderOne(*inFile, *outFile, *toStdout); err != nil {
			fmt.Fprintf(os.Stderr, "gen-docs: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *outFile != "" || *toStdout {
		fmt.Fprintln(os.Stderr, "gen-docs: -out and -stdout are only valid with -in")
		os.Exit(2)
	}
	if err := renderDir(*dir); err != nil {
		fmt.Fprintf(os.Stderr, "gen-docs: %v\n", err)
		os.Exit(1)
	}
}

func renderOne(inFile, outFile string, toStdout bool) error {
	data, err := os.ReadFile(inFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", inFile, err)
	}
	s, err := dsconfig.ParseAndResolveSchemaJSON(data)
	if err != nil {
		return fmt.Errorf("parse %s: %w", inFile, err)
	}
	out, err := dsconfig.RenderMarkdownDocs(s)
	if err != nil {
		return fmt.Errorf("render %s: %w", inFile, err)
	}

	if toStdout {
		_, err := os.Stdout.WriteString(out)
		return err
	}

	target := outFile
	if target == "" {
		target = filepath.Join(filepath.Dir(inFile), defaultOutputName)
	}
	if err := os.WriteFile(target, []byte(out), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", target, err)
	}
	fmt.Printf("wrote %s\n", target)
	return nil
}

func renderDir(root string) error {
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("stat %s: %w", root, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", root)
	}

	var files []string
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) == "dsconfig.json" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk %s: %w", root, err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no dsconfig.json files found under %s", root)
	}

	var failures int
	for _, f := range files {
		if err := renderOne(f, "", false); err != nil {
			fmt.Fprintf(os.Stderr, "gen-docs: %v\n", err)
			failures++
		}
	}
	if failures > 0 {
		return fmt.Errorf("%d dsconfig.json file(s) failed to render", failures)
	}
	return nil
}
