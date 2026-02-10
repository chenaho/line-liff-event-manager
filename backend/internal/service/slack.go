package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const SettingsKeySlackWebhookURL = "slack_webhook_url"

// SlackService handles Slack webhook notifications
type SlackService struct {
	Settings *SettingsService
}

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
func NewSlackService(settings *SettingsService) *SlackService {
	return &SlackService{Settings: settings}
}

// getWebhookURL reads the webhook URL from database settings
func (s *SlackService) getWebhookURL() string {
	if s.Settings == nil {
		return ""
	}
	url, err := s.Settings.Get(context.Background(), SettingsKeySlackWebhookURL)
	if err != nil {
		log.Printf("[Slack] Failed to get webhook URL from settings: %v", err)
		return ""
	}
	return url
}

// SendLineUpNotification sends a notification for lineup status changes
func (s *SlackService) SendLineUpNotification(eventTitle, userName, action, status string, note string) error {
	webhookURL := s.getWebhookURL()
	if webhookURL == "" {
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

	return s.sendMessage(webhookURL, msg)
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

func (s *SlackService) sendMessage(webhookURL string, msg SlackMessage) error {
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
		log.Printf("[Slack] Webhook returned status: %d, body: %s", resp.StatusCode, string(body))
		return fmt.Errorf("slack webhook returned status: %d", resp.StatusCode)
	}

	log.Println("[Slack] Notification sent successfully")
	return nil
}
