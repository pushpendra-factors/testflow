package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	C "factors/config"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	session "factors/session/store"
	U "factors/util"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var connections = map[string]bool{
	"google-oauth2": true,
}

type Authenticator struct {
	*oidc.Provider
	oauth2.Config
}

type Auth0Values struct {
	Subject   string    `json:"sub"`
	IssuedAt  uint64    `json:"iat"`
	ExpiresAt uint64    `json:"exp"`
	UpdatedAt time.Time `json:"updated_at"`
}

const SIGNUP_FLOW = "signup"
const SIGNIN_FLOW = "login"
const ACTIVATE_FLOW = "activate"

func NewAuth() (*Authenticator, error) {
	auth := C.GetAuth0Info()
	provider, err := oidc.NewProvider(
		context.Background(),
		"https://"+auth.Domain+"/",
	)
	if err != nil {
		return nil, err
	}

	conf := oauth2.Config{
		ClientID:     auth.ClientId,
		ClientSecret: auth.ClientSecret,
		RedirectURL:  auth.CallbackUrl,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &Authenticator{
		Provider: provider,
		Config:   conf,
	}, nil
}

func (a *Authenticator) verifyIDToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: a.ClientID,
	}

	return a.Verifier(oidcConfig).Verify(ctx, rawIDToken)
}

func ExternalAuthentication(auth *Authenticator, flow string) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn := c.Query("connection")
		if !connections[conn] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection"})
		}
		connection := oauth2.SetAuthURLParam("connection", conn)
		state, err := generateRandomState(flow)
		if err != nil {
			log.WithError(err).Error("Failed to generate random state")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		err = session.GetSessionStore().SetValue(c, C.GetAuth0StateCookieName(), state)
		if err != nil {
			log.WithError(err).Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set state cookie"})
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, auth.AuthCodeURL(state, connection))
	}
}

