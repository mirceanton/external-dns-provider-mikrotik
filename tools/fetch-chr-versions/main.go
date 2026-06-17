package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type matrixEntry struct {
	Version      string `json:"version"`
	Channel      string `json:"channel"`
	DownloadURL  string `json:"download_url"`
	AllowFailure bool   `json:"allow_failure"`
}

type rssDoc struct {
	XMLName xml.Name `xml:"rss"`
	Items   []struct {
		Title string `xml:"title"`
	} `xml:"channel>item"`
}

var versionRe = regexp.MustCompile(`RouterOS\s+(\S+)`)

func fetchVersion(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var doc rssDoc
	if err := xml.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return "", err
	}
	if len(doc.Items) == 0 {
		return "", fmt.Errorf("no items in feed")
	}

	m := versionRe.FindStringSubmatch(doc.Items[0].Title)
	if m == nil {
		return "", fmt.Errorf("no version found in title: %q", doc.Items[0].Title)
	}
	return m[1], nil
}

func main() {
	includeUnstable := os.Getenv("INCLUDE_UNSTABLE") == "true"

	type feedInfo struct {
		url     string
		channel string
		stable  bool
	}

	allFeeds := []feedInfo{
		{"https://cdn.mikrotik.com/routeros/latest-stable.rss", "stable", true},
		{"https://cdn.mikrotik.com/routeros/latest-long-term.rss", "long-term", true},
		{"https://cdn.mikrotik.com/routeros/latest-testing.rss", "testing", false},
		{"https://cdn.mikrotik.com/routeros/latest-development.rss", "development", false},
	}

	var entries []matrixEntry
	seen := map[string]bool{}

	for _, feed := range allFeeds {
		if !feed.stable && !includeUnstable {
			continue
		}

		version, err := fetchVersion(feed.url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to fetch %s: %v\n", feed.url, err)
			continue
		}

		if !strings.HasPrefix(version, "7.") {
			continue
		}

		if seen[version] {
			continue
		}
		seen[version] = true

		entries = append(entries, matrixEntry{
			Version:     version,
			Channel:     feed.channel,
			DownloadURL: fmt.Sprintf("https://download.mikrotik.com/routeros/%s/chr-%s.img.zip", version, version),
			// Stable channels (stable, long-term) must pass; testing/development are allowed to fail.
			AllowFailure: !feed.stable,
		})
	}

	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "error: no CHR versions found")
		os.Exit(1)
	}

	out, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to marshal output: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}
