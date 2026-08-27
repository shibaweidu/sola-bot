package bot

import (
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func TestLockedContentReason(t *testing.T) {
	tests := []struct {
		name string
		cfg  ChatModerationConfig
		msg  *gotgbot.Message
		text string
		want string
	}{
		{
			name: "blocked link",
			cfg:  ChatModerationConfig{BlockLinks: true},
			msg:  &gotgbot.Message{},
			text: "visit https://spam.example now",
			want: "link",
		},
		{
			name: "whitelisted link",
			cfg:  ChatModerationConfig{BlockLinks: true, LinkWhitelist: []string{"example.com"}},
			msg:  &gotgbot.Message{},
			text: "visit https://example.com/help",
		},
		{
			name: "forward",
			cfg:  ChatModerationConfig{BlockForwards: true},
			msg:  &gotgbot.Message{ForwardOrigin: gotgbot.MessageOriginHiddenUser{}},
			want: "forward",
		},
		{
			name: "media",
			cfg:  ChatModerationConfig{BlockMedia: true},
			msg:  &gotgbot.Message{Photo: []gotgbot.PhotoSize{{}}},
			want: "media",
		},
		{
			name: "locks disabled",
			cfg:  ChatModerationConfig{},
			msg:  &gotgbot.Message{Photo: []gotgbot.PhotoSize{{}}},
			text: "https://spam.example",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lockedContentReason(tt.cfg, tt.msg, tt.text); got != tt.want {
				t.Fatalf("lockedContentReason() = %q, want %q", got, tt.want)
			}
		})
	}
}
