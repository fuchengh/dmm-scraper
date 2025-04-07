package translate

import (
	"bytes"
	"dmm-scraper/pkg/config"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	Enable       bool
	apiUrl       string
	apiKey       string
	model        string
	temp         float64
	freqPenalty  float64
	topP         float64
	maxTokens    int64
	SystemPrompt string
}

func New() *Client {
	return &Client{
		Enable:       false,
		apiUrl:       "",
		apiKey:       "",
		model:        "",
		temp:         0.0,
		topP:         0.0,
		maxTokens:    0,
		freqPenalty:  0.0,
		SystemPrompt: "",
	}
}

func (d *Client) InitTranslateApi(conf *config.Translate) error {
	if conf.ApiUrl == "" || conf.ApiKey == "" {
		return nil
	}
	d.Enable = conf.Enable
	d.apiUrl = conf.ApiUrl
	d.apiKey = conf.ApiKey
	d.model = conf.Model
	d.temp = conf.Temparature
	d.topP = conf.TopP
	d.maxTokens = conf.MaxTokens
	d.freqPenalty = conf.FreqPenalty
	d.SystemPrompt = conf.SystemPrompt
	return nil
}

func (d *Client) Translate(text string) (string, error) {
	if d.apiUrl == "" || d.apiKey == "" {
		return text, nil
	}

	// send request to translation API
	payload := map[string]interface{}{
		"model":             d.model,
		"messages":          []map[string]string{{"role": "user", "content": d.SystemPrompt + text}},
		"temperature":       d.temp,
		"frequency_penalty": d.freqPenalty,
		"top_p":             d.topP,
		"max_tokens":        d.maxTokens,
		"reasoning": map[string]interface{}{
			"exclude": true,
		},
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v", err)
		return text, err
	}
	req, err := http.NewRequest("POST", d.apiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v", err)
		return text, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return text, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error sending request: %v, status code: %d\n", err, resp.StatusCode)
		return text, nil
	}

	// decode response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		return text, err
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		fmt.Printf("Error decoding response: ok is %v, len(choices) is %d\n", ok, len(choices))
		return text, nil
	}

	finish_reason := choices[0].(map[string]interface{})["finish_reason"]
	if finish_reason != "stop" {
		fmt.Printf("Error decoding response: finish_reason is %v\n", finish_reason)
		return text, nil
	}

	translatedText, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	if !ok || translatedText == "" {
		fmt.Printf("Error decoding response: ok is %vtranslatedText is %v\n", ok, translatedText)
		return text, nil
	}

	return translatedText, nil
}
