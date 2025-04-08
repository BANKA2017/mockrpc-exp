package wsrpc

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/BANKA2017/mockrpc-exp/message/grpc/grpcpb"
	"github.com/BANKA2017/mockrpc-exp/shared"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var Protocols = []string{"json", "protobuf"}

func PassOriginCheck(r *http.Request) bool {
	return true
}

func (context *WsConnContext) Close(noCheck bool) {
	if !noCheck {
		<-context.Ctx.Done()
	}
	context.WsRWMutex.Lock()
	defer context.WsRWMutex.Unlock()
	connID := shared.AppendStrings(context.ConnType, ":", strconv.Itoa(int(context.ID)))
	if !noCheck {
		context.Ext.WebsocketConnPool.Delete(connID)
	}
	log.Println(connID, context.Conn.RemoteAddr().String(), "close")
	if !context.IsClosed {
		context.IsClosed = true
		close(context.InChan)
		close(context.OutChan)
		context.Conn.Close()
	}

	if context.Export {
		context.Ext.LatestConn = nil
	}
	context.Ext.OnDisConnected(context)
}

func (context *WsConnContext) InLoop() {
	var (
		messageType int
		message     []byte
		err         error
	)

	for {
		messageType, message, err = context.Conn.ReadMessage()
		if err != nil || messageType == -1 {
			context.Cancel()
			log.Println("read:", err, context)
			return
		} else if slices.Compare(message, context.ExtLatestInMessage) == 0 {
			continue
		}
		context.ExtLatestInMessage = message
		select {
		case <-context.Ctx.Done():
			log.Println(context.Conn.RemoteAddr().String(), "in loop close")
			return
		case context.InChan <- message:
			go context.ReadIn()
		}
	}
}

func (context *WsConnContext) ReadIn() {
	message, ok := <-context.InChan
	if ok && !context.IsClosed {
		var response []byte
		var wsRpc = new(grpcpb.MockJSONRPCMessage)
		var err error
		switch context.ConnType {
		case "node", "client":
			err = MockRPCDataDecode(context.Protocol, message, wsRpc)
			if err != nil {
				log.Println(err)
				response, _ = MockRPCDataEncode(context.Protocol, &grpcpb.MockJSONRPCMessage{
					Error: &grpcpb.MockJSONRPCError{
						Code:    -32700,
						Message: "Parse error",
					},
					ID: wsRpc.ID,
				})
				// json.Marshal(MockWebsocketJsonRPC{
				// 	JsonRPC: "2.0",
				// 	Error: &MockWebsocketJsonRPCError{
				// 		Code:    -32700,
				// 		Message: "Parse error",
				// 	},
				// 	ID: wsRpc.ID,
				// })
			} else {

				// if !context.IsLogin && wsRpc.Method != "login" {
				// 	context.CloseChan <- 1
				// 	return
				// }

				// node only
				response, err = context.Ext.WsRPCSwitch(context, wsRpc)

				log.Println(wsRpc)
			}
		}

		if err != nil {
			log.Println(err)
		}
		if len(response) > 0 {
			log.Println(response)
			context.WriteOut(response)
		}
	}
}

func (context *WsConnContext) OutLoop() {
	var err error
	for {
		select {
		case <-context.Ctx.Done():
			log.Println(context.Conn.RemoteAddr().String(), "out loop close")
			return
		case message, ok := <-context.OutChan:
			if ok {
				context.WsRWMutex.Lock()
				err = context.Conn.WriteMessage(websocket.TextMessage, message)
				context.WsRWMutex.Unlock()
				if err != nil {
					context.Cancel()
					log.Println("write:", err)
					return
				}
			}
		}
	}
}

func (context *WsConnContext) WriteOut(data []byte) {
	if !context.IsClosed {
		context.OutChan <- data
	}
}

