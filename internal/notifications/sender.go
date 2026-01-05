package notifications

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/containrrr/shoutrrr"
)

// Sender sends notifications via configured services.
type Sender struct {
	logger   *slog.Logger
	services []ServiceConfig
	urls     []string // Shoutrrr URLs for enabled services
}

// NewSender creates a new notification sender with the given services.
func NewSender(logger *slog.Logger, services []ServiceConfig) (*Sender, error) {
	s := &Sender{
		logger:   logger,
		services: services,
	}

	// Build URLs for all enabled services
	for _, svc := range services {
		if !svc.Enabled {
			continue
		}
		u, err := BuildShoutrrrURL(svc)
		if err != nil {
			logger.Warn("skipping invalid service config",
				slog.String("service_id", svc.ID),
				slog.String("service_name", svc.Name),
				slog.String("error", err.Error()),
			)
			continue
		}
		s.urls = append(s.urls, u)
	}

	return s, nil
}

// Send sends a notification to all enabled services.
func (s *Sender) Send(ctx context.Context, title, message string) error {
	if len(s.urls) == 0 {
		s.logger.Debug("no notification services configured, skipping send")
		return nil
	}

	// Format message with title
	fullMessage := message
	if title != "" {
		fullMessage = fmt.Sprintf("%s\n\n%s", title, message)
	}

	// Create a fresh router per send to avoid race conditions.
	// Shoutrrr's ServiceRouter is not safe for concurrent use when modifying Timeout.
	sr, err := shoutrrr.CreateSender(s.urls...)
	if err != nil {
		return fmt.Errorf("creating shoutrrr router: %w", err)
	}

	// Set router timeout from context deadline if available.
	// Shoutrrr doesn't support context directly, so we approximate by setting
	// the router's timeout field based on the context's remaining time.
	if deadline, ok := ctx.Deadline(); ok {
		if remaining := time.Until(deadline); remaining > 0 {
			sr.Timeout = remaining
		}
	}

	errs := sr.Send(fullMessage, nil)

	// Collect all errors
	var errMsgs []string
	for _, err := range errs {
		if err != nil {
			errMsgs = append(errMsgs, err.Error())
		}
	}

	if len(errMsgs) > 0 {
		return fmt.Errorf("sending notifications: %s", strings.Join(errMsgs, "; "))
	}

	s.logger.Info("notification sent successfully",
		slog.String("title", title),
	)

	return nil
}

// BuildShoutrrrURL converts a ServiceConfig to a Shoutrrr URL format.
func BuildShoutrrrURL(svc ServiceConfig) (string, error) {
	switch svc.Type {
	case ServiceTypeTelegram:
		return buildTelegramURL(svc.Config)
	case ServiceTypeSMTP:
		return buildSMTPURL(svc.Config)
	case ServiceTypeGeneric:
		return buildGenericURL(svc.Config)
	default:
		return "", fmt.Errorf("unknown service type: %s", svc.Type)
	}
}

// buildTelegramURL creates a Telegram Shoutrrr URL.
// Format: telegram://token@telegram?chats=chatid
func buildTelegramURL(config map[string]string) (string, error) {
	token := config["token"]
	chatID := config["chat_id"]

	if token == "" {
		return "", fmt.Errorf("telegram token is required")
	}
	if chatID == "" {
		return "", fmt.Errorf("telegram chat_id is required")
	}

	return fmt.Sprintf("telegram://%s@telegram?chats=%s", token, url.QueryEscape(chatID)), nil
}

// buildSMTPURL creates an SMTP Shoutrrr URL.
// Format: smtp://user:pass@host:port/?from=x&to=y
func buildSMTPURL(config map[string]string) (string, error) {
	host := config["host"]
	port := config["port"]
	user := config["user"]
	pass := config["password"]
	from := config["from"]
	to := config["to"]

	if host == "" {
		return "", fmt.Errorf("smtp host is required")
	}
	if from == "" {
		return "", fmt.Errorf("smtp from address is required")
	}
	if to == "" {
		return "", fmt.Errorf("smtp to address is required")
	}

	// Build the URL
	var userInfo string
	if user != "" {
		if pass != "" {
			userInfo = fmt.Sprintf("%s:%s@", url.QueryEscape(user), url.QueryEscape(pass))
		} else {
			userInfo = fmt.Sprintf("%s@", url.QueryEscape(user))
		}
	}

	hostPort := host
	if port != "" {
		hostPort = fmt.Sprintf("%s:%s", host, port)
	}

	return fmt.Sprintf("smtp://%s%s/?from=%s&to=%s",
		userInfo,
		hostPort,
		url.QueryEscape(from),
		url.QueryEscape(to),
	), nil
}

// buildGenericURL uses the URL directly from config.
func buildGenericURL(config map[string]string) (string, error) {
	urlStr := config["url"]
	if urlStr == "" {
		return "", fmt.Errorf("generic url is required")
	}
	return urlStr, nil
}

// TestService tests a single service configuration by sending a test message.
func TestService(ctx context.Context, logger *slog.Logger, svc ServiceConfig) error {
	u, err := BuildShoutrrrURL(svc)
	if err != nil {
		return fmt.Errorf("invalid service configuration: %w", err)
	}

	sr, err := shoutrrr.CreateSender(u)
	if err != nil {
		return fmt.Errorf("creating sender: %w", err)
	}

	errs := sr.Send("Quantlete test notification - configuration verified!", nil)

	for _, e := range errs {
		if e != nil {
			return fmt.Errorf("sending test notification: %w", e)
		}
	}

	logger.Info("test notification sent successfully",
		slog.String("service_id", svc.ID),
		slog.String("service_name", svc.Name),
	)

	return nil
}
