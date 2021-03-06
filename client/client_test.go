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

package midgardclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/gaia"
	"go.aporeto.io/midgard-lib/ldaputils"
)

func TestClient_NewClient(t *testing.T) {

	Convey("Given I create a new Client with a valid URL", t, func() {

		cl := NewClient("http://com.com")

		Convey("Then client should be correctly initialized", func() {
			So(cl, ShouldNotBeNil)
		})

		Convey("Then client url should be set", func() {
			So(cl.url, ShouldEqual, "http://com.com")
		})
	})

	Convey("Given I create a new Client with a missing URL", t, func() {

		Convey("Then it should panic", func() {
			So(func() { NewClient("") }, ShouldPanic)
		})
	})
}

func TestClient_Authentify(t *testing.T) {

	Convey("Given I have a Client and some valid http header", t, func() {

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, `{
                "claims": {
                   "aud": "aporeto.com",
                   "data": {
                       "commonName": "superadmin",
                       "organization": "aporeto.com",
                       "organizationalUnit": "SuperAdmin"
                   },
                   "exp": 1475083201,
                   "iat": 1474996801,
                   "iss": "midgard.aporeto.com",
                   "realm": "certificate",
                   "sub": "10237207344299343489"
               }
            }`)
		}))
		defer ts.Close()

		cl := NewClient(ts.URL)

		Convey("When I call Authentify", func() {

			n, err := cl.Authentify(context.Background(), "thetoken")

			Convey("Then I should get valid normalization", func() {
				So(n, ShouldContain, "@auth:subject=10237207344299343489")
			})

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given I have a Client and some valid http header but Midgard doesn't respond", t, func() {

		cl := NewClient("http://sdfjdfjkshfjkhdskfhsdjkfhsdkfhsdkjfhsdjjshsjkgdsg.gsdjghdjgfdfjghdhfgdfjhg.dfgj")

		Convey("When I call Authentify", func() {

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			n, err := cl.Authentify(ctx, "thetoken")

			Convey("Then normalization should be nil", func() {
				So(n, ShouldBeNil)
			})

			Convey("Then err should be not nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given I have a Client and some valid http header but Midgard doesn't approve", t, func() {

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(403)
			fmt.Fprintln(w, `{
                "claims": null
            }`)
		}))
		defer ts.Close()

		cl := NewClient(ts.URL)

		Convey("When I call Authentify", func() {

			n, err := cl.Authentify(context.Background(), "thetoken")

			Convey("Then normalization should be nil", func() {
				So(n, ShouldBeNil)
			})

			Convey("Then err should be not nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given I have a Client and some valid http header but Midgard return garbage json", t, func() {

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, `{
                "claims": null
            }`)
		}))
		defer ts.Close()

		cl := NewClient(ts.URL)

		Convey("When I call Authentify", func() {

			n, err := cl.Authentify(context.Background(), "thetoken")

			Convey("Then normalization should be nil", func() {
				So(n, ShouldBeNil)
			})

			Convey("Then err should be not nil", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "error 401 (midgard-lib): Unauthorized: No claims returned. Token may be invalid")
			})
		})
	})

	Convey("Given I have a Client and some valid http header but Midgard return no claims", t, func() {

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, `{
                "claims
            }`)
		}))
		defer ts.Close()

		cl := NewClient(ts.URL)

		Convey("When I call Authentify", func() {

			n, err := cl.Authentify(context.Background(), "thetoken")

			Convey("Then normalization should be nil", func() {
				So(n, ShouldBeNil)
			})

			Convey("Then err should be not nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestClient_IssueFromGoogle(t *testing.T) {

	Convey("Given I have a client and a fake working server", t, func() {

		expectedRequest := gaia.NewIssue()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(expectedRequest); err != nil {
				panic(err)
			}
			fmt.Fprintln(w, `{"data": "","realm": "google","token": "yeay!"}`)
		}))
		defer ts.Close()

		cl := NewClient(ts.URL)

		Convey("When I call IssueFromGoogle", func() {

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			token, err := cl.IssueFromGoogle(ctx, "token", 1*time.Minute, OptQuota(1))

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the issue request should be correct", func() {
				So(expectedRequest.Realm, ShouldEqual, "Google")
				So(expectedRequest.Data, ShouldEqual, "token")
			})

			Convey("Then token should be correct", func() {
				So(token, ShouldEqual, "yeay!")
			})
		})
	})
}

