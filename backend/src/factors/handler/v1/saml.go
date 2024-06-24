package v1

import (
	// "errors"
	// mid "factors/middleware"
	// "crypto/x509"
	// "encoding/pem"
	C "factors/config"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	// saml2 "github.com/russellhaering/gosaml2"
	// "github.com/russellhaering/gosaml2/types"
	"net/http"
	"net/url"
	"strconv"
	"time"
	// dsig "github.com/russellhaering/goxmldsig"
	saml "factors/saml"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func IsValidSamlRequestHandler(c *gin.Context) {
	email := c.Query("email")

	agent, status := store.GetStore().GetAgentByEmail(email)

	if status != http.StatusFound {
		c.AbortWithStatusJSON(http.StatusBadRequest, "user does not exist.")
		return
	}
	agentUUID := agent.UUID

	pAM, status := store.GetStore().GetProjectAgentMappingsByAgentUUID(agentUUID)

	if status != http.StatusFound {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Project Agent Mappings"})
		return
	}

	var samlConfigExists bool
	var samlConfig *postgres.Jsonb

	for _, pam := range pAM {
		setting, status := store.GetStore().GetProjectSetting(pam.ProjectID)
		if status != http.StatusFound {
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		if setting.SSOState == model.SSO_STATE_SAML_ENABLED && setting.SamlConfiguration != nil {
			if samlConfigExists {
				// multiple saml configs found
				// need to figure out handling later
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Multiple SAML Configurations exist."})
				return
			}
			samlConfigExists = true
			samlConfig = setting.SamlConfiguration
		}
	}

	if !samlConfigExists {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "SAML SSO not enabled for this user"})
		return
	}

	var samlConfiguration model.SAMLConfiguration
	// Generate SAML request and redirect logic
	err := U.DecodePostgresJsonbToStructType(samlConfig, &samlConfiguration)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to get saml configuration"})
		return
	}
	c.Status(http.StatusOK)
}

func SamlLoginRequestHandler(c *gin.Context) {
	email := c.Query("email")

	agent, status := store.GetStore().GetAgentByEmail(email)

	if status != http.StatusFound {
		c.AbortWithStatusJSON(http.StatusBadRequest, "user does not exist.")
		return
	}
	agentUUID := agent.UUID

	pAM, status := store.GetStore().GetProjectAgentMappingsByAgentUUID(agentUUID)

	if status != http.StatusFound {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Project Agent Mappings"})
		return
	}

	var samlConfigExists bool
	var samlConfig *postgres.Jsonb
	var spConfig model.SAMLConfiguration
	var err error
	var projectID int64

	for _, pam := range pAM {
		setting, status := store.GetStore().GetProjectSetting(pam.ProjectID)
		if status != http.StatusFound {
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		if setting.SSOState == model.SSO_STATE_SAML_ENABLED && setting.SamlConfiguration != nil {
			if samlConfigExists {
				// multiple saml configs found
				// need to figure out handling later
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Multiple SAML Configurations exist."})
				return
			}
			samlConfigExists = true
			samlConfig = setting.SamlConfiguration
			spConfig, err = getSAMLConfigurationFromProjectSettings(*setting)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to build SAML request"})
			}
		}
	}

	logCtx := log.Fields{
		"project_id": projectID,
	}

	if !samlConfigExists {
		// should we redirect to login home page with error?
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "SAML SSO not enabled for this user"})
		return
	}

	var samlConfiguration model.SAMLConfiguration
	// Generate SAML request and redirect logic
	err = U.DecodePostgresJsonbToStructType(samlConfig, &samlConfiguration)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to get saml configuration"})
		return
	}

	// redirectURL, err := GetSAMLRedirectURL(samlConfiguration)
	// if err != nil {
	// 	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to generate SAML Request"})
	// 	return
	// }

	expectedDestinationID := fmt.Sprintf(C.GetProtocol()+C.GetAPIDomain()+"/project/%d/saml/acs", projectID)
	// saml authn request building
	sp := saml.GetSamlServiceProvider(projectID, spConfig, expectedDestinationID)
	url, err := sp.BuildAuthURL("")
	if err != nil {
		log.WithFields(logCtx).WithError(err).Error("Failed to build authn URL")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to build redirect url"})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, url)
}

func GetSAMLRedirectURL(samlConfig model.SAMLConfiguration) (url string, err error) {
	// TODO
	// RelayState parameter is meant to be an opaque identifier that is passed back without any modification or inspection

	// TODO : Generate SAML Request using library
	SamlRequest := ""

	RelayState := ""

	url = fmt.Sprintf(`%s?SAMLRequest=%s&RelayState=%s`, samlConfig.LoginURL, SamlRequest, RelayState)

	return url, nil
}

