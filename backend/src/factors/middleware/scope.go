package middleware

import (
	M "factors/model"
	U "factors/util"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// SetProjectScopeByTokenMiddleware sets projectId scope to the request context
// based on token on the 'Authorization' header.
func SetProjectScopeByTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		token = strings.TrimSpace(token)
		if token == "" {
			errorMessage := "Missing authorization header"
			log.WithFields(log.Fields{"error": errorMessage}).Error("Request failed with auth failure.")
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]string{"error": errorMessage})
			return
		}

		project, errCode := M.GetProjectByToken(token)
		if errCode != M.DB_SUCCESS {
			errorMessage := "Invalid token"
			log.WithFields(log.Fields{"error": errorMessage}).Error("Request failed because of invalid token.")
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]string{"error": errorMessage})
			return
		}
		U.SetScope(c, "projectId", project.ID)

		c.Next()
	}
}
