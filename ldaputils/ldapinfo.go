package ldaputils

import (
	"fmt"
	"strings"
)

const (
	userQueryString = "{USERNAME}"
)

// LDAPInfo holds information to authenticate a user using an LDAP Server.
type LDAPInfo struct {
	Address              string                 `json:"LDAPAddress"`
	BindDN               string                 `json:"LDAPBindDN"`
	BindPassword         string                 `json:"LDAPBindPassword"`
	BindSearchFilter     string                 `json:"LDAPBindSearchFilter"`
	SubjectKey           string                 `json:"LDAPSubjectKey"`
	IgnoreKeys           map[string]interface{} `json:"LDAPIgnoredKeys"`
	BaseDN               string                 `json:"LDAPBaseDN"`
	ConnSecurityProtocol string                 `json:"LDAPConnSecurityProtocol"`
	Username             string                 `json:"LDAPUsername"`
	Password             string                 `json:"LDAPPassword"`
}

// NewLDAPInfo returns a new LDAPInfo, or an error
func NewLDAPInfo(metadata map[string]interface{}) (*LDAPInfo, error) {

	if metadata == nil {
		return nil, fmt.Errorf("You must provide at least metadata or defaultMetdata")
	}

	info := &LDAPInfo{}

	var err error

	info.Address, err = findLDAPKey(LDAPAddressKey, metadata)
	if err != nil {
		return nil, err
	}

	info.BindDN, err = findLDAPKey(LDAPBindDNKey, metadata)
	if err != nil {
		return nil, err
	}

	info.BindPassword, err = findLDAPKey(LDAPBindPasswordKey, metadata)
	if err != nil {
		return nil, err
	}

	info.BindSearchFilter, err = findLDAPKey(LDAPBindSearchFilterKey, metadata)
	if err != nil {
		return nil, err
	}

	info.SubjectKey, err = findLDAPKey(LDAPSubjectKey, metadata)
	if err != nil {
		return nil, err
	}

	info.IgnoreKeys, err = findLDAPKeyMap(LDAPIgnoredKeys, metadata)
	if err != nil {
		return nil, err
	}

	info.ConnSecurityProtocol, err = findLDAPKey(LDAPConnSecurityProtocolKey, metadata)
	if err != nil {
		return nil, err
	}

	info.Username, err = findLDAPKey(LDAPUsernameKey, metadata)
	if err != nil {
		return nil, err
	}

	info.Password, err = findLDAPKey(LDAPPasswordKey, metadata)
	if err != nil {
		return nil, err
	}

	info.BaseDN, err = findLDAPKey(LDAPBaseDNKey, metadata)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// ToMap convert the LDAPInfo into a map[string]interface{}.
func (i *LDAPInfo) ToMap() map[string]interface{} {

	return map[string]interface{}{
		LDAPAddressKey:              i.Address,
		LDAPBindDNKey:               i.BindDN,
		LDAPBindPasswordKey:         i.BindPassword,
		LDAPBindSearchFilterKey:     i.BindSearchFilter,
		LDAPSubjectKey:              i.SubjectKey,
		LDAPIgnoredKeys:             i.IgnoreKeys,
		LDAPUsernameKey:             i.Username,
		LDAPPasswordKey:             i.Password,
		LDAPBaseDNKey:               i.BaseDN,
		LDAPConnSecurityProtocolKey: i.ConnSecurityProtocol,
	}
}

// GetUserQueryString returns the query string based on the filter and username provided.
func (i *LDAPInfo) GetUserQueryString() string {

	return strings.Replace(i.BindSearchFilter, userQueryString, i.Username, -1)
}
