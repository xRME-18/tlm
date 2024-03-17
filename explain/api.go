package explain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	Stable   string = "stable"
	Balanced        = "balanced"
	Creative        = "creative"
)

func (e *Explain) getParametersFor(preference string) map[string]interface{} {
	switch preference {
	case Stable:
		return map[string]interface{}{
			"temperature": 0.1,
			"top_p":       0.25,
		}

	case Balanced:
		return map[string]interface{}{
			"temperature": 0.5,
			"top_p":       0.4,
		}

	case Creative:
		return map[string]interface{}{
			"temperature": 0.9,
			"top_p":       0.7,
		}

	default:
		return map[string]interface{}{}
	}
}

// func (e *Explain) StreamExplanationFor(mode, prompt string) error {
// 	onResponseFunc := func(res ollama.GenerateResponse) error {
// 		fmt.Print(res.Response)
// 		return nil
// 	}

// 	err := e.api.Generate(context.Background(), &ollama.GenerateRequest{
// 		Model:   "explain:7b",
// 		Prompt:  "Explain command: " + prompt,
// 		Options: e.getParametersFor(mode),
// 	}, onResponseFunc)

// 	if err != nil {
// 		fmt.Println("Error during generation:", err)
// 	}
// 	return nil
// }

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

func (e *Explain) StreamExplanationFor(mode, prompt string) error {
	reqBody := &Request{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a helpful assistant.",
			},
			{
				Role:    "user",
				Content: "Explain command: " + prompt,
			},
		},
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	apiKey := os.Getenv("TLM_OPENAI_API_KEY")
	req.Header.Set("Authorization", "Bearer "+apiKey) // replace with your OpenAI API key

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	for _, choice := range response.Choices {
		fmt.Println("Message:", choice.Message.Content)
	}

	return nil
}
