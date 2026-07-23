package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zalando/go-keyring"
)

const (
	deepSeekModel             = "deepseek-chat"
	deepSeekChatCompletionsURL = "https://api.deepseek.com/chat/completions"
	httpTimeout               = 30 * time.Second
	illustrationCooldown      = 30 * time.Second
	illustrationMinChars      = 100 // minimum paragraph length for illustration analysis
)

// AI is the client for the DeepSeek API.
type AI struct {
	apiKey               string
	unsplashAccessKey    string
	httpClient           *http.Client
	lastIllustrationTime time.Time // rate-limiting for illustration calls
}

// Analysis is the result of analyzing a paragraph of text.
type Analysis struct {
	WordErrors   []WordError `json:"wordErrors"`
	Theme        string      `json:"theme"`
	Font         string      `json:"font"`
	Illustration string      `json:"illustration"`
}

// WordError is a single word error.
type WordError struct {
	Incorrect string `json:"incorrect"`
	Correct   string `json:"correct"`
}

// chatMessage is a message in a DeepSeek chat completion request.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatCompletionRequest is the request body for a chat completion.
type chatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

// chatCompletionResponse is the response from a chat completion.
type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// NewAI creates a new AI client.
func NewAI() (*AI, error) {
	apiKey, err := keyring.Get(credentialService, deepSeekKeyAccount)
	if err != nil {
		return nil, err
	}
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, errors.New("DeepSeek API key not found")
	}

	// Unsplash key is optional — illustration will still work with text descriptions
	unsplashKey, _ := keyring.Get(credentialService, unsplashKeyAccount)
	unsplashKey = strings.TrimSpace(unsplashKey)

	return &AI{
		apiKey:            apiKey,
		unsplashAccessKey: unsplashKey,
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
	}, nil
}

