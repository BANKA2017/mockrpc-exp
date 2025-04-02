package httpapi

import (
	"log"
	"net/http"
	"time"

	"github.com/BANKA2017/mockrpc-exp/center/functions"
	"github.com/BANKA2017/mockrpc-exp/shared"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

func SetHeaders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if shared.TestMode {
			c.Response().Header().Add("Access-Control-Allow-Origin", "*")
		} else if functions.AllowOrigin != "" {
			c.Response().Header().Add("Access-Control-Allow-Origin", functions.AllowOrigin)
		}

		c.Response().Header().Add("X-Powered-By", "twitter monitor pro")
		c.Response().Header().Add("Access-Control-Allow-Methods", "*")
		// c.Response().Header().Add("Access-Control-Allow-Credentials", "true")
		c.Response().Header().Add("Access-Control-Allow-Headers", "Authorization")
		return next(c)
	}
}

func PreCheckAuthorization(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("request_date", time.Now())
		method := c.Request().Method
		path := c.Path()
		log.Println(method, path, c.Request().URL.Path, c.QueryString())

		if PreCheckPassListExists(path) {
			return next(c)
		}

		// <- force UID:1
		c.Set("id", "6")
		c.Set("role", "node")
		c.Set("uid", "1")

		return next(c)
	}
}

func RateLimit(_rate int, expiersIn time.Duration) echo.MiddlewareFunc {
	config := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{Rate: rate.Limit(_rate), Burst: 0, ExpiresIn: expiersIn},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return context.JSON(http.StatusServiceUnavailable, apiTemplate(503, "service unavailable", echoEmptyObject, "tmpro"))
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.String(http.StatusTooManyRequests, "")
		},
	}

	return middleware.RateLimiterWithConfig(config)
}