func TestClient_IssueFromCertificate(t *testing.T) {

	Convey("Given I have a client and a fake working server", t, func() {

		expectedRequest := gaia.NewIssue()
		var expectedCert *x509.Certificate

		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if err := json.NewDecoder(r.Body).Decode(expectedRequest); err != nil {
				panic(err)
			}

			expectedCert = r.TLS.PeerCertificates[0]

			fmt.Fprintln(w, `{
                "data": "",
                "realm": "google",
                "token": "yeay!"
            }`)
		}))
		defer ts.Close()

		ts.TLS.ClientAuth = tls.RequireAnyClientCert

		cert, _ := tls.LoadX509KeyPair("./fixtures/client-cert.pem", "./fixtures/client-key.pem")

		cl := NewClientWithTLS(ts.URL, &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		})

		Convey("When I call IssueFromCertificate", func() {

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			token, err := cl.IssueFromCertificate(ctx, 1*time.Minute, OptQuota(1))

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the cert should have been sent", func() {
				So(expectedCert.SerialNumber.String(), ShouldEqual, "135383296740973442198818964228093856486")
			})

			Convey("Then the issue request should be correct", func() {
				So(expectedRequest.Realm, ShouldEqual, "Certificate")
			})

			Convey("Then token should be correct", func() {
				So(token, ShouldEqual, "yeay!")
			})
		})
	})
}

func TestClient_IssueFromLDAP(t *testing.T) {

	Convey("Given I have a client and a fake working server", t, func() {

		expectedRequest := gaia.NewIssue()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if err := json.NewDecoder(r.Body).Decode(expectedRequest); err != nil {
				panic(err)
			}

			fmt.Fprintln(w, `{
                "data": "",
                "realm": "google",
                "token": "yeay!"
            }`)
		}))
		defer ts.Close()

		cl := NewClient(ts.URL)

		Convey("When I call IssueFromCertificate", func() {

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			linfo := &ldaputils.LDAPInfo{
				Address:      "Address",
				BindDN:       "BindDN",
				BindPassword: "BindPassword",
				BaseDN:       "BaseDN",
				Username:     "Username",
				Password:     "Password",
			}

			token, err := cl.IssueFromLDAP(ctx, linfo, "namespace", "provider", 1*time.Minute, OptQuota(1))

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the issue request should be correct", func() {
				So(expectedRequest.Realm, ShouldEqual, "LDAP")
				So(expectedRequest.Metadata["namespace"], ShouldEqual, "namespace")
				So(expectedRequest.Metadata["provider"], ShouldEqual, "provider")
				So(expectedRequest.Metadata["address"], ShouldEqual, "Address")
				So(expectedRequest.Metadata["bindDN"], ShouldEqual, "BindDN")
				So(expectedRequest.Metadata["bindPassword"], ShouldEqual, "BindPassword")
				So(expectedRequest.Metadata["baseDN"], ShouldEqual, "BaseDN")
				So(expectedRequest.Metadata["username"], ShouldEqual, "Username")
				So(expectedRequest.Metadata["password"], ShouldEqual, "Password")
			})

			Convey("Then token should be correct", func() {
				So(token, ShouldEqual, "yeay!")
			})
		})
	})
}

