package health

import "sync"

// Remediation is the structured, surfaced guidance for a failure.
type Remediation struct {
	Summary string   `json:"summary"`
	Steps   []string `json:"steps,omitempty"`
	DocsURL string   `json:"docsUrl,omitempty"`
	Fields  []string `json:"fields,omitempty"`
}

func (r Remediation) isZero() bool {
	return r.Summary == "" && len(r.Steps) == 0 && r.DocsURL == "" && len(r.Fields) == 0
}

// Match is a declarative condition set. Empty fields are wildcards; a Match with
// all-empty fields matches everything. Conditions are ANDed.
type Match struct {
	Code         Code
	ProviderCode string
	TLSKind      TLSKind
	TimeoutKind  TimeoutKind
	BodyKind     BodyKind
	AuthType     string
	Deployment   string
}

// Rule pairs a Match with the Remediation to give when it matches. Rules are
// authored as plain Go literals and registered with RegisterRules (ADR-003).
type Rule struct {
	When Match
	Give Remediation
}

var (
	rulesMu sync.RWMutex
	rules   []Rule
)

// RegisterRules adds remediation rules to the global resolver. Safe for use from
// init(). Resolution is by specificity, not registration order (ADR-004).
func RegisterRules(rs ...Rule) {
	rulesMu.Lock()
	defer rulesMu.Unlock()
	rules = append(rules, rs...)
}

// Field weights for specificity scoring. ProviderCode and TLSKind are the most
// discriminating, so they weigh most (RFC §6.5).
const (
	weightProviderCode = 8
	weightTLSKind      = 8
	weightTimeoutKind  = 4
	weightBodyKind     = 4
	weightAuthType     = 2
	weightDeployment   = 2
	weightCode         = 1
)

// matches reports whether d satisfies every non-empty condition in m, and the
// summed specificity weight of the conditions that had to match.
func (m Match) matches(d Diagnosis) (bool, int) {
	score := 0
	if m.Code != "" {
		if m.Code != d.Code {
			return false, 0
		}
		score += weightCode
	}
	if m.ProviderCode != "" {
		if m.ProviderCode != d.ProviderCode {
			return false, 0
		}
		score += weightProviderCode
	}
	if m.TLSKind != "" {
		if m.TLSKind != d.TLSKind {
			return false, 0
		}
		score += weightTLSKind
	}
	if m.TimeoutKind != "" {
		if m.TimeoutKind != d.TimeoutKind {
			return false, 0
		}
		score += weightTimeoutKind
	}
	if m.BodyKind != "" {
		if m.BodyKind != d.BodyKind {
			return false, 0
		}
		score += weightBodyKind
	}
	if m.AuthType != "" {
		if m.AuthType != d.Context.AuthType {
			return false, 0
		}
		score += weightAuthType
	}
	if m.Deployment != "" {
		if m.Deployment != d.Context.Deployment {
			return false, 0
		}
		score += weightDeployment
	}
	return true, score
}

// resolveRemediation picks the highest-specificity matching rule, falling back to
// the catalog's generic remediation when no rule matches (RFC §6.5). The docs
// base URL from Context is applied to a relative DocsURL.
func resolveRemediation(d Diagnosis) Remediation {
	rulesMu.RLock()
	best := -1
	var chosen Remediation
	for _, r := range rules {
		ok, score := r.When.matches(d)
		if ok && score > best {
			best = score
			chosen = r.Give
		}
	}
	rulesMu.RUnlock()

	if best < 0 || chosen.isZero() {
		chosen = entryFor(d.Code).Fallback
	}
	return chosen
}
