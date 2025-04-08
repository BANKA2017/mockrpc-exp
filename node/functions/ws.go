package functions

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/BANKA2017/mockrpc-exp/message/grpc/grpcpb"
	rtcrpc "github.com/BANKA2017/mockrpc-exp/message/webrtc"
	wsrpc "github.com/BANKA2017/mockrpc-exp/message/websocket"
	"github.com/BANKA2017/mockrpc-exp/shared"
	"github.com/pion/webrtc/v4"
)

func init() {
	WsRPC.WsRPCSwitch = NodeWsRPCSwitch
	WsRPC.OnConnected = OnConnected
	WsRPC.OnDisConnected = OnDisConnected
}

var WsCnnectTimeout = 5

type tokenResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Type   string `json:"type"`
		Token  string `json:"token"`
		Expire int    `json:"expire"`
	} `json:"data"`
	Version string `json:"version"`
}

func GetAccessToken(addr string, refreshToken string) string {
	_, _, resp, err := shared.Fetch(addr, "POST", nil, strings.NewReader(shared.AppendStrings("token=", refreshToken)), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}, map[string]string{})
	if err != nil {
		return ""
	}

	tokenResponse := new(tokenResponse)
	err = json.Unmarshal(resp, &tokenResponse)
	if err != nil {
		return ""
	}
	return tokenResponse.Data.Token
}

var WsRPC = new(wsrpc.WsConn)
var wsConnConnected = make(chan struct{}, 10)

func OnConnected(wsConn *wsrpc.WsConnContext) error {
	wsConnConnected <- struct{}{}
	return nil
}
func OnDisConnected(wsConn *wsrpc.WsConnContext) error {
	return nil
}

// TODO fix jam
func ConnectToWebsocket(addr string, token string) error {
	err := WsRPC.WebsocketClient(context.Background(), addr, token, "protobuf")
	if err != nil {
		return err
	}

	// timeout...
	ticker := time.After(time.Second * time.Duration(WsCnnectTimeout))

	select {
	case <-wsConnConnected:
		if WsConnReconnectTicker != WsConnReconnectSinceTicker {
			WsConnReconnectTicker = WsConnReconnectSinceTicker
			WsConnReconnect.Reset(time.Second * WsConnReconnectTicker)
		}
		return nil
	case <-ticker:
		return errors.New("connect-to-center-websocket: time out")
	}
	//wsrpc.SendWebScoketMessage([]byte("{\"jsonrpc\":\"2.0\",\"method\":\"login\",\"params\":[\""+shared.Key+"\"],\"id\":\""+strconv.FormatInt(rand.Int63(), 10)+"\"}"), wsrpc.WsConn)
}

var methodMaps = map[string]wsrpc.TypeRPCFunc{
	"push_node_status_ack": pushNodeStatusAck,
	"ack":                  ack,
	"webrtc":               _webrtc,
}

func NodeWsRPCSwitch(context *wsrpc.WsConnContext, wsRpc *grpcpb.MockJSONRPCMessage) ([]byte, error) {
	//TODO response
	return wsrpc.RPCFunc(context, wsRpc, methodMaps)
}

func pushNodeStatusAck(context *wsrpc.WsConnContext, wsRpc *grpcpb.MockJSONRPCMessage) ([]byte, error) {
	//log.Println("center: receive new node status")
	in := wsRpc.GetBoolStatus()

	log.Println(context, in)
	shared.DoneCount += 1

	if NewTokenOrCrashCtx.Err() == nil {
		NewTokenOrCrashCtxCancel()
	}

	return []byte{}, nil
}

func ack(context *wsrpc.WsConnContext, wsRpc *grpcpb.MockJSONRPCMessage) ([]byte, error) {
	// do nothing... right?

	return []byte{}, nil
}

func _webrtc(wsctx *wsrpc.WsConnContext, wsRpc *grpcpb.MockJSONRPCMessage) ([]byte, error) {
	signal := wsRpc.GetWebRTC()

	if signal.Type == "ack" {
		return []byte{}, nil
	}

	if !slices.Contains([]string{"offer", "ice_candidate"}, signal.Type) {
		return []byte{}, errors.New("invalid signal type")
	}

	var res = &grpcpb.WebRTC{
		Type: "ack",
	}

	if RTCRPC.ExportedConn != nil {
		ctx := RTCRPC.ExportedConn
		switch signal.Type {
		// case "init":
		// 	rtcctx := RTCRPC.InitWebRTC(context.Background(), wsctx, false, nodeID, "node")
		// 	if rtcctx != nil {
		// 		r, err := rtcctx.CreateOffer()
		// 		if err != nil {
		// 			return []byte{}, err
		// 		}
		// 		res.Type = r.Type
		// 		res.Sdp = r.Sdp
		// 	} else {
		// 		return []byte{}, errors.New("init webrtc peer/channel failed")
		// 	}
		case "offer":
			ctx.SetRemoteDescription(&rtcrpc.RTCSignal{
				Type: signal.Type,
				Sdp:  signal.SDP,
			})
			r, err := ctx.CreateAnswer()
			if err != nil {
				return []byte{}, err
			}
			res.Type = r.Type
			res.SDP = r.Sdp
		//case "answer":
		//	ctx.SetRemoteDescription(&rtcrpc.RTCSignal{
		//		Type: signal.Type,
		//		Sdp:  signal.Sdp,
		//	})
		// 	r, err := ctx.CreateAnswer()
		// 	if err != nil {
		// 		return []byte{}, err
		// 	}
		// 	res.Type = r.Type
		// 	res.Sdp = r.Sdp
		case "ice_candidate":
			err := ctx.AddICECandidate(&rtcrpc.RTCSignal{
				Type: signal.Type,
				ICECandidate: webrtc.ICECandidateInit{
					Candidate:        signal.IceCandidate.Candidate,
					SDPMid:           signal.IceCandidate.SDPMid,
					SDPMLineIndex:    shared.VariablePtrWrapper(uint16(*signal.IceCandidate.SDPMLineIndex)),
					UsernameFragment: signal.IceCandidate.UsernameFragment,
				},
			})
			if err != nil {
				return []byte{}, err
			}
		}
	}

	return wsrpc.MockRPCDataEncode("protobuf", &grpcpb.MockJSONRPCMessage{
		//JsonRPC: "2.0",
		Method: "webrtc",
		Data: &grpcpb.MockJSONRPCMessage_WebRTC{
			WebRTC: res,
		},
		ID: wsRpc.ID,
	})
}
