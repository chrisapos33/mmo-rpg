package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"

	"github.com/chrisapos3/mmo-rpg/internal/domain"
)

const cvSystemPrompt = `You are a precise data extractor for a professional identity platform.

Extract structured information from the CV/resume text provided.
Return ONLY a valid JSON object — no markdown fences, no explanation, no preamble.

Required schema:
{
  "full_name": "string",
  "email": "string or null",
  "location": "string or null",
  "summary": "string or null",
  "experiences": [
    {
      "company": "string",
      "title": "string",
      "start_date": "YYYY-MM or YYYY",
      "end_date": "YYYY-MM or YYYY or null if current role",
      "is_current": true or false,
      "description": "concise string summarising responsibilities or null"
    }
  ],
  "skills": ["array of individual skill strings, be specific"],
  "education": [
    {
      "institution": "string",
      "degree": "string or null",
      "field": "string or null",
      "year": "graduation year string or null"
    }
  ],
  "languages": ["programming languages and spoken languages"],
  "inferred_specializations": [
    "2 to 4 concise professional specialization labels inferred from the full profile,
     e.g. Platform Engineering, Frontend Architecture, ML Infrastructure, DevOps"
  ]
}`

// ExtractTextFromPDF reads all text content from a PDF file.
func ExtractTextFromPDF(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening pdf: %w", err)
	}
	defer f.Close()

	reader, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("extracting pdf text: %w", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		return "", fmt.Errorf("reading pdf text: %w", err)
	}

	text := strings.TrimSpace(buf.String())
	if text == "" {
		return "", fmt.Errorf("no text content found in PDF")
	}
	return text, nil
}

// ParseCV sends CV text to Claude and returns structured CVData.
func ParseCV(ctx context.Context, client *Client, cvText string) (*domain.CVData, error) {
	// Truncate very long CVs — Claude has a context limit and CVs rarely need more than this
	if len(cvText) > 20000 {
		cvText = cvText[:20000]
	}

	userMsg := "Extract structured data from this CV:\n\n" + cvText

	raw, err := client.Complete(ctx, cvSystemPrompt, userMsg, 4096)
	if err != nil {
		return nil, fmt.Errorf("claude api: %w", err)
	}

	// Strip accidental markdown fences
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var data domain.CVData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, fmt.Errorf("parsing claude response: %w (raw: %.200s)", err, raw)
	}

	return &data, nil
}
