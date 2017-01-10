package claims

import (
	"fmt"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	ldap "gopkg.in/ldap.v2"
)

// LDAPClaims represents the claims used by a LDAP.
type LDAPClaims struct {
	Attributes map[string]string

	jwt.StandardClaims
}

// NewLDAPClaims returns a new CertificateClaims.
func NewLDAPClaims() *LDAPClaims {

	return &LDAPClaims{
		StandardClaims: jwt.StandardClaims{},
		Attributes:     map[string]string{},
	}
}

// FromMetadata verifies and returns the ldap claims for the given metadata.
func (c *LDAPClaims) FromMetadata(metadata map[string]interface{}) error {

	var LDAPAddress string
	var bindDN string
	var bindPassword string
	var baseDN string
	var username string
	var password string

	if _, ok := metadata["LDAPAddress"]; !ok {
		return fmt.Errorf("Metadata must contain the key 'LDAPAddress'")
	}
	LDAPAddress = metadata["LDAPAddress"].(string)

	if _, ok := metadata["bindDN"]; !ok {
		return fmt.Errorf("Metadata must contain the key 'bindDN'")
	}
	bindDN = metadata["bindDN"].(string)

	if _, ok := metadata["bindPassword"]; !ok {
		return fmt.Errorf("Metadata must contain the key 'bindPassword'")
	}
	bindPassword = metadata["bindPassword"].(string)

	if _, ok := metadata["username"]; !ok {
		return fmt.Errorf("Metadata must contain the key 'username'")
	}
	username = metadata["username"].(string)

	if _, ok := metadata["password"]; !ok {
		return fmt.Errorf("Metadata must contain the key 'password'")
	}
	password = metadata["password"].(string)

	if _, ok := metadata["baseDN"]; !ok {
		return fmt.Errorf("Metadata must contain the key 'baseDN'")
	}
	baseDN = metadata["baseDN"].(string)

	l, err := ldap.Dial("tcp", LDAPAddress)
	if err != nil {
		return err
	}
	defer l.Close()

	if err = l.Bind(bindDN, bindPassword); err != nil {
		return err
	}

	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(&(uid=%s))", username),
		nil,
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return err
	}

	if len(sr.Entries) != 1 {
		return fmt.Errorf("User does not exist or too many entries returned")
	}

	entry := sr.Entries[0]
	if err = l.Bind(entry.DN, password); err != nil {
		return err
	}

	dns, err := ldap.ParseDN(entry.DN)
	if err != nil {
		return err
	}

	var subOrgs []string
	var organization string

	for _, rdn := range dns.RDNs {
		attr := rdn.Attributes[0]
		if attr.Type == "ou" {
			subOrgs = append(subOrgs, attr.Value)
		}
		if attr.Type == "dc" {
			if len(organization) == 0 {
				organization = attr.Value
			} else {
				organization = organization + "." + attr.Value
			}
		}
	}

	if len(subOrgs) > 0 {
		c.Attributes["organizationalUnit"] = subOrgs[0]
	}

	if organization != "" {
		c.Attributes["organization"] = organization
	}

	c.Attributes["dn"] = strings.Replace(entry.DN, " ", "_", -1)

	for _, attr := range entry.Attributes {
		if attr.Name == "userPassword" || attr.Name == "objectClass" {
			continue
		}

		if attr.Values[0] == "" {
			continue
		}

		c.Attributes[attr.Name] = strings.Replace(attr.Values[0], " ", "_", -1)
	}

	return nil
}

// ToMidgardClaims returns the MidgardClaims from google claims.
func (c *LDAPClaims) ToMidgardClaims() *MidgardClaims {

	now := time.Now()

	return &MidgardClaims{
		StandardClaims: jwt.StandardClaims{
			Audience:  JWTAudience,
			Issuer:    JWTIssuer,
			ExpiresAt: now.Add(JWTValidity).Unix(),
			IssuedAt:  now.Unix(),
			Subject:   c.Subject,
		},
		Realm: "LDAP",
		Data:  c.Attributes,
	}
}
