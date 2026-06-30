package categorize

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type LLMConfig struct {
	BaseURL string
	APIKey  string
	Model   string
}

type LLM struct {
	cfg LLMConfig
	hc  *http.Client
}

func NewLLM(cfg LLMConfig) *LLM {
	return &LLM{cfg: cfg, hc: &http.Client{Timeout: 60 * time.Second}}
}

type chatReq struct {
	Model    string    `json:"model"`
	Messages []chatMsg `json:"messages"`
	Stream   bool      `json:"stream"`
}
type chatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type chatResp struct {
	Choices []struct {
		Message chatMsg `json:"message"`
	} `json:"choices"`
}

type classifyResponse struct {
	Category string `json:"category"`
	Reason   string `json:"reason"`
}

// Classify asks the LLM to pick exactly one category for a transaction
// partner and explain why, then snaps the answer to a known category name.
func (l *LLM) Classify(ctx context.Context, partner string, categories []string) (string, string, error) {
	prompt := fmt.Sprintf(
		"You categorise bank transactions. Choose exactly ONE category from this list "+
			"that best matches the merchant/partner, and give a one-sentence reason. "+
			"Reply with ONLY a JSON object of the form {\"category\":\"<name>\",\"reason\":\"<reason>\"}, nothing else.\n"+
			"Categories: %s\nPartner: %s",
		strings.Join(categories, ", "), partner)

	body, _ := json.Marshal(chatReq{
		Model:    l.cfg.Model,
		Stream:   false,
		Messages: []chatMsg{{Role: "user", Content: prompt}},
	})

	url := strings.TrimRight(l.cfg.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if l.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+l.cfg.APIKey)
	}

	resp, err := l.hc.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("llm http %d", resp.StatusCode)
	}

	var cr chatResp
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return "", "", err
	}
	if len(cr.Choices) == 0 {
		return "Uncategorized", "", nil
	}

	answer, reason := parseClassifyContent(cr.Choices[0].Message.Content)
	for _, c := range categories {
		if strings.EqualFold(answer, c) {
			return c, reason, nil
		}
	}
	return "Uncategorized", reason, nil
}

// parseClassifyContent extracts the category and reason from the LLM's reply.
// It expects {"category":"...","reason":"..."}, optionally wrapped in a
// ```json ... ``` fence, but falls back to treating the raw trimmed content
// as a bare category name (with no reason) if the model didn't follow the
// JSON instruction.
func parseClassifyContent(content string) (category string, reason string) {
	trimmed := strings.TrimSpace(content)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)

	var cr classifyResponse
	if err := json.Unmarshal([]byte(trimmed), &cr); err == nil && cr.Category != "" {
		return cr.Category, cr.Reason
	}
	return strings.Trim(strings.TrimSpace(content), `"'.`), ""
}

type PingResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type modelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// Ping checks whether the configured LLM server is reachable and whether the
// configured model is present in its model list.
func (l *LLM) Ping(ctx context.Context) PingResult {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	url := strings.TrimRight(l.cfg.BaseURL, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return PingResult{Status: "unreachable", Message: err.Error()}
	}
	if l.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+l.cfg.APIKey)
	}

	resp, err := l.hc.Do(req)
	if err != nil {
		return PingResult{Status: "unreachable", Message: err.Error()}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return PingResult{Status: "unreachable", Message: fmt.Sprintf("http %d", resp.StatusCode)}
	}

	var mr modelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&mr); err != nil {
		return PingResult{Status: "unreachable", Message: err.Error()}
	}

	for _, m := range mr.Data {
		if m.ID == l.cfg.Model {
			return PingResult{Status: "ok", Message: fmt.Sprintf("%d models available", len(mr.Data))}
		}
	}
	return PingResult{
		Status:  "model_not_found",
		Message: fmt.Sprintf("model not in server's list (%d available)", len(mr.Data)),
	}
}
