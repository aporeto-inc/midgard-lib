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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLDAPUtils_LDAPInfo(t *testing.T) {

	Convey("Given I create a new LDAPInfo with invalid metadata", t, func() {

		i, err := NewLDAPInfo(nil)

		Convey("Then err should not be nil", func() {
			So(err, ShouldNotBeNil)
		})

		Convey("Then info should be nil", func() {
			So(i, ShouldBeNil)
		})
	})

	Convey("Given I create a new LDAPInfo with valid metadata", t, func() {

		i, err := NewLDAPInfo(map[string]interface{}{
			LDAPAddressKey:              "123:123",
			LDAPBindDNKey:               "cn=admin,dc=toto,dc=com",
			LDAPBindPasswordKey:         "toto",
			LDAPBindSearchFilterKey:     "uid={USERNAME}",
			LDAPSubjectKey:              "uid",
			LDAPIgnoredKeys:             []string{"comment"},
			LDAPConnSecurityProtocolKey: "TLS",
			LDAPUsernameKey:             "lskywalker",
			LDAPPasswordKey:             "secret",
			LDAPBaseDNKey:               "ou=zoupla,dc=toto,dc=com",
		})

		Convey("Then err should be nil", func() {
			So(err, ShouldBeNil)
		})

		Convey("Then info should not be nil", func() {
			So(i, ShouldNotBeNil)
		})

		Convey("Then info should be correct", func() {
			So(i.Address, ShouldEqual, "123:123")
			So(i.BindDN, ShouldEqual, "cn=admin,dc=toto,dc=com")
			So(i.BindPassword, ShouldEqual, "toto")
			So(i.BindSearchFilter, ShouldEqual, "uid={USERNAME}")
			So(i.SubjectKey, ShouldEqual, "uid")
			So(i.IgnoreKeys, ShouldContainKey, "comment")
			So(i.ConnSecurityProtocol, ShouldEqual, "TLS")
			So(i.Username, ShouldEqual, "lskywalker")
			So(i.Password, ShouldEqual, "secret")
			So(i.BaseDN, ShouldEqual, "ou=zoupla,dc=toto,dc=com")
		})
	})
}

