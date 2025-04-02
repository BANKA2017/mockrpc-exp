package httpapi

import (
	"context"
	"net/http"
	"strconv"

	"github.com/BANKA2017/mockrpc-exp/center/functions"
	"github.com/labstack/echo/v4"
)

func WebSocketWrapper(c echo.Context) error {
	id := c.Get("id").(string)
	role := c.Get("role").(string)
	uid := c.Get("uid").(string)

	numID, err := strconv.ParseInt(id, 10, 64)

	if err != nil || numID <= 0 {
		return c.JSON(http.StatusUnauthorized, apiTemplate(403, "Invalid ws connect", echoEmptyObject, "tmpro"))
	}

	ctx := context.WithValue(context.Background(), ("tmpro_ws_node_id"), int32(numID))
	ctx = context.WithValue(ctx, ("tmpro_ws_conn_type"), role)

	var numUID int64 = 0
	if uid != "_" && role == "push" {
		numUID, _ = strconv.ParseInt(uid, 10, 64)
	}

	ctx = context.WithValue(ctx, ("tmpro_ws_conn_uid"), numUID)

	return functions.WsRPC.WebsocketServer(ctx, c.Response().Writer, c.Request())
}
