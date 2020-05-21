module go.aporeto.io/midgard-lib

go 1.13

require (
	go.aporeto.io/elemental v1.100.1-0.20200507180645-f7ef7a598da7
	go.aporeto.io/gaia v1.94.1-0.20200617164623-a2fa2783eb05
	go.aporeto.io/tg v1.34.1-0.20200407170715-afab00a55eba
)

require (
	cloud.google.com/go v0.53.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/opentracing/opentracing-go v1.1.0
	github.com/smartystreets/goconvey v1.6.4
	go.uber.org/zap v1.14.0
	golang.org/x/tools v0.0.0-20200226171234-020676185e75 // indirect
)

replace go.aporeto.io/gaia => go.aporeto.io/gaia v1.94.1-0.20200520061514-ef2c396bd7c2
