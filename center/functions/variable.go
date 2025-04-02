package functions

import (
	"sync"
)

type HttpAuthRefreshTokenMapItemType int

const (
	HttpAuthRefreshTokenMapItemUID HttpAuthRefreshTokenMapItemType = iota
	HttpAuthRefreshTokenMapItemToken
)

type HttpAuthRefreshTokenMapItemStruct struct {
	Content string
	Type    HttpAuthRefreshTokenMapItemType
	Expire  int64
}

var HttpAuthRefreshTokenMap sync.Map // int -> *HttpAuthRefreshTokenMapItemStruct

var AllowOrigin string