func (wsconn *WsConn) InitWebSocket(_ctx context.Context, c *websocket.Conn, export bool, nodeID int32, connType string, protocol string) *WsConnContext {
	ctx, cancel := context.WithCancel(_ctx)
	wsConnContext := &WsConnContext{
		Conn:    c,
		InChan:  make(chan []byte, 1000),
		OutChan: make(chan []byte, 1000),
		// CloseChan: make(chan byte, 1),
		Export:   export,
		Addr:     c.RemoteAddr().String(),
		ID:       nodeID,
		ConnType: connType,
		// IsLogin:   autoLogin,
		Ext:      wsconn,
		Ctx:      ctx,
		Cancel:   cancel,
		Protocol: MockRPCDataProtocol(protocol),
	}

	if wsConnContext.Export {
		wsconn.LatestConn = wsConnContext
	}

	connKey := shared.AppendStrings(wsConnContext.ConnType, ":", strconv.Itoa(int(wsConnContext.ID)))

	// kick out exists connect
	times := 20
	for {
		if existsConn, ok := wsconn.WebsocketConnPool.Load(connKey); ok && times > 0 {
			existsConn.(*WsConnContext).Cancel()
			log.Println("wsconn: kick out exists conn", connKey)
			time.Sleep(time.Millisecond)
		} else if times <= 0 {
			log.Println("connect-to-websocket: times exhausted")
			wsConnContext.Close(true)
			return nil
		} else {
			break
		}
		times--
	}

	wsconn.WebsocketConnPool.Store(connKey, wsConnContext)

	go wsConnContext.InLoop()
	go wsConnContext.OutLoop()
	go wsConnContext.Close(false)
	wsConnContext.Ext.OnConnected(wsConnContext)
	return wsConnContext
}

func (wsconn *WsConn) InitWebsocketUpgrader() {
	wsconn.WsUpgrader = websocket.Upgrader{
		//ReadBufferSize:  256,
		//WriteBufferSize: 256,
		CheckOrigin:  PassOriginCheck,
		Subprotocols: Protocols,
	}
}

func (wsconn *WsConn) WebsocketServer(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	nodeID := ctx.Value("tmpro_ws_node_id").(int32)
	connType := ctx.Value("tmpro_ws_conn_type").(string)

	c, err := wsconn.WsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		//log.Print("upgrade:", err)
		c.Close()
		return err
	}
	//defer c.Close()
	log.Println(c.RemoteAddr().String(), "connected")

	go wsconn.InitWebSocket(context.Background(), c, false, nodeID, connType, c.Subprotocol())
	return nil
}

func (wsconn *WsConn) WebsocketClient(ctx context.Context, url string, authorization string, protocol string) error {
	var requestHeader = make(http.Header)
	requestHeader.Set("Sec-WebSocket-Protocol", MockRPCDataProtocol(protocol))
	if authorization != "" {
		requestHeader.Set("Authorization", "Bearer "+authorization)
	}

	c, _, err := websocket.DefaultDialer.Dial(url, requestHeader)
	if err != nil {
		log.Println("ws:", err)
		return err
	} else {
		//defer c.Close()
		log.Println(c.RemoteAddr().String(), "connected")

		wsconn.InitWebSocket(ctx, c, true, -1, "client", protocol)
	}
	return nil
}

func (wsconn *WsConn) SendMockRPCMessage(message *grpcpb.MockJSONRPCMessage, wsConnContext *WsConnContext) error {
	if wsConnContext == nil {
		return errors.New("ws: wsconn is nil")
	}
	data, err := MockRPCDataEncode(wsConnContext.Protocol, message)
	if err != nil {
		return err
	}
	return wsconn.SendWebScoketMessage(data, wsConnContext)
}

func (wsconn *WsConn) SendWebScoketMessage(data []byte, wsConnContext *WsConnContext) error {
	if wsConnContext != nil && !wsConnContext.IsClosed {
		wsConnContext.OutChan <- data
	} else if wsConnContext == nil {
		log.Println("ws: wsconn is nil!", string(data))
	}
	return nil
}

