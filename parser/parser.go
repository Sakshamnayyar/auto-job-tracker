package parser

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"text/template"
	"time"

	"autojobtracker/llm"
	"autojobtracker/models"
)

func InitLLM() {
	llm.InitLLM()
}

func ParseEmail(subject, body, email string, emailDate time.Time) *models.Job {
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

	if job.Stage != "Applied" {
		t := time.Now()
		job.ResponseDate = &t
	}

	return &job
}

func parseWithLLM(subject, body, email string) (*llm.LLMResponse, error) {
	prompt, err := loadPromptTemplate(subject, body, email)
	if err != nil {
		return nil, fmt.Errorf("prompt load failed: %w", err)
	}
	return llm.ParseWithLLM(prompt)
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
		"EMAIL":   email,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func fallbackParse(subject, body string) *llm.LLMResponse {
	fullText := strings.ToLower(subject + " " + body)
	llm := llm.LLMResponse{
		Stage:    "Applied",
		Referral: strings.Contains(fullText, "referred") || strings.Contains(fullText, "referral"),
	}

	urlRegex := regexp.MustCompile(`https?://[^\s]+`)
	urls := urlRegex.FindAllString(body, -1)
	if len(urls) > 0 {
		llm.JobURL = urls[0]
	}

	switch {
	case strings.Contains(fullText, "rejected"),
		strings.Contains(fullText, "not selected"),
		strings.Contains(fullText, "unfortunately"),
		strings.Contains(fullText, "declined"):
		llm.Stage = "Rejected"
	case strings.Contains(fullText, "interview"),
		strings.Contains(fullText, "phone screen"),
		strings.Contains(fullText, "zoom"):
		llm.Stage = "Interview"
	}

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
