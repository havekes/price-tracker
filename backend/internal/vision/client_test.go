package vision

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/havekes/price-tracker/internal/config"
)

func TestExtractReceipt(t *testing.T) {
	tests := []struct {
		name          string
		responseCode  int
		responseBody  string
		expectError   bool
		expectedItems int
	}{
		{
			name:         "Success",
			responseCode: http.StatusOK,
			responseBody: `{
				"choices": [
					{
						"message": {
							"role": "assistant",
							"content": "{\"items\":[{\"raw_text\":\"Milk\",\"total_price\":3.99,\"raw_quantity\":1,\"suggested_product_name\":\"Milk\",\"suggested_base_unit\":\"l\"}]}"
						}
					}
				]
			}`,
			expectError:   false,
			expectedItems: 1,
		},
		{
			name:         "Malformed JSON in content",
			responseCode: http.StatusOK,
			responseBody: `{
				"choices": [
					{
						"message": {
							"role": "assistant",
							"content": "{invalid json}"
						}
					}
				]
			}`,
			expectError: true,
		},
		{
			name:         "Unauthorized API Error",
			responseCode: http.StatusUnauthorized,
			responseBody: `{"error": {"message": "Unauthorized"}}`,
			expectError:  true,
		},
		{
			name:         "Empty items list",
			responseCode: http.StatusOK,
			responseBody: `{
				"choices": [
					{
						"message": {
							"role": "assistant",
							"content": "{\"items\":[]}"
						}
					}
				]
			}`,
			expectError: true,
		},
		{
			name:         "Negative price",
			responseCode: http.StatusOK,
			responseBody: `{
				"choices": [
					{
						"message": {
							"role": "assistant",
							"content": "{\"items\":[{\"raw_text\":\"Discount\",\"total_price\":-3.99,\"raw_quantity\":1,\"suggested_product_name\":\"Discount\",\"suggested_base_unit\":\"\"}]}"
						}
					}
				]
			}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer test-key" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			cfg := config.Config{
				VisionAPIBaseURL: server.URL,
				VisionAPIKey:     "test-key",
			}
			client := NewClient(cfg)

			receipt, err := client.ExtractReceipt(context.Background(), []byte("fake-image-data"), "image/jpeg")

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if len(receipt.Items) != tt.expectedItems {
					t.Errorf("expected %d items, got %d", tt.expectedItems, len(receipt.Items))
				}
			}
		})
	}
}

func TestExtractReceipt_MarkdownFormatting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"choices": [
				{
					"message": {
						"role": "assistant",
						"content": "` + "```" + `json\n{\"items\":[{\"raw_text\":\"Milk\",\"total_price\":3.99,\"raw_quantity\":1,\"suggested_product_name\":\"Milk\",\"suggested_base_unit\":\"l\"}]}\n` + "```" + `"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	cfg := config.Config{
		VisionAPIBaseURL: server.URL,
		VisionAPIKey:     "test-key",
	}
	client := NewClient(cfg)

	receipt, err := client.ExtractReceipt(context.Background(), []byte("fake-image-data"), "image/jpeg")
	if err != nil {
		t.Fatalf("expected no error handling markdown, got %v", err)
	}
	if len(receipt.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(receipt.Items))
	}
}
