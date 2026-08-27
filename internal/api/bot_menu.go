package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *Server) ListBotMenu(c *gin.Context) {
	if !s.requireBotMenuAccess(c) {
		return
	}
	if s.deps.BotMenu == nil {
		writeError(c, http.StatusInternalServerError, "bot menu service is not configured")
		return
	}
	role := strings.TrimSpace(c.Query("role"))
	if role == "" {
		role = "member"
	}
	items, err := s.deps.BotMenu.List(c.Request.Context(), role)
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"role": role, "items": items})
}

func (s *Server) ReplaceBotMenu(c *gin.Context) {
	if !s.requireBotMenuAccess(c) {
		return
	}
	if s.deps.BotMenu == nil {
		writeError(c, http.StatusInternalServerError, "bot menu service is not configured")
		return
	}
	var req BotMenuReplaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	items, err := s.deps.BotMenu.Replace(c.Request.Context(), strings.TrimSpace(req.Role), req.Items)
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"role": strings.TrimSpace(req.Role), "items": items})
}

func (s *Server) ResetBotMenu(c *gin.Context) {
	if !s.requireBotMenuAccess(c) {
		return
	}
	if s.deps.BotMenu == nil {
		writeError(c, http.StatusInternalServerError, "bot menu service is not configured")
		return
	}
	role := strings.TrimSpace(c.Query("role"))
	if role == "" {
		role = "member"
	}
	items, err := s.deps.BotMenu.Reset(c.Request.Context(), role)
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"role": role, "items": items})
}

func (s *Server) requireBotMenuAccess(c *gin.Context) bool {
	claims, ok := CurrentAdminClaims(c)
	if !ok || claims == nil {
		writeError(c, http.StatusUnauthorized, "missing admin claims")
		return false
	}
	if claims.IsSuperAdmin() {
		return true
	}
	role := strings.ToLower(strings.TrimSpace(claims.Role))
	if role == "owner" || role == "owner_admin" {
		return true
	}
	writeError(c, http.StatusForbidden, "无权管理机器人菜单")
	return false
}
