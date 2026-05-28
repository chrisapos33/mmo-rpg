package domain

import (
	"time"

	"github.com/lib/pq"
)

type GitHubConnection struct {
	ID                int64          `db:"id"                 json:"id"`
	UserID            int64          `db:"user_id"            json:"user_id"`
	GitHubUsername    string         `db:"github_username"    json:"github_username"`
	GitHubUserID      int64          `db:"github_user_id"     json:"github_user_id"`
	AccessToken       string         `db:"access_token"       json:"-"`
	AvatarURL         *string        `db:"avatar_url"         json:"avatar_url"`
	RepoCount         int            `db:"repo_count"         json:"repo_count"`
	StarCount         int            `db:"star_count"         json:"star_count"`
	Followers         int            `db:"followers"          json:"followers"`
	TopLanguages      pq.StringArray `db:"top_languages"      json:"top_languages"`
	ContributionScore int            `db:"contribution_score" json:"contribution_score"`
	SyncedAt          *time.Time     `db:"synced_at"          json:"synced_at"`
	CreatedAt         time.Time      `db:"created_at"         json:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at"         json:"updated_at"`
}
