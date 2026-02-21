// Package tools provides the searchInternet tool for the agent.
package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const searchInternetTimeout = 15 * time.Second
const defaultNumResults = 10

// SearchInternetDefinition is the tool that searches the internet and returns result titles and URLs.
var SearchInternetDefinition = ToolDefinition{
	Name:        "searchInternet",
	Description: "Search the internet and return a list of result titles, URLs, and snippets. Use this when you need to find current information, documentation, or web pages. No API key required.",
	InputSchema: SearchInternetInputSchema,
	Function:    SearchInternet,
}

// SearchInternetInput is the JSON shape for the searchInternet tool.
type SearchInternetInput struct {
	Query       string `json:"query" jsonschema_description:"The search query."`
	NumResults  int    `json:"numResults" jsonschema_description:"Optional maximum number of results to return (default 10)."`
}

// SearchInternetInputSchema is the Anthropic tool input schema for searchInternet.
var SearchInternetInputSchema = GenerateSchema[SearchInternetInput]()

// SearchInternet implements the searchInternet tool using DuckDuckGo HTML search.
func SearchInternet(input json.RawMessage) (string, error) {
	var in SearchInternetInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("searchInternet input: %w", err)
	}
	query := strings.TrimSpace(in.Query)
	if query == "" {
		return "", fmt.Errorf("searchInternet: query is required")
	}
	numResults := in.NumResults
	if numResults <= 0 {
		numResults = defaultNumResults
	}

	searchURL := "https://html.duckduckgo.com/html/?q=" + url.QueryEscape(query)
	client := &http.Client{Timeout: searchInternetTimeout}
	req, err := http.NewRequest(http.MethodGet, searchURL, nil)
	if err != nil {
		return "", fmt.Errorf("searchInternet: request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; agentExample/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("searchInternet: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("searchInternet: HTTP status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("searchInternet: parse HTML: %w", err)
	}

	var lines []string
	count := 0
	doc.Find(".result").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if count >= numResults {
			return false
		}
		link := s.Find("a.result__a")
		if link.Length() == 0 {
			link = s.Find("a").First()
		}
		href, _ := link.Attr("href")
		title := strings.TrimSpace(link.Text())
		if href == "" && title == "" {
			return true
		}
		if href == "" {
			href = "(no URL)"
		}
		snippetSel := s.Find(".result__snippet")
		if snippetSel.Length() == 0 {
			snippetSel = s.Find(".result__body")
		}
		snippet := strings.TrimSpace(snippetSel.First().Text())
		count++
		if snippet != "" {
			if len(snippet) > 200 {
				snippet = snippet[:200] + "..."
			}
			lines = append(lines, fmt.Sprintf("%d. %s\n   %s\n   %s", count, title, href, snippet))
		} else {
			lines = append(lines, fmt.Sprintf("%d. %s\n   %s", count, title, href))
		}
		return true
	})

	if len(lines) == 0 {
		return "No results found for the query.", nil
	}
	return strings.Join(lines, "\n\n"), nil
}