func TestLDAPUtils_LDAPInfoMissingKeys(t *testing.T) {
	Convey("Given I create a new LDAPInfo with metadata and missing LDAPAddress", t, func() {

		i, err := NewLDAPInfo(map[string]interface{}{
			LDAPBindDNKey:               "cn=admin,dc=toto,dc=com",
			LDAPBindPasswordKey:         "toto",
			LDAPBindSearchFilterKey:     "uid={USERNAME}",
			LDAPSubjectKey:              "uid",
			LDAPIgnoredKeys:             []string{"comment"},
			LDAPConnSecurityProtocolKey: "TLS",
			LDAPUsernameKey:             "lskywalker",
			LDAPPasswordKey:             "secret",
			LDAPBaseDNKey:               "ou=zoupla,dc=toto,dc=com",
		})

		Convey("Then err should not be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "metadata must contain the key 'address'")
		})

		Convey("Then info should be nil", func() {
			So(i, ShouldBeNil)
		})
	})

	Convey("Given I create a new LDAPInfo with metadata and missing bindDN", t, func() {

		i, err := NewLDAPInfo(map[string]interface{}{
			LDAPAddressKey:              "123:123",
			LDAPBindPasswordKey:         "toto",
			LDAPBindSearchFilterKey:     "uid={USERNAME}",
			LDAPIgnoredKeys:             []string{"comment"},
			LDAPConnSecurityProtocolKey: "TLS",
			LDAPUsernameKey:             "lskywalker",
			LDAPPasswordKey:             "secret",
			LDAPBaseDNKey:               "ou=zoupla,dc=toto,dc=com",
		})

		Convey("Then err should not be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "metadata must contain the key 'bindDN'")
		})

		Convey("Then info should be nil", func() {
			So(i, ShouldBeNil)
		})
	})

	Convey("Given I create a new LDAPInfo with metadata and missing bindPassword", t, func() {

		i, err := NewLDAPInfo(map[string]interface{}{
			LDAPAddressKey:              "123:123",
			LDAPBindDNKey:               "cn=admin,dc=toto,dc=com",
			LDAPBindSearchFilterKey:     "uid={USERNAME}",
			LDAPSubjectKey:              "uid",
			LDAPIgnoredKeys:             []string{"comment"},
			LDAPConnSecurityProtocolKey: "TLS",
			LDAPUsernameKey:             "lskywalker",
			LDAPPasswordKey:             "secret",
			LDAPBaseDNKey:               "ou=zoupla,dc=toto,dc=com",
		})

		Convey("Then err should not be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "metadata must contain the key 'bindPassword'")
		})

		Convey("Then info should be nil", func() {
			So(i, ShouldBeNil)
		})
	})

	Convey("Given I create a new LDAPInfo with metadata and missing bindSearchFilter", t, func() {

		i, err := NewLDAPInfo(map[string]interface{}{
			LDAPAddressKey:              "123:123",
			LDAPBindPasswordKey:         "toto",
			LDAPBindDNKey:               "cn=admin,dc=toto,dc=com",
			LDAPConnSecurityProtocolKey: "TLS",
			LDAPIgnoredKeys:             []string{"comment"},
			LDAPUsernameKey:             "lskywalker",
			LDAPPasswordKey:             "secret",
			LDAPBaseDNKey:               "ou=zoupla,dc=toto,dc=com",
		})

		Convey("Then err should not be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "metadata must contain the key 'bindSearchFilter'")
		})

		Convey("Then info should be nil", func() {
			So(i, ShouldBeNil)
		})
	})

	Convey("Given I create a new LDAPInfo with metadata and missing subjectKey", t, func() {

		i, err := NewLDAPInfo(map[string]interface{}{
			LDAPAddressKey:              "123:123",
			LDAPBindPasswordKey:         "toto",
			LDAPBindDNKey:               "cn=admin,dc=toto,dc=com",
			LDAPBindSearchFilterKey:     "uid={USERNAME}",
			LDAPIgnoredKeys:             []string{"comment"},
			LDAPConnSecurityProtocolKey: "TLS",
			LDAPUsernameKey:             "lskywalker",
			LDAPPasswordKey:             "secret",
			LDAPBaseDNKey:               "ou=zoupla,dc=toto,dc=com",
		})

		Convey("Then err should not be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "metadata must contain the key 'subjectKey'")
		})

		Convey("Then info should be nil", func() {
			So(i, ShouldBeNil)
		})
	})

	Convey("Given I create a new LDAPInfo with metadata and missing ignoreKeys", t, func() {

		i, err := NewLDAPInfo(map[string]interface{}{
			LDAPAddressKey:              "123:123",
			LDAPBindPasswordKey:         "toto",
			LDAPBindDNKey:               "cn=admin,dc=toto,dc=com",
			LDAPBindSearchFilterKey:     "uid={USERNAME}",
			LDAPConnSecurityProtocolKey: "TLS",
			LDAPSubjectKey:              "uid",
			LDAPUsernameKey:             "lskywalker",
			LDAPPasswordKey:             "secret",
			LDAPBaseDNKey:               "ou=zoupla,dc=toto,dc=com",
		})

		Convey("Then err should not be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "metadata must contain the key 'ignoredKeys'")
		})

		Convey("Then info should be nil", func() {
			So(i, ShouldBeNil)
		})
	})

	Convey("Given I create a new LDAPInfo with metadata and missing connSecurityProtocol", t, func() {

		i, err := NewLDAPInfo(map[string]interface{}{
			LDAPAddressKey:          "123:123",
			LDAPBindPasswordKey:     "toto",
			LDAPBindDNKey:           "cn=admin,dc=toto,dc=com",
			LDAPBindSearchFilterKey: "uid={USERNAME}",
			LDAPSubjectKey:          "uid",
			LDAPIgnoredKeys:         []string{"comment"},
			LDAPUsernameKey:         "lskywalker",
			LDAPPasswordKey:         "secret",
			LDAPBaseDNKey:           "ou=zoupla,dc=toto,dc=com",
		})

		Convey("Then err should not be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "metadata must contain the key 'connSecurityProtocol'")
		})

		Convey("Then info should be nil", func() {
			So(i, ShouldBeNil)
		})
	})

	Convey("Given I create a new LDAPInfo with metadata and missing username", t, func() {

		i, err := NewLDAPInfo(map[string]interface{}{
			LDAPAddressKey:              "123:123",
			LDAPBindDNKey:               "cn=admin,dc=toto,dc=com",
			LDAPBindPasswordKey:         "toto",
			LDAPBindSearchFilterKey:     "uid={USERNAME}",
			LDAPSubjectKey:              "uid",
			LDAPIgnoredKeys:             []string{"comment"},
			LDAPConnSecurityProtocolKey: "TLS",
			LDAPPasswordKey:             "secret",
			LDAPBaseDNKey:               "ou=zoupla,dc=toto,dc=com",
		})

		Convey("Then err should not be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "metadata must contain the key 'username'")
		})

		Convey("Then info should be nil", func() {
			So(i, ShouldBeNil)
		})
	})

	Convey("Given I create a new LDAPInfo with metadata and missing password", t, func() {

		i, err := NewLDAPInfo(map[string]interface{}{
			LDAPAddressKey:              "123:123",
			LDAPBindDNKey:               "cn=admin,dc=toto,dc=com",
			LDAPBindPasswordKey:         "toto",
			LDAPBindSearchFilterKey:     "uid={USERNAME}",
			LDAPSubjectKey:              "uid",
			LDAPIgnoredKeys:             []string{"comment"},
			LDAPConnSecurityProtocolKey: "TLS",
			LDAPUsernameKey:             "lskywalker",
			LDAPBaseDNKey:               "ou=zoupla,dc=toto,dc=com",
		})

		Convey("Then err should not be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "metadata must contain the key 'password'")
		})

		Convey("Then info should be nil", func() {
			So(i, ShouldBeNil)
		})
	})

	Convey("Given I create a new LDAPInfo with metadata and missing baseDN", t, func() {

		i, err := NewLDAPInfo(map[string]interface{}{
			LDAPAddressKey:              "123:123",
			LDAPBindDNKey:               "cn=admin,dc=toto,dc=com",
			LDAPBindPasswordKey:         "toto",
			LDAPBindSearchFilterKey:     "uid={USERNAME}",
			LDAPSubjectKey:              "uid",
			LDAPIgnoredKeys:             []string{"comment"},
			LDAPConnSecurityProtocolKey: "TLS",
			LDAPUsernameKey:             "lskywalker",
			LDAPPasswordKey:             "secret",
		})

		Convey("Then err should not be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "metadata must contain the key 'baseDN'")
		})

		Convey("Then info should be nil", func() {
			So(i, ShouldBeNil)
		})
	})

	Convey("Given I create a new LDAPInfo nothing", t, func() {

		i, err := NewLDAPInfo(nil)

		Convey("Then err should not be be nil", func() {
			So(err, ShouldNotBeNil)
		})

		Convey("Then LDAPInfo should be nil", func() {
			So(i, ShouldBeNil)
		})
	})
}

