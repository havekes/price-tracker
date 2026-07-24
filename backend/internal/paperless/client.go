package paperless

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

var (
	ErrUnauthorized = errors.New("paperless: unauthorized")
	ErrNotFound     = errors.New("paperless: not found")
)

// Client is a REST API client for Paperless-ngx.
type Client struct {
	baseURL    string
	authToken  string
	httpClient *http.Client
}

// NewClient creates a new Paperless-ngx client.
func NewClient(baseURL, token string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:    baseURL,
		authToken:  token,
		httpClient: httpClient,
	}
}

// Document represents a Paperless document metadata.
type Document struct {
	ID              int    `json:"id"`
	Title           string `json:"title"`
	Created         string `json:"created"`
	CorrespondentID *int   `json:"correspondent"`
	Tags            []int  `json:"tags"`
}

// Correspondent represents a Paperless correspondent.
type Correspondent struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// DocumentListResponse represents a paginated list of documents.
type DocumentListResponse struct {
	Count    int        `json:"count"`
	Next     *string    `json:"next"`
	Previous *string    `json:"previous"`
	Results  []Document `json:"results"`
}

// GetDocumentsParams defines filter parameters for listing documents.
type GetDocumentsParams struct {
	Tags            []int // Example: filtering by multiple tag IDs
	CorrespondentID *int  // Example: filtering by a correspondent
	Query           string
}

func (c *Client) doRequest(ctx context.Context, method, path string, query url.Values) (*http.Response, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	u.Path = url.PathEscape(u.Path) // Handle any existing path in base URL (if not already escaped)
	u = u.JoinPath(path)
	if query != nil {
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+c.authToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		switch resp.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			return nil, ErrUnauthorized
		case http.StatusNotFound:
			return nil, ErrNotFound
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("paperless API error: status %d: %s", resp.StatusCode, string(body))
		}
	}

	return resp, nil
}

// GetDocuments returns a list of documents matching the specified parameters.
func (c *Client) GetDocuments(ctx context.Context, params GetDocumentsParams) ([]Document, error) {
	q := url.Values{}
	if params.Query != "" {
		q.Set("query", params.Query)
	}
	if params.CorrespondentID != nil {
		q.Set("correspondent__id", strconv.Itoa(*params.CorrespondentID))
	}
	if len(params.Tags) > 0 {
		var tagsStr string
		for i, tag := range params.Tags {
			if i > 0 {
				tagsStr += ","
			}
			tagsStr += strconv.Itoa(tag)
		}
		q.Set("tags__id__all", tagsStr)
	}

	resp, err := c.doRequest(ctx, http.MethodGet, "/api/documents/", q)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var listResponse DocumentListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return listResponse.Results, nil
}

// GetDocument returns a single document's metadata by its ID.
func (c *Client) GetDocument(ctx context.Context, docID int) (*Document, error) {
	path := fmt.Sprintf("/api/documents/%d/", docID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var doc Document
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &doc, nil
}

// GetCorrespondent returns a single correspondent by its ID.
func (c *Client) GetCorrespondent(ctx context.Context, correspondentID int) (*Correspondent, error) {
	path := fmt.Sprintf("/api/correspondents/%d/", correspondentID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var corr Correspondent
	if err := json.NewDecoder(resp.Body).Decode(&corr); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &corr, nil
}

// DownloadDocument fetches the raw document payload (PDF/Image bytes) and its content type.
func (c *Client) DownloadDocument(ctx context.Context, docID int) ([]byte, string, error) {
	path := fmt.Sprintf("/api/documents/%d/download/", docID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("reading response body: %w", err)
	}

	return bodyBytes, contentType, nil
}
