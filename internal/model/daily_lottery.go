package model

import (
	"time"

	"github.com/google/uuid"
)

type DailyLotteryConfig struct {
	ChatID          int64     `gorm:"primaryKey" json:"chat_id"`
	Enabled         bool      `gorm:"not null;default:false" json:"enabled"`
	DailyAttempts   int       `gorm:"not null;default:3" json:"daily_attempts"`
	CostPoints      int       `gorm:"not null;default:0" json:"cost_points"`
	GuaranteeOnLast bool      `gorm:"not null;default:true" json:"guarantee_on_last"`
	CreatedAt       time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt       time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

type DailyLotteryPrize struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ChatID    int64     `gorm:"not null;index;uniqueIndex:idx_daily_lottery_prize_chat_amount" json:"chat_id"`
	Amount    int       `gorm:"not null;uniqueIndex:idx_daily_lottery_prize_chat_amount" json:"amount"`
	Weight    int       `gorm:"not null;default:1" json:"weight"`
	Enabled   bool      `gorm:"not null;default:true" json:"enabled"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

type DailyLotteryAttempt struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ChatID     int64      `gorm:"not null;uniqueIndex:idx_daily_lottery_attempt,priority:1;index" json:"chat_id"`
	UserID     int64      `gorm:"not null;uniqueIndex:idx_daily_lottery_attempt,priority:2;index" json:"user_id"`
	DrawDate   string     `gorm:"size:10;not null;uniqueIndex:idx_daily_lottery_attempt,priority:3" json:"draw_date"`
	AttemptNo  int        `gorm:"not null;uniqueIndex:idx_daily_lottery_attempt,priority:4" json:"attempt_no"`
	Result     string     `gorm:"size:16;not null" json:"result"`
	Amount     int        `gorm:"not null;default:0" json:"amount"`
	CodeID     *uuid.UUID `gorm:"type:uuid;index" json:"code_id,omitempty"`
	CostPoints int        `gorm:"not null;default:0" json:"cost_points"`
	CreatedAt  time.Time  `gorm:"not null;default:now()" json:"created_at"`
}
