package dsconfig

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// RenderMarkdownDocs renders a consumer-facing Markdown configuration guide
// for a datasource from its dsconfig schema. The output is intended for
// operators provisioning the plugin in Grafana. It contains:
//
//   - a header with the plugin name, ID and (optional) documentation link;
//   - a single "Fields" table listing every storage field with its dotted
//     path, type, target, whether it is required, and description;
//   - a "Provisioning examples" section with a Grafana YAML block and a
//     Terraform HCL block per authentication scenario (or a single default
//     scenario when the schema has no auth discriminator).
//
// The schema is expected to have base fields already resolved (via
// ResolveBaseFields or ParseAndResolveSchemaJSON).
func RenderMarkdownDocs(s *Schema) (string, error) {
	if s == nil {
		return "", fmt.Errorf("schema is nil")
	}
	if s.PluginName == "" || s.PluginType == "" {
		return "", fmt.Errorf("schema is missing pluginName or pluginType")
	}

	var b strings.Builder

	// Header.
	fmt.Fprintf(&b, "# %s configuration\n\n", s.PluginName)
	fmt.Fprintf(&b, "Configuration reference for the **%s** data source (`%s`) in Grafana.\n\n",
		s.PluginName, s.PluginType)
	if s.DocURL != "" {
		fmt.Fprintf(&b, "For more information, see the [official documentation](%s).\n\n", s.DocURL)
	}
	b.WriteString("> Generated from [`dsconfig.json`](dsconfig.json). ")
	b.WriteString("Do not edit by hand — run `go generate ./...` to refresh.\n\n")

	// Fields reference table.
	renderFieldsTable(&b, s)

	// Provisioning examples per auth scenario.
	renderExamples(&b, s)

	return b.String(), nil
}

// -------------------------------------------------------------------------
// Fields table
// -------------------------------------------------------------------------

// renderFieldsTable writes the "Fields" section: one row per storage field
// in the schema (including nested array item fields). Virtual fields are
// skipped because they have no storage.
func renderFieldsTable(b *strings.Builder, s *Schema) {
	rows := collectFieldRows(s.Fields, "", "")
	if len(rows) == 0 {
		b.WriteString("## Fields\n\n_This data source has no configurable fields._\n\n")
		return
	}
	b.WriteString("## Fields\n\n")
	b.WriteString("| Field | Type | Target | Required | Description |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(b, "| %s | %s | %s | %s | %s |\n",
			r.field, r.typ, r.target, r.required, r.description)
	}
	b.WriteString("\n")
}

type fieldRow struct {
	field       string // e.g. "jsonData.timeout" or "jsonData.headers[].name"
	typ         string // e.g. "string", "select (GET, POST)"
	target      string // "root", "jsonData", "secureJsonData"
	required    string // "yes", "conditional", ""
	description string // one-line, escaped for tables
}

func collectFieldRows(fields []ConfigField, parentPath, inheritTarget string) []fieldRow {
	var rows []fieldRow
	for i := range fields {
		f := &fields[i]
		if f.Kind == VirtualField {
			continue
		}
		target := inheritTarget
		if f.Target != nil {
			target = string(*f.Target)
		}
		path := fieldPath(f, parentPath, target)
		rows = append(rows, fieldRow{
			field:       "`" + path + "`" + secretMark(target),
			typ:         fieldTypeLabel(f),
			target:      target,
			required:    requiredLabel(f),
			description: tableEscape(fieldDescription(f)),
		})
		if f.Item != nil && len(f.Item.Fields) > 0 {
			rows = append(rows, collectFieldRows(f.Item.Fields, path+"[]", target)...)
		}
	}
	return rows
}

// fieldPath returns the dotted storage path of a field, e.g.
// "jsonData.services.catchpoint.auth.id" or "secureJsonData.catchpoint.token"
// or "url" (for root fields).
func fieldPath(f *ConfigField, parentPath, target string) string {
	if parentPath != "" {
		if f.Key == "" {
			return parentPath
		}
		return parentPath + "." + f.Key
	}
	// Top-level field. Build target.section.key (skipping empties).
	var parts []string
	if target != "" && target != string(RootTarget) {
		parts = append(parts, target)
	}
	if f.Section != "" {
		parts = append(parts, f.Section)
	}
	if f.Key != "" {
		parts = append(parts, f.Key)
	}
	return strings.Join(parts, ".")
}

