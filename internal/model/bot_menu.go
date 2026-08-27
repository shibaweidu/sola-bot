package model

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	BotMenuRoleMember = "member"
	BotMenuRoleAdmin  = "admin"

	BotMenuActionBuiltin = "builtin"
	BotMenuActionLink    = "link"
	BotMenuActionCustom  = "custom"

	BotMenuLinkModeText    = "text"
	BotMenuLinkModeButtons = "buttons"
)

type BotMenuInlineLink struct {
	Label       string `json:"label"`
	URL         string `json:"url"`
	RowIndex    int    `json:"row_index"`
	ColumnIndex int    `json:"column_index"`
}

type BotMenuItem struct {
	ID              uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	TelegramBotID   int64     `gorm:"not null;index:idx_bot_menu_bot_role_order,priority:1" json:"telegram_bot_id"`
	Role            string    `gorm:"size:16;not null;index:idx_bot_menu_bot_role_order,priority:2" json:"role"`
	Label           string    `gorm:"size:128;not null" json:"label"`
	Icon            string    `gorm:"size:32;not null;default:''" json:"icon"`
	ActionType      string    `gorm:"size:16;not null" json:"action_type"`
	ActionKey       string    `gorm:"size:64;not null;default:''" json:"action_key"`
	ActionValue     string    `gorm:"size:1024;not null;default:''" json:"action_value"`
	MessageText     string    `gorm:"type:text;not null;default:''" json:"message_text"`
	LinkMode        string    `gorm:"size:16;not null;default:'text'" json:"link_mode"`
	LinkButtonsJSON string    `gorm:"column:link_buttons_json;type:jsonb;not null;default:'[]'" json:"-"`
	RowIndex        int       `gorm:"not null;default:0;index:idx_bot_menu_bot_role_order,priority:3" json:"row_index"`
	ColumnIndex     int       `gorm:"not null;default:0;index:idx_bot_menu_bot_role_order,priority:4" json:"column_index"`
	Enabled         bool      `gorm:"not null;default:true" json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (item BotMenuItem) RenderLabel() string {
	label := strings.TrimSpace(item.Label)
	icon := strings.TrimSpace(item.Icon)
	if icon == "" || strings.HasPrefix(label, icon) {
		return label
	}
	return strings.TrimSpace(icon + " " + label)
}

func IsValidBotMenuRole(role string) bool {
	return role == BotMenuRoleMember || role == BotMenuRoleAdmin
}

func IsValidBotMenuAction(action string) bool {
	return action == BotMenuActionBuiltin || action == BotMenuActionLink || action == BotMenuActionCustom
}

func IsValidBotMenuLink(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return false
	}
	return true
}

func NormalizeBotMenuLinkMode(mode string) string {
	mode = strings.TrimSpace(mode)
	if mode == "" {
		return BotMenuLinkModeText
	}
	return mode
}

func IsValidBotMenuLinkMode(mode string) bool {
	mode = NormalizeBotMenuLinkMode(mode)
	return mode == BotMenuLinkModeText || mode == BotMenuLinkModeButtons
}

func ParseBotMenuInlineLinks(raw string) ([]BotMenuInlineLink, error) {
	if strings.TrimSpace(raw) == "" {
		return []BotMenuInlineLink{}, nil
	}
	var links []BotMenuInlineLink
	if err := json.Unmarshal([]byte(raw), &links); err != nil {
		return nil, err
	}
	if links == nil {
		links = []BotMenuInlineLink{}
	}
	return links, nil
}

func EncodeBotMenuInlineLinks(links []BotMenuInlineLink) (string, error) {
	if links == nil {
		links = []BotMenuInlineLink{}
	}
	data, err := json.Marshal(links)
	return string(data), err
}