func (wsconn *WsConn) BroadcastToWebSocket(data []byte, connTypeFilter string) {
	wsconn.WebsocketConnPool.Range(func(k any, v any) bool {
		conn := (v).(*WsConnContext)
		if connTypeFilter != "" && conn.ConnType != connTypeFilter {
			return true
		}
		log.Print(k.(string), data)
		wsconn.SendWebScoketMessage(data, conn)
		return true
	})
}

func RPCFunc(context *WsConnContext, wsRpc *grpcpb.MockJSONRPCMessage, methodMaps map[string]TypeRPCFunc) ([]byte, error) {
	if handler, exists := methodMaps[wsRpc.Method]; exists {
		return handler(context, wsRpc)
	}

	return MockRPCDataEncode(context.Protocol, &grpcpb.MockJSONRPCMessage{
		Error: &grpcpb.MockJSONRPCError{
			Code:    -32601,
			Message: "method not found",
		},
		ID: wsRpc.ID,
	})
	//return json.Marshal(MockWebsocketJsonRPC{
	//	JsonRPC: "2.0",
	//	Error: &MockWebsocketJsonRPCError{
	//		Code:    -32601,
	//		Message: "method not found",
	//	},
	//	ID: wsRpc.ID,
	//})
}

func MockRPCDataDecode(protocol string, data []byte, template proto.Message) error {
	if protocol == "json" {
		return protojson.Unmarshal(data, template)
	} else {
		return proto.Unmarshal(data, template)
	}
}

func MockRPCDataEncode(protocol string, _struct proto.Message) ([]byte, error) {
	if protocol == "json" {
		return json.Marshal(_struct)
	} else {
		return proto.Marshal(_struct)
	}
}

func MockRPCDataProtocol(protocol string) string {
	if !slices.Contains(Protocols, protocol) {
		protocol = "json"
	}
	return protocol
}

var SyncRPCCallbackChannelMap sync.Map // id -> SyncWsRPCType

type MockRPCConn interface {
	SendMockRPCMessage(message *grpcpb.MockJSONRPCMessage, _ctx any) error
}

func SendSyncRPCMessage[T MockRPCConn](message *grpcpb.MockJSONRPCMessage, conn T, _ctx any) (*grpcpb.MockJSONRPCMessage, error) {
	if message == nil {
		return nil, errors.New("mockrpc: message is nil")
	}
	var id string
	if message.ID != "" {
		id = message.ID
	} else {
		id = uuid.NewString()
	}

	ctx, cancel := context.WithCancel(context.Background())

	syncRPCStruct := &SyncWsRPCType{
		Ctx:       ctx,
		ID:        id,
		Chan:      make(chan *grpcpb.MockJSONRPCMessage),
		ExpiresAt: time.Now().Add(time.Second * 10).Unix(),
		Timeout:   time.After(time.Second * 11),
	}

	SyncRPCCallbackChannelMap.Store(id, syncRPCStruct)
	defer func() {
		SyncRPCCallbackChannelMap.Delete(id)
		close(syncRPCStruct.Chan)
		cancel()
	}()

	err := conn.SendMockRPCMessage(message, _ctx)

	if err != nil {
		return nil, err
	}
	select {
	case response := <-syncRPCStruct.Chan:
		return response, nil
	case <-syncRPCStruct.Timeout:
		return nil, errors.New("mockrpc: timeout")
	}
}

func ReceiveSyncRPCMessage(message *grpcpb.MockJSONRPCMessage) (*grpcpb.MockJSONRPCMessage, error) {
	if message == nil {
		return message, errors.New("mockrpc: message is nil")
	}
	id := message.GetID()
	if id == "" {
		return message, errors.New("mockrpc: broadcast message")
	}
	if syncRPCStruct, ok := SyncRPCCallbackChannelMap.Load(id); ok && syncRPCStruct != nil {
		s, ok := syncRPCStruct.(*SyncWsRPCType)
		if !ok {
			return message, errors.New("mockrpc: invalid channel map type")
		}
		if s.ExpiresAt <= time.Now().Unix() {
			return message, errors.New("mockrpc: timeout")
		}
		s.Chan <- message
		return message, nil
	} else {
		return message, errors.New("mockrpc: not a sync message")
	}
}
