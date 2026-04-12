package alerting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type TelegramAlerter struct {
	token  string
	chatID string
}

func NewTelegramAlerter(token, chatID string) *TelegramAlerter {
	return &TelegramAlerter{
		token:  token,
		chatID: chatID,
	}
}

func (t *TelegramAlerter) SendAlert(message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)

	payload := map[string]string{
		"chat_id": t.chatID,
		"text":    message,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send telegram alert, status: %s", resp.Status)
	}

	return nil
}
