package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var apiUrl string
var apiKey string
var aiModel string

func init() {
	bts, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	var config map[string]any
	err = json.Unmarshal(bts, &config)
	if err != nil {
		panic(err)
	}
	apiUrl = config["openai"].(map[string]any)["apiUrl"].(string)
	apiKey = config["openai"].(map[string]any)["key"].(string)
	aiModel = config["openai"].(map[string]any)["model"].(string)
}

func Chat(content, key string, model string) (any, error) {
	if model == "" {
		model = aiModel
	}
	if key == "" {
		key = apiKey
	}
	var body = map[string]any{
		"model": model,
		"messages": []map[string]any{
			{
				"role":    "user",
				"content": content,
			},
		},
		"stream":     false,
		"max_tokens": 4096,
	}
	bts, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", apiUrl), bytes.NewReader(bts))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+key)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bts, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, err

	}
	var m map[string]any
	err = json.Unmarshal(bts, &m)
	if err != nil {
		return nil, err
	}

	code := m["code"]
	if code != nil {
		message := m["message"].(string)
		log.Println(message)
		return nil, fmt.Errorf("接口错误:%f %s", code.(float64), message)
	}
	message := m["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)
	respContent := message["content"]
	var ret = map[string]any{}
	reasoning_content := message["reasoning_content"]
	if reasoning_content != nil {
		ret["reasoning_content"] = reasoning_content
	}
	if respContent != nil {
		ret["content"] = respContent
	}

	log.Println(respContent)
	return ret, nil

}
