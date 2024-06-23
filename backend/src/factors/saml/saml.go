package saml

import (
	"crypto/x509"
	"encoding/pem"
	C "factors/config"
	saml2 "github.com/russellhaering/gosaml2"
	// "github.com/russellhaering/gosaml2/types"
	"factors/model/model"
	"fmt"
	dsig "github.com/russellhaering/goxmldsig"
)

func GetSamlServiceProvider(projectID int64, samlConfig model.SAMLConfiguration, destinationID string) *saml2.SAMLServiceProvider {
	block, _ := pem.Decode([]byte(samlConfig.Certificate))

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil
	}

	certStore := dsig.MemoryX509CertificateStore{
		Roots: []*x509.Certificate{cert},
	}
	IdentityProviderURL := fmt.Sprintf(C.GetProtocol()+"%d/factors.app", projectID)
	return &saml2.SAMLServiceProvider{
		IdentityProviderIssuer:      IdentityProviderURL,
		IdentityProviderSSOURL:      samlConfig.LoginURL,
		AssertionConsumerServiceURL: destinationID,
		IDPCertificateStore:         &certStore,
		NameIdFormat:                saml2.NameIdFormatPersistent,
	}
}
