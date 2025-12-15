package notify

import (
	"dnsbin/util"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	maxRetries = 3
	retryDelay = time.Second * 2
)

// Telegram handles sending notifications to Telegram
type Telegram struct {
	botToken string
	chatIDs  []int64
	interval time.Duration
	client   *tgbotapi.BotAPI
}

// NewTelegram creates a new Telegram notification handler
func NewTelegram(config util.Config) (*Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(config.Notify[0].BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
	}

	chatIDs, err := convertChatIDs(config.Notify[0].ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chat IDs: %w", err)
	}

	return &Telegram{
		botToken: config.Notify[0].BotToken,
		chatIDs:  chatIDs,
		interval: config.Interval,
		client:   bot,
	}, nil
}

// retryWithBackoff executes the given function with exponential backoff retry
func retryWithBackoff(operation func() (tgbotapi.Message, error)) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		_, err = operation()
		if err == nil {
			return nil
		}

		// Check if error is an APIError with ResponseParameters
		var apiErr *tgbotapi.Error
		if errors.As(err, &apiErr) {
			if apiErr.ResponseParameters.RetryAfter > 0 {
				time.Sleep(time.Duration(apiErr.ResponseParameters.RetryAfter) * time.Second)
				continue
			}
		}

		// If no ResponseParameters or RetryAfter is 0, use our default backoff
		if i < maxRetries-1 {
			time.Sleep(retryDelay * time.Duration(i+1))
		}
	}
	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
}

func (t *Telegram) SendText(content string) error {
	msg := tgbotapi.NewMessage(t.chatIDs[0], content)
	msg.ParseMode = tgbotapi.ModeHTML

	err := retryWithBackoff(func() (tgbotapi.Message, error) {
		return t.client.Send(msg)
	})
	if err != nil {
		return fmt.Errorf("failed to send message to Telegram chat %d: %w", t.chatIDs[0], err)
	}

	if t.interval != 0 {
		time.Sleep(t.interval)
	}

	return nil
}

// SendMarkdown sends a Markdown-formatted message to the single configured chat ID
func (t *Telegram) SendMarkdown(title, content string) error {
	fullMessage := title + "\n" + content

	telegramMsg := tgbotapi.NewMessage(t.chatIDs[0], fullMessage)
	telegramMsg.ParseMode = tgbotapi.ModeMarkdown

	err := retryWithBackoff(func() (tgbotapi.Message, error) {
		return t.client.Send(telegramMsg)
	})
	if err != nil {
		return fmt.Errorf("failed to send message to Telegram chat %d: %w", t.chatIDs[0], err)
	}

	if t.interval != 0 {
		time.Sleep(t.interval)
	}

	return nil
}

// convertChatIDs parses a comma-separated list of chat IDs
func convertChatIDs(rawIDs string) ([]int64, error) {
	ids := strings.Split(rawIDs, ",")
	var chatIDs []int64
	for _, id := range ids {
		chatID := strings.TrimSpace(id)
		if chatID == "" {
			continue
		}
		id64, err := strconv.ParseInt(chatID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to convert chatID %q to int64: %w", id, err)
		}
		chatIDs = append(chatIDs, id64)
	}
	if len(chatIDs) == 0 {
		return nil, fmt.Errorf("no valid chatIDs found")
	}
	return chatIDs, nil
}
