package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	C "factors/config"
	"factors/handler/helpers"
	M "factors/model"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
)

// scope constants.
const SCOPE_PROJECT_ID = "projectId"
const SCOPE_PROJECT_TOKEN = "projectToken"
const SCOPE_PROJECT_PRIVATE_TOKEN = "projectPrivateToken"
const SCOPE_AUTHORIZED_PROJECTS = "authorizedProjects"
const SCOPE_LOGGEDIN_AGENT_UUID = "loggedInAgentUUID"
const SCOPE_REQ_ID = "requestId"
const SCOPE_SHOPIFY_HASH_EMAIL = "shopifyHashEmail"

// cors prefix constants.
const PREFIX_PATH_SDK = "/sdk/"
const PREFIX_PATH_INTEGRATIONS = "/integrations"
const SUB_ROUTE_SHOPIFY_INTEGRATION_SDK = "/shopify_sdk"

const ADMIN_LOGIN_TOKEN_SEP = ":"

// SetScopeProjectIdByToken - Request scope set by token on 'Authorization' header.
func SetScopeProjectIdByToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		token = strings.TrimSpace(token)
		if token == "" {
			errorMessage := "Missing authorization header"
			log.WithFields(log.Fields{"error": errorMessage}).Error("Request failed with auth failure.")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
			return
		}

		if C.IsBlockedSDKRequestProjectToken(token) {
			c.AbortWithStatusJSON(http.StatusOK,
				gin.H{"error": "Request failed. Blocked."})
			return
		}

		project, errCode := M.GetProjectByToken(token)
		if errCode != http.StatusFound {
			errorMessage := "Invalid token"
			log.WithFields(log.Fields{"error": errorMessage}).Error("Request failed because of invalid token.")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
			return
		}
		U.SetScope(c, SCOPE_PROJECT_ID, project.ID)

		c.Next()
	}
}

func SetScopeProjectToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		token = strings.TrimSpace(token)
		if token == "" {
			errorMessage := "Missing authorization header"
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
			return
		}

		if C.IsBlockedSDKRequestProjectToken(token) {
			c.AbortWithStatusJSON(http.StatusOK,
				gin.H{"error": "Request failed. Blocked."})
			return
		}

		U.SetScope(c, SCOPE_PROJECT_TOKEN, token)
		c.Next()
	}
}

func SetScopeProjectPrivateToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		token = strings.TrimSpace(token)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "Invalid authorization token"})
			return
		}

		if C.IsBlockedSDKRequestProjectToken(token) {
			c.AbortWithStatusJSON(http.StatusOK,
				gin.H{"error": "Request failed. Blocked."})
			return
		}

		U.SetScope(c, SCOPE_PROJECT_PRIVATE_TOKEN, token)
		c.Next()
	}
}

// SetScopeProjectIdByPrivateToken - Set project id scope by private
// token on 'Authorization' header.
func SetScopeProjectIdByPrivateToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		token = strings.TrimSpace(token)
		if token == "" {
			errorMessage := "Missing authorization header"
			log.WithFields(log.Fields{"error": errorMessage}).Error("Request failed with auth failure.")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
			return
		}

		if C.IsBlockedSDKRequestProjectToken(token) {
			c.AbortWithStatusJSON(http.StatusOK,
				gin.H{"error": "Request failed. Blocked."})
			return
		}

		project, errCode := M.GetProjectByPrivateToken(token)
		if errCode != http.StatusFound {
			errorMessage := "Invalid token"
			log.WithFields(log.Fields{"error": errorMessage}).Error("Request failed because of invalid private token.")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
			return
		}
		U.SetScope(c, SCOPE_PROJECT_ID, project.ID)

		c.Next()
	}
}

func decodeBasicAuthToken(basicAuthToken string) (string, error) {
	basicAuthToken = strings.TrimSpace(basicAuthToken)
	if basicAuthToken == "" {
		return "", errors.New("invalid authorization header")
	}

	base64TokenWithColon := strings.TrimPrefix(basicAuthToken, "Basic ")
	tokenWithColon, err := base64.StdEncoding.DecodeString(base64TokenWithColon)
	if err != nil {
		return "", errors.New("invalid basic auth token")
	}

	token := strings.TrimSuffix(string(tokenWithColon), ":")
	return token, nil
}

func SetScopeProjectPrivateTokenUsingBasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := decodeBasicAuthToken(c.Request.Header.Get("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "Invalid authorization token"})
			return
		}

		U.SetScope(c, SCOPE_PROJECT_PRIVATE_TOKEN, token)
		c.Next()
	}
}