func secretMark(target string) string {
	if target == string(SecureJSONTarget) {
		return " 🔒"
	}
	return ""
}

// fieldTypeLabel returns a short type description used in the fields table.
// Select/multiselect fields include their allowed values inline.
func fieldTypeLabel(f *ConfigField) string {
	base := string(f.ValueType)
	if base == "" {
		base = "string"
	}
	// Prefer UI component names when they are more informative.
	if f.UI != nil {
		switch f.UI.Component {
		case UISelect, UIRadio:
			opts := optionValues(f.UI.Options)
			if len(opts) > 0 {
				return "enum (" + strings.Join(opts, ", ") + ")"
			}
			return "enum"
		case UIMultiselect:
			opts := optionValues(f.UI.Options)
			if len(opts) > 0 {
				return "list<enum> (" + strings.Join(opts, ", ") + ")"
			}
			return "list<enum>"
		case UICheckbox, UISwitch:
			return "boolean"
		case UITextarea:
			return "string (multiline)"
		case UICode:
			if f.UI.Language != "" {
				return f.UI.Language
			}
			return "string (code)"
		case UIKeyValue:
			return "map<string,string>"
		case UIList:
			return "list"
		case UIFileUpload:
			return "file upload"
		}
	}
	if base == "array" {
		return "list"
	}
	// Look for an allowedValues validation.
	for _, v := range f.Validations {
		if v.Type == AllowedValuesValidation && len(v.Values) > 0 {
			vals := make([]string, 0, len(v.Values))
			for _, x := range v.Values {
				vals = append(vals, fmt.Sprintf("%v", x))
			}
			return "enum (" + strings.Join(vals, ", ") + ")"
		}
	}
	return base
}

func optionValues(opts []FieldOption) []string {
	out := make([]string, 0, len(opts))
	for _, o := range opts {
		out = append(out, fmt.Sprintf("%v", o.Value))
	}
	return out
}

func requiredLabel(f *ConfigField) string {
	switch {
	case f.Required:
		return "yes"
	case f.RequiredWhen != "":
		return "conditional"
	default:
		return ""
	}
}

// fieldDescription picks the most useful one-line description for the field.
func fieldDescription(f *ConfigField) string {
	if f.Description != "" {
		return f.Description
	}
	if f.Label != "" {
		return f.Label
	}
	return ""
}

// tableEscape sanitises a string for inclusion in a Markdown table cell:
// replaces `|` and newlines to keep the row intact.
func tableEscape(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "|", `\|`)
	return strings.TrimSpace(s)
}

// -------------------------------------------------------------------------
// Provisioning examples
// -------------------------------------------------------------------------

// scenario describes one authentication (or default) configuration for which
// an example is generated.
type scenario struct {
	title  string        // e.g. "******", "No authentication", "Default configuration"
	discID string        // the discriminator field ID, or "" if none
	value  string        // the discriminator value, or "" if none
	fields []*ConfigField // fields to include in this scenario, in order
}

// renderExamples writes the "Provisioning examples" section. It emits one
// subsection per authentication scenario (based on the auth discriminator's
// options) with a Grafana YAML block and a Terraform HCL block.
func renderExamples(b *strings.Builder, s *Schema) {
	scenarios := buildScenarios(s)
	if len(scenarios) == 0 {
		return
	}
	b.WriteString("## Provisioning examples\n\n")
	b.WriteString("Each scenario below shows how to provision the data source in Grafana ")
	b.WriteString("using a YAML file (loaded by Grafana's [file provisioner]")
	b.WriteString("(https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) ")
	b.WriteString("and using the [Grafana Terraform provider]")
	b.WriteString("(https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).\n\n")
	b.WriteString("Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.\n\n")

	for _, sc := range scenarios {
		fmt.Fprintf(b, "### %s\n\n", sc.title)

		cfg := buildScenarioConfig(s, sc)

		b.WriteString("**Grafana provisioning YAML**\n\n")
		b.WriteString("```yaml\n")
		b.WriteString(renderYAML(s, cfg))
		b.WriteString("```\n\n")

		b.WriteString("**Terraform**\n\n")
		b.WriteString("```hcl\n")
		b.WriteString(renderTerraform(s, sc, cfg))
		b.WriteString("```\n\n")
	}
}

