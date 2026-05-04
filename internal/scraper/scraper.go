package scraper

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/temoto/robotstxt"
)

const (
	baseURL        = "https://www.postgresql.org"
	listPath       = "/docs/release/"
	robotsPath     = "/robots.txt"
	botName        = "pg-release-scraper"
	userAgent      = botName + " (+https://github.com/hidetzu/pg-release-scraper)"
	requestDelay   = 500 * time.Millisecond
	maxRetries     = 3
	initialBackoff = time.Second
)

var (
	httpClient = &http.Client{Timeout: 30 * time.Second}
	wsRe       = regexp.MustCompile(`\s+`)
)

type Release struct {
	Version string
	Detail  string
}

type Version struct {
	Parts []int
	Raw   string
}

func ParseVersion(s string) (Version, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Version{}, fmt.Errorf("empty version")
	}
	parts := strings.Split(s, ".")
	out := make([]int, len(parts))
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return Version{}, fmt.Errorf("invalid version %q: %w", s, err)
		}
		out[i] = n
	}
	return Version{Parts: out, Raw: s}, nil
}

func (v Version) Compare(o Version) int {
	n := len(v.Parts)
	if len(o.Parts) > n {
		n = len(o.Parts)
	}
	for i := 0; i < n; i++ {
		var a, b int
		if i < len(v.Parts) {
			a = v.Parts[i]
		}
		if i < len(o.Parts) {
			b = o.Parts[i]
		}
		if a != b {
			return a - b
		}
	}
	return 0
}

func Fetch(start, end string) ([]Release, error) {
	s, err := ParseVersion(start)
	if err != nil {
		return nil, err
	}
	e, err := ParseVersion(end)
	if err != nil {
		return nil, err
	}
	if s.Compare(e) > 0 {
		return nil, fmt.Errorf("start (%s) is greater than end (%s)", start, end)
	}

	robots, err := fetchRobots()
	if err != nil {
		return nil, fmt.Errorf("fetch robots.txt: %w", err)
	}
	if !robots.TestAgent(listPath, botName) {
		return nil, fmt.Errorf("robots.txt disallows %s for %s", listPath, botName)
	}

	doc, err := fetchDoc(baseURL + listPath)
	if err != nil {
		return nil, fmt.Errorf("fetch release list: %w", err)
	}

	type entry struct {
		ver  Version
		href string
	}
	var entries []entry

	doc.Find(".release-notes-list a").Each(func(_ int, sel *goquery.Selection) {
		text := strings.TrimSpace(sel.Text())
		href, ok := sel.Attr("href")
		if !ok {
			return
		}
		v, err := ParseVersion(text)
		if err != nil {
			return
		}
		if v.Compare(s) >= 0 && v.Compare(e) <= 0 {
			entries = append(entries, entry{v, strings.TrimSpace(href)})
		}
	})

	// Listing page is newest-first; iterate oldest → newest.
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	var releases []Release
	for i, ent := range entries {
		if !robots.TestAgent(ent.href, botName) {
			return nil, fmt.Errorf("robots.txt disallows %s for %s", ent.href, botName)
		}
		if i > 0 {
			time.Sleep(requestDelay)
		}
		items, err := fetchReleaseDetail(baseURL+ent.href, ent.ver.Raw)
		if err != nil {
			return nil, fmt.Errorf("fetch %s: %w", ent.ver.Raw, err)
		}
		releases = append(releases, items...)
	}
	return releases, nil
}

func fetchRobots() (*robotstxt.RobotsData, error) {
	resp, err := httpGet(baseURL + robotsPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return robotstxt.FromResponse(resp)
}

func fetchReleaseDetail(url, version string) ([]Release, error) {
	doc, err := fetchDoc(url)
	if err != nil {
		return nil, err
	}
	// Strip PostgreSQL's commit-link icons (rendered as bare "§" in extracted text).
	doc.Find(".listitem a.ulink").Remove()
	var items []Release
	doc.Find(".listitem").Each(func(_ int, sel *goquery.Selection) {
		var paras []string
		sel.Find("p").Each(func(_ int, p *goquery.Selection) {
			text := strings.TrimSpace(wsRe.ReplaceAllString(p.Text(), " "))
			if text != "" {
				paras = append(paras, text)
			}
		})
		detail := strings.Join(paras, "\n")
		if detail == "" {
			detail = strings.TrimSpace(wsRe.ReplaceAllString(sel.Text(), " "))
		}
		items = append(items, Release{Version: version, Detail: detail})
	})
	return items, nil
}

func fetchDoc(url string) (*goquery.Document, error) {
	resp, err := httpGet(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}
	return goquery.NewDocumentFromReader(resp.Body)
}

// httpGet performs a GET with exponential backoff on transient failures
// (network errors, 429, 5xx). Other non-2xx responses are returned as-is
// so callers can decide how to handle them.
func httpGet(url string) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(initialBackoff << (attempt - 1))
		}
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", userAgent)
		resp, err := httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		if !shouldRetry(resp.StatusCode) {
			return resp, nil
		}
		resp.Body.Close()
		lastErr = fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}
	return nil, fmt.Errorf("after %d retries: %w", maxRetries, lastErr)
}

func shouldRetry(status int) bool {
	return status == http.StatusTooManyRequests || (status >= 500 && status < 600)
}