func TestClient_IssueFromVince(t *testing.T) {

	Convey("Given I have a client and a fake working server", t, func() {

		expectedRequest := gaia.NewIssue()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(expectedRequest); err != nil {
				panic(err)
			}
			fmt.Fprintln(w, `{"data": "","realm": "google","token": "yeay!"}`)
		}))
		defer ts.Close()

		cl := NewClient(ts.URL)

		Convey("When I call IssueFromVince", func() {

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			token, err := cl.IssueFromVince(ctx, "account", "password", "otp", 1*time.Minute,
				OptQuota(1),
				OptRestrictNamespace("/ns1"),
				OptRestrictPermissions([]string{"@auth:role=toto"}),
				OptRestrictNetworks([]string{"127.0.0.0/8"}),
			)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the issue request should be correct", func() {
				So(expectedRequest.Realm, ShouldEqual, "Vince")
				So(expectedRequest.Metadata["vinceAccount"], ShouldEqual, "account")
				So(expectedRequest.Metadata["vincePassword"], ShouldEqual, "password")
				So(expectedRequest.Metadata["vinceOTP"], ShouldEqual, "otp")
				So(expectedRequest.RestrictedPermissions, ShouldResemble, []string{"@auth:role=toto"})
				So(expectedRequest.RestrictedNamespace, ShouldEqual, "/ns1")
				So(expectedRequest.RestrictedNetworks, ShouldResemble, []string{"127.0.0.0/8"})
			})

			Convey("Then token should be correct", func() {
				So(token, ShouldEqual, "yeay!")
			})
		})
	})
}

func TestClient_IssueFromAporetoIdentityToken(t *testing.T) {

	Convey("Given I have a client and a fake working server", t, func() {

		expectedRequest := gaia.NewIssue()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(expectedRequest); err != nil {
				panic(err)
			}
			fmt.Fprintln(w, `{"data": "","realm": "aporetoidentitytoken","token": "yeay!"}`)
		}))
		defer ts.Close()

		cl := NewClient(ts.URL)

		Convey("When I call IssueFromAporetoIdentityToken", func() {

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			token, err := cl.IssueFromAporetoIdentityToken(ctx, "in-token", 1*time.Minute,
				OptQuota(1),
				OptRestrictNamespace("/ns1"),
				OptRestrictPermissions([]string{"@auth:role=toto"}),
				OptRestrictNetworks([]string{"127.0.0.0/8"}),
			)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the issue request should be correct", func() {
				So(expectedRequest.Realm, ShouldEqual, "AporetoIdentityToken")
				So(expectedRequest.Metadata["token"], ShouldEqual, "in-token")
				So(expectedRequest.RestrictedPermissions, ShouldResemble, []string{"@auth:role=toto"})
				So(expectedRequest.RestrictedNamespace, ShouldEqual, "/ns1")
				So(expectedRequest.RestrictedNetworks, ShouldResemble, []string{"127.0.0.0/8"})
			})

			Convey("Then token should be correct", func() {
				So(token, ShouldEqual, "yeay!")
			})
		})
	})
}

func TestClient_IssueFromGCPIdentityToken(t *testing.T) {

	Convey("Given I have a client and a fake working server", t, func() {

		expectedRequest := gaia.NewIssue()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(expectedRequest); err != nil {
				panic(err)
			}
			fmt.Fprintln(w, `{"data": "","realm": "google","token": "yeay!"}`)
		}))
		defer ts.Close()

		cl := NewClient(ts.URL)

		Convey("When I call IssueFromGCPIdentityToken", func() {

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			token, err := cl.IssueFromGCPIdentityToken(ctx, "doc", 1*time.Minute,
				OptQuota(1),
				OptRestrictNamespace("/ns1"),
				OptRestrictPermissions([]string{"@auth:role=toto"}),
				OptRestrictNetworks([]string{"127.0.0.0/8"}),
			)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the issue request should be correct", func() {
				So(expectedRequest.Realm, ShouldEqual, "GCPIdentityToken")
				So(expectedRequest.Metadata["token"], ShouldEqual, "doc")
				So(expectedRequest.RestrictedPermissions, ShouldResemble, []string{"@auth:role=toto"})
				So(expectedRequest.RestrictedNamespace, ShouldEqual, "/ns1")
				So(expectedRequest.RestrictedNetworks, ShouldResemble, []string{"127.0.0.0/8"})
			})

			Convey("Then token should be correct", func() {
				So(token, ShouldEqual, "yeay!")
			})
		})
	})
}