// buildScenarios inspects the schema and returns the list of scenarios to
// document. If the schema has exactly one `auth.discriminator` field, one
// scenario is produced per allowed value of that field. Otherwise a single
// "Default configuration" scenario is returned.
func buildScenarios(s *Schema) []scenario {
	var disc *ConfigField
	discCount := 0
	for i := range s.Fields {
		f := &s.Fields[i]
		if f.Role == RoleAuthDiscriminator {
			disc = f
			discCount++
		}
	}
	if discCount != 1 || disc == nil {
		return []scenario{{
			title:  "Default configuration",
			fields: collectScenarioFields(s, "", ""),
		}}
	}

	values := discriminatorValues(disc)
	if len(values) == 0 {
		return []scenario{{
			title:  "Default configuration",
			fields: collectScenarioFields(s, "", ""),
		}}
	}

	out := make([]scenario, 0, len(values))
	for _, v := range values {
		title := authScenarioTitle(disc, v)
		out = append(out, scenario{
			title:  title,
			discID: disc.ID,
			value:  v,
			fields: collectScenarioFields(s, disc.ID, v),
		})
	}
	return out
}

// discriminatorValues returns the list of allowed values for the given
// auth-discriminator field, preferring UI options for stable ordering.
func discriminatorValues(disc *ConfigField) []string {
	if disc.UI != nil && len(disc.UI.Options) > 0 {
		out := make([]string, 0, len(disc.UI.Options))
		for _, o := range disc.UI.Options {
			out = append(out, fmt.Sprintf("%v", o.Value))
		}
		return out
	}
	for _, v := range disc.Validations {
		if v.Type == AllowedValuesValidation {
			out := make([]string, 0, len(v.Values))
			for _, x := range v.Values {
				out = append(out, fmt.Sprintf("%v", x))
			}
			return out
		}
	}
	return nil
}

// authScenarioTitle returns a readable title for an auth-method scenario.
func authScenarioTitle(disc *ConfigField, value string) string {
	// Use the option label if available.
	if disc.UI != nil {
		for _, o := range disc.UI.Options {
			if fmt.Sprintf("%v", o.Value) == value {
				if o.Label != "" && o.Label != value {
					return fmt.Sprintf("%s (`%s`)", o.Label, value)
				}
				return "`" + value + "`"
			}
		}
	}
	return "`" + value + "`"
}

// collectScenarioFields returns the top-level fields that should appear in a
// scenario's example. A field is included when:
//   - it is a storage field (not virtual), and
//   - it has no `dependsOn` clause, OR the clause is satisfied by the
//     scenario's discriminator value, and
//   - it is required OR requiredWhen is satisfied OR it carries a default.
//
// The discriminator field itself is always included when a scenario has one.
func collectScenarioFields(s *Schema, discID, discValue string) []*ConfigField {
	var out []*ConfigField
	for i := range s.Fields {
		f := &s.Fields[i]
		if f.Kind == VirtualField {
			continue
		}
		// Filter by dependsOn.
		if f.DependsOn != "" && !conditionMatches(f.DependsOn, discID, discValue) {
			continue
		}
		// Discriminator itself is always included.
		if discID != "" && f.ID == discID {
			out = append(out, f)
			continue
		}
		// Include when required outright, required for this scenario, or has
		// a default (so operators see a complete, minimally-valid example).
		if f.Required || (f.RequiredWhen != "" && conditionMatches(f.RequiredWhen, discID, discValue)) || f.DefaultValue != nil {
			out = append(out, f)
		}
	}
	return out
}

// simpleCompareRE matches a single "<id> == '<value>'" clause.
var simpleCompareRE = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_.]*)\s*==\s*'([^']*)'`)

