package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// FetchRegistryDependents queries the package registry appropriate for the given
// primary language and returns a download-based proxy for dependency impact.
// Returns 0 if the package is not found, the language has no supported registry,
// or the registry is unreachable.
func FetchRegistryDependents(ctx context.Context, primaryLang, repoName string) int {
	switch primaryLang {
	case "JavaScript", "TypeScript":
		return FetchNPMDownloads(ctx, repoName)
	case "Rust":
		return FetchCratesDownloads(ctx, repoName)
	case "Python":
		return FetchPyPIDownloads(ctx, repoName)
	default:
		return 0
	}
}

// FetchNPMDownloads returns the total recent download count for an npm package via npms.io.
func FetchNPMDownloads(ctx context.Context, pkg string) int {
	url := fmt.Sprintf("https://api.npms.io/v2/package/%s", pkg)
	var resp struct {
		Collected struct {
			NPM struct {
				Downloads []struct {
					Count int `json:"count"`
				} `json:"downloads"`
			} `json:"npm"`
		} `json:"collected"`
	}
	if err := registryGet(ctx, url, &resp); err != nil {
		return 0
	}
	total := 0
	for _, d := range resp.Collected.NPM.Downloads {
		total += d.Count
	}
	return total
}

// FetchCratesDownloads returns the recent download count for a crates.io crate.
func FetchCratesDownloads(ctx context.Context, crate string) int {
	url := fmt.Sprintf("https://crates.io/api/v1/crates/%s", crate)
	var resp struct {
		Crate struct {
			RecentDownloads int `json:"recent_downloads"`
		} `json:"crate"`
	}
	if err := registryGet(ctx, url, &resp); err != nil {
		return 0
	}
	return resp.Crate.RecentDownloads
}

// FetchPyPIDownloads returns the last-month download count for a PyPI package.
func FetchPyPIDownloads(ctx context.Context, pkg string) int {
	url := fmt.Sprintf("https://pypistats.org/api/packages/%s/recent", pkg)
	var resp struct {
		Data struct {
			LastMonth int `json:"last_month"`
		} `json:"data"`
	}
	if err := registryGet(ctx, url, &resp); err != nil {
		return 0
	}
	return resp.Data.LastMonth
}

func registryGet(ctx context.Context, url string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "mmo-rpg-scorer/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("registry GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registry GET %s returned %d", url, resp.StatusCode)
	}
	return json.Unmarshal(raw, dest)
}
