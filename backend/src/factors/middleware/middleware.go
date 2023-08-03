package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	C "factors/config"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	"factors/model/store/memsql"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httputil"
	"os"
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

var HOSTED_DOMAINS = []string{
	"api.factors.ai",
	"staging-api.factors.ai",
}

const (
	FEATURE_UNAVAILABLE = 0
	FEATURE_AVAILABLE   = 1
	FEATURE_ENABLED     = 2
)

// BlockRequestGracefully - Blocks HTTP requests from proceeding
// further with StatusOK response, on mounted routes.
func BlockRequestGracefully() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sending empty JSON to avoid null object
		// reference on client.
		c.AbortWithStatusJSON(http.StatusOK, gin.H{})
		return
	}
}

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

		project, errCode := store.GetStore().GetProjectByToken(token)
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

func AddSecurityResponseHeadersToCustomDomain() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Note: When custom domain is used the c.Request.Host will contain
		// Internal IP of loadbalancer. This makes it not possible to enable
		// it by selected custom domain.
		if !U.StringValueIn(c.Request.Host, HOSTED_DOMAINS) {
			c.Header("Strict-Transport-Security", "max-age=31536000;includeSubDomains")
			c.Header("X-Frame-Options", "SAMEORIGIN")
			c.Header("X-Content-Type-Options", "nosniff")
		}
		c.Next()
	}
}

func IsBlockedIPByProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := U.GetScopeByKeyAsString(c, SCOPE_PROJECT_TOKEN)
		if !C.IsIPBlockingFeatureEnabled(token) {
			return
		}

		if token == "" {
			log.WithFields(log.Fields{"error": "Invalid token", "token": token}).
				Error("Request failed because of invalid token.")
			return
		}

		logCtx := log.WithFields(log.Fields{"token": token})

		settings, errCode := store.GetStore().GetProjectSettingByTokenWithCacheAndDefault(token)
		if errCode != http.StatusFound {
			logCtx.Error("Request failed. Project info not found.")
			return
		}
		checkListJson := settings.FilterIps
		if checkListJson == nil {
			c.Next()
			return
		}
		var filterIpsMap model.FilterIps
		err := U.DecodePostgresJsonbToStructType(checkListJson, &filterIpsMap)
		if err != nil {
			logCtx.WithError(err).Error("Internal server error. Couldn't decode json.")
			return
		}
		// Checks for IP-Address string in http-request and for IP Address
		isIpBlocked := memsql.IsBlockedIP(c.Request.RemoteAddr, c.ClientIP(), filterIpsMap)
		if isIpBlocked {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
			return
		}
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

		project, errCode := store.GetStore().GetProjectByPrivateToken(token)
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

		project, errCode := store.GetStore().GetProjectByPrivateToken(token)
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

		projectId, secret, shouldHashEmail, errCode := store.GetStore().GetProjectDetailsByShopifyDomain(shopifyDomain)
		if errCode != http.StatusFound {
			errorMessage := "Invalid Domain"
			log.WithFields(log.Fields{"error": errorMessage, "domain": shopifyDomain}).Error(
				"Request failed with invalid domain.")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
			return
		}

		project, errCode := store.GetStore().GetProject(projectId)
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
					"https://staging-app-old.factors.ai",
					"https://flash-staging.factors.ai",
					"https://sloth-staging.factors.ai",
				}
			} else {
				corsConfig.AllowOrigins = []string{
					"http://app.factors.ai",
					"https://app.factors.ai",
					"https://tufte-prod.factors.ai",
					"https://app-old.factors.ai",
					"http://localhost:3000",
					"http://factors-dev.com:3000",
					"https://flash.factors.ai",
					"https://sloth.factors.ai",
				}
			}

			corsConfig.AllowCredentials = true
			corsConfig.AddAllowHeaders("Access-Control-Allow-Headers")
			corsConfig.AddAllowHeaders("Access-Control-Allow-Origin")
			corsConfig.AddAllowHeaders("content-type")
			corsConfig.AddAllowHeaders("Authorization")
			corsConfig.AddAllowHeaders(model.QueryCacheRequestInvalidatedCacheHeader)
			corsConfig.AddAllowHeaders(model.QueryFunnelV2)
			corsConfig.AddAllowHeaders(helpers.HeaderUserFilterOptForProfiles)
			corsConfig.AddAllowHeaders(helpers.HeaderUserFilterOptForEventsAndUsers)
		}
		// Applys custom cors and proceed
		cors.New(corsConfig)(c)
		c.Next()
	}
}