// SetScopeProjectIdByPrivateTokenUsingBasicAuth - Set project id scope by private
// token on header 'Authorization': 'Basic <TOKEN>:'
func SetScopeProjectIdByPrivateTokenUsingBasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := decodeBasicAuthToken(c.Request.Header.Get("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "Invalid authorization token"})
			return
		}

		project, errCode := M.GetProjectByPrivateToken(token)
		if errCode != http.StatusFound {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "Invalid authorization token"})
			return
		}

		U.SetScope(c, SCOPE_PROJECT_ID, project.ID)
		c.Next()
	}
}

// SetScopeProjectIdByPrivateTokenUsingBasicAuth - Set project id scope by private
// token on header 'Authorization': 'Basic <TOKEN>:'
func SetScopeProjectIdByStoreAndSecret() gin.HandlerFunc {
	return func(c *gin.Context) {
		actualMacString := c.Request.Header.Get("X-Shopify-Hmac-Sha256")
		actualMacString = strings.TrimSpace(actualMacString)
		if actualMacString == "" {
			errorMessage := "Missing Shopify Hmac header"
			log.WithFields(log.Fields{"error": errorMessage}).Error("Request failed with missing Mac.")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
			return
		}

		shopifyDomain := c.Request.Header.Get("X-Shopify-Shop-Domain")
		shopifyDomain = strings.TrimSpace(shopifyDomain)
		if shopifyDomain == "" {
			errorMessage := "Missing Shopify Shop Domain header"
			log.WithFields(log.Fields{"error": errorMessage}).Error("Request failed with missing domain.")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
			return
		}

		projectId, secret, shouldHashEmail, errCode := M.GetProjectDetailsByShopifyDomain(shopifyDomain)
		if errCode != http.StatusFound {
			errorMessage := "Invalid Domain"
			log.WithFields(log.Fields{"error": errorMessage, "domain": shopifyDomain}).Error(
				"Request failed with invalid domain.")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
			return
		}

		project, errCode := M.GetProject(projectId)
		if errCode != http.StatusFound {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "Request failed. Invalid project."})
			return
		}

		if C.IsBlockedSDKRequestProjectToken(project.Token) {
			c.AbortWithStatusJSON(http.StatusOK,
				gin.H{"error": "Request failed. Blocked."})
			return
		}

		// Read the body content to verify token and restore it for processing later.
		// https://stackoverflow.com/questions/47186741/how-to-get-the-json-from-the-body-of-a-request-on-go/47295689#47295689
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
		}
		// Restore the io.ReadCloser to its original state
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(bodyBytes)
		macSum := mac.Sum(nil)
		expectedMac := []byte(base64.StdEncoding.EncodeToString(macSum))
		actualMac := []byte(actualMacString)
		if !hmac.Equal(actualMac, expectedMac) {
			errorMessage := fmt.Sprintf("Invalid Token. Expected: %s, Actual: %s", expectedMac, actualMac)
			log.WithFields(log.Fields{"error": errorMessage}).Error(
				"Request failed with invalid domain.")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
			return
		}

		U.SetScope(c, SCOPE_PROJECT_ID, projectId)
		U.SetScope(c, SCOPE_SHOPIFY_HASH_EMAIL, shouldHashEmail)

		c.Next()
	}
}

func isSDKRequest(path string) bool {
	return strings.HasPrefix(path, PREFIX_PATH_SDK) ||
		strings.Contains(path, SUB_ROUTE_SHOPIFY_INTEGRATION_SDK)
}

func isIntergrationsRequest(path string) bool {
	return strings.HasPrefix(path, PREFIX_PATH_INTEGRATIONS)
}

const SAMEORIGIN = "SAMEORIGIN"

func AddSecurityHeadersForAppRoutes() gin.HandlerFunc {
	return func(c *gin.Context) {

		if !isSDKRequest(c.Request.URL.Path) && !isIntergrationsRequest(c.Request.URL.Path) {
			c.Header("X-Frame-Options", SAMEORIGIN)
		}
		c.Next()
	}
}

