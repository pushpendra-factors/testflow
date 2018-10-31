package middleware

import (
	C "factors/config"
	M "factors/model"
	U "factors/util"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// scope constants.
const SCOPE_PROJECT = "projectId"

// cors prefix constants.
const PREFIX_PATH_SDK = "/sdk/"

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
		U.SetScope(c, SCOPE_PROJECT, project.ID)

		c.Next()
	}
}

// CustomCorsMiddleware for customised cors configuration based on conditions.
func CustomCorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		corsConfig := cors.DefaultConfig()

		if strings.HasPrefix(c.Request.URL.Path, PREFIX_PATH_SDK) {
			log.Info(c.Request.URL.Path)
			corsConfig.AllowAllOrigins = true
			corsConfig.AddAllowHeaders("Authorization")
			cors.New(corsConfig)(c)
		} else {
			if C.IsDevelopment() {
				log.Info("Running in development..")
				corsConfig.AllowOrigins = []string{"http://localhost:8080", "http://localhost:3000"}
			}
		}

		// Applys custom cors and proceed.
		cors.New(corsConfig)(c)
		c.Next()
	}
}
