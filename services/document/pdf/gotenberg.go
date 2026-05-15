package pdf

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type GotenbergClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewGotenbergClient(baseURL string) *GotenbergClient {
	return &GotenbergClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// HTMLToPDF ส่ง HTML ไปให้ Gotenberg แปลงเป็น PDF คืน bytes
func (g *GotenbergClient) HTMLToPDF(ctx context.Context, htmlContent string) ([]byte, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// เพิ่ม index.html (ชื่อไฟล์ต้องเป็น index.html เสมอสำหรับ Gotenberg)
	part, err := writer.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.WriteString(part, htmlContent); err != nil {
		return nil, fmt.Errorf("write html: %w", err)
	}

	// ตั้งค่า paper size A4
	_ = writer.WriteField("paperWidth", "8.27")
	_ = writer.WriteField("paperHeight", "11.69")
	_ = writer.WriteField("marginTop", "0.5")
	_ = writer.WriteField("marginBottom", "0.5")
	_ = writer.WriteField("marginLeft", "0.5")
	_ = writer.WriteField("marginRight", "0.5")
	_ = writer.WriteField("printBackground", "true")

	writer.Close()

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		g.baseURL+"/forms/chromium/convert/html",
		&buf,
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call gotenberg: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gotenberg error %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}