func TestClient_IssueFromAzureIdentityToken(t *testing.T) {

	Convey("Given I have a client and a fake working server", t, func() {

		expectedRequest := gaia.NewIssue()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(expectedRequest); err != nil {
				panic(err)
			}
			fmt.Fprintln(w, `{"data": "","realm": "google","token": "yeay!"}`)
		}))
		defer ts.Close()

		cl := NewClient(ts.URL)

		Convey("When I call IssueFromGCPIdentityToken", func() {

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			token, err := cl.IssueFromAzureIdentityToken(ctx, "doc", 1*time.Minute,
				OptQuota(1),
				OptRestrictNamespace("/ns1"),
				OptRestrictPermissions([]string{"@auth:role=toto"}),
				OptRestrictNetworks([]string{"127.0.0.0/8"}),
			)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the issue request should be correct", func() {
				So(expectedRequest.Realm, ShouldEqual, "AzureIdentityToken")
				So(expectedRequest.Metadata["token"], ShouldEqual, "doc")
				So(expectedRequest.RestrictedPermissions, ShouldResemble, []string{"@auth:role=toto"})
				So(expectedRequest.RestrictedNamespace, ShouldEqual, "/ns1")
				So(expectedRequest.RestrictedNetworks, ShouldResemble, []string{"127.0.0.0/8"})
			})

			Convey("Then token should be correct", func() {
				So(token, ShouldEqual, "yeay!")
			})
		})
	})
}

func TestClient_IssueFromOIDCStep1(t *testing.T) {

	Convey("Given I have a client and a fake working server", t, func() {

		expectedRequest := gaia.NewIssue()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(expectedRequest); err != nil {
				panic(err)
			}

			w.Header().Set("Location", "http://laba")
			w.WriteHeader(http.StatusFound)

		}))
		defer ts.Close()

		cl := NewClient(ts.URL)

		Convey("When I call IssueFromOIDCStep1(", func() {

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			url, err := cl.IssueFromOIDCStep1(ctx, "aporeto", "okta", "http://ici")

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the issue request should be correct", func() {
				So(expectedRequest.Realm, ShouldEqual, "OIDC")
			})

			Convey("Then url should be correct", func() {
				So(url, ShouldEqual, "http://laba")
			})
		})
	})
}

func TestClient_IssueFromOIDCStep2(t *testing.T) {

	Convey("Given I have a client and a fake working server", t, func() {

		expectedRequest := gaia.NewIssue()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(expectedRequest); err != nil {
				panic(err)
			}

			fmt.Fprintln(w, `{"data": "","realm": "oidc","token": "token"}`)

		}))
		defer ts.Close()

		cl := NewClient(ts.URL)

		Convey("When I call IssueFromOIDCStep2", func() {

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			token, err := cl.IssueFromOIDCStep2(ctx, "code", "state", 1*time.Minute,
				OptQuota(1),
				OptRestrictNamespace("/ns1"),
				OptRestrictPermissions([]string{"@auth:role=toto"}),
				OptRestrictNetworks([]string{"127.0.0.0/8"}),
			)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the issue request should be correct", func() {
				So(expectedRequest.Realm, ShouldEqual, "OIDC")
				So(expectedRequest.Metadata["code"], ShouldEqual, "code")
				So(expectedRequest.Metadata["state"], ShouldEqual, "state")
				So(expectedRequest.RestrictedPermissions, ShouldResemble, []string{"@auth:role=toto"})
				So(expectedRequest.RestrictedNamespace, ShouldEqual, "/ns1")
				So(expectedRequest.RestrictedNetworks, ShouldResemble, []string{"127.0.0.0/8"})
			})

			Convey("Then token should be correct", func() {
				So(token, ShouldEqual, "token")
			})
		})
	})
}

