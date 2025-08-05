package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"google.golang.org/api/option"
	gemini "github.com/google/generative-ai-go/genai"
)

type LLMResponse struct {
	Company  string `json:"company"`
	Position string `json:"position"`
	Stage    string `json:"stage"`
	Referral bool   `json:"referral"`
	JobURL   string `json:"job_url,omitempty"`
}

type LLMClient interface {
	Parse(prompt string) (*LLMResponse, error)
}

var client LLMClient

func InitLLM() {
	useGemini := os.Getenv("USE_GEMINI")
	if useGemini == "true" {
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			log.Fatal("GEMINI_API_KEY missing")
		}
		client = NewGeminiClient(apiKey)
	} else {
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			log.Fatal("OPENAI_API_KEY missing")
		}
		client = NewOpenAIClient(apiKey)
	}
}

func ParseWithLLM(prompt string) (*LLMResponse, error) {
	return client.Parse(prompt)
}

// ---------------- OPENAI ----------------

type OpenAIClient struct {
	client *openai.Client
}

func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		client: openai.NewClient(apiKey),
	}
}

func (c *OpenAIClient) Parse(prompt string) (*LLMResponse, error) {
	resp, err := c.client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You're a helpful assistant that extracts job application data from emails. Always respond with only JSON, no explanation.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0,
	})
	if err != nil {
		return nil, err
	}

	return parseJSONFromOutput(resp.Choices[0].Message.Content)
}

// ---------------- GEMINI ----------------

type GeminiClient struct {
	client *gemini.Client
	model  *gemini.GenerativeModel
}

func NewGeminiClient(apiKey string) *GeminiClient {
	ctx := context.Background()
	client, err := gemini.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}
	return &GeminiClient{
		client: client,
		model:  client.GenerativeModel("gemini-2.0-flash-lite"),
	}
}

func (c *GeminiClient) Parse(prompt string) (*LLMResponse, error) {
	resp, err := c.model.GenerateContent(context.Background(), gemini.Text(prompt))
	if err != nil {
		return nil, err
	}
	if len(resp.Candidates) == 0 {
		return nil, errors.New("Gemini returned no candidates")
	}

	var sb strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(gemini.Text); ok {
			sb.WriteString(string(text))
		}
	}

	if sb.Len() == 0 {
		return nil, errors.New("Gemini response contained no text parts")
	}

	return parseJSONFromOutput(sb.String())
}


// ---------------- SHARED JSON CLEANER ----------------

func parseJSONFromOutput(output string) (*LLMResponse, error) {
	var result LLMResponse

	raw := strings.TrimSpace(output)
	re := regexp.MustCompile("(?s)```(?:json)?\\s*(\\{.*?\\})\\s*```")
	match := re.FindStringSubmatch(raw)

	var jsonPart string
	if len(match) >= 2 {
		jsonPart = match[1]
	} else {
		re := regexp.MustCompile(`(?s)\{.*\}`)
		jsonPart = re.FindString(raw)
	}

	if jsonPart == "" {
		return nil, fmt.Errorf("no valid JSON block found in LLM output")
	}

	if err := json.Unmarshal([]byte(jsonPart), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}
