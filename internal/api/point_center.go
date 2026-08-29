package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *Server) GetPointCenterConfig(c *gin.Context) {
	if s.deps.PointCenter == nil {
		writeError(c, http.StatusInternalServerError, "point center service is not configured")
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
	result, err := s.deps.PointCenter.GetConfig(c.Request.Context(), chatID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) UpdatePointCenterConfig(c *gin.Context) {
	if s.deps.PointCenter == nil {
		writeError(c, http.StatusInternalServerError, "point center service is not configured")
		return
	}
	var req PointCenterConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.ChatID == 0 {
		writeError(c, http.StatusBadRequest, "chat_id is required")
		return
	}
	if !s.ensureChatAllowed(c, req.ChatID) {
		return
	}
	result, err := s.deps.PointCenter.UpdateConfig(c.Request.Context(), req)
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) ResetPointCenterReferral(c *gin.Context) {
	if s.deps.PointCenter == nil {
		writeError(c, http.StatusInternalServerError, "point center service is not configured")
		return
	}
	var req PointCenterReferralResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.ChatID == 0 || req.UserID == 0 {
		writeError(c, http.StatusBadRequest, "chat_id and user_id are required")
		return
	}
	if !s.ensureChatAllowed(c, req.ChatID) {
		return
	}
	if err := s.deps.PointCenter.ResetReferralForTesting(c.Request.Context(), req.ChatID, req.UserID); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) ListExchangeCodes(c *gin.Context) {
	if s.deps.PointCenter == nil {
		writeError(c, http.StatusInternalServerError, "point center service is not configured")
		return
	}
	if !s.requireBotMenuAccess(c) {
		return
	}
	var query ExchangeCodeListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	if query.Amount != nil && *query.Amount <= 0 {
		writeError(c, http.StatusBadRequest, "amount must be greater than zero")
		return
	}
	page, err := s.deps.PointCenter.ListExchangeCodesPage(c.Request.Context(), query)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, page)
}

func (s *Server) ImportExchangeCodes(c *gin.Context) {
	if s.deps.PointCenter == nil {
		writeError(c, http.StatusInternalServerError, "point center service is not configured")
		return
	}
	if !s.requireBotMenuAccess(c) {
		return
	}
	var req ExchangeCodeImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := s.deps.PointCenter.ImportExchangeCodes(c.Request.Context(), req)
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) SummarizeExchangeCodes(c *gin.Context) {
	if s.deps.PointCenter == nil {
		writeError(c, http.StatusInternalServerError, "point center service is not configured")
		return
	}
	if !s.requireBotMenuAccess(c) {
		return
	}
	items, err := s.deps.PointCenter.SummarizeExchangeCodes(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) BatchUpdateExchangeCodes(c *gin.Context) {
	if s.deps.PointCenter == nil {
		writeError(c, http.StatusInternalServerError, "point center service is not configured")
		return
	}
	if !s.requireBotMenuAccess(c) {
		return
	}
	var req ExchangeCodeBatchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := s.deps.PointCenter.BatchUpdateExchangeCodes(c.Request.Context(), req)
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) BatchDeleteExchangeCodes(c *gin.Context) {
	if s.deps.PointCenter == nil {
		writeError(c, http.StatusInternalServerError, "point center service is not configured")
		return
	}
	if !s.requireBotMenuAccess(c) {
		return
	}
	var req ExchangeCodeBatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := s.deps.PointCenter.BatchDeleteExchangeCodes(c.Request.Context(), req)
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}
