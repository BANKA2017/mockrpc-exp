package httpapi

import (
	"net/http"
	"slices"
	"time"

	"github.com/BANKA2017/mockrpc-exp/center/functions"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

var echoEmptyObject = make(map[string]any, 0)

type ApiTemplate struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Version string `json:"version"`
}

func apiTemplate[T any](code int, message string, data T, version string) ApiTemplate {
	return ApiTemplate{
		Code:    code,
		Message: message,
		Data:    data,
		Version: version,
	}
}

func echoReject(c echo.Context) error {
	return c.JSON(http.StatusForbidden, apiTemplate(403, "Invalid requests", echoEmptyObject, "tmpro"))
}

func echoNoContent(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func echoRobots(c echo.Context) error {
	return c.String(http.StatusOK, "User-agent: *\nDisallow: /*")
}

var PreCheckPassList = []string{
	"/api/account/token",
}

func PreCheckPassListExists(path string) bool {
	return slices.Contains(PreCheckPassList, path)
}

func jwtBuilder(uid string, expire int, aud string) string {
	// expire
	if expire > 30*24*60*60 {
		// 30 days
		expire = 10 * 24 * 60 * 60
	} else if expire < 30 {
		// 30 seconds
		expire = 30
	}

	// Create the Claims
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   uid,
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expire) * time.Second)),
		Audience:  []string{aud},
		Issuer:    "tmpro",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	ss, _ := token.SignedString(functions.HttpAuthBearerTokenKey)

	return ss
}
