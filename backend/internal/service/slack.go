package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// SlackService handles Slack webhook notifications
type SlackService struct{}

// SlackMessage represents a Slack webhook message
type SlackMessage struct {
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment represents a Slack message attachment
type Attachment struct {
	Color      string `json:"color,omitempty"`
	Title      string `json:"title,omitempty"`
	Text       string `json:"text,omitempty"`
	Footer     string `json:"footer,omitempty"`
	FooterIcon string `json:"footer_icon,omitempty"`
}

// NewSlackService creates a new SlackService
func NewSlackService() *SlackService {
	url := os.Getenv("SLACK_WEBHOOK_URL")
	if url == "" {
		log.Println("[Slack] WARNING: SLACK_WEBHOOK_URL not set, notifications will be disabled")
	} else {
		log.Println("[Slack] Service initialized with webhook URL configured")
	}
	return &SlackService{}
}

// getWebhookURL reads the webhook URL from env each time so updates take effect without restart
func (s *SlackService) getWebhookURL() string {
	return os.Getenv("SLACK_WEBHOOK_URL")
}

// SendLineUpNotification sends a notification for lineup status changes
func (s *SlackService) SendLineUpNotification(eventTitle, userName, action, status string, note string) error {
	if s.getWebhookURL() == "" {
		log.Println("[Slack] No webhook URL configured, skipping notification")
		return nil
	}

	var color, emoji string
	switch action {
	case "register":
		if status == "SUCCESS" {
			color = "#36a64f" // green
			emoji = "✅"
		} else if status == "WAITLIST" {
			color = "#ff9800" // orange
			emoji = "⏳"
		}
	case "cancel":
		color = "#e91e63" // pink
		emoji = "❌"
	}

	statusText := status
	if status == "SUCCESS" {
		statusText = "正取"
	} else if status == "WAITLIST" {
		statusText = "候補"
	} else if action == "cancel" {
		statusText = "已取消"
	}

	text := fmt.Sprintf("%s *%s* %s報名", emoji, userName, getActionVerb(action))
	if note != "" {
		text += fmt.Sprintf(" (備註: %s)", note)
	}

	msg := SlackMessage{
		Attachments: []Attachment{
			{
				Color:  color,
				Title:  eventTitle,
				Text:   text,
				Footer: fmt.Sprintf("狀態: %s", statusText),
			},
		},
	}

	return s.sendMessage(msg)
}

func getActionVerb(action string) string {
	switch action {
	case "register":
		return "已"
	case "cancel":
		return "取消"
	default:
		return ""
	}
}

func (s *SlackService) sendMessage(msg SlackMessage) error {
	webhookURL := s.getWebhookURL()
	if webhookURL == "" {
		log.Println("[Slack] No webhook URL configured, skipping")
		return nil
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[Slack] Failed to marshal message: %v", err)
		return err
	}

	log.Printf("[Slack] Sending notification to webhook...")
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[Slack] Failed to send message: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[Slack] Webhook returned status: %d, body: %s, url_prefix: %s", resp.StatusCode, string(body), webhookURL[:50])
		return fmt.Errorf("slack webhook returned status: %d", resp.StatusCode)
	}

	log.Println("[Slack] Notification sent successfully")
	return nil
}