func ValidateLoggedInAgentHasAccessToRequestProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		urlParamProjectId, err := strconv.ParseInt(c.Params.ByName("project_id"), 10, 64)
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

		for _, pid := range authorizedProjects.([]int64) {
			if urlParamProjectId == pid {
				// Set scope projectId. This has to be used by other
				// handlers for projectId.
				U.SetScope(c, SCOPE_PROJECT_ID, pid)

				c.Next()
				return
			}
		}

		if C.IsDemoProject(urlParamProjectId) && C.EnableDemoReadAccess() {
			U.SetScope(c, SCOPE_PROJECT_ID, urlParamProjectId)

			c.Next()
			return

		}

		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Unauthorized access. No projects found."})
		return
	}
}

func validateAuthData(authDataStr string, cookieExpiry int64) (*model.Agent, string, int) {
	if authDataStr == "" {
		return nil, "error parsing auth data empty", http.StatusBadRequest
	}
	authData, err := helpers.ParseAuthData(authDataStr)
	if err != nil {
		return nil, "error parsing auth data", http.StatusUnauthorized
	}

	agent, errCode := store.GetStore().GetAgentByUUID(authData.AgentUUID)
	if errCode == http.StatusNotFound {
		return nil, "agent not found", http.StatusUnauthorized
	} else if errCode == http.StatusInternalServerError {
		return nil, "error fetching agent", http.StatusInternalServerError
	}

	var passwordCreatedAt int64
	if agent.PasswordCreatedAt != nil {
		passwordCreatedAt = agent.PasswordCreatedAt.Unix()
	} else {
		passwordCreatedAt = 0
	}
	email, _, err := helpers.ParseAndDecryptProtectedFields(agent.Salt, agent.LastLoggedOut, passwordCreatedAt, authData.ProtectedFields, cookieExpiry)
	if err != nil {
		return nil, "error parsing protected fields", http.StatusUnauthorized
	}

	if email != agent.Email {
		return nil, "token email and agent email do not match", http.StatusUnauthorized
	}
	return agent, "", http.StatusOK
}

// // Function checking for black-listed IP Addresses
func IsBlockedIP(c *gin.Context) bool {
	checkString0 := c.Request.RemoteAddr // Checks for IP-Address-like sub-string in http-request
	checkString1 := c.ClientIP()         // Checks for IP Address similarity
	checkList := C.GetConfig().BlockedIPList
	BadRequest := false
	for _, blockedIP := range checkList {
		if strings.Contains(checkString0, blockedIP) ||
			strings.Contains(checkString1, blockedIP) {
			BadRequest = true
			return BadRequest
		}
	}
	return BadRequest
}

func IsBlockedEmail(c *gin.Context) bool {

	cookieStr, _ := c.Cookie(C.GetFactorsCookieName())

	agent, errMsg, errCode := validateAuthData(cookieStr, helpers.SecondsInOneMonth)
	if errCode != http.StatusOK {
		c.AbortWithStatusJSON(errCode, gin.H{"error": errMsg})
		return true
	}

	email := agent.Email
	checkList := C.GetConfig().BlockedEmailList
	for _, blockedEmail := range checkList {
		if strings.Contains(email, blockedEmail) {
			c.AbortWithStatus(http.StatusBadRequest)
			return true
		}
	}

	return false
}

