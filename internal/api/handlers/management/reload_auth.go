package management

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ReloadAuthFiles triggers a manual reload of all authentication files.
// POST /v0/management/reload-auth-files
func (h *Handler) ReloadAuthFiles(c *gin.Context) {
	if h.reloadAuthFilesFunc == nil {
		log.Warn("reload auth files function not configured")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "reload functionality not available"})
		return
	}

	log.Info("manual auth files reload requested via management API")
	h.reloadAuthFilesFunc()

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "auth files reload triggered successfully",
	})
}
