package dsconfig

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// RenderMarkdownDocs renders a consumer-facing Markdown configuration guide
// for a datasource from its dsconfig schema. The output is intended for
// operators configuring the plugin in Grafana, not for plugin authors — it
// deliberately omits internal details like storage keys, roles, storage
// mappings, effects, and LLM-tagged instructions.
//
// The schema is expected to have base fields already resolved (via
// ResolveBaseFields or ParseAndResolveSchemaJSON). If it has not, unresolved
// base field references are ignored and only the plugin-declared fields
// appear in the output.
func RenderMarkdownDocs(s *Schema) (string, error) {
	if s == nil {
		return "", fmt.Errorf("schema is nil")
	}
	if s.PluginName == "" || s.PluginType == "" {
		return "", fmt.Errorf("schema is missing pluginName or pluginType")
	}

	var b strings.Builder

	// Title & intro.
	fmt.Fprintf(&b, "# %s configuration\n\n", s.PluginName)
	fmt.Fprintf(&b, "How to configure the **%s** data source (`%s`) in Grafana.\n\n",
		s.PluginName, s.PluginType)
	if s.DocURL != "" {
		fmt.Fprintf(&b, "For more information, see the [official documentation](%s).\n\n", s.DocURL)
	}
	b.WriteString("> This page is generated from [`dsconfig.json`](dsconfig.json). ")
	b.WriteString("Do not edit it by hand — run `go generate ./...` to refresh.\n\n")

	// Index fields for quick lookup and detect item-only fields (fields that
	// only appear nested inside another field's array item schema).
	fieldsByID := map[string]*ConfigField{}
	for i := range s.Fields {
		f := &s.Fields[i]
		fieldsByID[f.ID] = f
	}
	itemOnly := map[string]bool{}
	for i := range s.Fields {
		f := &s.Fields[i]
		if f.Item != nil {
			for j := range f.Item.Fields {
				itemOnly[f.Item.Fields[j].ID] = true
			}
		}
	}

	// Partition fields into groups and "ungrouped".
	grouped := map[string]bool{}
	for _, g := range s.Groups {
		for _, ref := range g.FieldRefs {
			grouped[ref] = true
		}
	}

	// Groups (sorted by Order, then declaration order).
	groups := make([]ConfigGroup, len(s.Groups))
	copy(groups, s.Groups)
	sort.SliceStable(groups, func(i, j int) bool {
		oi, oj := 0, 0
		if groups[i].Order != nil {
			oi = *groups[i].Order
		}
		if groups[j].Order != nil {
			oj = *groups[j].Order
		}
		return oi < oj
	})

	if len(groups) > 0 {
		b.WriteString("## Configuration sections\n\n")
		for _, g := range groups {
			anchor := anchorize(g.Title)
			fmt.Fprintf(&b, "- [%s](#%s)", groupTitle(g), anchor)
			if g.Optional {
				b.WriteString(" — _optional_")
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Render each group.
	for _, g := range groups {
		fmt.Fprintf(&b, "## %s\n\n", groupTitle(g))
		if g.Description != "" {
			b.WriteString(g.Description)
			b.WriteString("\n\n")
		}
		if g.Optional {
			b.WriteString("_This section is optional._\n\n")
		}

		visible := 0
		for _, ref := range g.FieldRefs {
			f, ok := fieldsByID[ref]
			if !ok {
				continue
			}
			if !isConsumerVisible(f) {
				continue
			}
			renderField(&b, f, s, "###")
			visible++
		}
		if visible == 0 {
			b.WriteString("_No user-configurable fields in this section._\n\n")
		}
	}

	// Ungrouped fields, if any.
	var ungrouped []*ConfigField
	for i := range s.Fields {
		f := &s.Fields[i]
		if grouped[f.ID] || itemOnly[f.ID] {
			continue
		}
		if !isConsumerVisible(f) {
			continue
		}
		ungrouped = append(ungrouped, f)
	}
	if len(ungrouped) > 0 {
		b.WriteString("## Other settings\n\n")
		for _, f := range ungrouped {
			renderField(&b, f, s, "###")
		}
	}

	// If there are no visible fields at all, still produce a useful note.
	if len(groups) == 0 && len(ungrouped) == 0 {
		b.WriteString("_This data source has no user-configurable fields._\n\n")
	}

	return b.String(), nil
}

// groupTitle falls back to the group ID when title is empty.
func groupTitle(g ConfigGroup) string {
	if g.Title != "" {
		return g.Title
	}
	return g.ID
}

// isConsumerVisible reports whether a field should appear in the
// consumer-facing document. Virtual fields (no storage) that carry no UI —
// typically internal discriminators — are hidden.
func isConsumerVisible(f *ConfigField) bool {
	if f == nil {
		return false
	}
	if f.Kind == VirtualField && f.UI == nil {
		return false
	}
	// A single-value select whose only option is a discriminator (e.g. the
	// only auth method is "none") adds no user choice — but it still
	// documents the fact that no auth is used, so keep it visible when it
	// has a label. If it has no label at all, skip.
	if f.Label == "" && f.Description == "" && f.Help == nil {
		return false
	}
	return true
}

// renderField writes a single field entry using the given heading prefix
// (e.g. "###" for group children, "####" for nested item fields).
func renderField(b *strings.Builder, f *ConfigField, s *Schema, headingPrefix string) {
	label := f.Label
	if label == "" {
		label = f.Key
	}
	fmt.Fprintf(b, "%s %s\n\n", headingPrefix, label)

	// One-line summary badges.
	var badges []string
	if isSecret(f) {
		badges = append(badges, "🔒 secret (write-only)")
	}
	switch {
	case f.Required:
		badges = append(badges, "**required**")
	case f.RequiredWhen != "":
		badges = append(badges, "conditionally required")
	default:
		badges = append(badges, "optional")
	}
	if t := humanValueType(f); t != "" {
		badges = append(badges, t)
	}
	if len(badges) > 0 {
		b.WriteString("_" + strings.Join(badges, " · ") + "_\n\n")
	}

	if f.Description != "" {
		b.WriteString(f.Description)
		if !strings.HasSuffix(f.Description, ".") {
			b.WriteString(".")
		}
		b.WriteString("\n\n")
	}

	// Details table.
	rows := [][2]string{}
	if def := formatDefault(f.DefaultValue); def != "" {
		rows = append(rows, [2]string{"Default", def})
	}
	if p := placeholder(f); p != "" {
		rows = append(rows, [2]string{"Example", "`" + p + "`"})
	}
	if opts := formatOptions(f); opts != "" {
		rows = append(rows, [2]string{"Allowed values", opts})
	}
	if allowed := formatAllowedValues(f); allowed != "" {
		rows = append(rows, [2]string{"Allowed values", allowed})
	}
	if r := formatRange(f); r != "" {
		rows = append(rows, [2]string{"Range", r})
	}
	if p := formatPattern(f); p != "" {
		rows = append(rows, [2]string{"Must match", p})
	}
	if f.DependsOn != "" {
		rows = append(rows, [2]string{"Shown when", humanCondition(f.DependsOn, s)})
	}
	if f.RequiredWhen != "" && f.RequiredWhen != f.DependsOn {
		rows = append(rows, [2]string{"Required when", humanCondition(f.RequiredWhen, s)})
	}
	if f.DisabledWhen != "" {
		rows = append(rows, [2]string{"Disabled when", humanCondition(f.DisabledWhen, s)})
	}
	if len(rows) > 0 {
		b.WriteString("| | |\n|---|---|\n")
		for _, r := range rows {
			fmt.Fprintf(b, "| %s | %s |\n", r[0], r[1])
		}
		b.WriteString("\n")
	}

	// Help block.
	if f.Help != nil && f.Help.Markdown != "" {
		if f.Help.Title != "" {
			fmt.Fprintf(b, "**%s**\n\n", f.Help.Title)
		}
		b.WriteString(f.Help.Markdown)
		if !strings.HasSuffix(f.Help.Markdown, "\n") {
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Per-field doc URL.
	if f.DocURL != "" {
		fmt.Fprintf(b, "[Learn more](%s)\n\n", f.DocURL)
	}

	// Array item schema: describe the shape of each item.
	if f.Item != nil && len(f.Item.Fields) > 0 {
		b.WriteString("Each item has the following fields:\n\n")
		for i := range f.Item.Fields {
			itemF := &f.Item.Fields[i]
			if !isConsumerVisible(itemF) {
				continue
			}
			renderField(b, itemF, s, headingPrefix+"#")
		}
	}
}

// isSecret returns true if the field is stored in secureJsonData.
func isSecret(f *ConfigField) bool {
	return f.Target != nil && *f.Target == SecureJSONTarget
}

// humanValueType renders a short human label for the field's value type.
func humanValueType(f *ConfigField) string {
	if f.UI != nil {
		switch f.UI.Component {
		case UISelect:
			return "select"
		case UIMultiselect:
			return "multi-select"
		case UIRadio:
			return "radio"
		case UICheckbox, UISwitch:
			return "toggle"
		case UITextarea:
			return "multiline text"
		case UICode:
			if f.UI.Language != "" {
				return f.UI.Language + " code"
			}
			return "code"
		case UIKeyValue:
			return "key/value pairs"
		case UIList:
			return "list"
		case UIFileUpload:
			return "file upload"
		}
	}
	switch f.ValueType {
	case "string":
		return "string"
	case "number":
		return "number"
	case "integer":
		return "integer"
	case "boolean":
		return "boolean"
	case "array":
		return "list"
	case "object":
		return "object"
	}
	return ""
}

// formatDefault renders the default value as inline markdown code.
func formatDefault(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		if t == "" {
			return "`\"\"`"
		}
		return "`" + t + "`"
	case bool:
		if t {
			return "`true`"
		}
		return "`false`"
	case float64:
		return fmt.Sprintf("`%v`", t)
	}
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return "`" + string(data) + "`"
}

// placeholder returns the UI placeholder as an example, if any.
func placeholder(f *ConfigField) string {
	if f.UI == nil {
		return ""
	}
	return f.UI.Placeholder
}

// formatOptions renders UI-provided select/radio options as a list.
func formatOptions(f *ConfigField) string {
	if f.UI == nil || len(f.UI.Options) == 0 {
		return ""
	}
	parts := make([]string, 0, len(f.UI.Options))
	for _, o := range f.UI.Options {
		val := fmt.Sprintf("%v", o.Value)
		label := o.Label
		if label == "" || label == val {
			parts = append(parts, "`"+val+"`")
		} else {
			parts = append(parts, fmt.Sprintf("`%s` (%s)", val, label))
		}
	}
	return strings.Join(parts, ", ")
}

// formatAllowedValues renders any allowedValues validation rules as a list.
// If UI options already describe the same set, prefer those and return "".
func formatAllowedValues(f *ConfigField) string {
	if f.UI != nil && len(f.UI.Options) > 0 {
		return ""
	}
	var parts []string
	for _, v := range f.Validations {
		if v.Type != AllowedValuesValidation {
			continue
		}
		for _, val := range v.Values {
			parts = append(parts, fmt.Sprintf("`%v`", val))
		}
	}
	return strings.Join(parts, ", ")
}

// formatRange renders numeric range / length constraints.
func formatRange(f *ConfigField) string {
	for _, v := range f.Validations {
		if v.Type != RangeValidation && v.Type != LengthValidation && v.Type != ItemCountValidation {
			continue
		}
		unit := ""
		switch v.Type {
		case LengthValidation:
			unit = " (characters)"
		case ItemCountValidation:
			unit = " (items)"
		}
		switch {
		case v.Min != nil && v.Max != nil:
			return fmt.Sprintf("%v – %v%s", *v.Min, *v.Max, unit)
		case v.Min != nil:
			return fmt.Sprintf("at least %v%s", *v.Min, unit)
		case v.Max != nil:
			return fmt.Sprintf("at most %v%s", *v.Max, unit)
		}
	}
	return ""
}

// formatPattern returns the first pattern validation as a code block.
func formatPattern(f *ConfigField) string {
	for _, v := range f.Validations {
		if v.Type == PatternValidation && v.Pattern != "" {
			return "`" + v.Pattern + "`"
		}
	}
	return ""
}

// humanCondition converts a simple CEL-style condition like
// "jsonData_services_x_auth_id == 'bearer_token'" into a more readable
// form using field labels and option labels where possible.
// For compound expressions (e.g. containing `||`, `&&`, `!=`, `<`, `>`),
// the raw expression is returned wrapped in a code block.
func humanCondition(expr string, s *Schema) string {
	trimmed := strings.TrimSpace(expr)
	// Only apply the pretty renderer to simple single-comparison expressions.
	if strings.ContainsAny(trimmed, "|&!<>") {
		return "`" + expr + "`"
	}
	idx := strings.Index(trimmed, "==")
	if idx <= 0 || strings.Count(trimmed, "==") != 1 {
		return "`" + expr + "`"
	}
	left := strings.TrimSpace(trimmed[:idx])
	right := strings.TrimSpace(trimmed[idx+2:])
	right = strings.Trim(right, "'\"")
	f := findField(s, left)
	if f == nil {
		return "`" + expr + "`"
	}
	label := f.Label
	if label == "" {
		label = f.Key
	}
	if f.UI != nil {
		for _, o := range f.UI.Options {
			if fmt.Sprintf("%v", o.Value) == right && o.Label != "" && o.Label != right {
				return fmt.Sprintf("**%s** is **%s** (`%s`)", label, o.Label, right)
			}
		}
	}
	return fmt.Sprintf("**%s** is `%s`", label, right)
}

// findField locates a field by ID inside the top-level schema (does not
// recurse into item field schemas).
func findField(s *Schema, id string) *ConfigField {
	for i := range s.Fields {
		if s.Fields[i].ID == id {
			return &s.Fields[i]
		}
	}
	return nil
}

// anchorize converts a heading title into a GitHub-style anchor slug.
func anchorize(title string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(title) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			b.WriteRune('-')
		}
	}
	return b.String()
}
