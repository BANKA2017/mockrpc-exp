package functions

import (
	"time"
)

var UpdateServerStatusTicker *time.Ticker

// 5, 10, 20, 40, 80, 160
const WsConnReconnectSinceTicker = 5
const WsConnReconnectMaxTicker = 160

var WsConnReconnectTicker time.Duration = WsConnReconnectSinceTicker

var WsConnReconnect *time.Ticker

func InitTicker() {
	UpdateServerStatusTicker = time.NewTicker(time.Second * 1) // Minute
	WsConnReconnect = time.NewTicker(time.Second * WsConnReconnectTicker)
}
