package line

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://api.line.me/v2/bot/message"

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type textMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type pushPayload struct {
	To       string        `json:"to"`
	Messages []textMessage `json:"messages"`
}

// PushText ส่งข้อความไปหา user หรือ group โดยใช้ token ของบริษัทนั้น
func (c *Client) PushText(ctx context.Context, channelToken, to, text string) error {
	payload := pushPayload{
		To:       to,
		Messages: []textMessage{{Type: "text", Text: text}},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/push", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+channelToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("line api error %d: %s", resp.StatusCode, string(b))
	}

	return nil
}
