package httpapi

import (
	"time"

	"github.com/labstack/echo/v4"
)

func Api(address string) {
	// api
	e := echo.New()
	e.IPExtractor = echo.ExtractIPFromXFFHeader(
		echo.TrustLoopback(false),   // e.g. ipv4 start with 127.
		echo.TrustLinkLocal(false),  // e.g. ipv4 start with 169.254
		echo.TrustPrivateNet(false), // e.g. ipv4 start with 10. or 192.168
	)
	//e.Use(middleware.Logger())
	e.Use(SetHeaders)

	e.Any("/*", echoReject)

	api := e.Group("/api")

	// pre-check
	api.Use(PreCheckAuthorization)

	// ws
	ws := api.Group("/ws")
	ws.Use(RateLimit(1, time.Second))
	ws.GET("", WebSocketWrapper)

	// WTF...
	api.OPTIONS("/*", echoNoContent)

	// account
	account := api.Group("/account")
	account.OPTIONS("/*", echoNoContent)
	account.POST("/token", AuthGetAccessToken, RateLimit(5, time.Second))
	e.Logger.Fatal(e.Start(address))
}
