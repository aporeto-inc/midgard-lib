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
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"reflect"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/gaia"
)

func TestUtils_extractJWT(t *testing.T) {

	Convey("Given I have some http Header", t, func() {

		h := http.Header{}

		Convey("When I extract the token of a valid Authorization header", func() {

			h.Add("Authorization", "Bearer thetoken")
			token, err := ExtractJWTFromHeader(h)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then token should be thetoken", func() {
				So(token, ShouldEqual, "thetoken")
			})
		})

		Convey("When I extract the token of a missing Authorization header", func() {

			token, err := ExtractJWTFromHeader(h)

			Convey("Then err should be not nil", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then err.Error should be correct", func() {
				So(err.Error(), ShouldEqual, "missing authorization header")
			})

			Convey("Then token should be empty", func() {
				So(token, ShouldBeEmpty)
			})
		})

		Convey("When I extract the token of a malformed Authorization header", func() {

			h.Add("Authorization", "Bearer")
			token, err := ExtractJWTFromHeader(h)

			Convey("Then err should be not nil", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then err.Error should be correct", func() {
				So(err.Error(), ShouldEqual, "invalid authorization header")
			})

			Convey("Then token should be empty", func() {
				So(token, ShouldBeEmpty)
			})
		})

		Convey("When I extract the token of a invalid type Authorization header", func() {

			h.Add("Authorization", "NotBeaer thetoken")
			token, err := ExtractJWTFromHeader(h)

			Convey("Then err should be not nil", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then err.Error should be correct", func() {
				So(err.Error(), ShouldEqual, "invalid authorization header")
			})

			Convey("Then token should be empty", func() {
				So(token, ShouldBeEmpty)
			})
		})
	})
}

func TestUtils_NormalizeAuth(t *testing.T) {

	Convey("Given I have a Auth object", t, func() {

		auth := gaia.NewAuthn()
		auth.Claims.Realm = "realm"
		auth.Claims.Subject = "subject"
		auth.Claims.Data["d1"] = "v1"
		auth.Claims.Data["d2"] = "v2"
		auth.Claims.Data["subject"] = "subject"

		Convey("When I normalize it", func() {

			v := NormalizeAuth(auth.Claims)

			Convey("Then the subject should be correct", func() {
				So(v, ShouldContain, "@auth:subject=subject")
			})

			Convey("Then the d1 should be correct", func() {
				So(v, ShouldContain, "@auth:d1=v1")
			})

			Convey("Then the d2 should be correct", func() {
				So(v, ShouldContain, "@auth:d2=v2")
			})
		})

		Convey("When I normalize nil claims", func() {

			v := NormalizeAuth(nil)

			Convey("Then the subject should be correct", func() {
				So(len(v), ShouldEqual, 0)
			})
		})
	})
}

