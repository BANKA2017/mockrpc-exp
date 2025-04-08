package functions

import (
	"context"
	"errors"
	"log"
	"slices"
	"strconv"
	"time"

	"github.com/BANKA2017/mockrpc-exp/message/grpc/grpcpb"
	rtcrpc "github.com/BANKA2017/mockrpc-exp/message/webrtc"
	wsrpc "github.com/BANKA2017/mockrpc-exp/message/websocket"
	"github.com/BANKA2017/mockrpc-exp/shared"
	"github.com/pion/webrtc/v4"
)

func init() {
	WsRPC.InitWebsocketUpgrader()
	WsRPC.WsRPCSwitch = CenterWsRPCSwitch
	WsRPC.OnConnected = WsRPCOnConnected
	WsRPC.OnDisConnected = WsRPCOnDisConnected
}

var WsRPC = new(wsrpc.WsConn)

func WsRPCOnConnected(wsConn *wsrpc.WsConnContext) error {
	return nil
}
func WsRPCOnDisConnected(wsConn *wsrpc.WsConnContext) error {
	return nil
}

var nodeMethodMaps = map[string]wsrpc.TypeRPCFunc{
	"restart_center_signal": restart_center_signal,
	"push_node_status":      centerPushNodeStatus,
	"ack":                   ack,
	"webrtc":                _webrtc,
}

func CenterWsRPCSwitch(context *wsrpc.WsConnContext, wsRpc *grpcpb.MockJSONRPCMessage) ([]byte, error) {
	//TODO response
	switch context.ConnType {
	case "node":
		return wsrpc.RPCFunc(context, wsRpc, nodeMethodMaps)
	default:
		return wsrpc.RPCFunc(context, wsRpc, map[string]wsrpc.TypeRPCFunc{})
	}
}

func centerPushNodeStatus(context *wsrpc.WsConnContext, wsRpc *grpcpb.MockJSONRPCMessage) ([]byte, error) {
	//log.Println("center: receive new node status")
	in := wsRpc.GetServer()
	if in == nil {
		return []byte{}, errors.New("invalid request")
	}
	log.Println(context.Addr, "->", "send:", in.SendCount, ", active:", in.ActiveCount, ", done:", in.DoneCount, ", failed:", in.FailedCount, ", goroutine:", in.GoroutineCount, ", interval:", in.Interval, ", start_at:", time.Unix(in.StartAt, 0))
	// log.Println(context, in)
	return wsrpc.MockRPCDataEncode("protobuf", &grpcpb.MockJSONRPCMessage{
		//JsonRPC: "2.0",
		Method: "push_node_status_ack",
		Data: &grpcpb.MockJSONRPCMessage_BoolStatus{
			BoolStatus: true,
		},
		ID: wsRpc.ID,
	})
}

func restart_center_signal(context *wsrpc.WsConnContext, wsRpc *grpcpb.MockJSONRPCMessage) ([]byte, error) {
	//log.Println("center: receive new node status")

	return []byte{}, nil
}

func ack(context *wsrpc.WsConnContext, wsRpc *grpcpb.MockJSONRPCMessage) ([]byte, error) {
	// do nothing... right?

	return []byte{}, nil
}

func _webrtc(wsctx *wsrpc.WsConnContext, wsRpc *grpcpb.MockJSONRPCMessage) ([]byte, error) {
	nodeID := wsctx.ID
	signal := wsRpc.GetWebRTC()
	if signal == nil {
		return []byte{}, errors.New("invalid request")
	}

	if signal.Type == "ack" {
		return []byte{}, nil
	}

	if !slices.Contains([]string{"init", "answer", "ice_candidate"}, signal.Type) {
		return []byte{}, errors.New("invalid signal type")
	}

	var res = &grpcpb.WebRTC{
		Type: "ack",
	}

	if signal.Type == "init" {
		rtcctx := RTCRPC.InitWebRTC(context.Background(), wsctx, false, nodeID, "node")
		if rtcctx != nil {
			r, err := rtcctx.CreateOffer()
			if err != nil {
				return []byte{}, err
			}
			res.Type = r.Type
			res.SDP = r.Sdp
		} else {
			return []byte{}, errors.New("init webrtc peer/channel failed")
		}
	} else if _ctx, ok := RTCRPC.WebRTCConnPool.Load(shared.AppendStrings("node:", strconv.Itoa(int(nodeID)))); ok && signal.Type != "ack" {
		ctx := _ctx.(*rtcrpc.RTCConnContext)
		switch signal.Type {
		// case "offer":
		// 	ctx.SetRemoteDescription(&rtcrpc.RTCSignal{
		// 		Type: signal.Type,
		// 		Sdp:  signal.Sdp,
		// 	})
		// 	r, err := ctx.CreateAnswer()
		// 	if err != nil {
		// 		return []byte{}, err
		// 	}
		// 	res.Type = r.Type
		// 	res.Sdp = r.Sdp
		case "answer":
			ctx.SetRemoteDescription(&rtcrpc.RTCSignal{
				Type: signal.Type,
				Sdp:  signal.SDP,
			})
			// r, err := ctx.CreateAnswer()
			// if err != nil {
			// 	return []byte{}, err
			// }
			// res.Type = r.Type
			// res.Sdp = r.Sdp
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
