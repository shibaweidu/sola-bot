package model

import "time"

const (
	WelcomeButtonDailyLottery = "daily_lottery"
	WelcomeButtonLink         = "link"
)

type WelcomeButton struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ChatID      int64     `gorm:"not null;uniqueIndex:idx_welcome_buttons_chat_label,priority:1;index" json:"chat_id"`
	Label       string    `gorm:"type:text;not null;uniqueIndex:idx_welcome_buttons_chat_label,priority:2" json:"label"`
	ActionType  string    `gorm:"size:24;not null" json:"action_type"`
	ActionValue string    `gorm:"type:text;not null;default:''" json:"action_value"`
	Enabled     bool      `gorm:"not null;default:true" json:"enabled"`
	SortOrder   int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;default:now()" json:"updated_at"`
}