func CallbackHandler(auth *Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		state := session.GetSessionStore().GetValueAsString(c, C.GetAuth0StateCookieName())
		if state == "" {
			log.Error("Error in auth0 callback handler, No State")
			c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, "", "NO_STATE"))
			return
		}

		if state != c.Query("state") {
			log.Error("Error in auth0 callback handler, Invalid State")
			c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, "", "INVALID_STATE"))
			return
		}

		err := session.GetSessionStore().DeleteValue(c, C.GetAuth0StateCookieName())
		if err != nil {
			log.WithError(err).Error(err.Error())
			log.Error("Error in auth0 callback handler, Session Error")
			c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, "", "SESSION_ERROR"))
			return
		}

		token, err := auth.Exchange(c.Request.Context(), c.Query("code"))
		if err != nil {
			log.Error("Error in auth0 callback handler, Token Exchange Error")
			c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, "", "TOKEN_ERROR"))
			return
		}

		idToken, err := auth.verifyIDToken(c.Request.Context(), token)
		if err != nil {
			log.Error("Error in auth0 callback handler, Token ID verification Error")
			c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, "", "VERIFY_ERROR"))
			return
		}

		profile := model.Auth0Profile{}
		if err := idToken.Claims(&profile); err != nil {
			log.Error("Error in auth0 callback handler, Token Error")
			c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, "", "TOKEN_ERROR"))
			return
		}

		flow, err := decodeState(state)
		if err != nil {
			log.Error("Error in auth0 callback handler, Invalid State")
			c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, "", "INVALID_STATE"))
			return
		}

		var existingAgent *model.Agent
		var errCode int

		if flow == SIGNUP_FLOW {

			if U.IsPersonalEmail(strings.TrimSpace(profile.Email)) {
				log.WithField("email", profile.Email).Error("Failed To Create Agent, Personal Email Provided")
				c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, flow, "INVALID_PERSONAL_EMAIL"))
				return
			}
			alreadyExists := false
			if existingAgent, errCode = store.GetStore().GetAgentByEmail(profile.Email); errCode == http.StatusInternalServerError {
				log.WithField("email", profile.Email).Error("Failed To Get Agent By Email")
				c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, flow, "DB_ERROR"))
				return
			} else if errCode == http.StatusFound {
				alreadyExists = true
			}

			value, err := generateValueBytes(profile.Subject, profile.IssuedAt, profile.ExpiresAt, profile.UpdatedAt)
			if err != nil {
				log.WithField("email", profile.Email).Error("Failed To Generate Value bytes")
				c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, flow, "SERVER_ERROR"))
				return
			}

			var createAgentResp *model.CreateAgentResponse

			if !alreadyExists {
				createAgentParams := model.CreateAgentParams{
					Agent: &model.Agent{
						Email:               profile.Email,
						LastName:            profile.LastName,
						FirstName:           profile.FirstName,
						IsEmailVerified:     profile.IsEmailVerified,
						IsAuth0User:         true,
						Value:               value,
						SubscribeNewsletter: true,
					},
					PlanCode: model.FreePlanCode,
				}
				createAgentResp, errCode = store.GetStore().CreateAgentWithDependencies(&createAgentParams)
				if errCode == http.StatusInternalServerError {
					log.WithField("email", profile.Email).Error("Failed To Create Agent")
					c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, flow, "SERVER_ERROR"))
					return
				}
			} else {
				errCode = store.GetStore().UpdateAgentEmailVerificationDetails(existingAgent.UUID, true)
				if errCode != http.StatusAccepted {
					log.WithField("email", profile.Email).Error("Failed To Update Agent Email Verification Details")
					c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, flow, "SERVER_ERROR"))
				}
			}
			if !alreadyExists {
				existingAgent = createAgentResp.Agent

				errCode = onboardingSlackAPICall(existingAgent)
				if errCode != http.StatusOK {
					log.WithField("email", existingAgent.Email).
						WithField("status_code", errCode).
						Error("Failed To Send Onboarding Slack")
				}
			}

		} else if flow == SIGNIN_FLOW {
			var errCode int
			existingAgent, errCode = store.GetStore().GetAgentByEmail(profile.Email)
			if errCode != http.StatusFound {
				log.WithField("email", profile.Email).Error("Failed To Sign In, Invalid Agent")
				c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, flow, "INVALID_AGENT"))
				return
			}

			if !existingAgent.IsEmailVerified {
				log.WithField("email", profile.Email).Error("Failed To Sign In,Agent not active ")
				c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, flow, "AGENT_NOT_ACTIVE"))
				return
			}

		} else if flow == ACTIVATE_FLOW {
			var errCode int
			existingAgent, errCode = store.GetStore().GetAgentByEmail(profile.Email)
			if errCode != http.StatusFound {
				log.WithField("email", profile.Email).Error("Failed To Activate Agent, Invalid Agent")
				c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, flow, "INVALID_AGENT"))
				return
			} else if existingAgent.IsEmailVerified {
				log.WithField("email", profile.Email).Error("Failed To Activate Agent, Aleady Active")
				c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, flow, "ALREADY_ACTIVE"))
				return
			}

			value, err := generateValueBytes(profile.Subject, profile.IssuedAt, profile.ExpiresAt, profile.UpdatedAt)
			if err != nil {
				log.WithField("email", profile.Email).Error("Failed To Activate Agent, Failed at generate value bytes")
				c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, flow, "SERVER_ERROR"))
				return
			}

			errCode = store.GetStore().UpdateAgentVerificationDetailsFromAuth0(existingAgent.UUID, profile.FirstName, profile.LastName, profile.IsEmailVerified, value)
			if errCode != http.StatusAccepted {
				log.WithField("email", existingAgent.Email).Error("Failed to update Agent verification details")
				c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, flow, "SERVER_ERROR"))
				return
			}
		} else {
			log.WithField("email", profile.Email).Error("Invalid Flow")
			c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, "", "INVALID_FLOW"))
			return
		}

		// Auth0 hence no password check
		ts := time.Now().UTC()
		errCode = store.GetStore().UpdateAgentLastLoginInfo(existingAgent.UUID, ts)
		if errCode != http.StatusAccepted {
			log.WithField("email", existingAgent.Email).Error("Failed to update Agent lastLoginInfo")
			c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, flow, "SERVER_ERROR"))
			return
		}

		cookieData, err := helpers.GetAuthData(existingAgent.Email, existingAgent.UUID, existingAgent.Salt, helpers.SecondsInOneMonth*time.Second)
		if err != nil {
			log.WithField("email", profile.Email).Error("Failed in auth0 callback, Failed to generate cookie data")
			c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, flow, "SERVER_ERROR"))
			return
		}
		domain := C.GetCookieDomian()

		cookie := C.UseSecureCookie()
		httpOnly := C.UseHTTPOnlyCookie()
		if C.IsDevBox() {
			cookie = true
			httpOnly = true
			c.SetSameSite(http.SameSiteNoneMode)
		}
		c.SetCookie(C.GetFactorsCookieName(), cookieData, helpers.SecondsInOneMonth, "/", domain, cookie, httpOnly)
		c.Redirect(http.StatusPermanentRedirect, buildRedirectURL(c, "", ""))
	}
}

func generateRandomState(flow string) (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	copy(b, flow+"|")
	state := base64.StdEncoding.EncodeToString(b)

	return state, nil
}

func decodeState(state string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(state)
	if err != nil {
		return "", err
	}
	return strings.Split(string(decoded), "|")[0], nil
}

func generateValueBytes(Subject string, IssuedAt, ExpiresAt uint64, UpdatedAt time.Time) (*postgres.Jsonb, error) {

	valueBytes, err := json.Marshal(Auth0Values{
		Subject:   Subject,
		IssuedAt:  IssuedAt,
		ExpiresAt: ExpiresAt,
		UpdatedAt: UpdatedAt,
	})
	if err != nil {
		return nil, err
	}
	value := postgres.Jsonb{RawMessage: json.RawMessage(valueBytes)}
	return &value, nil
}

func buildRedirectURL(c *gin.Context, flow string, errMsg string) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	u := url.URL{
		Scheme: scheme,
		Host:   C.GetAPPDomain(),
		Path:   flow,
	}
	q := u.Query()
	q.Set("error", errMsg)
	q.Set("mode", "auth0")
	u.RawQuery = q.Encode()
	return u.String()
}
