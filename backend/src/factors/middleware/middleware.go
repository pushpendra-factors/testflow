package middleware

import (
	C "factors/config"
	M "factors/model"
	U "factors/util"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// scope constants.
const SCOPE_PROJECT = "projectId"
const SCOPE_AUTHORIZED_PROJECTS = "authorizedProjects"

// cors prefix constants.
const PREFIX_PATH_SDK = "/sdk/"

// SetScopeProjectIdByToken - Request scope set by token on 'Authorization' header.
func SetScopeProjectIdByToken() gin.HandlerFunc {
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
		if errCode != http.StatusFound {
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
func CustomCors() gin.HandlerFunc {
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
				corsConfig.AllowOrigins = []string{"http://localhost:8080", "http://localhost:3000", "http://localhost:8090"}
			}
		}

		// Applys custom cors and proceed.
		cors.New(corsConfig)(c)
		c.Next()
	}
}

// SetScopeAuthorizedProjectsBySubdomain - scope set by subdomain.
func SetScopeAuthorizedProjectsBySubdomain() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only requests from dev and localhost authorized to access all projects. For tests.
		if C.IsDevelopment() && U.IsRequestFromLocalhost(c.Request.Host) {
			allProjects, errCode := M.GetProjects()
			if errCode != http.StatusFound {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Dev envinoment failure. Failed to get projects."})
				return
			}

			var projectIds []uint64
			for _, project := range allProjects {
				projectIds = append(projectIds, project.ID)
			}

			U.SetScope(c, SCOPE_AUTHORIZED_PROJECTS, projectIds)

			c.Next()
			return
		}

		if C.IsTokenLoginEnabled() {
			loginTokenCache := C.GetLoginTokenCache().Map
			subdomain, err := U.GetRequestSubdomain(c.Request.Host)

			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access. Invalid subdomain."})
				return
			}

			projectIds, tokenExists := loginTokenCache[subdomain]
			if tokenExists {
				U.SetScope(c, SCOPE_AUTHORIZED_PROJECTS, projectIds)
			}
		}

		c.Next()
	}
}

// IsAuthorized - Authorizes request by validating authorized projects scope.
func IsAuthorized() gin.HandlerFunc {
	return func(c *gin.Context) {
		paramProjectId, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
		if err != nil || paramProjectId == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid project id on param."})
			return
		}

		authorizedProjects := U.GetScopeByKey(c, SCOPE_AUTHORIZED_PROJECTS)
		if authorizedProjects == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access. No projects found."})
			return
		}

		for _, pid := range authorizedProjects.([]uint64) {
			if paramProjectId == pid {
				// Set scope projectId. This has to be used by other
				// handlers for projectId.
				U.SetScope(c, SCOPE_PROJECT, pid)

				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access. No projects found."})
		return
	}
}

// DenyPublicAccess - Allows only localhost.
func DenyPublicAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !U.IsRequestFromLocalhost(c.Request.Host) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access. Restricted public access."})
			return
		}
		c.Next()
	}
}
