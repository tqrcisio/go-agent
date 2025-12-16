package agents

import (
	"context"
	"go-agent/internal/state"
	"sort"
	"strings"
)

type PrioritizeFindingsAgent struct{}

func NewPrioritizeFindingsAgent() *PrioritizeFindingsAgent {
	return &PrioritizeFindingsAgent{}
}

func (a *PrioritizeFindingsAgent) Name() string {
	return "PrioritizeFindingsAgent"
}

func (a *PrioritizeFindingsAgent) Run(ctx context.Context, s *state.State) error {
	// Sort findings by severity: HIGH > MEDIUM > LOW
	// We can use a helper map to assign integer values to severity levels
	severityScore := map[string]int{
		"HIGH":   3,
		"MEDIUM": 2,
		"LOW":    1,
	}

	sort.Slice(s.Findings, func(i, j int) bool {
		sev1 := strings.ToUpper(s.Findings[i].Severity)
		sev2 := strings.ToUpper(s.Findings[j].Severity)

		score1 := severityScore[sev1]
		score2 := severityScore[sev2]

		// If scores are different, sort by score descending
		if score1 != score2 {
			return score1 > score2
		}

		// If severity is the same, sort by file path for consistency
		if s.Findings[i].File != s.Findings[j].File {
			return s.Findings[i].File < s.Findings[j].File
		}

		return s.Findings[i].Line < s.Findings[j].Line
	})

	return nil
}
