// Package tools provides the fetchHtml tool for the agent.
package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const fetchHTMLMaxBytes = 5 * 1024 * 1024   // 5 MB read limit
const fetchHTMLMaxReturnChars = 50_000       // cap returned body to avoid blowing API context
const fetchHTMLTimeout = 30 * time.Second

// FetchHTMLDefinition is the tool that fetches the HTML or text body of a URL.
var FetchHTMLDefinition = ToolDefinition{
	Name:        "fetchHtml",
	Description: "Fetch the HTML or text body of a URL. Use this when you need to read the content of a web page. Returns the response body as text; for non-2xx status the body is still returned with a status line so you can reason about the response.",
	InputSchema: FetchHTMLInputSchema,
	Function:    FetchHTML,
}

// FetchHTMLInput is the JSON shape for the fetchHtml tool.
type FetchHTMLInput struct {
	URL string `json:"url" jsonschema_description:"The full URL to fetch (must be http or https)."`
}

// FetchHTMLInputSchema is the Anthropic tool input schema for fetchHtml.
var FetchHTMLInputSchema = GenerateSchema[FetchHTMLInput]()

// FetchHTML implements the fetchHtml tool: GETs the URL and returns the body as string.
func FetchHTML(input json.RawMessage) (string, error) {
	var in FetchHTMLInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("fetchHtml input: %w", err)
	}
	rawURL := strings.TrimSpace(in.URL)
	if rawURL == "" {
		return "", fmt.Errorf("fetchHtml: url is required")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("fetchHtml: invalid url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("fetchHtml: url must use http or https scheme")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("fetchHtml: url must have a host")
	}

	client := &http.Client{Timeout: fetchHTMLTimeout}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("fetchHtml: request: %w", err)
	}
	req.Header.Set("User-Agent", "agentExample/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetchHtml: %w", err)
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, fetchHTMLMaxBytes)
	body, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("fetchHtml: read body: %w", err)
	}
	bodyStr := string(body)
	if len(bodyStr) > fetchHTMLMaxReturnChars {
		bodyStr = bodyStr[:fetchHTMLMaxReturnChars] + "\n\n[Content truncated to " + fmt.Sprintf("%d", fetchHTMLMaxReturnChars) + " characters.]"
	}

	statusLine := fmt.Sprintf("HTTP status: %d %s", resp.StatusCode, resp.Status)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return bodyStr, nil
	}
	return statusLine + "\n\n" + bodyStr, nil
}