func TestClient_IssueFromSAMLStep1(t *testing.T) {

	Convey("Given I have a client and a fake working server", t, func() {

		expectedRequest := gaia.NewIssue()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(expectedRequest); err != nil {
				panic(err)
			}

			w.Header().Set("Location", "http://laba")
			w.WriteHeader(http.StatusFound)

		}))
		defer ts.Close()

		cl := NewClient(ts.URL)

		Convey("When I call IssueFromSAMLStep1(", func() {

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			url, err := cl.IssueFromSAMLStep1(ctx, "aporeto", "okta", "http://ici")

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the issue request should be correct", func() {
				So(expectedRequest.Realm, ShouldEqual, "SAML")
			})

			Convey("Then url should be correct", func() {
				So(url, ShouldEqual, "http://laba")
			})
		})
	})
}

func TestClient_IssueFromSAMLStep2(t *testing.T) {

	Convey("Given I have a client and a fake working server", t, func() {

		expectedRequest := gaia.NewIssue()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(expectedRequest); err != nil {
				panic(err)
			}

			fmt.Fprintln(w, `{"data": "","realm": "saml","token": "token"}`)

		}))
		defer ts.Close()

		cl := NewClient(ts.URL)

		Convey("When I call IssueFromSAMLStep2", func() {

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			token, err := cl.IssueFromSAMLStep2(ctx, "response", "state", 1*time.Minute,
				OptQuota(1),
				OptRestrictNamespace("/ns1"),
				OptRestrictPermissions([]string{"@auth:role=toto"}),
				OptRestrictNetworks([]string{"127.0.0.0/8"}),
			)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the issue request should be correct", func() {
				So(expectedRequest.Realm, ShouldEqual, "SAML")
				So(expectedRequest.Metadata["SAMLResponse"], ShouldEqual, "response")
				So(expectedRequest.Metadata["relayState"], ShouldEqual, "state")
				So(expectedRequest.RestrictedPermissions, ShouldResemble, []string{"@auth:role=toto"})
				So(expectedRequest.RestrictedNamespace, ShouldEqual, "/ns1")
				So(expectedRequest.RestrictedNetworks, ShouldResemble, []string{"127.0.0.0/8"})
			})

			Convey("Then token should be correct", func() {
				So(token, ShouldEqual, "token")
			})
		})
	})
}

func TestClient_IssueFromAWSSecurityToken(t *testing.T) {

	Convey("Given I have a fake working server", t, func() {

		expectedRequest := gaia.NewIssue()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if err := json.NewDecoder(r.Body).Decode(expectedRequest); err != nil {
				panic(err)
			}

			fmt.Fprintln(w, `{
                "data": "",
                "realm": "sts",
                "token": "yeay!"
            }`)
		}))
		defer ts.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		cl := NewClient(ts.URL)

		Convey("When I call IssueFromAWSSecurityToken with valid info", func() {

			_, err := cl.IssueFromAWSSecurityToken(ctx, "x", "y", "z", 1*time.Second,
				OptQuota(1),
				OptRestrictNamespace("/ns1"),
				OptRestrictPermissions([]string{"@auth:role=toto"}),
				OptRestrictNetworks([]string{"127.0.0.0/8"}),
			)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the request should be correct", func() {
				So(expectedRequest.Realm, ShouldEqual, gaia.IssueRealmAWSSecurityToken)
				So(expectedRequest.Metadata["accessKeyID"], ShouldEqual, "x")
				So(expectedRequest.Metadata["secretAccessKey"], ShouldEqual, "y")
				So(expectedRequest.Metadata["token"], ShouldEqual, "z")
				So(expectedRequest.RestrictedPermissions, ShouldResemble, []string{"@auth:role=toto"})
				So(expectedRequest.RestrictedNamespace, ShouldEqual, "/ns1")
				So(expectedRequest.RestrictedNetworks, ShouldResemble, []string{"127.0.0.0/8"})
			})
		})
	})
}