// conditionMatches returns true when the CEL-style expression is satisfied
// by the given discriminator (id, value) pair. It recognises single
// comparisons and OR-joined comparisons of the form:
//
//	discID == 'a' || discID == 'b'
//
// Any other shape is conservatively treated as "matches" so that fields
// tied to non-trivial conditions still show up in examples.
func conditionMatches(expr, discID, discValue string) bool {
	if strings.ContainsAny(expr, "&<>!") {
		return true
	}
	matches := simpleCompareRE.FindAllStringSubmatch(expr, -1)
	if len(matches) == 0 {
		return true
	}
	// All comparisons must reference discID; if any do, then require at
	// least one to match discValue.
	sawDisc := false
	for _, m := range matches {
		if m[1] != discID {
			// Non-discriminator variable — cannot evaluate, treat as match.
			return true
		}
		sawDisc = true
		if m[2] == discValue {
			return true
		}
	}
	if !sawDisc {
		return true
	}
	return false
}

// buildScenarioConfig produces the concrete config values for a scenario,
// organised by storage target. The `root` map holds top-level datasource
// keys (e.g. url, access), `jsonData` holds a nested tree, and
// `secureJsonData` holds flat dotted keys.
type scenarioConfig struct {
	root           map[string]any
	jsonData       map[string]any
	secureJSONData map[string]any // flat: dotted keys as-is
	haveHeaders    bool           // true when at least one field is populated
}

func buildScenarioConfig(s *Schema, sc scenario) scenarioConfig {
	cfg := scenarioConfig{
		root:           map[string]any{},
		jsonData:       map[string]any{},
		secureJSONData: map[string]any{},
	}
	for _, f := range sc.fields {
		if f.Target == nil {
			continue
		}
		val := scenarioValue(f, sc)
		if val == nil {
			continue
		}
		cfg.haveHeaders = true
		switch *f.Target {
		case RootTarget:
			cfg.root[f.Key] = val
		case JSONDataTarget:
			setNested(cfg.jsonData, joinPath(f.Section, f.Key), val)
		case SecureJSONTarget:
			cfg.secureJSONData[f.Key] = val
		}
	}
	return cfg
}

// scenarioValue picks a value for the field within the given scenario:
// discriminator → the scenario's value; else defaultValue when present;
// else a `<PLACEHOLDER>` derived from the label/key. Boolean and numeric
// scalars are preserved as-is; strings become placeholders when unknown.
func scenarioValue(f *ConfigField, sc scenario) any {
	if sc.discID != "" && f.ID == sc.discID {
		return sc.value
	}
	if f.DefaultValue != nil {
		return f.DefaultValue
	}
	// Prefer a value from ui.options[0] for select-like fields with no
	// default (rare, but keeps examples valid).
	if f.UI != nil && (f.UI.Component == UISelect || f.UI.Component == UIRadio) && len(f.UI.Options) > 0 {
		return fmt.Sprintf("%v", f.UI.Options[0].Value)
	}
	// Type-driven fallbacks.
	switch f.ValueType {
	case "boolean":
		return false
	case "integer":
		return 0
	case "number":
		return 0
	case "array":
		return []any{}
	case "object":
		return map[string]any{}
	}
	// String placeholder.
	if f.UI != nil && f.UI.Placeholder != "" &&
		f.Target != nil && *f.Target != SecureJSONTarget {
		return f.UI.Placeholder
	}
	return placeholderFor(f)
}

var nonWordRE = regexp.MustCompile(`[^A-Za-z0-9]+`)

// placeholderFor returns a "<YOUR_LABEL>" style placeholder derived from the
// field's label or key.
func placeholderFor(f *ConfigField) string {
	name := f.Label
	if name == "" {
		name = f.Key
	}
	name = nonWordRE.ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")
	if name == "" {
		name = "value"
	}
	return "<YOUR_" + strings.ToUpper(name) + ">"
}

// joinPath joins a dotted section prefix with a key.
func joinPath(section, key string) string {
	if section == "" {
		return key
	}
	if key == "" {
		return section
	}
	return section + "." + key
}

// setNested writes value into m at the dotted path, creating intermediate
// maps as needed.
func setNested(m map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	cur := m
	for i, p := range parts {
		if i == len(parts)-1 {
			cur[p] = value
			return
		}
		next, ok := cur[p].(map[string]any)
		if !ok {
			next = map[string]any{}
			cur[p] = next
		}
		cur = next
	}
}

// -------------------------------------------------------------------------
// YAML rendering
// -------------------------------------------------------------------------

