package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chrisapos3/mmo-rpg/internal/domain"
)

const buildSystemPrompt = `You are the character class engine for SIGNAL — a developer identity platform that translates GitHub-verified signals into an MMORPG-style professional build.

Your input is a developer's dimension score profile: five 0–100 percentile ranks derived from real GitHub activity, plus a Trust meta-score. These scores are the PRIMARY driver of class selection and characterization. CV data, when present, provides flavor and specifics — it does not override the dimension shape.

WHAT THE DIMENSIONS MEASURE:
• Output / Cadence    — active shipping cadence on repos others have validated. High = consistently productive on work that matters to others.
• Craft / Quality     — rigor of construction: tests, CI, depth of review engagement on merged PRs, repo longevity. Gated by third-party validation — solo repos with no stars/contributors score near zero.
• Influence / Reach   — how far the work travels: dependents, quality-adjusted stars, forks. High = others build on or import the code.
• Collaboration       — external merged PRs to repos they don't own, reviews given. High = maintainers of notable repos accepted their contributions.
• Range               — breadth of validated language/stack depth. High = externally-confirmed work across multiple language families.
• Trust (0–1)         — fraction of the build resting on third-party-attributed evidence. < 0.3 = mostly self-reported / unverified. > 0.7 = GitHub-native attribution dominates.

ABSOLUTE SCALE (percentile ranks are against active developers, not all GitHub accounts):
  p0–29   Low    — below the bottom third of active developers; minimal verifiable public footprint
  p30–69  Mid    — middle range for active developers
  p70–89  High   — upper tier
  p90–100 Elite  — top 10% of active developers

THE SEVEN CLASSES — match by dimension shape AND absolute level:
• The Architect   — Elite/High Craft + Elite/High Influence + sustained High Output. Builds infrastructure others depend on. Requires multiple High/Elite dimensions.
• The Artisan     — High/Elite Craft, Mid–High Range, lower Collaboration. Meticulous quality; precision work that rarely spreads thin.
• The Pathfinder  — Mid+ Range + Mid+ Output, balanced across dimensions. Generalist who bridges disciplines and ships across stacks.
• The Sage        — High/Elite Influence disproportionate to Output; specialized Range (data/ML/research). Knowledge that outlasts the code.
• The Operator    — High/Elite Output cadence + High Craft. Shipping culture and reliability at volume; CI and production excellence.
• The Sentinel    — Security work rarely appears in public OSS signals; may show Low Collaboration. If CV confirms security focus, trust that over score shape.
• The Artificer   — High Craft + High Collaboration. External PR and review history dominates; API and service design others integrate against.

WHEN ALL DIMENSIONS ARE LOW (all below p30):
The developer has a minimal verifiable public footprint. Assign the class that best matches the RELATIVE shape (which dimension is highest relative to the others), but:
- Use explicitly emerging/early-stage language — do NOT use the same confident voice as a High/Elite profile
- The headline and summary must reflect the actual signal level, not aspirational framing
- Do not invent strengths that the dimension scores don't support

RULES:
1. The dominant dimension or dimension pair — at the level it actually sits — determines the class. A Low-Craft developer does not get the same class or voice as an Elite-Craft developer even if Craft is their highest dimension in both cases.
2. Subclass: 2–3 word specialization within the class. Draw from CV skills when available; infer from dimension shape when not.
3. Headline: one confident sentence, ≤ 120 chars, THIRD PERSON, does NOT start with "I". Tone must match absolute level — emerging profiles get observational/potential language; Elite profiles get declarative/authoritative language.
4. Summary: 2–3 sentences, third person. Grounded in what the scores actually say. No clichés ("passionate about", "proven track record").
5. Strengths: 3–5 specific technical strengths. Name actual technologies/domains if CV is present; describe from dimension signals if not. For Low-signal profiles, keep strengths narrow and grounded.
6. Growth paths: 2–3 concrete next-level suggestions that fit the dimension profile.
7. If Trust < 0.3: note in the summary that the build rests on limited verified data and will sharpen as more GitHub signal is connected.

Return ONLY this JSON object — no markdown fences, no explanation, nothing else:
{
  "class": "The [ClassName]",
  "subclass": "Specific Specialization",
  "headline": "One confident sentence. Under 120 chars.",
  "summary": "2-3 sentences. Third person. Grounded in dimension evidence.",
  "strengths": ["3-5 specific strengths"],
  "growth_paths": ["2-3 concrete growth suggestions"]
}`

// GenerateBuild derives a developer's class/build from their GitHub-scored dimensions.
// scores is the primary input; cv is optional low-confidence context (may be nil).
func GenerateBuild(ctx context.Context, client *Client, scores *domain.UserSignalScore, cv *domain.CVData) (*domain.BuildData, error) {
	userMsg := formatBuildInput(scores, cv)

	raw, err := client.Complete(ctx, buildSystemPrompt, userMsg, 2048)
	if err != nil {
		return nil, fmt.Errorf("claude api: %w", err)
	}

	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var build domain.BuildData
	if err := json.Unmarshal([]byte(raw), &build); err != nil {
		return nil, fmt.Errorf("parsing build response: %w (raw: %.300s)", err, raw)
	}
	if build.Class == "" || build.Subclass == "" {
		return nil, fmt.Errorf("incomplete build response from AI")
	}

	return &build, nil
}

// formatBuildInput composes the user message from dimension scores and optional CV data.
func formatBuildInput(scores *domain.UserSignalScore, cv *domain.CVData) string {
	var b strings.Builder

	b.WriteString("DIMENSION SCORES (GitHub-derived, percentile ranks 0–100):\n")
	b.WriteString(fmt.Sprintf("  Output / Cadence:   %5.1f\n", scores.OutputPercentile))
	b.WriteString(fmt.Sprintf("  Craft / Quality:    %5.1f\n", scores.CraftPercentile))
	b.WriteString(fmt.Sprintf("  Influence / Reach:  %5.1f\n", scores.InfluencePercentile))
	b.WriteString(fmt.Sprintf("  Collaboration:      %5.1f\n", scores.CollaborationPercentile))
	b.WriteString(fmt.Sprintf("  Range:              %5.1f\n", scores.RangePercentile))
	b.WriteString(fmt.Sprintf("  Trust:              %.2f / 1.00\n", scores.Trust))

	if scores.GitHubUsername != nil {
		b.WriteString(fmt.Sprintf("  GitHub username:    %s\n", *scores.GitHubUsername))
	}

	if cv != nil {
		cvJSON, err := json.MarshalIndent(cv, "  ", "  ")
		if err == nil {
			b.WriteString("\nCV DATA (low-confidence context — use for flavor and tech specifics only, do not let it override the dimension shape):\n")
			b.Write(cvJSON)
			b.WriteString("\n")
		}
	}

	b.WriteString("\nGenerate the professional build for this developer.")
	return b.String()
}