func CustomCors() gin.HandlerFunc {
	return func(c *gin.Context) {
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "HEAD", "DELETE"}

		if isSDKRequest(c.Request.URL.Path) {
			corsConfig.AllowAllOrigins = true
			corsConfig.AddAllowHeaders("Authorization")
		} else {
			if C.IsDevelopment() || C.IsStaging() {
				log.Infof("Running custom cors in %s environment.", C.GetConfig().Env)
				corsConfig.AllowOrigins = []string{
					"http://localhost:8080",
					"http://localhost:3000",
					"http://localhost:8090",
					"http://127.0.0.1:3000",
					"http://factors-dev.com:3000",
					"http://staging-app.factors.ai",
					"https://staging-app.factors.ai",
					"https://tufte-staging.factors.ai",
				}
			} else {
				corsConfig.AllowOrigins = []string{
					"http://app.factors.ai",
					"https://app.factors.ai",
					"https://tufte-prod.factors.ai",
				}
			}

			corsConfig.AllowCredentials = true
			corsConfig.AddAllowHeaders("Access-Control-Allow-Headers")
			corsConfig.AddAllowHeaders("Access-Control-Allow-Origin")
			corsConfig.AddAllowHeaders("content-type")
			corsConfig.AddAllowHeaders("Authorization")
		}
		// Applys custom cors and proceed
		cors.New(corsConfig)(c)
		c.Next()
	}
}

func ValidateLoggedInAgentHasAccessToRequestProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		urlParamProjectId, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
		if err != nil || urlParamProjectId == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid project id on param."})
			return
		}

		authorizedProjects := U.GetScopeByKey(c, SCOPE_AUTHORIZED_PROJECTS)
		if authorizedProjects == nil {
			c.AbortWithStatusJSON(http.StatusForbidden,
				gin.H{"error": "Access Forbidden. No projects found."})
			return
		}

		for _, pid := range authorizedProjects.([]uint64) {
			if urlParamProjectId == pid {
				// Set scope projectId. This has to be used by other
				// handlers for projectId.
				U.SetScope(c, SCOPE_PROJECT_ID, pid)

				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Unauthorized access. No projects found."})
		return
	}
}

// returns *M.Agent, errString, errCode
func validateAuthData(authDataStr string) (*M.Agent, string, int) {
	if authDataStr == "" {
		return nil, "error parsing auth data empty", http.StatusBadRequest
	}
	authData, err := helpers.ParseAuthData(authDataStr)
	if err != nil {
		return nil, "error parsing auth data", http.StatusUnauthorized
	}

	agent, errCode := M.GetAgentByUUID(authData.AgentUUID)
	if errCode == http.StatusNotFound {
		return nil, "agent not found", http.StatusUnauthorized
	} else if errCode == http.StatusInternalServerError {
		return nil, "error fetching agent", http.StatusInternalServerError
	}

	email, err := helpers.ParseAndDecryptProtectedFields(agent.Salt, authData.ProtectedFields)
	if err != nil {
		return nil, "error parsing protected fields", http.StatusUnauthorized
	}

	if email != agent.Email {
		return nil, "token email and agent email do not match", http.StatusUnauthorized
	}
	return agent, "", http.StatusOK
}

func isAdminTokenLogin(token string) bool {
	configAdminEmail := C.GetConfig().AdminLoginEmail
	configAdminToken := C.GetConfig().AdminLoginToken
	token = strings.TrimSpace(token)

	return configAdminEmail != "" && configAdminToken != "" && token != "" &&
		// login token for admin would be <token>:<project_id>.
		strings.HasPrefix(token, configAdminToken+ADMIN_LOGIN_TOKEN_SEP)
}

func SetLoggedInAgent() gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginAgent *M.Agent

		loginAuthToken := c.Request.Header.Get("Authorization")
		loginAuthToken = strings.TrimSpace(loginAuthToken)
		if loginAuthToken != "" {
			// Admin token login.
			if isAdminTokenLogin(loginAuthToken) {
				agent, errCode := M.GetAgentByEmail(C.GetConfig().AdminLoginEmail)
				if errCode != http.StatusFound {
					c.AbortWithStatus(errCode)
					return
				}

				loginAgent = agent
			} else {
				// Agent token login.
				agentLoginTokenMap := C.GetConfig().LoginTokenMap
				for token, email := range agentLoginTokenMap {
					if loginAuthToken == token {
						agent, errCode := M.GetAgentByEmail(email)
						if errCode != http.StatusFound {
							c.AbortWithStatusJSON(errCode, gin.H{"error": "invalid token"})
							return
						}

						loginAgent = agent
						break
					}
				}
			}
		} else {
			// Cookie login.
			cookieStr, err := c.Cookie(C.GetFactorsCookieName())
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "session cookie not found"})
				return
			}

			if cookieStr == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "missing session cookie data"})
				return
			}

			agent, errMsg, errCode := validateAuthData(cookieStr)
			if errCode != http.StatusOK {
				c.AbortWithStatusJSON(errCode, gin.H{"error": errMsg})
				return
			}

			loginAgent = agent
		}

		// TODO
		// check if agent email is not verified
		// send to verification page

		if loginAgent == nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "unable to authenticate"})
			return
		}

		U.SetScope(c, SCOPE_LOGGEDIN_AGENT_UUID, loginAgent.UUID)
		c.Next()
	}
}

