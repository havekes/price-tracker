package vision

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/havekes/price-tracker/internal/config"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	model      string
	apiKey     string
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		httpClient: &http.Client{},
		baseURL:    cfg.VisionAPIBaseURL,
		model:      "gpt-4o",
		apiKey:     cfg.VisionAPIKey,
	}
}

// ExtractReceipt parses an image of a receipt using a vision LLM.
func (c *Client) ExtractReceipt(ctx context.Context, imageBytes []byte, mimeType string) (*ExtractedReceipt, error) {
	// Construct the data URI
	encodedImage := base64.StdEncoding.EncodeToString(imageBytes)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, encodedImage)

	systemPrompt := "You are a receipt extraction engine. Your job is to extract line items from the provided receipt image.\n" +
		"You must return the result as a raw JSON object conforming EXACTLY to this schema:\n" +
		"{\n" +
		"  \"items\": [\n" +
		"    {\n" +
		"      \"raw_text\": \"string (the exact text on the receipt line)\",\n" +
		"      \"total_price\": float (the total price for this line item),\n" +
		"      \"raw_quantity\": float (the quantity, e.g. 1.0 or 1.5. Default to 1.0 if not specified),\n" +
		"      \"suggested_product_name\": \"string (a clean, generic name for the product)\",\n" +
		"      \"suggested_base_unit\": \"string (the unit, e.g. 'kg', 'l', 'pack', 'piece')\"\n" +
		"    }\n" +
		"  ]\n" +
		"}\n" +
		"Do not include markdown blocks like ```json. Return only the raw JSON."

	reqBody := ChatCompletionRequest{
		Model: c.model,
		Messages: []Message{
			{
				Role: "system",
				Content: []ContentPart{
					{
						Type: "text",
						Text: systemPrompt,
					},
				},
			},
			{
				Role: "user",
				Content: []ContentPart{
					{
						Type: "image_url",
						ImageURL: &ImageURL{
							URL: dataURI,
						},
					},
				},
			},
		},
		Temperature: 0.1,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vision API returned status %d", resp.StatusCode)
	}

	var chatResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("vision API returned no choices")
	}

	content := chatResp.Choices[0].Message.Content

	// Clean up potential markdown formatting
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
	}
	content = strings.TrimSpace(content)

	var receipt ExtractedReceipt
	if err := json.Unmarshal([]byte(content), &receipt); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON output from vision model: %w", err)
	}

	// Validate
	if len(receipt.Items) == 0 {
		return nil, fmt.Errorf("vision API returned empty items list")
	}
	for i, item := range receipt.Items {
		if item.TotalPrice < 0 {
			return nil, fmt.Errorf("invalid total price %f for item %d", item.TotalPrice, i)
		}
	}

	return &receipt, nil
}

// OpenAI API types

type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
}

type Message struct {
	Role    string        `json:"role"`
	Content []ContentPart `json:"content"`
}

type ContentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL string `json:"url"`
}

type ChatCompletionResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message ResponseMessage `json:"message"`
}

type ResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
