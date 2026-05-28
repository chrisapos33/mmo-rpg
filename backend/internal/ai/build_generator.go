package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chrisapos3/mmo-rpg/internal/domain"
)

const buildSystemPrompt = `You are the identity engine for SIGNAL — a professional identity platform for engineers and builders.

Analyze the CV data and assign this professional to exactly one of the seven classes below. Then generate their professional build identity.

THE SEVEN CLASSES:
- The Architect   — Designs infrastructure, platforms, distributed systems, large-scale backend
- The Artisan     — Frontend engineering, design systems, UI/UX craft, interface quality
- The Pathfinder  — Full-stack generalist who bridges disciplines, moves fast, ships end-to-end
- The Sage        — ML, data engineering, AI systems, analytics, research, data science
- The Operator    — DevOps, SRE, platform reliability, cloud operations, production excellence
- The Sentinel    — Security engineering, penetration testing, compliance, threat modeling
- The Artificer   — Backend engineering, API design, service architecture, business logic

RULES:
- Pick the class that best reflects the dominant pattern of their entire career, not just latest role
- The subclass is a specific 2-3 word specialization within the class (e.g. Platform Engineer, ML Infrastructure, API Design, Frontend Performance)
- Base every claim on evidence from their actual experience — no generic statements
- Summary is third person, professional, no clichés, no phrases like "passionate about" or "proven track record"
- Headline does not start with "I" — it describes them as a practitioner

Return ONLY this JSON object — no markdown fences, no explanation, nothing else:
{
  "class": "The [ClassName]",
  "subclass": "Specific Specialization",
  "headline": "One confident sentence capturing their professional identity. Under 120 chars.",
  "summary": "2-3 sentences. Third person. Captures their depth, approach, and signal. Specific, not generic.",
  "strengths": ["3 to 5 specific technical strengths drawn from their actual experience"],
  "growth_paths": ["2 to 3 concrete growth trajectory suggestions based on their profile"]
}`

// GenerateBuild sends parsed CVData to Claude and returns a professional build identity.
func GenerateBuild(ctx context.Context, client *Client, cv *domain.CVData) (*domain.BuildData, error) {
	cvJSON, err := json.Marshal(cv)
	if err != nil {
		return nil, fmt.Errorf("marshalling cv data: %w", err)
	}

	userMsg := "Generate the professional build identity for this CV data:\n\n" + string(cvJSON)

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
