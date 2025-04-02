package httpapi

import (
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type AuthorizationColumn struct {
	Authorization string `gorm:"column:authorization;not null" json:"authorization"`
}

type tokenResponse struct {
	Type   string `json:"type"`
	Token  string `json:"token"`
	Expire int64  `json:"expire,omitempty"`
}

func AuthGetAccessToken(c echo.Context) error {
	// TODO get token from authorization?
	token := strings.Split(strings.TrimSpace(c.FormValue("token")), ":")
	// TODO static target
	if len(token) != 3 || !slices.Contains([]string{"node"}, token[0]) {
		return c.JSON(http.StatusOK, apiTemplate(401, "Invalid token", echoEmptyObject, "tmpro"))
	}

	numServiceID, err := strconv.ParseInt(token[1], 10, 64)
	if err != nil || numServiceID <= 0 {
		return c.JSON(http.StatusOK, apiTemplate(401, "Invalid token", echoEmptyObject, "tmpro"))
	}

	bearerToken := jwtBuilder(strconv.Itoa(int(numServiceID)), 30, token[0])
	log.Println(bearerToken)
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", tokenResponse{
		Type:   "bearer",
		Token:  bearerToken,
		Expire: time.Now().Add(time.Second * 30).Unix(),
	}, "tmpro"))
}