func TestUtils_AppCredsToTLSConfig(t *testing.T) {

	Convey("Given I have some valid appcred", t, func() {

		credsData := `{"certificate":"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJ2ekNDQVdXZ0F3SUJBZ0lRRGhjK0E2elNqUGlLbjQxZm82Z045REFLQmdncWhrak9QUVFEQWpCR01SQXcKRGdZRFZRUUtFd2RCY0c5eVpYUnZNUTh3RFFZRFZRUUxFd1poY0c5dGRYZ3hJVEFmQmdOVkJBTVRHRUZ3YjIxMQplQ0JRZFdKc2FXTWdVMmxuYm1sdVp5QkRRVEFlRncweE9ERXdNVFl4T1RVMk1qWmFGdzB4T1RFd01UWXlNRFUyCk1qWmFNRVl4Q2pBSUJnTlZCQW9UQVM4eE9EQTJCZ05WQkFNVEwyRndjRHBqY21Wa1pXNTBhV0ZzT2pWaVl6WTEKTURaaU4yUmtaakZtTnpVNE0yWmpZek5pTVRwMFpYTjBZWEJ3TUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowRApBUWNEUWdBRXBuZ0g2K2hIcXBpQ1ZHb1h0N2dWWXp6ZlJCSE92YVBtcU5LNHhNWHRUVjlzTUl4S0lwZDNBdlBOCko1amVlUkJGOFNOaTRzSHhSSDlCSjMzYjdMVnp6YU0xTURNd0RnWURWUjBQQVFIL0JBUURBZ1dnTUJNR0ExVWQKSlFRTU1Bb0dDQ3NHQVFVRkJ3TUNNQXdHQTFVZEV3RUIvd1FDTUFBd0NnWUlLb1pJemowRUF3SURTQUF3UlFJZwpPNDRQSS9TaG01bGxQUHRKbGllak0rdkN6WmowMk9QNEhWQTZEVllCdmpvQ0lRQ2pnUEw0WXZKYmRyTENUOE9hCmlLSGFGOWk2RjNPTjQ3dzRUMGtYV0ZLcUZ3PT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=","certificateAuthority":"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJyRENDQVZLZ0F3SUJBZ0lSQUtjMERhOUVRSHB4aGxickNvTmZ2T1F3Q2dZSUtvWkl6ajBFQXdJd1JERVEKTUE0R0ExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNUjh3SFFZRFZRUURFeFpCY0c5dApkWGdnU1c1MFpYSnRaV1JwWVhSbElFTkJNQjRYRFRFNE1EWXlNREl4TURNME1Gb1hEVEk0TURReU9ESXhNRE0wCk1Gb3dSakVRTUE0R0ExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNU0V3SHdZRFZRUUQKRXhoQmNHOXRkWGdnVUhWaWJHbGpJRk5wWjI1cGJtY2dRMEV3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQgpCd05DQUFUSlExeVRDVEpzQUx0N25UbjBZRVNpSGgvZ0xlWlBDWlBhb09nWEJIdU5icEltUTF5Z0xPb2wvMUc1CmZ3VzdJNVJTdXZqNCtwV0Nad3pTbmxRaFIwZ0tveU13SVRBT0JnTlZIUThCQWY4RUJBTUNBUVl3RHdZRFZSMFQKQVFIL0JBVXdBd0VCL3pBS0JnZ3Foa2pPUFFRREFnTklBREJGQWlCSlNJNlRjQTdTODhnWmhXb29oeXYxK0FxNQpuY0dybXN1SG9NdUN3WEJUelFJaEFNeVRaMW5lZFEwelQ1SkVIQTJoaFRmUjFCT01zQS9Ic3AwNWpPa1BJbVpnCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJvVENDQVVlZ0F3SUJBZ0lRU2VKS3pXNjV4elFhZzlBeEhPVGR2REFLQmdncWhrak9QUVFEQWpBOE1SQXcKRGdZRFZRUUtFd2RCY0c5eVpYUnZNUTh3RFFZRFZRUUxFd1poY0c5dGRYZ3hGekFWQmdOVkJBTVREa0Z3YjIxMQplQ0JTYjI5MElFTkJNQjRYRFRFNE1EWXlNREl4TURNME1Gb1hEVEk0TURReU9ESXhNRE0wTUZvd1JERVFNQTRHCkExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNUjh3SFFZRFZRUURFeFpCY0c5dGRYZ2cKU1c1MFpYSnRaV1JwWVhSbElFTkJNRmt3RXdZSEtvWkl6ajBDQVFZSUtvWkl6ajBEQVFjRFFnQUUvNXRrN3pSdgpNWDVuZ1l6dkhNUEh1ZXVOc2dkU1pWMzRkZk4va3UyakxjZUwrNi9FNUViQWpHdWYrY3RLT3dRamNha09oajE0Cllrb1dHL0svNzYvZzg2TWpNQ0V3RGdZRFZSMFBBUUgvQkFRREFnRUdNQThHQTFVZEV3RUIvd1FGTUFNQkFmOHcKQ2dZSUtvWkl6ajBFQXdJRFNBQXdSUUloQU5aT3ZUVDhicHp1Vk1FY2xORzBsaFlCdmt3L0dXYjFZVWxNTFJCeApHYjNFQWlCL3RCQTlPN1AyZXdQaU9hclhNb2FzZFVjNU83Ukk2QThUdTczQ28vamtmdz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJtVENDQVQrZ0F3SUJBZ0lRYVJId3B6NWw5blo2eEoyRVIwdkNHakFLQmdncWhrak9QUVFEQWpBOE1SQXcKRGdZRFZRUUtFd2RCY0c5eVpYUnZNUTh3RFFZRFZRUUxFd1poY0c5dGRYZ3hGekFWQmdOVkJBTVREa0Z3YjIxMQplQ0JTYjI5MElFTkJNQjRYRFRFNE1EWXlNREl4TURNME1Gb1hEVEk0TURReU9ESXhNRE0wTUZvd1BERVFNQTRHCkExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNUmN3RlFZRFZRUURFdzVCY0c5dGRYZ2cKVW05dmRDQkRRVEJaTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5QXdFSEEwSUFCQnYyMUhMM3pjWGROZERzK3RRcwpmZWl6eno3ODRjcXp0TE0zYXFPRWlqdkNraGNGOURmdFFnTlQ2cEMxMVNJZ1IzVkJBY2xFZFU3aGdnRnRGR3lrCmR1T2pJekFoTUE0R0ExVWREd0VCL3dRRUF3SUJCakFQQmdOVkhSTUJBZjhFQlRBREFRSC9NQW9HQ0NxR1NNNDkKQkFNQ0EwZ0FNRVVDSVFEZ0dQQ0FLMlpsMkwrcUkwRFd1YWd1ZmFXampBUE9YOWFqVkRIbDBsbkVwd0lnTVRCeAphaWo4TkpGRHphaHBsc0dWZUE3WFJld3Y2VjRCMW4zMCtaZHA4Tk09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K","certificateKey":"LS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1IY0NBUUVFSUxuMkFMN3FuMVRrK0VYNWNBU0gxdTljS1JzQ0tndnFmaVlFL3RDaGZYbm1vQW9HQ0NxR1NNNDkKQXdFSG9VUURRZ0FFcG5nSDYraEhxcGlDVkdvWHQ3Z1ZZenpmUkJIT3ZhUG1xTks0eE1YdFRWOXNNSXhLSXBkMwpBdlBOSjVqZWVSQkY4U05pNHNIeFJIOUJKMzNiN0xWenpRPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQo="}`

		Convey("When I call AppCredsToTLSConfig", func() {

			_, tlsConfig, err := ParseCredentials([]byte(credsData))

			Convey("Then the err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then tlsconfig should be correct", func() {
				So(tlsConfig.Certificates[0].Certificate, ShouldResemble, [][]byte{{48, 130, 1, 191, 48, 130, 1, 101, 160, 3, 2, 1, 2, 2, 16, 14, 23, 62, 3, 172, 210, 140, 248, 138, 159, 141, 95, 163, 168, 13, 244, 48, 10, 6, 8, 42, 134, 72, 206, 61, 4, 3, 2, 48, 70, 49, 16, 48, 14, 6, 3, 85, 4, 10, 19, 7, 65, 112, 111, 114, 101, 116, 111, 49, 15, 48, 13, 6, 3, 85, 4, 11, 19, 6, 97, 112, 111, 109, 117, 120, 49, 33, 48, 31, 6, 3, 85, 4, 3, 19, 24, 65, 112, 111, 109, 117, 120, 32, 80, 117, 98, 108, 105, 99, 32, 83, 105, 103, 110, 105, 110, 103, 32, 67, 65, 48, 30, 23, 13, 49, 56, 49, 48, 49, 54, 49, 57, 53, 54, 50, 54, 90, 23, 13, 49, 57, 49, 48, 49, 54, 50, 48, 53, 54, 50, 54, 90, 48, 70, 49, 10, 48, 8, 6, 3, 85, 4, 10, 19, 1, 47, 49, 56, 48, 54, 6, 3, 85, 4, 3, 19, 47, 97, 112, 112, 58, 99, 114, 101, 100, 101, 110, 116, 105, 97, 108, 58, 53, 98, 99, 54, 53, 48, 54, 98, 55, 100, 100, 102, 49, 102, 55, 53, 56, 51, 102, 99, 99, 51, 98, 49, 58, 116, 101, 115, 116, 97, 112, 112, 48, 89, 48, 19, 6, 7, 42, 134, 72, 206, 61, 2, 1, 6, 8, 42, 134, 72, 206, 61, 3, 1, 7, 3, 66, 0, 4, 166, 120, 7, 235, 232, 71, 170, 152, 130, 84, 106, 23, 183, 184, 21, 99, 60, 223, 68, 17, 206, 189, 163, 230, 168, 210, 184, 196, 197, 237, 77, 95, 108, 48, 140, 74, 34, 151, 119, 2, 243, 205, 39, 152, 222, 121, 16, 69, 241, 35, 98, 226, 193, 241, 68, 127, 65, 39, 125, 219, 236, 181, 115, 205, 163, 53, 48, 51, 48, 14, 6, 3, 85, 29, 15, 1, 1, 255, 4, 4, 3, 2, 5, 160, 48, 19, 6, 3, 85, 29, 37, 4, 12, 48, 10, 6, 8, 43, 6, 1, 5, 5, 7, 3, 2, 48, 12, 6, 3, 85, 29, 19, 1, 1, 255, 4, 2, 48, 0, 48, 10, 6, 8, 42, 134, 72, 206, 61, 4, 3, 2, 3, 72, 0, 48, 69, 2, 32, 59, 142, 15, 35, 244, 161, 155, 153, 101, 60, 251, 73, 150, 39, 163, 51, 235, 194, 205, 152, 244, 216, 227, 248, 29, 80, 58, 13, 86, 1, 190, 58, 2, 33, 0, 163, 128, 242, 248, 98, 242, 91, 118, 178, 194, 79, 195, 154, 136, 161, 218, 23, 216, 186, 23, 115, 141, 227, 188, 56, 79, 73, 23, 88, 82, 170, 23}})
			})
		})
	})

	Convey("Given I have some bad json appcred", t, func() {

		credsData := `nope`

		Convey("When I call AppCredsToTLSConfig", func() {

			_, tlsConfig, err := ParseCredentials([]byte(credsData))

			Convey("Then the err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then the tlsConfig should be nil", func() {
				So(tlsConfig, ShouldBeNil)
			})

			Convey("Then the err should be correct", func() {
				So(err.Error(), ShouldEqual, "unable to decode app credential: invalid character 'o' in literal null (expecting 'u')")
			})
		})
	})

	Convey("Given I have some bad encoded cert in appcred", t, func() {

		credsData := `{"certificate":"^^^NOTGOOD=","certificateAuthority":"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJyRENDQVZLZ0F3SUJBZ0lSQUtjMERhOUVRSHB4aGxickNvTmZ2T1F3Q2dZSUtvWkl6ajBFQXdJd1JERVEKTUE0R0ExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNUjh3SFFZRFZRUURFeFpCY0c5dApkWGdnU1c1MFpYSnRaV1JwWVhSbElFTkJNQjRYRFRFNE1EWXlNREl4TURNME1Gb1hEVEk0TURReU9ESXhNRE0wCk1Gb3dSakVRTUE0R0ExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNU0V3SHdZRFZRUUQKRXhoQmNHOXRkWGdnVUhWaWJHbGpJRk5wWjI1cGJtY2dRMEV3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQgpCd05DQUFUSlExeVRDVEpzQUx0N25UbjBZRVNpSGgvZ0xlWlBDWlBhb09nWEJIdU5icEltUTF5Z0xPb2wvMUc1CmZ3VzdJNVJTdXZqNCtwV0Nad3pTbmxRaFIwZ0tveU13SVRBT0JnTlZIUThCQWY4RUJBTUNBUVl3RHdZRFZSMFQKQVFIL0JBVXdBd0VCL3pBS0JnZ3Foa2pPUFFRREFnTklBREJGQWlCSlNJNlRjQTdTODhnWmhXb29oeXYxK0FxNQpuY0dybXN1SG9NdUN3WEJUelFJaEFNeVRaMW5lZFEwelQ1SkVIQTJoaFRmUjFCT01zQS9Ic3AwNWpPa1BJbVpnCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJvVENDQVVlZ0F3SUJBZ0lRU2VKS3pXNjV4elFhZzlBeEhPVGR2REFLQmdncWhrak9QUVFEQWpBOE1SQXcKRGdZRFZRUUtFd2RCY0c5eVpYUnZNUTh3RFFZRFZRUUxFd1poY0c5dGRYZ3hGekFWQmdOVkJBTVREa0Z3YjIxMQplQ0JTYjI5MElFTkJNQjRYRFRFNE1EWXlNREl4TURNME1Gb1hEVEk0TURReU9ESXhNRE0wTUZvd1JERVFNQTRHCkExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNUjh3SFFZRFZRUURFeFpCY0c5dGRYZ2cKU1c1MFpYSnRaV1JwWVhSbElFTkJNRmt3RXdZSEtvWkl6ajBDQVFZSUtvWkl6ajBEQVFjRFFnQUUvNXRrN3pSdgpNWDVuZ1l6dkhNUEh1ZXVOc2dkU1pWMzRkZk4va3UyakxjZUwrNi9FNUViQWpHdWYrY3RLT3dRamNha09oajE0Cllrb1dHL0svNzYvZzg2TWpNQ0V3RGdZRFZSMFBBUUgvQkFRREFnRUdNQThHQTFVZEV3RUIvd1FGTUFNQkFmOHcKQ2dZSUtvWkl6ajBFQXdJRFNBQXdSUUloQU5aT3ZUVDhicHp1Vk1FY2xORzBsaFlCdmt3L0dXYjFZVWxNTFJCeApHYjNFQWlCL3RCQTlPN1AyZXdQaU9hclhNb2FzZFVjNU83Ukk2QThUdTczQ28vamtmdz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJtVENDQVQrZ0F3SUJBZ0lRYVJId3B6NWw5blo2eEoyRVIwdkNHakFLQmdncWhrak9QUVFEQWpBOE1SQXcKRGdZRFZRUUtFd2RCY0c5eVpYUnZNUTh3RFFZRFZRUUxFd1poY0c5dGRYZ3hGekFWQmdOVkJBTVREa0Z3YjIxMQplQ0JTYjI5MElFTkJNQjRYRFRFNE1EWXlNREl4TURNME1Gb1hEVEk0TURReU9ESXhNRE0wTUZvd1BERVFNQTRHCkExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNUmN3RlFZRFZRUURFdzVCY0c5dGRYZ2cKVW05dmRDQkRRVEJaTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5QXdFSEEwSUFCQnYyMUhMM3pjWGROZERzK3RRcwpmZWl6eno3ODRjcXp0TE0zYXFPRWlqdkNraGNGOURmdFFnTlQ2cEMxMVNJZ1IzVkJBY2xFZFU3aGdnRnRGR3lrCmR1T2pJekFoTUE0R0ExVWREd0VCL3dRRUF3SUJCakFQQmdOVkhSTUJBZjhFQlRBREFRSC9NQW9HQ0NxR1NNNDkKQkFNQ0EwZ0FNRVVDSVFEZ0dQQ0FLMlpsMkwrcUkwRFd1YWd1ZmFXampBUE9YOWFqVkRIbDBsbkVwd0lnTVRCeAphaWo4TkpGRHphaHBsc0dWZUE3WFJld3Y2VjRCMW4zMCtaZHA4Tk09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K","certificateKey":"LS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1IY0NBUUVFSUxuMkFMN3FuMVRrK0VYNWNBU0gxdTljS1JzQ0tndnFmaVlFL3RDaGZYbm1vQW9HQ0NxR1NNNDkKQXdFSG9VUURRZ0FFcG5nSDYraEhxcGlDVkdvWHQ3Z1ZZenpmUkJIT3ZhUG1xTks0eE1YdFRWOXNNSXhLSXBkMwpBdlBOSjVqZWVSQkY4U05pNHNIeFJIOUJKMzNiN0xWenpRPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQo="}`

		Convey("When I call AppCredsToTLSConfig", func() {

			_, tlsConfig, err := ParseCredentials([]byte(credsData))

			Convey("Then the err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then the tlsConfig should be nil", func() {
				So(tlsConfig, ShouldBeNil)
			})

			Convey("Then the err should be correct", func() {
				So(err.Error(), ShouldEqual, "unable to derive tls config from creds: unable to decode certificate: illegal base64 data at input byte 0")
			})
		})
	})

	Convey("Given I have some bad encoded key in appcred", t, func() {

		credsData := `{"certificate":"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJ2ekNDQVdXZ0F3SUJBZ0lRRGhjK0E2elNqUGlLbjQxZm82Z045REFLQmdncWhrak9QUVFEQWpCR01SQXcKRGdZRFZRUUtFd2RCY0c5eVpYUnZNUTh3RFFZRFZRUUxFd1poY0c5dGRYZ3hJVEFmQmdOVkJBTVRHRUZ3YjIxMQplQ0JRZFdKc2FXTWdVMmxuYm1sdVp5QkRRVEFlRncweE9ERXdNVFl4T1RVMk1qWmFGdzB4T1RFd01UWXlNRFUyCk1qWmFNRVl4Q2pBSUJnTlZCQW9UQVM4eE9EQTJCZ05WQkFNVEwyRndjRHBqY21Wa1pXNTBhV0ZzT2pWaVl6WTEKTURaaU4yUmtaakZtTnpVNE0yWmpZek5pTVRwMFpYTjBZWEJ3TUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowRApBUWNEUWdBRXBuZ0g2K2hIcXBpQ1ZHb1h0N2dWWXp6ZlJCSE92YVBtcU5LNHhNWHRUVjlzTUl4S0lwZDNBdlBOCko1amVlUkJGOFNOaTRzSHhSSDlCSjMzYjdMVnp6YU0xTURNd0RnWURWUjBQQVFIL0JBUURBZ1dnTUJNR0ExVWQKSlFRTU1Bb0dDQ3NHQVFVRkJ3TUNNQXdHQTFVZEV3RUIvd1FDTUFBd0NnWUlLb1pJemowRUF3SURTQUF3UlFJZwpPNDRQSS9TaG01bGxQUHRKbGllak0rdkN6WmowMk9QNEhWQTZEVllCdmpvQ0lRQ2pnUEw0WXZKYmRyTENUOE9hCmlLSGFGOWk2RjNPTjQ3dzRUMGtYV0ZLcUZ3PT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=","certificateAuthority":"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJyRENDQVZLZ0F3SUJBZ0lSQUtjMERhOUVRSHB4aGxickNvTmZ2T1F3Q2dZSUtvWkl6ajBFQXdJd1JERVEKTUE0R0ExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNUjh3SFFZRFZRUURFeFpCY0c5dApkWGdnU1c1MFpYSnRaV1JwWVhSbElFTkJNQjRYRFRFNE1EWXlNREl4TURNME1Gb1hEVEk0TURReU9ESXhNRE0wCk1Gb3dSakVRTUE0R0ExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNU0V3SHdZRFZRUUQKRXhoQmNHOXRkWGdnVUhWaWJHbGpJRk5wWjI1cGJtY2dRMEV3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQgpCd05DQUFUSlExeVRDVEpzQUx0N25UbjBZRVNpSGgvZ0xlWlBDWlBhb09nWEJIdU5icEltUTF5Z0xPb2wvMUc1CmZ3VzdJNVJTdXZqNCtwV0Nad3pTbmxRaFIwZ0tveU13SVRBT0JnTlZIUThCQWY4RUJBTUNBUVl3RHdZRFZSMFQKQVFIL0JBVXdBd0VCL3pBS0JnZ3Foa2pPUFFRREFnTklBREJGQWlCSlNJNlRjQTdTODhnWmhXb29oeXYxK0FxNQpuY0dybXN1SG9NdUN3WEJUelFJaEFNeVRaMW5lZFEwelQ1SkVIQTJoaFRmUjFCT01zQS9Ic3AwNWpPa1BJbVpnCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJvVENDQVVlZ0F3SUJBZ0lRU2VKS3pXNjV4elFhZzlBeEhPVGR2REFLQmdncWhrak9QUVFEQWpBOE1SQXcKRGdZRFZRUUtFd2RCY0c5eVpYUnZNUTh3RFFZRFZRUUxFd1poY0c5dGRYZ3hGekFWQmdOVkJBTVREa0Z3YjIxMQplQ0JTYjI5MElFTkJNQjRYRFRFNE1EWXlNREl4TURNME1Gb1hEVEk0TURReU9ESXhNRE0wTUZvd1JERVFNQTRHCkExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNUjh3SFFZRFZRUURFeFpCY0c5dGRYZ2cKU1c1MFpYSnRaV1JwWVhSbElFTkJNRmt3RXdZSEtvWkl6ajBDQVFZSUtvWkl6ajBEQVFjRFFnQUUvNXRrN3pSdgpNWDVuZ1l6dkhNUEh1ZXVOc2dkU1pWMzRkZk4va3UyakxjZUwrNi9FNUViQWpHdWYrY3RLT3dRamNha09oajE0Cllrb1dHL0svNzYvZzg2TWpNQ0V3RGdZRFZSMFBBUUgvQkFRREFnRUdNQThHQTFVZEV3RUIvd1FGTUFNQkFmOHcKQ2dZSUtvWkl6ajBFQXdJRFNBQXdSUUloQU5aT3ZUVDhicHp1Vk1FY2xORzBsaFlCdmt3L0dXYjFZVWxNTFJCeApHYjNFQWlCL3RCQTlPN1AyZXdQaU9hclhNb2FzZFVjNU83Ukk2QThUdTczQ28vamtmdz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJtVENDQVQrZ0F3SUJBZ0lRYVJId3B6NWw5blo2eEoyRVIwdkNHakFLQmdncWhrak9QUVFEQWpBOE1SQXcKRGdZRFZRUUtFd2RCY0c5eVpYUnZNUTh3RFFZRFZRUUxFd1poY0c5dGRYZ3hGekFWQmdOVkJBTVREa0Z3YjIxMQplQ0JTYjI5MElFTkJNQjRYRFRFNE1EWXlNREl4TURNME1Gb1hEVEk0TURReU9ESXhNRE0wTUZvd1BERVFNQTRHCkExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNUmN3RlFZRFZRUURFdzVCY0c5dGRYZ2cKVW05dmRDQkRRVEJaTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5QXdFSEEwSUFCQnYyMUhMM3pjWGROZERzK3RRcwpmZWl6eno3ODRjcXp0TE0zYXFPRWlqdkNraGNGOURmdFFnTlQ2cEMxMVNJZ1IzVkJBY2xFZFU3aGdnRnRGR3lrCmR1T2pJekFoTUE0R0ExVWREd0VCL3dRRUF3SUJCakFQQmdOVkhSTUJBZjhFQlRBREFRSC9NQW9HQ0NxR1NNNDkKQkFNQ0EwZ0FNRVVDSVFEZ0dQQ0FLMlpsMkwrcUkwRFd1YWd1ZmFXampBUE9YOWFqVkRIbDBsbkVwd0lnTVRCeAphaWo4TkpGRHphaHBsc0dWZUE3WFJld3Y2VjRCMW4zMCtaZHA4Tk09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K","certificateKey":"BAD"}`

		Convey("When I call AppCredsToTLSConfig", func() {

			_, tlsConfig, err := ParseCredentials([]byte(credsData))

			Convey("Then the err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then the tlsConfig should be nil", func() {
				So(tlsConfig, ShouldBeNil)
			})

			Convey("Then the err should be correct", func() {
				So(err.Error(), ShouldEqual, "unable to derive tls config from creds: unable to decode key: illegal base64 data at input byte 0")
			})
		})
	})

	Convey("Given I have some bad encoded ca in appcred", t, func() {

		credsData := `{"certificate":"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJ2ekNDQVdXZ0F3SUJBZ0lRRGhjK0E2elNqUGlLbjQxZm82Z045REFLQmdncWhrak9QUVFEQWpCR01SQXcKRGdZRFZRUUtFd2RCY0c5eVpYUnZNUTh3RFFZRFZRUUxFd1poY0c5dGRYZ3hJVEFmQmdOVkJBTVRHRUZ3YjIxMQplQ0JRZFdKc2FXTWdVMmxuYm1sdVp5QkRRVEFlRncweE9ERXdNVFl4T1RVMk1qWmFGdzB4T1RFd01UWXlNRFUyCk1qWmFNRVl4Q2pBSUJnTlZCQW9UQVM4eE9EQTJCZ05WQkFNVEwyRndjRHBqY21Wa1pXNTBhV0ZzT2pWaVl6WTEKTURaaU4yUmtaakZtTnpVNE0yWmpZek5pTVRwMFpYTjBZWEJ3TUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowRApBUWNEUWdBRXBuZ0g2K2hIcXBpQ1ZHb1h0N2dWWXp6ZlJCSE92YVBtcU5LNHhNWHRUVjlzTUl4S0lwZDNBdlBOCko1amVlUkJGOFNOaTRzSHhSSDlCSjMzYjdMVnp6YU0xTURNd0RnWURWUjBQQVFIL0JBUURBZ1dnTUJNR0ExVWQKSlFRTU1Bb0dDQ3NHQVFVRkJ3TUNNQXdHQTFVZEV3RUIvd1FDTUFBd0NnWUlLb1pJemowRUF3SURTQUF3UlFJZwpPNDRQSS9TaG01bGxQUHRKbGllak0rdkN6WmowMk9QNEhWQTZEVllCdmpvQ0lRQ2pnUEw0WXZKYmRyTENUOE9hCmlLSGFGOWk2RjNPTjQ3dzRUMGtYV0ZLcUZ3PT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=","certificateAuthority":"BAD","certificateKey":"LS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1IY0NBUUVFSUxuMkFMN3FuMVRrK0VYNWNBU0gxdTljS1JzQ0tndnFmaVlFL3RDaGZYbm1vQW9HQ0NxR1NNNDkKQXdFSG9VUURRZ0FFcG5nSDYraEhxcGlDVkdvWHQ3Z1ZZenpmUkJIT3ZhUG1xTks0eE1YdFRWOXNNSXhLSXBkMwpBdlBOSjVqZWVSQkY4U05pNHNIeFJIOUJKMzNiN0xWenpRPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQo="}`

		Convey("When I call AppCredsToTLSConfig", func() {

			_, tlsConfig, err := ParseCredentials([]byte(credsData))

			Convey("Then the err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then the tlsConfig should be nil", func() {
				So(tlsConfig, ShouldBeNil)
			})

			Convey("Then the err should be correct", func() {
				So(err.Error(), ShouldEqual, "unable to derive tls config from creds: unable to decode ca: illegal base64 data at input byte 0")
			})
		})
	})

	Convey("Given I have some incorrect cert in appcred", t, func() {

		credsData := `{"certificate":"d29vcHM=","certificateAuthority":"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJyRENDQVZLZ0F3SUJBZ0lSQUtjMERhOUVRSHB4aGxickNvTmZ2T1F3Q2dZSUtvWkl6ajBFQXdJd1JERVEKTUE0R0ExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNUjh3SFFZRFZRUURFeFpCY0c5dApkWGdnU1c1MFpYSnRaV1JwWVhSbElFTkJNQjRYRFRFNE1EWXlNREl4TURNME1Gb1hEVEk0TURReU9ESXhNRE0wCk1Gb3dSakVRTUE0R0ExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNU0V3SHdZRFZRUUQKRXhoQmNHOXRkWGdnVUhWaWJHbGpJRk5wWjI1cGJtY2dRMEV3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQgpCd05DQUFUSlExeVRDVEpzQUx0N25UbjBZRVNpSGgvZ0xlWlBDWlBhb09nWEJIdU5icEltUTF5Z0xPb2wvMUc1CmZ3VzdJNVJTdXZqNCtwV0Nad3pTbmxRaFIwZ0tveU13SVRBT0JnTlZIUThCQWY4RUJBTUNBUVl3RHdZRFZSMFQKQVFIL0JBVXdBd0VCL3pBS0JnZ3Foa2pPUFFRREFnTklBREJGQWlCSlNJNlRjQTdTODhnWmhXb29oeXYxK0FxNQpuY0dybXN1SG9NdUN3WEJUelFJaEFNeVRaMW5lZFEwelQ1SkVIQTJoaFRmUjFCT01zQS9Ic3AwNWpPa1BJbVpnCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJvVENDQVVlZ0F3SUJBZ0lRU2VKS3pXNjV4elFhZzlBeEhPVGR2REFLQmdncWhrak9QUVFEQWpBOE1SQXcKRGdZRFZRUUtFd2RCY0c5eVpYUnZNUTh3RFFZRFZRUUxFd1poY0c5dGRYZ3hGekFWQmdOVkJBTVREa0Z3YjIxMQplQ0JTYjI5MElFTkJNQjRYRFRFNE1EWXlNREl4TURNME1Gb1hEVEk0TURReU9ESXhNRE0wTUZvd1JERVFNQTRHCkExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNUjh3SFFZRFZRUURFeFpCY0c5dGRYZ2cKU1c1MFpYSnRaV1JwWVhSbElFTkJNRmt3RXdZSEtvWkl6ajBDQVFZSUtvWkl6ajBEQVFjRFFnQUUvNXRrN3pSdgpNWDVuZ1l6dkhNUEh1ZXVOc2dkU1pWMzRkZk4va3UyakxjZUwrNi9FNUViQWpHdWYrY3RLT3dRamNha09oajE0Cllrb1dHL0svNzYvZzg2TWpNQ0V3RGdZRFZSMFBBUUgvQkFRREFnRUdNQThHQTFVZEV3RUIvd1FGTUFNQkFmOHcKQ2dZSUtvWkl6ajBFQXdJRFNBQXdSUUloQU5aT3ZUVDhicHp1Vk1FY2xORzBsaFlCdmt3L0dXYjFZVWxNTFJCeApHYjNFQWlCL3RCQTlPN1AyZXdQaU9hclhNb2FzZFVjNU83Ukk2QThUdTczQ28vamtmdz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJtVENDQVQrZ0F3SUJBZ0lRYVJId3B6NWw5blo2eEoyRVIwdkNHakFLQmdncWhrak9QUVFEQWpBOE1SQXcKRGdZRFZRUUtFd2RCY0c5eVpYUnZNUTh3RFFZRFZRUUxFd1poY0c5dGRYZ3hGekFWQmdOVkJBTVREa0Z3YjIxMQplQ0JTYjI5MElFTkJNQjRYRFRFNE1EWXlNREl4TURNME1Gb1hEVEk0TURReU9ESXhNRE0wTUZvd1BERVFNQTRHCkExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNUmN3RlFZRFZRUURFdzVCY0c5dGRYZ2cKVW05dmRDQkRRVEJaTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5QXdFSEEwSUFCQnYyMUhMM3pjWGROZERzK3RRcwpmZWl6eno3ODRjcXp0TE0zYXFPRWlqdkNraGNGOURmdFFnTlQ2cEMxMVNJZ1IzVkJBY2xFZFU3aGdnRnRGR3lrCmR1T2pJekFoTUE0R0ExVWREd0VCL3dRRUF3SUJCakFQQmdOVkhSTUJBZjhFQlRBREFRSC9NQW9HQ0NxR1NNNDkKQkFNQ0EwZ0FNRVVDSVFEZ0dQQ0FLMlpsMkwrcUkwRFd1YWd1ZmFXampBUE9YOWFqVkRIbDBsbkVwd0lnTVRCeAphaWo4TkpGRHphaHBsc0dWZUE3WFJld3Y2VjRCMW4zMCtaZHA4Tk09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K","certificateKey":"LS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1IY0NBUUVFSUxuMkFMN3FuMVRrK0VYNWNBU0gxdTljS1JzQ0tndnFmaVlFL3RDaGZYbm1vQW9HQ0NxR1NNNDkKQXdFSG9VUURRZ0FFcG5nSDYraEhxcGlDVkdvWHQ3Z1ZZenpmUkJIT3ZhUG1xTks0eE1YdFRWOXNNSXhLSXBkMwpBdlBOSjVqZWVSQkY4U05pNHNIeFJIOUJKMzNiN0xWenpRPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQo="}`

		Convey("When I call AppCredsToTLSConfig", func() {

			_, tlsConfig, err := ParseCredentials([]byte(credsData))

			Convey("Then the err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then the tlsConfig should be nil", func() {
				So(tlsConfig, ShouldBeNil)
			})

			Convey("Then the err should be correct", func() {
				So(err.Error(), ShouldEqual, "unable to derive tls config from creds: unable to parse certificate: tls: failed to find any PEM data in certificate input")
			})
		})
	})

	Convey("Given I have some incorrect key in appcred", t, func() {

		credsData := `{"certificate":"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJ2ekNDQVdXZ0F3SUJBZ0lRRGhjK0E2elNqUGlLbjQxZm82Z045REFLQmdncWhrak9QUVFEQWpCR01SQXcKRGdZRFZRUUtFd2RCY0c5eVpYUnZNUTh3RFFZRFZRUUxFd1poY0c5dGRYZ3hJVEFmQmdOVkJBTVRHRUZ3YjIxMQplQ0JRZFdKc2FXTWdVMmxuYm1sdVp5QkRRVEFlRncweE9ERXdNVFl4T1RVMk1qWmFGdzB4T1RFd01UWXlNRFUyCk1qWmFNRVl4Q2pBSUJnTlZCQW9UQVM4eE9EQTJCZ05WQkFNVEwyRndjRHBqY21Wa1pXNTBhV0ZzT2pWaVl6WTEKTURaaU4yUmtaakZtTnpVNE0yWmpZek5pTVRwMFpYTjBZWEJ3TUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowRApBUWNEUWdBRXBuZ0g2K2hIcXBpQ1ZHb1h0N2dWWXp6ZlJCSE92YVBtcU5LNHhNWHRUVjlzTUl4S0lwZDNBdlBOCko1amVlUkJGOFNOaTRzSHhSSDlCSjMzYjdMVnp6YU0xTURNd0RnWURWUjBQQVFIL0JBUURBZ1dnTUJNR0ExVWQKSlFRTU1Bb0dDQ3NHQVFVRkJ3TUNNQXdHQTFVZEV3RUIvd1FDTUFBd0NnWUlLb1pJemowRUF3SURTQUF3UlFJZwpPNDRQSS9TaG01bGxQUHRKbGllak0rdkN6WmowMk9QNEhWQTZEVllCdmpvQ0lRQ2pnUEw0WXZKYmRyTENUOE9hCmlLSGFGOWk2RjNPTjQ3dzRUMGtYV0ZLcUZ3PT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=","certificateAuthority":"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJyRENDQVZLZ0F3SUJBZ0lSQUtjMERhOUVRSHB4aGxickNvTmZ2T1F3Q2dZSUtvWkl6ajBFQXdJd1JERVEKTUE0R0ExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNUjh3SFFZRFZRUURFeFpCY0c5dApkWGdnU1c1MFpYSnRaV1JwWVhSbElFTkJNQjRYRFRFNE1EWXlNREl4TURNME1Gb1hEVEk0TURReU9ESXhNRE0wCk1Gb3dSakVRTUE0R0ExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNU0V3SHdZRFZRUUQKRXhoQmNHOXRkWGdnVUhWaWJHbGpJRk5wWjI1cGJtY2dRMEV3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQgpCd05DQUFUSlExeVRDVEpzQUx0N25UbjBZRVNpSGgvZ0xlWlBDWlBhb09nWEJIdU5icEltUTF5Z0xPb2wvMUc1CmZ3VzdJNVJTdXZqNCtwV0Nad3pTbmxRaFIwZ0tveU13SVRBT0JnTlZIUThCQWY4RUJBTUNBUVl3RHdZRFZSMFQKQVFIL0JBVXdBd0VCL3pBS0JnZ3Foa2pPUFFRREFnTklBREJGQWlCSlNJNlRjQTdTODhnWmhXb29oeXYxK0FxNQpuY0dybXN1SG9NdUN3WEJUelFJaEFNeVRaMW5lZFEwelQ1SkVIQTJoaFRmUjFCT01zQS9Ic3AwNWpPa1BJbVpnCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJvVENDQVVlZ0F3SUJBZ0lRU2VKS3pXNjV4elFhZzlBeEhPVGR2REFLQmdncWhrak9QUVFEQWpBOE1SQXcKRGdZRFZRUUtFd2RCY0c5eVpYUnZNUTh3RFFZRFZRUUxFd1poY0c5dGRYZ3hGekFWQmdOVkJBTVREa0Z3YjIxMQplQ0JTYjI5MElFTkJNQjRYRFRFNE1EWXlNREl4TURNME1Gb1hEVEk0TURReU9ESXhNRE0wTUZvd1JERVFNQTRHCkExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNUjh3SFFZRFZRUURFeFpCY0c5dGRYZ2cKU1c1MFpYSnRaV1JwWVhSbElFTkJNRmt3RXdZSEtvWkl6ajBDQVFZSUtvWkl6ajBEQVFjRFFnQUUvNXRrN3pSdgpNWDVuZ1l6dkhNUEh1ZXVOc2dkU1pWMzRkZk4va3UyakxjZUwrNi9FNUViQWpHdWYrY3RLT3dRamNha09oajE0Cllrb1dHL0svNzYvZzg2TWpNQ0V3RGdZRFZSMFBBUUgvQkFRREFnRUdNQThHQTFVZEV3RUIvd1FGTUFNQkFmOHcKQ2dZSUtvWkl6ajBFQXdJRFNBQXdSUUloQU5aT3ZUVDhicHp1Vk1FY2xORzBsaFlCdmt3L0dXYjFZVWxNTFJCeApHYjNFQWlCL3RCQTlPN1AyZXdQaU9hclhNb2FzZFVjNU83Ukk2QThUdTczQ28vamtmdz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJtVENDQVQrZ0F3SUJBZ0lRYVJId3B6NWw5blo2eEoyRVIwdkNHakFLQmdncWhrak9QUVFEQWpBOE1SQXcKRGdZRFZRUUtFd2RCY0c5eVpYUnZNUTh3RFFZRFZRUUxFd1poY0c5dGRYZ3hGekFWQmdOVkJBTVREa0Z3YjIxMQplQ0JTYjI5MElFTkJNQjRYRFRFNE1EWXlNREl4TURNME1Gb1hEVEk0TURReU9ESXhNRE0wTUZvd1BERVFNQTRHCkExVUVDaE1IUVhCdmNtVjBiekVQTUEwR0ExVUVDeE1HWVhCdmJYVjRNUmN3RlFZRFZRUURFdzVCY0c5dGRYZ2cKVW05dmRDQkRRVEJaTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5QXdFSEEwSUFCQnYyMUhMM3pjWGROZERzK3RRcwpmZWl6eno3ODRjcXp0TE0zYXFPRWlqdkNraGNGOURmdFFnTlQ2cEMxMVNJZ1IzVkJBY2xFZFU3aGdnRnRGR3lrCmR1T2pJekFoTUE0R0ExVWREd0VCL3dRRUF3SUJCakFQQmdOVkhSTUJBZjhFQlRBREFRSC9NQW9HQ0NxR1NNNDkKQkFNQ0EwZ0FNRVVDSVFEZ0dQQ0FLMlpsMkwrcUkwRFd1YWd1ZmFXampBUE9YOWFqVkRIbDBsbkVwd0lnTVRCeAphaWo4TkpGRHphaHBsc0dWZUE3WFJld3Y2VjRCMW4zMCtaZHA4Tk09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K","certificateKey":"d29vcHM="}`

		Convey("When I call AppCredsToTLSConfig", func() {

			_, tlsConfig, err := ParseCredentials([]byte(credsData))

			Convey("Then the err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then the tlsConfig should be nil", func() {
				So(tlsConfig, ShouldBeNil)
			})

			Convey("Then the err should be correct", func() {
				So(err.Error(), ShouldEqual, "unable to derive tls config from creds: unable to parse certificate: could not read key data from bytes: 'woops'")
			})
		})
	})

	// Convey("Given I have some incorrect ca in appcred", t, func() {

	// 	credsData := `{"certificate":"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJ2ekNDQVdXZ0F3SUJBZ0lRRGhjK0E2elNqUGlLbjQxZm82Z045REFLQmdncWhrak9QUVFEQWpCR01SQXcKRGdZRFZRUUtFd2RCY0c5eVpYUnZNUTh3RFFZRFZRUUxFd1poY0c5dGRYZ3hJVEFmQmdOVkJBTVRHRUZ3YjIxMQplQ0JRZFdKc2FXTWdVMmxuYm1sdVp5QkRRVEFlRncweE9ERXdNVFl4T1RVMk1qWmFGdzB4T1RFd01UWXlNRFUyCk1qWmFNRVl4Q2pBSUJnTlZCQW9UQVM4eE9EQTJCZ05WQkFNVEwyRndjRHBqY21Wa1pXNTBhV0ZzT2pWaVl6WTEKTURaaU4yUmtaakZtTnpVNE0yWmpZek5pTVRwMFpYTjBZWEJ3TUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowRApBUWNEUWdBRXBuZ0g2K2hIcXBpQ1ZHb1h0N2dWWXp6ZlJCSE92YVBtcU5LNHhNWHRUVjlzTUl4S0lwZDNBdlBOCko1amVlUkJGOFNOaTRzSHhSSDlCSjMzYjdMVnp6YU0xTURNd0RnWURWUjBQQVFIL0JBUURBZ1dnTUJNR0ExVWQKSlFRTU1Bb0dDQ3NHQVFVRkJ3TUNNQXdHQTFVZEV3RUIvd1FDTUFBd0NnWUlLb1pJemowRUF3SURTQUF3UlFJZwpPNDRQSS9TaG01bGxQUHRKbGllak0rdkN6WmowMk9QNEhWQTZEVllCdmpvQ0lRQ2pnUEw0WXZKYmRyTENUOE9hCmlLSGFGOWk2RjNPTjQ3dzRUMGtYV0ZLcUZ3PT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=","certificateAuthority":"d29vcHM=","certificateKey":"LS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1IY0NBUUVFSUxuMkFMN3FuMVRrK0VYNWNBU0gxdTljS1JzQ0tndnFmaVlFL3RDaGZYbm1vQW9HQ0NxR1NNNDkKQXdFSG9VUURRZ0FFcG5nSDYraEhxcGlDVkdvWHQ3Z1ZZenpmUkJIT3ZhUG1xTks0eE1YdFRWOXNNSXhLSXBkMwpBdlBOSjVqZWVSQkY4U05pNHNIeFJIOUJKMzNiN0xWenpRPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQo="}`

	// 	Convey("When I call AppCredsToTLSConfig", func() {

	// 		_, tlsConfig, err := ParseCredentials([]byte(credsData))

	// 		Convey("Then the err should not be nil", func() {
	// 			So(err, ShouldNotBeNil)
	// 		})

	// 		Convey("Then the tlsConfig should be nil", func() {
	// 			So(tlsConfig, ShouldBeNil)
	// 		})

	// 		Convey("Then the err should be correct", func() {
	// 			So(err.Error(), ShouldEqual, "unable to add ca to cert pool")
	// 		})
	// 	})
	// })
}

