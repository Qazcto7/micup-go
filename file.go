package micup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

// File represents a file attachment to be sent with a message.
type File struct {
	Name        string
	ContentType string
	Reader      io.Reader
}

// ChannelMessageSendFile sends a single file with optional text content.
func (c *RESTClient) ChannelMessageSendFile(ctx context.Context, channelID, filename string, reader io.Reader, content string) (*Message, error) {
	file := &File{
		Name:   filename,
		Reader: reader,
	}
	send := &MessageSend{
		Content: content,
	}
	return c.ChannelMessageSendFiles(ctx, channelID, []*File{file}, send)
}

// ChannelMessageSendFiles sends multiple file attachments with optional embeds and text.
func (c *RESTClient) ChannelMessageSendFiles(ctx context.Context, channelID string, files []*File, data *MessageSend) (*Message, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add payload_json part
	if data != nil {
		payloadPart, err := writer.CreateFormField("payload_json")
		if err != nil {
			return nil, fmt.Errorf("failed to create payload_json form field: %w", err)
		}
		if err := json.NewEncoder(payloadPart).Encode(data); err != nil {
			return nil, fmt.Errorf("failed to encode payload_json: %w", err)
		}
	}

	// Add files parts (files[0], files[1], etc.)
	for i, file := range files {
		fieldName := fmt.Sprintf("files[%d]", i)
		part, err := writer.CreateFormFile(fieldName, filepath.Base(file.Name))
		if err != nil {
			return nil, fmt.Errorf("failed to create form file for %s: %w", file.Name, err)
		}
		if _, err := io.Copy(part, file.Reader); err != nil {
			return nil, fmt.Errorf("failed to write file content for %s: %w", file.Name, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	baseURL := c.getBaseURL(ctx)
	reqURL := fmt.Sprintf("%s/channels/%s/messages", baseURL, channelID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", c.tokenPrefix+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("multipart request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr APIError
		respBytes, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(respBytes, &apiErr); err == nil && apiErr.Code != "" {
			apiErr.StatusCode = resp.StatusCode
			return nil, &apiErr
		}
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Code:       "HTTP_ERROR",
			Message:    string(respBytes),
		}
	}

	var msg Message
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &msg, nil
}

// Session convenience methods

// ChannelMessageSendFile sends a file to a channel.
func (s *Session) ChannelMessageSendFile(channelID, filename string, reader io.Reader, content string) (*Message, error) {
	return s.rest.ChannelMessageSendFile(context.Background(), channelID, filename, reader, content)
}

// ChannelMessageSendFiles sends multiple files with rich payload.
func (s *Session) ChannelMessageSendFiles(channelID string, files []*File, data *MessageSend) (*Message, error) {
	return s.rest.ChannelMessageSendFiles(context.Background(), channelID, files, data)
}