func TestClient_sendRequest(t *testing.T) {

	Convey("Given I have a client and a fake working server", t, func() {

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, `{
                "data": "",
                "realm": "google",
                "token": "yeay!"
            }`)
		}))
		defer ts.Close()

		cl := NewClient(ts.URL)

		Convey("When I call sendRequest with a valid token", func() {

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			jwt, err := cl.sendRequest(ctx, &gaia.Issue{Realm: "test"})

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then jwt be correct", func() {
				So(jwt, ShouldEqual, "yeay!")
			})
		})
	})

	Convey("Given I have a client with an invalid URL", t, func() {

		cl := NewClient("http:/ssaffsdf")

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		jwt, err := cl.sendRequest(ctx, &gaia.Issue{Realm: "test"})

		Convey("Then err should not be nil", func() {
			So(err, ShouldNotBeNil)
		})

		Convey("Then jwt be empty", func() {
			So(jwt, ShouldBeEmpty)
		})
	})

	Convey("Given I have a client and a working fake server ", t, func() {

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(403)
			fmt.Fprintln(w, `{
                "data": "",
                "realm": "google",
                "token": "yeay!"
            }`)
		}))
		defer ts.Close()

		cl := NewClient(ts.URL)

		Convey("When I call IssueFromGoogle with an invalid token", func() {

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			jwt, err := cl.sendRequest(ctx, &gaia.Issue{Realm: "test"})

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then jwt be empty", func() {
				So(jwt, ShouldBeEmpty)
			})
		})
	})

	Convey("Given I have a client and a fake server that returns garbage", t, func() {

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, `{
                "data": "
            }`)
		}))
		defer ts.Close()

		cl := NewClient(ts.URL)

		Convey("When I call IssueFromGoogle with a valid token", func() {

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			jwt, err := cl.sendRequest(ctx, &gaia.Issue{Realm: "test"})

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then jwt be empty", func() {
				So(jwt, ShouldBeEmpty)
			})
		})
	})
}

func TestTokenUtils_Snip(t *testing.T) {

	Convey("Given have a token and and error containing the token", t, func() {

		token := "token"
		err := errors.New("your token is token")

		Convey("When I call snipToken", func() {

			e := snipToken(err, token)

			Convey("Then err should have the reference to token snipped", func() {
				So(e.Error(), ShouldEqual, "your [snip] is [snip]")
			})
		})
	})

	Convey("Given have a token and and error that doesn't contain the token", t, func() {

		token := "token"
		err := errors.New("your secret is secret")

		Convey("When I call snipToken", func() {

			e := snipToken(err, token)

			Convey("Then err should have the reference to token snipped", func() {
				So(e.Error(), ShouldEqual, "your secret is secret")
			})
		})
	})

	Convey("Given I have a token and a nil error", t, func() {

		token := "token"

		Convey("When I call snipToken", func() {

			e := snipToken(nil, token)

			Convey("Then err should be nil", func() {
				So(e, ShouldBeNil)
			})
		})
	})
}

func TestClient_IssueFromPCIdentityToken(t *testing.T) {

	Convey("Given I have a client and a fake working server", t, func() {

		expectedRequest := gaia.NewIssue()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(expectedRequest); err != nil {
				panic(err)
			}
			fmt.Fprintln(w, `{"data": "","realm": "PCIdentityToken","token": "yeay!"}`)
		}))
		defer ts.Close()

		cl := NewClient(ts.URL)

		Convey("When I call IssueFromPCCToken", func() {

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			token, err := cl.IssueFromPCIdentityToken(ctx, "token", 1*time.Minute,
				OptQuota(1),
				OptRestrictNamespace("/ns1"),
				OptRestrictPermissions([]string{"@auth:role=toto"}),
				OptRestrictNetworks([]string{"127.0.0.0/8"}),
			)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the issue request should be correct", func() {
				So(expectedRequest.Realm, ShouldEqual, "PCIdentityToken")
				So(expectedRequest.Metadata["token"], ShouldEqual, "token")
				So(expectedRequest.RestrictedPermissions, ShouldResemble, []string{"@auth:role=toto"})
				So(expectedRequest.RestrictedNamespace, ShouldEqual, "/ns1")
				So(expectedRequest.RestrictedNetworks, ShouldResemble, []string{"127.0.0.0/8"})
			})

			Convey("Then token should be correct", func() {
				So(token, ShouldEqual, "yeay!")
			})
		})
	})
}