func TestLDAPUtils_LDAPInfoBadValues(t *testing.T) {

	Convey("Given I create a new LDAPInfo with metadata and bad LDAPAddressKey", t, func() {

		i, err := NewLDAPInfo(map[string]interface{}{
			LDAPAddressKey:              123,
			LDAPBindPasswordKey:         "toto",
			LDAPBindDNKey:               "cn=admin,dc=toto,dc=com",
			LDAPBindSearchFilterKey:     "uid={USERNAME}",
			LDAPIgnoredKeys:             []string{},
			LDAPConnSecurityProtocolKey: "TLS",
			LDAPSubjectKey:              "uid",
			LDAPUsernameKey:             "lskywalker",
			LDAPPasswordKey:             "secret",
			LDAPBaseDNKey:               "ou=zoupla,dc=toto,dc=com",
		})

		Convey("Then err should not be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "metadata must be a string for key 'address'")
		})

		Convey("Then info should be nil", func() {
			So(i, ShouldBeNil)
		})
	})

	Convey("Given I create a new LDAPInfo with metadata and bad ignoreKeys", t, func() {

		i, err := NewLDAPInfo(map[string]interface{}{
			LDAPAddressKey:              "123:123",
			LDAPBindPasswordKey:         "toto",
			LDAPBindDNKey:               "cn=admin,dc=toto,dc=com",
			LDAPBindSearchFilterKey:     "uid={USERNAME}",
			LDAPIgnoredKeys:             "",
			LDAPConnSecurityProtocolKey: "TLS",
			LDAPSubjectKey:              "uid",
			LDAPUsernameKey:             "lskywalker",
			LDAPPasswordKey:             "secret",
			LDAPBaseDNKey:               "ou=zoupla,dc=toto,dc=com",
		})

		Convey("Then err should not be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "metadata must be a list of strings for key 'ignoredKeys'")
		})

		Convey("Then info should be nil", func() {
			So(i, ShouldBeNil)
		})
	})
}