func SetLoggedInAgent() gin.HandlerFunc {
	return func(c *gin.Context) {
		if C.EnableIPBlocking() && IsBlockedIP(c) {
			c.AbortWithStatus(http.StatusBadGateway)
			return
		}

		var loginAgent *model.Agent
		loginAuthToken := c.Request.Header.Get("Authorization")
		loginAuthToken = strings.TrimSpace(loginAuthToken)
		if loginAuthToken != "" {
			agentLoginTokenMap := C.GetConfig().LoginTokenMap
			for token, email := range agentLoginTokenMap {
				if loginAuthToken == token {
					agent, errCode := store.GetStore().GetAgentByEmail(email)
					if errCode != http.StatusFound {
						c.AbortWithStatusJSON(errCode, gin.H{"error": "invalid token"})
						return
					}

					loginAgent = agent
					break
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

			agent, errMsg, errCode := validateAuthData(cookieStr, helpers.SecondsInOneMonth)
			if errCode != http.StatusOK {
				c.AbortWithStatusJSON(errCode, gin.H{"error": errMsg})
				return
			}

			// Block temporary and blocked emails.
			if C.EnableEmailDomainBlocking() && IsBlockedEmail(c) {
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

func MonitoringAPIMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow whitelisted agents.
		statusCode, errMessage1, agent := DefineSetLoggedInAgentInternalOnly(c)
		if statusCode == http.StatusOK {
			U.SetScope(c, SCOPE_LOGGEDIN_AGENT_UUID, agent.UUID)
			c.Next()
			return
		}

		// Allow pre-defined token/secret. Used for internal usage from services.
		statusCode, errMessage2 := TokenMiddleware(c)
		if statusCode == http.StatusOK {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(statusCode, gin.H{"err1": errMessage1, "err2": errMessage2})
		return
	}
}

// TokenMiddleware - This method is for token based authorization
func TokenMiddleware(c *gin.Context) (int, string) {
	token := c.Request.Header.Get("Authorization")
	token = strings.TrimSpace(token)

	statusCode := http.StatusOK
	var errorMessage string
	if token != C.GetConfig().MonitoringAPIToken {
		statusCode = http.StatusUnauthorized
		errorMessage = "invalid monitoring API token"
	}

	return statusCode, errorMessage
}

// DefineSetLoggedInAgentInternalOnly - This method is for definition of SetLoggedInAgentInternalOnly middleware
func DefineSetLoggedInAgentInternalOnly(c *gin.Context) (int, string, *model.Agent) {
	// Cookie login.
	cookieStr, err := c.Cookie(C.GetFactorsCookieName())
	statusCode := http.StatusOK
	var msg string
	if err != nil || cookieStr == "" {
		statusCode = http.StatusUnauthorized
		msg = "session cookie not found"
		return statusCode, msg, nil
	}

	agent, errMsg, errCode := validateAuthData(cookieStr, helpers.SecondsInOneMonth)
	if errCode != http.StatusOK {
		statusCode = errCode
		msg = errMsg
	} else if agent == nil {
		statusCode = http.StatusBadRequest
		msg = "unable to authenticate"
	} else if !C.IsLoggedInUserWhitelistedForProjectAnalytics(agent.UUID) {
		statusCode = http.StatusUnauthorized
		msg = "operation allowed for only admins"
	}

	return statusCode, msg, agent
}

func SetLoggedInAgentInternalOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Cookie login.
		statusCode, msg, agent := DefineSetLoggedInAgentInternalOnly(c)
		if statusCode != http.StatusOK || agent == nil {
			c.AbortWithStatusJSON(statusCode, gin.H{"error": msg})
			return
		}

		U.SetScope(c, SCOPE_LOGGEDIN_AGENT_UUID, agent.UUID)
		c.Next()
	}
}

func SetAuthorizedProjectsByLoggedInAgent() gin.HandlerFunc {
	return func(c *gin.Context) {
		loggedInAgentUUID := U.GetScopeByKeyAsString(c, SCOPE_LOGGEDIN_AGENT_UUID)

		var projectIds []int64

		projectAgentMappings, errCode := store.GetStore().GetProjectAgentMappingsByAgentUUID(loggedInAgentUUID)
		if errCode == http.StatusInternalServerError {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get projects."})
			return
		}

		for _, pam := range projectAgentMappings {
			projectIds = append(projectIds, pam.ProjectID)
		}

		U.SetScope(c, SCOPE_AUTHORIZED_PROJECTS, projectIds)
		c.Next()
	}
}

func ValidateAgentActivationRequest() gin.HandlerFunc {

	return func(c *gin.Context) {
		token := c.Query("token")

		agent, errMsg, errCode := validateAuthData(token, helpers.SecondsInFifteenDays)
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

		agent, errMsg, errCode := validateAuthData(token, helpers.SecondsInOneDay)
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

func ValidateAccessToSharedEntity(entityType int) gin.HandlerFunc {
	return func(c *gin.Context) {
		shareString := c.Query("query_id")
		urlParamProjectId, err := strconv.ParseInt(c.Params.ByName("project_id"), 10, 64)
		if err != nil || urlParamProjectId == 0 {
			log.WithError(err).Error("Failed to parse project_id")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Invalid project id on param."})
			return
		}

		var agentId string
		cookieStr, err := c.Cookie(C.GetFactorsCookieName())
		if err != nil {
			agentId = ""
		} else {
			agent, errMsg, errCode := validateAuthData(cookieStr, helpers.SecondsInOneMonth)
			if errCode != http.StatusOK {
				log.Error(errMsg + ": Failed to validate auth data.")
				c.AbortWithStatusJSON(errCode, gin.H{"error": errMsg})
				return
			}
			agentId = agent.UUID
			U.SetScope(c, SCOPE_LOGGEDIN_AGENT_UUID, agent.UUID)

			if C.EnableDemoReadAccess() && C.IsDemoProject(urlParamProjectId) {
				U.SetScope(c, SCOPE_PROJECT_ID, urlParamProjectId)
				c.Next()
				return
			}
		}

		// Check whether is part of the project, if yes than access allowed directly
		_, agentErrCode := store.GetStore().GetProjectAgentMapping(urlParamProjectId, agentId)
		if agentErrCode == http.StatusInternalServerError {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get agent info",
			})
			return
		} else if agentErrCode != http.StatusFound { // Not part of the project, check whether it is shared
			sharedEntity, errCode := store.GetStore().GetShareableURLWithShareStringWithLargestScope(urlParamProjectId, shareString, entityType)
			if errCode == http.StatusNotFound {
				log.Error("No access to entity.")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "cannot access entity",
				})
				return
			} else if errCode != http.StatusFound {
				log.Error("Failed to get shared entity.")
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "failed to access entity",
				})
				return
			}

			if entityType == model.ShareableURLEntityTypeQuery {
				_, errCode := store.GetStore().GetQueryWithQueryIdString(urlParamProjectId, shareString)
				if errCode != http.StatusFound {
					log.Error("Failed to get query.")
					c.AbortWithStatusJSON(errCode, gin.H{
						"error": "query not found",
					})
					return
				}
			}

			if sharedEntity.ShareType == model.ShareableURLShareTypePublic {
				// Public share, access allowed but agent is not part of the project, so add an audit log
				errCode = store.GetStore().CreateSharableURLAudit(sharedEntity, agentId)
				if errCode != http.StatusOK {
					log.Error("Failed to create audit for shared entity.")
					c.AbortWithStatusJSON(errCode, gin.H{
						"error": "failed to create audit",
					})
					return
				}
				// Add allowed users case
			} else {
				log.Error("Forbidden access to entity.")
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "cannot access entity",
				})
				return
			}
		}
		U.SetScope(c, SCOPE_PROJECT_ID, urlParamProjectId)
		c.Next()
	}
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				httprequest, _ := httputil.DumpRequest(c.Request, false)
				logFields := log.Fields{"http_request": string(httprequest)}
				U.NotifyOnPanicWithErrorLog("APIPanicRecoveryMid", C.GetConfig().Env, r, &logFields)

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

