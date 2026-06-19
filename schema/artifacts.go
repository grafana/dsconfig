package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	sdkSchema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

const (
	ArtifactPathSchema           = "schema.gen.json"
	ArtifactPathSettings         = "settings.gen.json"
	ArtifactPathSettingsExamples = "settings.examples.gen.json"
)

// WriteArtifacts writes the three canonical schema artifacts (full schema,
// settings schema, and settings examples) to disk at the shared dsconfig paths.
// It is the single source of truth used by plugin `go:generate` programs so the
// artifact layout and encoding stay consistent across every plugin built on the
// dsconfig schema. The same encoding is asserted in the conformance suite, so a
// generated artifact will always be in sync with its drift check.
func WriteArtifacts(schema *sdkSchema.PluginSchema) error {
	artifacts := []struct {
		path  string
		input any
	}{
		{ArtifactPathSchema, schema},
		{ArtifactPathSettings, schema.SettingsSchema},
		{ArtifactPathSettingsExamples, schema.SettingsExamples},
	}

	for _, a := range artifacts {
		data, err := marshal(a.input)
		if err != nil {
			return fmt.Errorf("marshal %s: %w", a.path, err)
		}
		if err := os.MkdirAll(filepath.Dir(a.path), 0o750); err != nil {
			return fmt.Errorf("create output dir for %s: %w", a.path, err)
		}
		if err := os.WriteFile(a.path, data, 0o600); err != nil {
			return fmt.Errorf("write %s: %w", a.path, err)
		}
	}
	return nil
}

func marshal(input any) ([]byte, error) {
	out, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal plugin schema artifacts: %w", err)
	}
	return append(out, '\n'), nil
}
