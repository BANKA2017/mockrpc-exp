package wsrpc

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/BANKA2017/mockrpc-exp/message/grpc/grpcpb"
	"github.com/gorilla/websocket"
)

// TODO use real jsonrpc
// type MockWebsocketJsonRPCRequest struct {
// 	JsonRPC string   `json:"jsonrpc"`
// 	Method  string   `json:"method"`
// 	Params  []string `json:"params"`
// 	ID      string   `json:"id"`
// }

type MockWebsocketJsonRPC struct {
	JsonRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	// req
	Params []any `json:"params,omitempty"`
	// res
	Result any                        `json:"result,omitempty"`
	Error  *MockWebsocketJsonRPCError `json:"error,omitempty"`

	ID string `json:"id,omitempty"`
}

type MockWebsocketJsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

type WsRPCFunc func(*WsConnContext, *grpcpb.MockJSONRPCMessage) ([]byte, error)
type WsRPCVerifyFunc func(*http.Request) (int32, string, bool)

type WsConnContext struct {
	Conn      *websocket.Conn
	WsRWMutex sync.RWMutex
	InChan    chan []byte
	OutChan   chan []byte
	// CloseChan chan byte
	IsClosed bool
	Export   bool
	Addr     string
	ID       int32
	ConnType string

	Ext                *WsConn
	ExtLatestInMessage []byte // what this is?
	Ctx                context.Context
	Cancel             context.CancelFunc
	Protocol           string
	// IsLogin   bool
}

type TypeRPCFunc func(*WsConnContext, *grpcpb.MockJSONRPCMessage) ([]byte, error)

type WsConn struct {
	// variables
	LatestConn *WsConnContext
	// WsRPCAuthorizationVerifyFunc WsRPCVerifyFunc
	WsUpgrader        websocket.Upgrader
	WebsocketConnPool sync.Map //make(map[string]*WsConnContext)

	// funcs
	WsRPCSwitch    WsRPCFunc
	OnConnected    func(*WsConnContext) error
	OnDisConnected func(*WsConnContext) error
}

type WsConnHooks interface {
	WsRPCSwitch() WsRPCFunc

	OnConnected() error
	OnDisConnected() error
}

type SyncWsRPCType struct {
	ID        string
	Chan      chan *grpcpb.MockJSONRPCMessage
	ExpiresAt int64
	Timeout   <-chan time.Time
	Ctx       context.Context
}