func SkipAPIWritesIfDisabled() gin.HandlerFunc {
	return func(c *gin.Context) {
		if C.DisableDBWrites() {
			isAllowedPath := strings.HasSuffix(c.Request.URL.Path, "/query") ||
				strings.HasSuffix(c.Request.URL.Path, "/query/web_analytics") ||
				strings.HasSuffix(c.Request.URL.Path, "/agents/signin") ||
				strings.HasSuffix(c.Request.URL.Path, "/profiles/query") ||
				strings.HasSuffix(c.Request.URL.Path, "/v1/factor") || strings.HasSuffix(c.Request.URL.Path, "/v1/explainV2") || strings.HasSuffix(c.Request.URL.Path, "/filter_values")
			if (c.Request.Method == http.MethodPost && !isAllowedPath) ||
				c.Request.Method == http.MethodDelete || c.Request.Method == http.MethodPut {
				c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{"error": "Writes are disabled for MQL"})
				return
			}
		}
		c.Next()
	}
}

func SkipDemoProjectWriteAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		if C.EnableDemoReadAccess() {
			projectId := U.GetScopeByKeyAsInt64(c, SCOPE_PROJECT_ID)
			agentId := U.GetScopeByKeyAsString(c, SCOPE_LOGGEDIN_AGENT_UUID)
			if !C.IsLoggedInUserWhitelistedForProjectAnalytics(agentId) && C.IsDemoProject(projectId) {
				c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{"error": "Operations disallowed for Non-Admin users"})
				return
			}
			c.Next()
		}
	}
}

