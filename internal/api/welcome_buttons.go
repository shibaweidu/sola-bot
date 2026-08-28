package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *Server) ListWelcomeButtons(c *gin.Context) {
	if s.deps.WelcomeButtons == nil {
		writeError(c, http.StatusInternalServerError, "welcome button service is not configured")
		return
	}
	chatID, err := strconv.ParseInt(c.Param("chatID"), 10, 64)
	if err != nil || chatID == 0 {
		writeError(c, http.StatusBadRequest, "invalid chat_id")
		return
	}
	if !s.ensureChatAllowed(c, chatID) {
		return
	}
	items, err := s.deps.WelcomeButtons.List(c.Request.Context(), chatID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) ReplaceWelcomeButtons(c *gin.Context) {
	if s.deps.WelcomeButtons == nil {
		writeError(c, http.StatusInternalServerError, "welcome button service is not configured")
		return
	}
	chatID, err := strconv.ParseInt(c.Param("chatID"), 10, 64)
	if err != nil || chatID == 0 {
		writeError(c, http.StatusBadRequest, "invalid chat_id")
		return
	}
	if !s.ensureChatAllowed(c, chatID) {
		return
	}
	var req struct {
		Items []WelcomeButtonInput `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	items, err := s.deps.WelcomeButtons.Replace(c.Request.Context(), chatID, req.Items)
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