func SetAuthorizedProjectsByLoggedInAgent() gin.HandlerFunc {
	return func(c *gin.Context) {

		loggedInAgentUUID := U.GetScopeByKeyAsString(c, SCOPE_LOGGEDIN_AGENT_UUID)

		loginAdminToken := c.Request.Header.Get("Authorization")
		loginAdminToken = strings.TrimSpace(loginAdminToken)

		var projectIds []uint64

		if isAdminTokenLogin(loginAdminToken) {
			// Set project with admin token.
			// Login token for admin would be like, <token>:<project_id>.
			splitToken := strings.Split(loginAdminToken, ADMIN_LOGIN_TOKEN_SEP)
			if len(splitToken) != 2 {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid token."})
				return
			}

			tokenProjectId := strings.TrimSpace(splitToken[1])
			projectId, err := strconv.ParseUint(tokenProjectId, 10, 64)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid token and project_id."})
				return
			}
			projectIds = append(projectIds, projectId)
		} else {
			// Set project with project agent mapping.

			projectAgentMappings, errCode := M.GetProjectAgentMappingsByAgentUUID(loggedInAgentUUID)
			if errCode == http.StatusInternalServerError {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get projects."})
				return
			}

			for _, pam := range projectAgentMappings {
				projectIds = append(projectIds, pam.ProjectID)
			}
		}

		U.SetScope(c, SCOPE_AUTHORIZED_PROJECTS, projectIds)
		c.Next()
	}
}

func ValidateAgentActivationRequest() gin.HandlerFunc {

	return func(c *gin.Context) {
		token := c.Query("token")

		agent, errMsg, errCode := validateAuthData(token)
		if errCode != http.StatusOK {
			c.AbortWithStatusJSON(errCode, gin.H{
				"error": errMsg,
			})
			return
		}

		if agent.IsEmailVerified {
			c.AbortWithStatusJSON(http.StatusIMUsed, gin.H{
				"error": "agent is already verified",
			})
			return
		}

		U.SetScope(c, SCOPE_LOGGEDIN_AGENT_UUID, agent.UUID)
		c.Next()
	}
}

func ValidateAgentSetPasswordRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")

		agent, errMsg, errCode := validateAuthData(token)
		if errCode != http.StatusOK {
			c.AbortWithStatusJSON(errCode, gin.H{
				"error": errMsg,
			})
			return
		}
		U.SetScope(c, SCOPE_LOGGEDIN_AGENT_UUID, agent.UUID)
		c.Next()
	}
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				httprequest, _ := httputil.DumpRequest(c.Request, false)

				buf := make([]byte, 1024)
				runtime.Stack(buf, false)

				msg := fmt.Sprintf("Panic CausedBy: %v\nStackTrace: %v\nHttpReq: %v\n", r, string(buf), string(httprequest))

				log.Errorf("Recovering from panic: %v", msg)

				err := U.NotifyThroughSNS("APIPanicRecoveryMid", C.GetConfig().Env, msg)
				if err != nil {
					log.WithError(err).Error("failed to send message to sns")
				}

				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

func RequestIdGenerator() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqId := xid.New().String()
		U.SetScope(c, SCOPE_REQ_ID, reqId)
		c.Header("X-Req-Id", reqId)
		c.Next()
	}
}

func Logger() gin.HandlerFunc {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknow"
	}
	return func(c *gin.Context) {
		// other handler can change c.Path so:
		path := c.Request.URL.Path
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		referer := c.Request.Referer()
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}

		entry := log.WithFields(log.Fields{
			"hostname":    hostname,
			"x-req-id":    U.GetScopeByKeyAsString(c, SCOPE_REQ_ID),
			"statusCode":  statusCode,
			"latency(ms)": latency,
			"clientIP":    clientIP,
			"method":      c.Request.Method,
			"path":        path,
			"referer":     referer,
			"dataLength":  dataLength,
			"userAgent":   clientUserAgent,
			"type":        "reqlog",
		})

		msg := fmt.Sprintf("%s - %s [%s] \"%s %s\" %d %d \"%s\" \"%s\" (%dms)", clientIP, hostname, time.Now().UTC(), c.Request.Method, path, statusCode, dataLength, referer, clientUserAgent, latency)
		entry.Info(msg)
	}
}
