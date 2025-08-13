package summarizer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type requestBody struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type responseBody struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func Summarize(apiKey, text string) (string, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=" + apiKey
	prompt := fmt.Sprintf("Matnni qisqa ilmiy tilda xulosa qil va kalit so‘zlarni chiqar.\nMatn:\n%s", text)

	reqData := requestBody{
		Contents: []Content{
			{Parts: []Part{{Text: prompt}}},
		},
	}

	b, _ := json.Marshal(reqData)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res responseBody
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	if len(res.Candidates) > 0 && len(res.Candidates[0].Content.Parts) > 0 {
		return res.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("no response from Gemini")
}

func IsOverLimit(tariff, text string) bool {
	switch tariff {
	case "Free":
		return len(text) > 1000
	case "Basic":
		return len(text) > 2000
	case "Pro":
		return false
	}
	return false
}

func LimitText(tariff, text string) string {
	limit := 1000
	if tariff == "Basic" {
		limit = 2000
	}
	if len(text) > limit {
		return text[:limit]
	}
	return text
}
