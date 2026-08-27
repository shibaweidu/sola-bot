package model

import "time"

// PointCenterConfig stores the private point-center rules and copy for one chat.
type PointCenterConfig struct {
	ChatID               int64     `gorm:"primaryKey" json:"chat_id"`
	InviteEnabled        bool      `gorm:"not null;default:true" json:"invite_enabled"`
	InviterReward        int       `gorm:"not null;default:100" json:"inviter_reward"`
	InviteeReward        int       `gorm:"not null;default:20" json:"invitee_reward"`
	SignEnabled          bool      `gorm:"not null;default:true" json:"sign_enabled"`
	SignReward           int       `gorm:"not null;default:1" json:"sign_reward"`
	ExchangeEnabled      bool      `gorm:"not null;default:true" json:"exchange_enabled"`
	ExchangeMinimum      int       `gorm:"not null;default:10" json:"exchange_minimum"`
	ExchangeRate         int       `gorm:"not null;default:1" json:"exchange_rate"`
	ExchangeURL          string    `gorm:"type:text;not null;default:''" json:"exchange_url"`
	ExchangeInstructions string    `gorm:"type:text;not null;default:''" json:"exchange_instructions"`
	PurchaseURL          string    `gorm:"type:text;not null;default:''" json:"purchase_url"`
	PurchaseText         string    `gorm:"type:text;not null;default:''" json:"purchase_text"`
	ShopURL              string    `gorm:"type:text;not null;default:''" json:"shop_url"`
	ShopText             string    `gorm:"type:text;not null;default:''" json:"shop_text"`
	InviteText           string    `gorm:"type:text;not null;default:''" json:"invite_text"`
	InvitePageTemplate   string    `gorm:"type:text;not null;default:''" json:"invite_page_template"`
	InviteJoinURL        string    `gorm:"type:text;not null;default:''" json:"invite_join_url"`
	InviteJoinText       string    `gorm:"type:text;not null;default:''" json:"invite_join_text"`
	InviteSuccessText    string    `gorm:"type:text;not null;default:''" json:"invite_success_text"`
	PointsText           string    `gorm:"type:text;not null;default:''" json:"points_text"`
	SignText             string    `gorm:"type:text;not null;default:''" json:"sign_text"`
	ExchangeText         string    `gorm:"type:text;not null;default:''" json:"exchange_text"`
	RankText             string    `gorm:"type:text;not null;default:''" json:"rank_text"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type ReferralCode struct {
	BaseModel
	Code      string `gorm:"size:32;not null;uniqueIndex" json:"code"`
	BotID     int64  `gorm:"not null;index" json:"bot_id"`
	ChatID    int64  `gorm:"not null;index" json:"chat_id"`
	InviterID int64  `gorm:"not null;index" json:"inviter_id"`
	UseCount  int    `gorm:"not null;default:0" json:"use_count"`
	IsActive  bool   `gorm:"not null;default:true" json:"is_active"`
}

type Referral struct {
	BaseModel
	CodeID      UUID       `gorm:"type:uuid;not null;index" json:"code_id"`
	BotID       int64      `gorm:"not null;index" json:"bot_id"`
	ChatID      int64      `gorm:"not null;index" json:"chat_id"`
	InviterID   int64      `gorm:"not null;index" json:"inviter_id"`
	InviteeID   int64      `gorm:"not null;index" json:"invitee_id"`
	Status      string     `gorm:"size:24;not null;default:'started';index" json:"status"`
	StartedAt   time.Time  `gorm:"not null" json:"started_at"`
	ActivatedAt *time.Time `json:"activated_at,omitempty"`
	RewardedAt  *time.Time `json:"rewarded_at,omitempty"`
}

type DailySignRecord struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ChatID    int64     `gorm:"not null;uniqueIndex:idx_daily_sign_user_chat_day,priority:1;index" json:"chat_id"`
	UserID    int64     `gorm:"not null;uniqueIndex:idx_daily_sign_user_chat_day,priority:2;index" json:"user_id"`
	SignDate  string    `gorm:"size:10;not null;uniqueIndex:idx_daily_sign_user_chat_day,priority:3" json:"sign_date"`
	Reward    int       `gorm:"not null" json:"reward"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
}

type ExchangeCode struct {
	BaseModel
	Code         string     `gorm:"type:text;not null;uniqueIndex" json:"code"`
	BatchName    string     `gorm:"size:128;not null;default:'';index" json:"batch_name"`
	Amount       int        `gorm:"not null;default:1" json:"amount"`
	RedeemURL    string     `gorm:"type:text;not null;default:''" json:"redeem_url"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	Status       string     `gorm:"size:24;not null;default:'available';index" json:"status"`
	AssignedUser int64      `gorm:"index" json:"assigned_user"`
	AssignedChat int64      `gorm:"index" json:"assigned_chat"`
	AssignedAt   *time.Time `json:"assigned_at,omitempty"`
}

type ExchangeOrder struct {
	BaseModel
	UserID      int64  `gorm:"not null;index" json:"user_id"`
	ChatID      int64  `gorm:"not null;index" json:"chat_id"`
	PointsSpent int    `gorm:"not null" json:"points_spent"`
	Amount      int    `gorm:"not null" json:"amount"`
	CodeID      UUID   `gorm:"type:uuid;not null;index" json:"code_id"`
	Status      string `gorm:"size:24;not null;default:'assigned';index" json:"status"`
	RedeemURL   string `gorm:"type:text;not null;default:''" json:"redeem_url"`
}
