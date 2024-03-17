package suggest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"runtime"
	"strings"

	"os"

	"github.com/yusufcanb/tlm/shell"
)

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

const (
	Stable   string = "stable"
	Balanced        = "balanced"
	Creative        = "creative"
)

func (s *Suggest) getParametersFor(preference string) map[string]interface{} {
	switch preference {
	case Stable:
		return map[string]interface{}{
			"seed":        42,
			"temperature": 0.1,
			"top_p":       0.25,
		}

	case Balanced:
		return map[string]interface{}{
			"seed":        21,
			"temperature": 0.5,
			"top_p":       0.4,
		}

	case Creative:
		return map[string]interface{}{
			"seed":        0,
			"temperature": 0.9,
			"top_p":       0.7,
		}

	default:
		return map[string]interface{}{}
	}
}

func (s *Suggest) extractCommandsFromResponse(response string) []string {
	re := regexp.MustCompile("```([^\n]*)\n([^\n]*)\n```")

	matches := re.FindAllStringSubmatch(response, -1)

	if len(matches) == 0 {
		return nil
	}

	var codeSnippets []string
	for _, match := range matches {
		if len(match) == 3 {
			codeSnippets = append(codeSnippets, match[2])
		}
	}

	return codeSnippets
}

func (s *Suggest) getCommandSuggestionFor(mode, term string, prompt string) (string, error) {
	var responseText string

	builder := strings.Builder{}
	builder.WriteString(prompt)

	usingTerminalStr := ". I'm using %s terminal"
	onOperatingSystemStr := "on operating system: %s"

	switch term {
	case "zsh":
		builder.WriteString(fmt.Sprintf(usingTerminalStr, term))
		builder.WriteString(fmt.Sprintf(onOperatingSystemStr, "macOS"))
	case "bash":
		builder.WriteString(fmt.Sprintf(usingTerminalStr, term))
		builder.WriteString(fmt.Sprintf(onOperatingSystemStr, "Linux"))
	case "powershell":
		builder.WriteString(fmt.Sprintf(usingTerminalStr, term))
		builder.WriteString(fmt.Sprintf(onOperatingSystemStr, "Windows"))

	default:
		builder.WriteString(fmt.Sprintf(usingTerminalStr, shell.GetShell()))
		builder.WriteString(fmt.Sprintf(onOperatingSystemStr, runtime.GOOS))
	}

	// stream := false
	// req := &ollama.GenerateRequest{
	// 	Model:   "suggest:7b",
	// 	Prompt:  builder.String(),
	// 	Stream:  &stream,
	// 	Options: s.getParametersFor(mode),
	// }

	// onResponse := func(res ollama.GenerateResponse) error {
	// 	responseText = res.Response
	// 	return nil
	// }

	reqBody := &Request{
		Model: "gpt-4-0125-preview",
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a helpful assistant.",
			},
			{
				Role:    "user",
				Content: builder.String(),
			},
		},
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Set the OpenAI API key as an environment variable

	// Fetch the OpenAI API key from environment variables
	apiKey := os.Getenv("TLM_OPENAI_API_KEY")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}

	for _, choice := range response.Choices {
		responseText = choice.Message.Content
	}

	fmt.Println("using openAI")

	// err := s.api.Generate(context.Background(), req, onResponse)
	// if err != nil {
	// 	return "", err
	// }

	return responseText, nil
}
