package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
	"io/ioutil"
	"text/template"

	"autojobtracker/models"

	openai "github.com/sashabaranov/go-openai"
)

var openaiClient *openai.Client

func Init(apiKey string) {
    if apiKey == "" {
        log.Fatal("OPENAI_API_KEY is missing")
    }
    openaiClient = openai.NewClient(apiKey)
}

type LLMResponse struct {
	Company  string `json:"company"`
	Position string `json:"position"`
	Stage    string `json:"stage"` // Applied, Interview, Rejected
	Referral bool   `json:"referral"`
	JobURL   string `json:"job_url,omitempty"`
}

func ParseEmail(subject, body, email string, emailDate time.Time) *models.Job {
	// First try LLM
	llmResult, err := parseWithLLM(subject, body, email)
	if err != nil {
		log.Printf("LLM parse failed: %v", err)
		llmResult = fallbackParse(subject, body)
	}

	job := models.Job{
		Company:   llmResult.Company,
		Position:  llmResult.Position,
		Stage:     llmResult.Stage,
		Referral:  llmResult.Referral,
		JobURL:    llmResult.JobURL,
		ApplyDate: emailDate,
	}

	// Set response date if stage is not "Applied"
	if job.Stage != "Applied" {
		t := time.Now()
		job.ResponseDate = &t
	}

	return &job
}

func loadPromptTemplate(subject, body, email string) (string, error) {
	raw, err := ioutil.ReadFile("prompt.txt")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("prompt").Option("missingkey=error").Parse(string(raw))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]string{
		"SUBJECT": subject,
		"BODY":    body,
		"EMAIL":	email,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func parseWithLLM(subject, body, email string) (*LLMResponse, error) {
	ctx := context.Background()

	prompt, err := loadPromptTemplate(subject, body, email)
	if err != nil {
		return nil, fmt.Errorf("prompt load failed: %w",err)
	}
	req := openai.ChatCompletionRequest{
		Model: openai.GPT4o, // use GPT-4o or GPT-3.5
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You're a helpful assistant that extracts job application data from emails. Always respond with only JSON, no explanation.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0,
	}

	resp, err := openaiClient.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	var llm LLMResponse

	raw := strings.TrimSpace(resp.Choices[0].Message.Content)

	// Handle markdown-style code block (```json ... ```)
	re := regexp.MustCompile("(?s)```(?:json)?\\s*(\\{.*?\\})\\s*```")
	match := re.FindStringSubmatch(raw)

	var jsonPart string
	if len(match) >= 2 {
		jsonPart = match[1] // extract content inside code block
	} else {
		// fallback: try to find first valid JSON object
		re := regexp.MustCompile(`(?s)\{.*\}`)
		jsonPart = re.FindString(raw)
	}

	if jsonPart == "" {
		return nil, fmt.Errorf("no valid JSON block found in LLM output")
	}

	if err := json.Unmarshal([]byte(jsonPart), &llm); err != nil {
		log.Printf("JSON unmarshal failed: %v", err)
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return &llm, nil
}

func fallbackParse(subject, body string) *LLMResponse {
	fullText := strings.ToLower(subject + " " + body)
	llm := LLMResponse{
		Stage:    "Applied",
		Referral: strings.Contains(fullText, "referred") || strings.Contains(fullText, "referral"),
	}

	// Job URL
	urlRegex := regexp.MustCompile(`https?://[^\s]+`)
	urls := urlRegex.FindAllString(body, -1)
	if len(urls) > 0 {
		llm.JobURL = urls[0]
	}

	// Stage
	switch {
	case strings.Contains(fullText, "rejected"),
		strings.Contains(fullText, "not selected"),
		strings.Contains(fullText, "unfortunately"),
		strings.Contains(fullText, "declined"):
		llm.Stage = "Rejected"
	case strings.Contains(fullText, "interview"),
		strings.Contains(fullText, "call scheduled"),
		strings.Contains(fullText, "recruiter will reach out"):
		llm.Stage = "Interview"
	}

	// Basic position + company
	re := regexp.MustCompile(`(?i)application.*?for (.+?) position at (.+)`)
	match := re.FindStringSubmatch(subject)
	if len(match) == 3 {
		llm.Position = strings.TrimSpace(match[1])
		llm.Company = strings.TrimSpace(match[2])
	}

	if llm.Company == "" {
		re := regexp.MustCompile(`(?i)application (?:at|to) ([\w\s\-]+)`)
		match := re.FindStringSubmatch(subject)
		if len(match) == 2 {
			llm.Company = strings.TrimSpace(match[1])
		}
	}

	if llm.Company == "" && llm.JobURL != "" {
		re := regexp.MustCompile(`https?://(?:www\.)?([a-zA-Z0-9\-]+)\.`)
		match := re.FindStringSubmatch(llm.JobURL)
		if len(match) >= 2 {
			llm.Company = strings.TrimSpace(match[1])
		}
	}

	return &llm
}