func SamlCallbackHandler(c *gin.Context) {

	var projectID int64
	var err error

	projectIdString := c.Param("project_id")
	if projectIdString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, "misconfigured acs url")
		return
	}
	projectID, err = strconv.ParseInt(projectIdString, 10, 64)
	if err != nil {
		log.WithField("projectID ", projectIdString).Error("failed to parse project ID")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// get settings

	setting, status := store.GetStore().GetProjectSetting(projectID)
	if status != http.StatusFound {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// get saml configuration
	samlConfig, err := getSAMLConfigurationFromProjectSettings(*setting)
	if err != nil {
		log.WithField("projectID ", projectIdString).Error("failed to get saml configuration")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// certBlock, _ := pem.Decode([]byte(samlConfig.Certificate))
	// cert, err := x509.ParseCertificate(certBlock.Bytes)
	// if err != nil {
	// 	log.WithField("projectID ", projectIdString).WithError(err).Error("failed to parse x509 certificate from existing configuration")
	// 	c.AbortWithStatus(http.StatusInternalServerError)
	// 	return
	// }

	// This is the destination ID we expect to see in the SAML assertion. We
	// verify this to make sure that this SAML assertion is meant for us, and
	// not some other SAML application in the identity provider
	expectedDestinationID := fmt.Sprintf(C.GetProtocol()+C.GetAPIDomain()+"/project/%s/saml/acs", projectIdString)

	// Get the raw SAML response, and verify it.
	rawSAMLResponse := c.Request.FormValue("SAMLResponse")

	// get the saml config of the original person done?
	// get the user from relay state ?

	sp := saml.GetSamlServiceProvider(projectID, samlConfig, expectedDestinationID)

	_, err = sp.ValidateEncodedResponse(rawSAMLResponse)
	if err != nil {
		log.WithField("projectID ", projectIdString).WithError(err).Error("failed to validate saml response")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	assertion, err := sp.RetrieveAssertionInfo(rawSAMLResponse)
	if err != nil {
		log.WithField("projectID ", projectIdString).WithError(err).Error("failed to retrireve assertion from saml response")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// samlResponse, err := saml2.Verify(rawSAMLResponse, "", cert, expectedDestinationID, time.Now())
	// if err != nil {
	// 	log.WithField("projectID ", projectIdString).WithError(err).Error("failed to verify SAML Response")
	// 	c.AbortWithStatus(http.StatusInternalServerError)
	// 	return
	// }

	// samlUserID will contain the user ID from the identity provider.
	//
	// If a user with that saml_id already exists in our database, we'll log the
	// user in as them. If no such user already exists, we'll create one first.
	samlUserEmailID := assertion.NameID

	agent, status := store.GetStore().GetAgentByEmail(samlUserEmailID)
	if status != http.StatusFound {
		// TODO : decide on just in time user provisioning?
		log.WithField("projectID ", projectIdString).Error("user does not exist.")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// log them in (ADD PROJECT LEVEL RESTRICTIONS)
	ts := time.Now().UTC()
	errCode := store.GetStore().UpdateAgentLastLoginInfo(agent.UUID, ts)
	if errCode != http.StatusAccepted {
		log.WithField("email", agent.Email).Error("Failed to update Agent lastLoginInfo")
		c.Redirect(http.StatusPermanentRedirect, buildRedirectURLforLogin(c, "login", "SERVER_ERROR"))
		return
	}

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, helpers.SecondsInOneMonth*time.Second)
	if err != nil {
		log.WithField("email", samlUserEmailID).Error("Failed in SAML callback, Failed to generate cookie data")
		c.Redirect(http.StatusPermanentRedirect, buildRedirectURLforLogin(c, "login", "SERVER_ERROR"))
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
	c.Redirect(http.StatusSeeOther, buildRedirectURLforLogin(c, "login", ""))

}

func getSAMLConfigurationFromProjectSettings(setting model.ProjectSetting) (model.SAMLConfiguration, error) {
	var samlConfig *postgres.Jsonb
	if setting.SSOState == model.SSO_STATE_SAML_ENABLED && setting.SamlConfiguration != nil {
		samlConfig = setting.SamlConfiguration
	}
	var samlConfiguration model.SAMLConfiguration
	err := U.DecodePostgresJsonbToStructType(samlConfig, &samlConfiguration)
	if err != nil {
		return samlConfiguration, err
	}
	return samlConfiguration, nil
}

func buildRedirectURLforLogin(c *gin.Context, flow string, errMsg string) string {
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
