package domain

import (
	"encoding/json"
	"time"
)

const (
	CVStatusProcessing = "processing"
	CVStatusDone       = "done"
	CVStatusFailed     = "failed"
)

type CVUpload struct {
	ID            int64           `db:"id"             json:"id"`
	UserID        int64           `db:"user_id"        json:"user_id"`
	StoragePath   string          `db:"storage_path"   json:"-"`
	OriginalName  string          `db:"original_name"  json:"original_name"`
	Status        string          `db:"status"         json:"status"`
	ExtractedData *json.RawMessage `db:"extracted_data" json:"extracted_data,omitempty"`
	ErrorMessage  *string         `db:"error_message"  json:"error_message,omitempty"`
	CreatedAt     time.Time       `db:"created_at"     json:"created_at"`
	ProcessedAt   *time.Time      `db:"processed_at"   json:"processed_at,omitempty"`
}

// CVData is the structured output from AI CV parsing.
type CVData struct {
	FullName                string         `json:"full_name"`
	Email                   *string        `json:"email"`
	Location                *string        `json:"location"`
	Summary                 *string        `json:"summary"`
	Experiences             []CVExperience `json:"experiences"`
	Skills                  []string       `json:"skills"`
	Education               []CVEducation  `json:"education"`
	Languages               []string       `json:"languages"`
	InferredSpecializations []string       `json:"inferred_specializations"`
}

type CVExperience struct {
	Company     string  `json:"company"`
	Title       string  `json:"title"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date"`
	IsCurrent   bool    `json:"is_current"`
	Description *string `json:"description"`
}

type CVEducation struct {
	Institution string  `json:"institution"`
	Degree      *string `json:"degree"`
	Field       *string `json:"field"`
	Year        *string `json:"year"`
}