// renderYAML renders a scenarioConfig as a Grafana provisioning YAML block.
func renderYAML(s *Schema, cfg scenarioConfig) string {
	var b strings.Builder
	b.WriteString("apiVersion: 1\n")
	b.WriteString("datasources:\n")
	fmt.Fprintf(&b, "  - name: %s\n", yamlScalar(s.PluginName))
	fmt.Fprintf(&b, "    type: %s\n", yamlScalar(s.PluginType))
	// Grafana defaults; harmless when unused.
	b.WriteString("    access: proxy\n")

	// Root fields as top-level keys, sorted for determinism.
	for _, k := range sortedKeys(cfg.root) {
		if k == "name" || k == "type" || k == "access" {
			continue
		}
		writeYAMLKV(&b, "    ", k, cfg.root[k])
	}

	if len(cfg.jsonData) > 0 {
		b.WriteString("    jsonData:\n")
		writeYAMLMap(&b, "      ", cfg.jsonData)
	}
	if len(cfg.secureJSONData) > 0 {
		b.WriteString("    secureJsonData:\n")
		for _, k := range sortedKeys(cfg.secureJSONData) {
			writeYAMLKV(&b, "      ", k, cfg.secureJSONData[k])
		}
	}
	return b.String()
}

func writeYAMLMap(b *strings.Builder, indent string, m map[string]any) {
	for _, k := range sortedKeys(m) {
		writeYAMLKV(b, indent, k, m[k])
	}
}

func writeYAMLKV(b *strings.Builder, indent, key string, v any) {
	switch t := v.(type) {
	case map[string]any:
		if len(t) == 0 {
			fmt.Fprintf(b, "%s%s: {}\n", indent, yamlKey(key))
			return
		}
		fmt.Fprintf(b, "%s%s:\n", indent, yamlKey(key))
		writeYAMLMap(b, indent+"  ", t)
	case []any:
		if len(t) == 0 {
			fmt.Fprintf(b, "%s%s: []\n", indent, yamlKey(key))
			return
		}
		fmt.Fprintf(b, "%s%s:\n", indent, yamlKey(key))
		for _, item := range t {
			fmt.Fprintf(b, "%s- %s\n", indent, yamlScalar(item))
		}
	default:
		fmt.Fprintf(b, "%s%s: %s\n", indent, yamlKey(key), yamlScalar(v))
	}
}

// yamlKey quotes a key when it contains characters that would be
// misinterpreted (dots are fine unquoted in YAML mapping keys, but we quote
// anything with whitespace or a leading special char to be safe).
func yamlKey(k string) string {
	if k == "" {
		return `""`
	}
	if strings.ContainsAny(k, ` :#&*!|>'"%@` + "`") {
		return strconvQuote(k)
	}
	return k
}

// yamlScalar renders a value as a YAML scalar. Strings are quoted when they
// contain characters that could be misparsed.
func yamlScalar(v any) string {
	switch t := v.(type) {
	case string:
		if needsYAMLQuote(t) {
			return strconvQuote(t)
		}
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case json.Number:
		return string(t)
	case float64:
		// Preserve integers when possible.
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%v", t)
	case int, int32, int64:
		return fmt.Sprintf("%d", t)
	case nil:
		return "null"
	}
	data, err := json.Marshal(v)
	if err != nil {
		return `""`
	}
	return string(data)
}

func needsYAMLQuote(s string) bool {
	if s == "" {
		return true
	}
	// Reserved YAML tokens.
	switch strings.ToLower(s) {
	case "true", "false", "yes", "no", "null", "~", "on", "off":
		return true
	}
	// Leading/trailing whitespace, or characters that trigger special
	// parsing.
	if s != strings.TrimSpace(s) {
		return true
	}
	if strings.ContainsAny(s, ":#&*!|>'\"%@`\n\t") {
		return true
	}
	// Numbers must be quoted to remain strings.
	if _, err := parseNumber(s); err == nil {
		return true
	}
	return false
}

func parseNumber(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%g", &f)
	return f, err
}

// strconvQuote wraps a string in double quotes and escapes double quotes
// and backslashes.
func strconvQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

func sortedKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// -------------------------------------------------------------------------
// Terraform rendering
// -------------------------------------------------------------------------

// terraformResourceName sanitises a plugin type into a valid Terraform
// resource-name identifier.
var tfNameRE = regexp.MustCompile(`[^A-Za-z0-9_]`)

