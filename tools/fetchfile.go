// Package tools provides the fetchFile tool for the agent.
package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const fetchFileMaxBytes = 5 * 1024 * 1024   // 5 MB when reading into memory
const fetchFileMaxReturnChars = 50_000      // cap returned text to avoid blowing API context
const fetchFileTimeout = 60 * time.Second

// FetchFileDefinition is the tool that downloads a file from a URL; optionally saves to a path, otherwise returns content or a short description for large/binary files.
var FetchFileDefinition = ToolDefinition{
	Name:        "fetchFile",
	Description: "Download a file from a URL. If savePath is provided, saves the response to that path (relative to working directory) and returns a summary. Otherwise returns the body as text for text-like Content-Types, or a message for binary/large responses; use savePath to download binary or large files to disk.",
	InputSchema: FetchFileInputSchema,
	Function:    FetchFile,
}

// FetchFileInput is the JSON shape for the fetchFile tool.
type FetchFileInput struct {
	URL      string `json:"url" jsonschema_description:"The URL of the file to fetch (must be http or https)."`
	SavePath string `json:"savePath" jsonschema_description:"Optional path to save the file to, relative to the working directory."`
}

// FetchFileInputSchema is the Anthropic tool input schema for fetchFile.
var FetchFileInputSchema = GenerateSchema[FetchFileInput]()

// FetchFile implements the fetchFile tool.
func FetchFile(input json.RawMessage) (string, error) {
	var in FetchFileInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("fetchFile input: %w", err)
	}
	rawURL := strings.TrimSpace(in.URL)
	if rawURL == "" {
		return "", fmt.Errorf("fetchFile: url is required")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("fetchFile: invalid url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("fetchFile: url must use http or https scheme")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("fetchFile: url must have a host")
	}

	client := &http.Client{Timeout: fetchFileTimeout}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("fetchFile: request: %w", err)
	}
	req.Header.Set("User-Agent", "agentExample/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetchFile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("fetchFile: HTTP status %d %s", resp.StatusCode, resp.Status)
	}

	savePath := strings.TrimSpace(in.SavePath)
	if savePath != "" {
		savePath = filepath.Clean(savePath)
		dir := filepath.Dir(savePath)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return "", fmt.Errorf("fetchFile: mkdir: %w", err)
			}
		}
		f, err := os.Create(savePath)
		if err != nil {
			return "", fmt.Errorf("fetchFile: create file: %w", err)
		}
		n, err := io.Copy(f, resp.Body)
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		if err != nil {
			os.Remove(savePath)
			return "", fmt.Errorf("fetchFile: write: %w", err)
		}
		return fmt.Sprintf("Saved to %s, %d bytes", savePath, n), nil
	}

	contentType := resp.Header.Get("Content-Type")
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	textLike := strings.HasPrefix(contentType, "text/") ||
		contentType == "application/json" ||
		contentType == "application/xml" ||
		contentType == "application/javascript"

	if textLike {
		limited := io.LimitReader(resp.Body, fetchFileMaxBytes)
		body, err := io.ReadAll(limited)
		if err != nil {
			return "", fmt.Errorf("fetchFile: read body: %w", err)
		}
		s := string(body)
		if len(s) > fetchFileMaxReturnChars {
			s = s[:fetchFileMaxReturnChars] + "\n\n[Content truncated to " + fmt.Sprintf("%d", fetchFileMaxReturnChars) + " characters.]"
		}
		return s, nil
	}

	limited := io.LimitReader(resp.Body, fetchFileMaxBytes)
	n, err := io.Copy(io.Discard, limited)
	if err != nil {
		return "", fmt.Errorf("fetchFile: read: %w", err)
	}
	return fmt.Sprintf("Binary response, %d bytes; use savePath to download to disk", n), nil
}
