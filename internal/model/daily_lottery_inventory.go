package model

import (
	"time"

	"github.com/google/uuid"
)

// DailyLotteryCode is inventory reserved exclusively for daily quota lottery prizes.
type DailyLotteryCode struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	ChatID       int64      `gorm:"not null;index:idx_daily_lottery_codes_lookup,priority:1" json:"chat_id"`
	Code         string     `gorm:"type:text;not null;uniqueIndex:idx_daily_lottery_codes_chat_code,priority:1" json:"code"`
	Amount       int        `gorm:"not null;index:idx_daily_lottery_codes_lookup,priority:2" json:"amount"`
	BatchName    string     `gorm:"type:text;not null;default:''" json:"batch_name"`
	RedeemURL    string     `gorm:"type:text;not null;default:''" json:"redeem_url"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	Status       string     `gorm:"size:16;not null;default:'available';index:idx_daily_lottery_codes_lookup,priority:3" json:"status"`
	AssignedUser int64      `gorm:"not null;default:0" json:"assigned_user"`
	AssignedAt   *time.Time `json:"assigned_at,omitempty"`
	CreatedAt    time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"not null;default:now()" json:"updated_at"`
}