func TestUnsecureClaimsFromToken(t *testing.T) {

	validToken := `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyZWFsbSI6IlZpbmNlIiwiZGF0YSI6eyJhY2NvdW50IjoiYXBvbXV4IiwiZW1haWwiOiJhZG1pbkBhcG9tdXguY29tIiwiaWQiOiI1YTZhNTUxMTdkZGYxZjIxMmY4ZWIwY2UiLCJvcmdhbml6YXRpb24iOiJhcG9tdXgiLCJyZWFsbSI6InZpbmNlIn0sImF1ZCI6ImFwb3JldG8uY29tIiwiZXhwIjoxNTIwNjQ5MTAyLCJpYXQiOjE1MTgwNTcxMDIsImlzcyI6Im1pZGdhcmQuYXBvbXV4LmNvbSIsInN1YiI6ImFwb211eCJ9.jvh034mNSV-Fy--GIGnnYeWouluV6CexC9_8IHJ-IR4`

	type args struct {
		token string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			"simple",
			args{
				validToken,
			},
			[]string{
				"@auth:account=apomux",
				"@auth:email=admin@apomux.com",
				"@auth:id=5a6a55117ddf1f212f8eb0ce",
				"@auth:organization=apomux",
				"@auth:realm=vince",
				"@auth:subject=apomux",
			},
			false,
		},
		{
			"invalid token",
			args{
				`nope`,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnsecureClaimsFromToken(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnsecureClaimsFromToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnsecureClaimsFromToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

var signerCert = []byte(`-----BEGIN CERTIFICATE-----
MIIBPzCB56ADAgECAhEAlRc7rgkYskDa/lxWVs/dLzAKBggqhkjOPQQDAjARMQ8w
DQYDVQQDEwZzaWduZXIwHhcNMTgwMzA3MTkzNTM3WhcNMjgwMTE0MTkzNTM3WjAR
MQ8wDQYDVQQDEwZzaWduZXIwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARNXtH6
Oppa77mBMd5FJV+lkCPG7BQlOWIxWDw0UoefDGR34lCu1Dv9aZRLwb9VSMw/VLMp
Q2wJTNZuzYeGo8XmoyAwHjAOBgNVHQ8BAf8EBAMCB4AwDAYDVR0TAQH/BAIwADAK
BggqhkjOPQQDAgNHADBEAiAZk088o0RxnDNnixJceFqlKWBErpGLNH1K1rZpcpk2
kQIgSgmXP0fMXE3JhAAa70npHrptiUKFedU631t1ebfbs/E=
-----END CERTIFICATE-----`)

var signerKey = []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIBL+5RFSepzRuQi/qLhUKp9JZvNqjuXZ1WJH3eNZJJ3GoAoGCCqGSM49
AwEHoUQDQgAETV7R+jqaWu+5gTHeRSVfpZAjxuwUJTliMVg8NFKHnwxkd+JQrtQ7
/WmUS8G/VUjMP1SzKUNsCUzWbs2HhqPF5g==
-----END EC PRIVATE KEY-----`)

var wrongSignerKey = []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEILv+8L9D/fyQIg2t+y8+abHpKBjgr+NOd1ykmTdeYdE1oAoGCCqGSM49
AwEHoUQDQgAEyPxsSGqLEH6yyKemBOCgED1y/0voTiAPQs0aRSi+Uto0ParJC+AN
zXSz0haUGzMJoobuLTgnninur98NJhPftg==
-----END EC PRIVATE KEY-----`)

func cert(data []byte) *x509.Certificate {

	b, _ := pem.Decode(data)
	cert, err := x509.ParseCertificate(b.Bytes)
	if err != nil {
		panic(err)
	}

	return cert
}

func key(data []byte) crypto.PrivateKey {

	b, _ := pem.Decode(data)
	k, err := x509.ParseECPrivateKey(b.Bytes)
	if err != nil {
		panic(err)
	}

	return k
}

func makeToken(claims jwt.Claims, signMethod jwt.SigningMethod, key crypto.PrivateKey) string {

	token := jwt.NewWithClaims(signMethod, claims)
	t, err := token.SignedString(key)
	if err != nil {
		panic(err)
	}

	return t
}

func TestVerifyToken(t *testing.T) {

	Convey("Given I verify a valid token", t, func() {

		token := makeToken(
			&jwt.StandardClaims{Subject: "sub"},
			jwt.SigningMethodES256,
			key(signerKey),
		)

		claims, err := VerifyToken(token, cert(signerCert))

		Convey("Then err should be nil", func() {
			So(err, ShouldBeNil)
		})

		Convey("Then claims should be correct", func() {
			So(claims.Valid(), ShouldBeNil)
		})
	})

	Convey("Given I verify a valid token with wrong signature", t, func() {

		token := makeToken(
			&jwt.StandardClaims{Subject: "sub"},
			jwt.SigningMethodES256,
			key(wrongSignerKey),
		)

		claims, err := VerifyToken(token, cert(signerCert))

		Convey("Then err should be nil", func() {
			So(err, ShouldNotBeNil)
		})

		Convey("Then claims should be nil", func() {
			So(claims, ShouldBeNil)
		})
	})
}

func TestVerifyTokenSignature(t *testing.T) {

	Convey("Given I verify a valid token", t, func() {

		token := makeToken(
			&jwt.StandardClaims{Subject: "sub"},
			jwt.SigningMethodES256,
			key(signerKey),
		)

		claims, err := VerifyTokenSignature(token, cert(signerCert))

		Convey("Then err should be nil", func() {
			So(err, ShouldBeNil)
		})

		Convey("Then claims should be correct", func() {
			So(claims, ShouldResemble, []string{"@auth:subject=sub"})
		})
	})

	Convey("Given I verify a valid token with wrong signature", t, func() {

		token := makeToken(
			&jwt.StandardClaims{Subject: "sub"},
			jwt.SigningMethodES256,
			key(wrongSignerKey),
		)

		claims, err := VerifyTokenSignature(token, cert(signerCert))

		Convey("Then err should be nil", func() {
			So(err, ShouldNotBeNil)
		})

		Convey("Then claims should be nil", func() {
			So(claims, ShouldBeNil)
		})
	})
}
