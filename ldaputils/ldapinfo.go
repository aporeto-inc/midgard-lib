// Copyright 2019 Aporeto Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	Address              string                 `msgpack:"address" json:"address"`
	BindDN               string                 `msgpack:"bindDN" json:"bindDN"`
	BindPassword         string                 `msgpack:"bindPassword" json:"bindPassword"`
	BindSearchFilter     string                 `msgpack:"bindSearchFilter" json:"bindSearchFilter"`
	SubjectKey           string                 `msgpack:"subjectKey" json:"subjectKey"`
	IgnoreKeys           map[string]interface{} `msgpack:"ignoredKeys" json:"ignoredKeys"`
	BaseDN               string                 `msgpack:"baseDN" json:"baseDN"`
	ConnSecurityProtocol string                 `msgpack:"connSecurityProtocol" json:"connSecurityProtocol"`
	Username             string                 `msgpack:"username" json:"username"`
	Password             string                 `msgpack:"password" json:"password"`
}

// NewLDAPInfo returns a new LDAPInfo, or an error
func NewLDAPInfo(metadata map[string]interface{}) (*LDAPInfo, error) {

	if metadata == nil {
		return nil, fmt.Errorf("you must provide at least metadata or defaultMetadata")
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