func BlockMaliciousPayload() gin.HandlerFunc {
	return func(c *gin.Context) {
		var bodyBytes []byte
		if c.Request.Body != nil {
			var err error
			bodyBytes, err = ioutil.ReadAll(c.Request.Body)
			if err != nil {
				log.WithError(err).Error("Failed to read request paylaod.")
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
		}

		if exists, code := U.HasMaliciousContent(string(bodyBytes)); exists {
			log.WithField("client_ip", c.ClientIP()).
				WithField("user_agent", c.Request.UserAgent()).
				WithError(code).Error("Malicious content on payload.")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		// Restore the io.ReadCloser to its original state
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		c.Next()
	}
}

// Feature gate middleware
// func FeatureMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// Get the name of the handler function
// 		if !C.IsEnabledFeatureGates() {
// 			c.Next()
// 			return
// 		}
// 		handlerName := runtime.FuncForPC(reflect.ValueOf(c.Handler()).Pointer()).Name()
// 		projectID := U.GetScopeByKeyAsInt64(c, SCOPE_PROJECT_ID)
// 		handlerFeatures := GetFeatureMap()
// 		features, ok := handlerFeatures[handlerName]
// 		if !ok {
// 			log.Error("Handler is not mapped to any feature", handlerName)
// 			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Handler is not mapped to any feature" + handlerName})
// 			return
// 		}

// 		for _, feature := range features {
// 			status, err := store.GetStore().GetFeatureStatusForProject(projectID, feature)
// 			if err != nil {
// 				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feature status for this project " + feature})
// 				return
// 			}
// 			if isFeatureAvailable(status) {
// 				c.Next()
// 				return
// 			}

// 			// if !isFeatureEnabled(status) {
// 			// 	c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{"error": "Feature not enabled for this project "})
// 			// 	return

// 			// }
// 		}
// 		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{"error": "Feature not available for this project "})
// 	}
// }

// Feature gate middleware new
func FeatureMiddleware(features []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !C.IsEnabledFeatureGatesV2() {
			c.Next()
			return
		}
		projectID := U.GetScopeByKeyAsInt64(c, SCOPE_PROJECT_ID)

		for _, feature := range features {
			status, err := store.GetStore().GetFeatureStatusForProjectV2(projectID, feature)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feature status for this project " + feature})
				return
			}
			if status {
				c.Next()
				return
			}

		}
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{"error": "Feature not available for this project "})
	}
}

// Check if a feature is available
func isFeatureAvailable(status int) bool {

	return status != FEATURE_UNAVAILABLE
}

// Check if a feature is enabled
func isFeatureEnabled(status int) bool {
	return status == FEATURE_ENABLED
}