func TestLDAPUtils_ToMap(t *testing.T) {

	Convey("Given I create a new LDAPInfo with valid metadata", t, func() {

		i, err := NewLDAPInfo(map[string]interface{}{
			LDAPAddressKey:              "123:123",
			LDAPBindDNKey:               "cn=admin,dc=toto,dc=com",
			LDAPBindPasswordKey:         "toto",
			LDAPBindSearchFilterKey:     "uid={USERNAME}",
			LDAPSubjectKey:              "uid",
			LDAPIgnoredKeys:             []string{"comment"},
			LDAPConnSecurityProtocolKey: "TLS",
			LDAPUsernameKey:             "lskywalker",
			LDAPPasswordKey:             "secret",
			LDAPBaseDNKey:               "ou=zoupla,dc=toto,dc=com",
		})

		Convey("Then err should be nil", func() {
			So(err, ShouldBeNil)
		})

		Convey("Then info should not be nil", func() {
			So(i, ShouldNotBeNil)
		})

		m := i.ToMap()

		Convey("Then map should not be nil", func() {
			So(m, ShouldNotBeNil)
		})

		Convey("Then map should have all keys", func() {
			temp, ok := m[LDAPAddressKey].(string)
			So(ok, ShouldEqual, true)
			So(temp, ShouldEqual, "123:123")
			temp, ok = m[LDAPBindDNKey].(string)
			So(ok, ShouldEqual, true)
			So(temp, ShouldEqual, "cn=admin,dc=toto,dc=com")
			temp, ok = m[LDAPBindPasswordKey].(string)
			So(ok, ShouldEqual, true)
			So(temp, ShouldEqual, "toto")
			temp, ok = m[LDAPBindSearchFilterKey].(string)
			So(ok, ShouldEqual, true)
			So(temp, ShouldEqual, "uid={USERNAME}")
			tempMap, ok := m[LDAPIgnoredKeys].(map[string]interface{})
			So(ok, ShouldEqual, true)
			So(tempMap, ShouldContainKey, "comment")
			temp, ok = m[LDAPConnSecurityProtocolKey].(string)
			So(ok, ShouldEqual, true)
			So(temp, ShouldEqual, "TLS")
			temp, ok = m[LDAPUsernameKey].(string)
			So(ok, ShouldEqual, true)
			So(temp, ShouldEqual, "lskywalker")
			temp, ok = m[LDAPPasswordKey].(string)
			So(ok, ShouldEqual, true)
			So(temp, ShouldEqual, "secret")
			temp, ok = m[LDAPBaseDNKey].(string)
			So(ok, ShouldEqual, true)
			So(temp, ShouldEqual, "ou=zoupla,dc=toto,dc=com")
		})
	})
}

func TestLDAPUtils_GetUserQueryString(t *testing.T) {

	Convey("Given I create a new LDAPInfo with valid metadata", t, func() {

		i, err := NewLDAPInfo(map[string]interface{}{
			LDAPAddressKey:              "123:123",
			LDAPBindDNKey:               "cn=admin,dc=toto,dc=com",
			LDAPBindPasswordKey:         "toto",
			LDAPBindSearchFilterKey:     "uid={USERNAME}",
			LDAPSubjectKey:              "uid",
			LDAPIgnoredKeys:             []string{"comment"},
			LDAPConnSecurityProtocolKey: "TLS",
			LDAPUsernameKey:             "lskywalker",
			LDAPPasswordKey:             "secret",
			LDAPBaseDNKey:               "ou=zoupla,dc=toto,dc=com",
		})

		Convey("Then err should be nil", func() {
			So(err, ShouldBeNil)
		})

		Convey("Then info should not be nil", func() {
			So(i, ShouldNotBeNil)
		})

		Convey("Then info should be correct", func() {
			So(i.GetUserQueryString(), ShouldEqual, "uid=lskywalker")
		})
	})

	Convey("Given I create a new LDAPInfo with valid metadata", t, func() {

		i, err := NewLDAPInfo(map[string]interface{}{
			LDAPAddressKey:              "123:123",
			LDAPBindDNKey:               "cn=admin,dc=toto,dc=com",
			LDAPBindPasswordKey:         "toto",
			LDAPBindSearchFilterKey:     "uid={USERNAME},khg={USERNAME}",
			LDAPSubjectKey:              "uid",
			LDAPIgnoredKeys:             []string{"comment"},
			LDAPConnSecurityProtocolKey: "TLS",
			LDAPUsernameKey:             "lskywalker",
			LDAPPasswordKey:             "secret",
			LDAPBaseDNKey:               "ou=zoupla,dc=toto,dc=com",
		})

		Convey("Then err should be nil", func() {
			So(err, ShouldBeNil)
		})

		Convey("Then info should not be nil", func() {
			So(i, ShouldNotBeNil)
		})

		Convey("Then info should be correct", func() {
			So(i.GetUserQueryString(), ShouldEqual, "uid=lskywalker,khg=lskywalker")
		})
	})
}