func terraformResourceName(pluginType string) string {
	name := tfNameRE.ReplaceAllString(pluginType, "_")
	if name == "" {
		name = "datasource"
	}
	// Terraform resource names cannot start with a digit.
	if name[0] >= '0' && name[0] <= '9' {
		name = "_" + name
	}
	return name
}

// renderTerraform renders a scenarioConfig as a `grafana_data_source` HCL
// block using the grafana/grafana Terraform provider.
func renderTerraform(s *Schema, sc scenario, cfg scenarioConfig) string {
	var b strings.Builder
	resName := terraformResourceName(s.PluginType)
	if sc.value != "" {
		resName = resName + "_" + tfNameRE.ReplaceAllString(sc.value, "_")
	}
	fmt.Fprintf(&b, "resource \"grafana_data_source\" %q {\n", resName)
	fmt.Fprintf(&b, "  type = %q\n", s.PluginType)
	fmt.Fprintf(&b, "  name = %q\n", s.PluginName)

	// Root fields become top-level provider attributes when they are
	// recognised (url, access, basic_auth_enabled, ...); otherwise they are
	// nested inside http_headers or ignored. To keep this generator generic
	// we emit them as commented hints so operators can map them.
	rootKeys := sortedKeys(cfg.root)
	knownRoot := map[string]string{
		"url":      "url",
		"access":   "access",
		"user":     "basic_auth_username",
		"database": "database_name",
	}
	for _, k := range rootKeys {
		if attr, ok := knownRoot[k]; ok {
			fmt.Fprintf(&b, "  %s = %s\n", attr, hclScalar(cfg.root[k]))
		}
	}
	// Fallback: if url wasn't in root, use empty placeholder to make the
	// example resource valid HCL.
	if _, ok := cfg.root["url"]; !ok {
		fmt.Fprintf(&b, "  url = %q\n", "https://example.com")
	}

	// jsonData → json_data_encoded.
	if len(cfg.jsonData) > 0 {
		b.WriteString("\n  json_data_encoded = jsonencode(")
		writeHCLValue(&b, cfg.jsonData, "  ")
		b.WriteString(")\n")
	}
	// secureJsonData → secure_json_data_encoded (sensitive).
	if len(cfg.secureJSONData) > 0 {
		b.WriteString("\n  secure_json_data_encoded = jsonencode(")
		writeHCLValue(&b, mapAny(cfg.secureJSONData), "  ")
		b.WriteString(")\n")
	}
	b.WriteString("}\n")
	return b.String()
}

func mapAny(m map[string]any) map[string]any { return m }

// writeHCLValue renders any Go value as an HCL literal — objects use
// `{ key = value }` syntax and lists use `[...]`.
func writeHCLValue(b *strings.Builder, v any, indent string) {
	switch t := v.(type) {
	case map[string]any:
		if len(t) == 0 {
			b.WriteString("{}")
			return
		}
		b.WriteString("{\n")
		next := indent + "  "
		for _, k := range sortedKeys(t) {
			fmt.Fprintf(b, "%s%s = ", next, hclKey(k))
			writeHCLValue(b, t[k], next)
			b.WriteString("\n")
		}
		fmt.Fprintf(b, "%s}", indent)
	case []any:
		if len(t) == 0 {
			b.WriteString("[]")
			return
		}
		b.WriteString("[\n")
		next := indent + "  "
		for _, item := range t {
			b.WriteString(next)
			writeHCLValue(b, item, next)
			b.WriteString(",\n")
		}
		fmt.Fprintf(b, "%s]", indent)
	default:
		b.WriteString(hclScalar(v))
	}
}

// hclScalar renders scalars for HCL/jsonencode.
func hclScalar(v any) string {
	switch t := v.(type) {
	case string:
		return fmt.Sprintf("%q", t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%v", t)
	case int, int32, int64:
		return fmt.Sprintf("%d", t)
	case nil:
		return "null"
	}
	data, err := json.Marshal(v)
	if err != nil {
		return `""`
	}
	return string(data)
}

// hclKey quotes a key when it is not a bare HCL identifier.
var hclBareKeyRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)

func hclKey(k string) string {
	if hclBareKeyRE.MatchString(k) {
		return k
	}
	return fmt.Sprintf("%q", k)
}
