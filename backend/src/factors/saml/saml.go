package saml

import (
	"crypto/x509"
	"encoding/pem"
	saml2 "github.com/russellhaering/gosaml2"
	// "github.com/russellhaering/gosaml2/types"
	dsig "github.com/russellhaering/goxmldsig"
	"factors/model/model"
)

func GetSamlServiceProvider(samlConfig model.SAMLConfiguration, destinationID string) *saml2.SAMLServiceProvider {
	block, _ := pem.Decode([]byte(samlConfig.Certificate))
	
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil
	}

	certStore := dsig.MemoryX509CertificateStore{
		Roots: []*x509.Certificate{cert},
	}

	return &saml2.SAMLServiceProvider{
		IdentityProviderSSOURL:      samlConfig.LoginURL,
		IdentityProviderIssuer:      "http://www.okta.com/exkggqcydkRkzGU2G5d7",
		AssertionConsumerServiceURL: destinationID,
		SignAuthnRequests:           true,
		AudienceURI:                 "123",
		IDPCertificateStore:         &certStore,
		NameIdFormat:                saml2.NameIdFormatPersistent,
	}
}