// AnalyzeParagraph analyzes a paragraph of text and returns a summary.
func (a *AI) AnalyzeParagraph(ctx context.Context, text string) (*Analysis, error) {
	if a.apiKey == "" {
		return nil, errors.New("AI client not initialized")
	}

	// Fire off four calls concurrently for speed
	type result struct {
		wordErrors   []WordError
		wordErr      error
		theme        string
		themeErr     error
		font         string
		fontErr      error
		illustration string
		illustErr    error
	}

	ch := make(chan result, 1)
	go func() {
		var r result
		r.wordErrors, r.wordErr = a.getWordErrors(ctx, text)
		r.theme, r.themeErr = a.getTheme(ctx, text)
		r.font, r.fontErr = a.getFont(ctx, text)
		r.illustration, r.illustErr = a.getIllustrationKeywords(ctx, text)
		ch <- r
	}()

	select {
	case r := <-ch:
		return &Analysis{
			WordErrors:   r.wordErrors,
			Theme:        r.theme,
			Font:         r.font,
			Illustration: r.illustration,
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// CompleteWord returns suggestions to complete the partial word based on preceding text.
func (a *AI) CompleteWord(ctx context.Context, partialWord string, precedingText string) ([]string, error) {
	if a.apiKey == "" {
		return nil, errors.New("AI client not initialized")
	}

	prompt := fmt.Sprintf(wordAutocompletePrompt, partialWord)
	resp, err := a.createChatCompletion(ctx, prompt, precedingText)
	if err != nil {
		return nil, err
	}

	var suggestions struct {
		Words []string `json:"words"`
	}
	err = json.Unmarshal([]byte(resp), &suggestions)
	if err != nil {
		return nil, err
	}

	return suggestions.Words, nil
}

// CompleteParagraph returns a suggested continuation sentence for the preceding text.
func (a *AI) CompleteParagraph(ctx context.Context, precedingText string) (string, error) {
	if a.apiKey == "" {
		return "", errors.New("AI client not initialized")
	}

	resp, err := a.createChatCompletion(ctx, paragraphAutocompletePrompt, precedingText)
	if err != nil {
		return "", err
	}

	var suggestion struct {
		Continuation string `json:"continuation"`
	}
	err = json.Unmarshal([]byte(resp), &suggestion)
	if err != nil {
		return "", err
	}

	return suggestion.Continuation, nil
}

func (a *AI) getWordErrors(ctx context.Context, text string) ([]WordError, error) {
	resp, err := a.createChatCompletion(ctx, wordErrorPrompt, text)
	if err != nil {
		return nil, err
	}

	var wordErrors []WordError
	err = json.Unmarshal([]byte(resp), &wordErrors)
	if err != nil {
		return nil, err
	}

	return wordErrors, nil
}

func (a *AI) getTheme(ctx context.Context, text string) (string, error) {
	resp, err := a.createChatCompletion(ctx, themePrompt, text)
	if err != nil {
		return "", err
	}

	var theme struct {
		Theme string `json:"theme"`
	}
	err = json.Unmarshal([]byte(resp), &theme)
	if err != nil {
		return "", err
	}

	return theme.Theme, nil
}

func (a *AI) getFont(ctx context.Context, text string) (string, error) {
	resp, err := a.createChatCompletion(ctx, fontPrompt, text)
	if err != nil {
		return "", err
	}

	var font struct {
		Font string `json:"font"`
	}
	err = json.Unmarshal([]byte(resp), &font)
	if err != nil {
		return "", err
	}

	return font.Font, nil
}

// CanIllustrate checks whether the given paragraph qualifies for illustration analysis.
// It enforces minimum length and a rate-limiting cooldown window.
func (a *AI) CanIllustrate(text string) bool {
	if a.apiKey == "" {
		return false
	}
	if len(strings.TrimSpace(text)) < illustrationMinChars {
		return false
	}
	if time.Since(a.lastIllustrationTime) < illustrationCooldown {
		return false
	}
	return true
}

// GetIllustration returns an image URL for the given paragraph text.
// It first uses DeepSeek to extract search keywords, then searches Unsplash for a matching photo.
// Falls back to a text description if Unsplash is not configured or returns no results.
func (a *AI) GetIllustration(ctx context.Context, text string) (string, error) {
	if a.apiKey == "" {
		return "", errors.New("AI client not initialized")
	}

	a.lastIllustrationTime = time.Now()

	// Step 1: Ask DeepSeek for search keywords
	keywords, err := a.getIllustrationKeywords(ctx, text)
	if err != nil {
		return "", err
	}

	// Step 2: Search Unsplash for an actual image
	if a.unsplashAccessKey != "" {
		imageURL, err := a.searchUnsplash(ctx, keywords)
		if err == nil && imageURL != "" {
			return imageURL, nil
		}
		// Fall through to text description if Unsplash fails
	}

	// Fallback: return the keyword description as text
	return keywords, nil
}

func (a *AI) getIllustrationKeywords(ctx context.Context, text string) (string, error) {
	resp, err := a.createChatCompletion(ctx, illustrationPrompt, text)
	if err != nil {
		return "", err
	}

	var illustration struct {
		Illustration string `json:"illustration"`
	}
	err = json.Unmarshal([]byte(resp), &illustration)
	if err != nil {
		return "", err
	}

	return illustration.Illustration, nil
}

// unsplashSearchResponse is the response from the Unsplash search API.
type unsplashSearchResponse struct {
	Results []struct {
		URLs struct {
			Regular string `json:"regular"`
			Small   string `json:"small"`
		} `json:"urls"`
		Description string `json:"description"`
	} `json:"results"`
}

func (a *AI) searchUnsplash(ctx context.Context, query string) (string, error) {
	url := fmt.Sprintf("https://api.unsplash.com/search/photos?query=%s&per_page=1&orientation=landscape", strings.ReplaceAll(query, " ", "+"))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Client-ID "+a.unsplashAccessKey)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Unsplash API returned status %d", resp.StatusCode)
	}

	var result unsplashSearchResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	if len(result.Results) == 0 {
		return "", nil
	}

	// Prefer regular size; fall back to small
	imageURL := result.Results[0].URLs.Regular
	if imageURL == "" {
		imageURL = result.Results[0].URLs.Small
	}

	return imageURL, nil
}

func (a *AI) createChatCompletion(ctx context.Context, systemPromptContent, userContent string) (string, error) {
	reqBody := chatCompletionRequest{
		Model: deepSeekModel,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: systemPromptContent + "\n\n" + userContent},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, deepSeekChatCompletionsURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("DeepSeek API returned status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp chatCompletionResponse
	err = json.NewDecoder(resp.Body).Decode(&chatResp)
	if err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", errors.New("no choices returned from API")
	}

	return chatResp.Choices[0].Message.Content, nil
}