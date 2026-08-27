package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *Server) GetDailyLotteryConfig(c *gin.Context) {
	if s.deps.DailyLottery == nil {
		writeError(c, http.StatusInternalServerError, "daily lottery service is not configured")
		return
	}
	chatID, err := strconv.ParseInt(strings.TrimSpace(c.Query("chat_id")), 10, 64)
	if err != nil || chatID == 0 {
		writeError(c, http.StatusBadRequest, "invalid chat_id")
		return
	}
	if !s.ensureChatAllowed(c, chatID) {
		return
	}
	item, err := s.deps.DailyLottery.GetConfig(c.Request.Context(), chatID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) UpdateDailyLotteryConfig(c *gin.Context) {
	if s.deps.DailyLottery == nil {
		writeError(c, http.StatusInternalServerError, "daily lottery service is not configured")
		return
	}
	var req DailyLotteryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	if !s.ensureChatAllowed(c, int64(req.ChatID)) {
		return
	}
	item, err := s.deps.DailyLottery.UpdateConfig(c.Request.Context(), req)
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) ListDailyLotteryPrizes(c *gin.Context) {
	if s.deps.DailyLottery == nil {
		writeError(c, http.StatusInternalServerError, "daily lottery service is not configured")
		return
	}
	chatID, err := strconv.ParseInt(strings.TrimSpace(c.Query("chat_id")), 10, 64)
	if err != nil || chatID == 0 {
		writeError(c, http.StatusBadRequest, "invalid chat_id")
		return
	}
	if !s.ensureChatAllowed(c, chatID) {
		return
	}
	items, err := s.deps.DailyLottery.ListPrizes(c.Request.Context(), chatID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) ReplaceDailyLotteryPrizes(c *gin.Context) {
	if s.deps.DailyLottery == nil {
		writeError(c, http.StatusInternalServerError, "daily lottery service is not configured")
		return
	}
	var req struct {
		ChatID int                      `json:"chat_id" binding:"required"`
		Items  []DailyLotteryPrizeInput `json:"items" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	if !s.ensureChatAllowed(c, int64(req.ChatID)) {
		return
	}
	items, err := s.deps.DailyLottery.ReplacePrizes(c.Request.Context(), int64(req.ChatID), req.Items)
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
