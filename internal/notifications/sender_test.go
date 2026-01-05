package notifications

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestBuildShoutrrrURL(t *testing.T) {
	tests := []struct {
		name    string
		svc     ServiceConfig
		want    string
		wantErr string
	}{
		{
			name: "telegram url",
			svc: ServiceConfig{
				Type: ServiceTypeTelegram,
				Config: map[string]string{
					"token":   "123456:ABC-DEF",
					"chat_id": "-1001234567890",
				},
			},
			want: "telegram://123456:ABC-DEF@telegram?chats=-1001234567890",
		},
		{
			name: "telegram with special chars in chat_id",
			svc: ServiceConfig{
				Type: ServiceTypeTelegram,
				Config: map[string]string{
					"token":   "123456:ABC-DEF",
					"chat_id": "@my_channel",
				},
			},
			want: "telegram://123456:ABC-DEF@telegram?chats=%40my_channel",
		},
		{
			name: "telegram missing token",
			svc: ServiceConfig{
				Type: ServiceTypeTelegram,
				Config: map[string]string{
					"chat_id": "-1001234567890",
				},
			},
			wantErr: "telegram token is required",
		},
		{
			name: "telegram missing chat_id",
			svc: ServiceConfig{
				Type: ServiceTypeTelegram,
				Config: map[string]string{
					"token": "123456:ABC-DEF",
				},
			},
			wantErr: "telegram chat_id is required",
		},
		{
			name: "smtp basic url",
			svc: ServiceConfig{
				Type: ServiceTypeSMTP,
				Config: map[string]string{
					"host": "smtp.example.com",
					"from": "sender@example.com",
					"to":   "receiver@example.com",
				},
			},
			want: "smtp://smtp.example.com/?from=sender%40example.com&to=receiver%40example.com",
		},
		{
			name: "smtp with port",
			svc: ServiceConfig{
				Type: ServiceTypeSMTP,
				Config: map[string]string{
					"host": "smtp.example.com",
					"port": "587",
					"from": "sender@example.com",
					"to":   "receiver@example.com",
				},
			},
			want: "smtp://smtp.example.com:587/?from=sender%40example.com&to=receiver%40example.com",
		},
		{
			name: "smtp with auth",
			svc: ServiceConfig{
				Type: ServiceTypeSMTP,
				Config: map[string]string{
					"host":     "smtp.example.com",
					"port":     "587",
					"user":     "myuser",
					"password": "mypass",
					"from":     "sender@example.com",
					"to":       "receiver@example.com",
				},
			},
			want: "smtp://myuser:mypass@smtp.example.com:587/?from=sender%40example.com&to=receiver%40example.com",
		},
		{
			name: "smtp user without password",
			svc: ServiceConfig{
				Type: ServiceTypeSMTP,
				Config: map[string]string{
					"host": "smtp.example.com",
					"user": "myuser",
					"from": "sender@example.com",
					"to":   "receiver@example.com",
				},
			},
			want: "smtp://myuser@smtp.example.com/?from=sender%40example.com&to=receiver%40example.com",
		},
		{
			name: "smtp missing host",
			svc: ServiceConfig{
				Type: ServiceTypeSMTP,
				Config: map[string]string{
					"from": "sender@example.com",
					"to":   "receiver@example.com",
				},
			},
			wantErr: "smtp host is required",
		},
		{
			name: "smtp missing from",
			svc: ServiceConfig{
				Type: ServiceTypeSMTP,
				Config: map[string]string{
					"host": "smtp.example.com",
					"to":   "receiver@example.com",
				},
			},
			wantErr: "smtp from address is required",
		},
		{
			name: "smtp missing to",
			svc: ServiceConfig{
				Type: ServiceTypeSMTP,
				Config: map[string]string{
					"host": "smtp.example.com",
					"from": "sender@example.com",
				},
			},
			wantErr: "smtp to address is required",
		},
		{
			name: "generic url",
			svc: ServiceConfig{
				Type: ServiceTypeGeneric,
				Config: map[string]string{
					"url": "https://api.example.com/webhook",
				},
			},
			want: "https://api.example.com/webhook",
		},
		{
			name: "generic missing url",
			svc: ServiceConfig{
				Type:   ServiceTypeGeneric,
				Config: map[string]string{},
			},
			wantErr: "generic url is required",
		},
		{
			name: "unknown service type",
			svc: ServiceConfig{
				Type:   "unknown",
				Config: map[string]string{},
			},
			wantErr: "unknown service type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildShoutrrrURL(tt.svc)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("BuildShoutrrrURL() error = nil, want error containing %q", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("BuildShoutrrrURL() error = %v, want error containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("BuildShoutrrrURL() error = %v, want nil", err)
				return
			}
			if got != tt.want {
				t.Errorf("BuildShoutrrrURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewSender(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name      string
		services  []ServiceConfig
		wantEmpty bool // whether urls should be empty
	}{
		{
			name:      "no services",
			services:  nil,
			wantEmpty: true,
		},
		{
			name:      "empty services",
			services:  []ServiceConfig{},
			wantEmpty: true,
		},
		{
			name: "all disabled services",
			services: []ServiceConfig{
				{
					ID:      "test-1",
					Type:    ServiceTypeTelegram,
					Name:    "Disabled Telegram",
					Enabled: false,
					Config: map[string]string{
						"token":   "123456:ABC",
						"chat_id": "-100123",
					},
				},
			},
			wantEmpty: true,
		},
		{
			name: "invalid service config skipped",
			services: []ServiceConfig{
				{
					ID:      "test-1",
					Type:    ServiceTypeTelegram,
					Name:    "Invalid Telegram",
					Enabled: true,
					Config:  map[string]string{}, // Missing required fields
				},
			},
			wantEmpty: true, // Invalid service is skipped, no valid services left
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender, err := NewSender(logger, tt.services)
			if err != nil {
				t.Errorf("NewSender() error = %v, want nil", err)
				return
			}
			if sender == nil {
				t.Error("NewSender() returned nil sender")
				return
			}
			if tt.wantEmpty && len(sender.urls) > 0 {
				t.Error("NewSender() urls should be empty")
			}
			if !tt.wantEmpty && len(sender.urls) == 0 {
				t.Error("NewSender() urls should not be empty")
			}
		})
	}
}

func TestSender_SendNoRouter(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	sender, err := NewSender(logger, nil)
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}

	// Send should succeed silently when no router is configured
	err = sender.Send(context.Background(), "Test", "Message")
	if err != nil {
		t.Errorf("Send() error = %v, want nil", err)
	}
}
