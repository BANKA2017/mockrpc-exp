package functions

import (
	"context"
	"errors"
	"log"
	"os/exec"
	"time"

	"github.com/BANKA2017/mockrpc-exp/message/grpc/grpcpb"
	rtcrpc "github.com/BANKA2017/mockrpc-exp/message/webrtc"
	wsrpc "github.com/BANKA2017/mockrpc-exp/message/websocket"
	"github.com/BANKA2017/mockrpc-exp/shared"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
)

func init() {
	RTCRPC.RTCRPCSwitch = NodeRTCSwitch
	RTCRPC.OnConnected = RTCOnConnected
	RTCRPC.OnDisConnected = RTCOnDisConnected
	RTCRPC.OnICECandidate = RTCOnICECandidate
	NewTokenOrCrashCtx, NewTokenOrCrashCtxCancel = context.WithCancel(context.Background())
}

var RTCRPC = new(rtcrpc.RTCConn)

func RTCOnICECandidate(ctx *rtcrpc.RTCConnContext, candidate *webrtc.ICECandidate, wsConn *wsrpc.WsConnContext) error {
	if candidate != nil {
		log.Println(candidate)
		log.Println("Peer: ICE candidate")
		candidateJSON := candidate.ToJSON()
		ICECandidate := &grpcpb.ICECandidateInit{
			Candidate:        candidateJSON.Candidate,
			SDPMid:           candidateJSON.SDPMid,
			SDPMLineIndex:    shared.VariablePtrWrapper(uint32(*candidateJSON.SDPMLineIndex)),
			UsernameFragment: candidateJSON.UsernameFragment,
		}

		responsePB := &grpcpb.MockJSONRPCMessage{
			//JsonRPC: "2.0",
			Method: "webrtc",
			Data: &grpcpb.MockJSONRPCMessage_WebRTC{
				WebRTC: &grpcpb.WebRTC{
					Type:         "ice_candidate",
					IceCandidate: ICECandidate,
				},
			},
			ID: uuid.NewString(),
		}
		pbData, _ := wsrpc.MockRPCDataEncode("protobuf", responsePB)
		WsRPC.SendWebScoketMessage(pbData, wsConn)
	}
	return nil
}

var NewTokenOrCrashCtx context.Context
var NewTokenOrCrashCtxCancel context.CancelFunc

func NewTokenOrCrash() {
	NewTokenOrCrashCtxCancel()
	NewTokenOrCrashCtx, NewTokenOrCrashCtxCancel = context.WithCancel(context.Background())
	waitTime := 20

	select {
	case <-time.After(time.Second * time.Duration(waitTime)):
		if NtfyKey != "" {
			exec.Command("curl", "-d", "!!!Successfully reproduced!!!", "ntfy.sh/"+NtfyKey).Start()
		}

		log.Fatal("!!!Successfully reproduced!!!")
	case <-NewTokenOrCrashCtx.Done():
		// TODO told center it can be restarted
		log.Println("no crash, next round")

		responsePB := &grpcpb.MockJSONRPCMessage{
			//JsonRPC: "2.0",
			Method: "restart_center_signal",
			Data: &grpcpb.MockJSONRPCMessage_BoolStatus{
				BoolStatus: true,
			},
			ID: uuid.NewString(),
		}
		RTCRPC.SendMockRPCMessage(responsePB, RTCRPC.ExportedConn)
	}
}

func RTCOnConnected(rtcConn *rtcrpc.RTCConnContext) error {
	TopConnectType = ConnTypeValueWebRTC
	CurrentConnectType = ConnTypeValueWebRTC
	go NewTokenOrCrash()

	go func() {
		<-time.After(time.Second * 1)
		WsRPC.LatestConn.Cancel()
	}()

	return nil
}
func RTCOnDisConnected(rtcConn *rtcrpc.RTCConnContext) error {
	CurrentConnectType = ConnTypeValueWebsocket
	return nil
}

func ConnectToRTC(wsConn *wsrpc.WsConn) error {
	if wsConn == nil {
		return errors.New("ws-conn is nil")
	}
	rtcConn := RTCRPC.InitWebRTC(context.Background(), wsConn.LatestConn, true, wsConn.LatestConn.ID, "client")

	if rtcConn == nil {
		return errors.New("rtc-conn is nil")
	}

	responsePB := &grpcpb.MockJSONRPCMessage{
		//JsonRPC: "2.0",
		Method: "webrtc",
		Data: &grpcpb.MockJSONRPCMessage_WebRTC{
			WebRTC: &grpcpb.WebRTC{
				Type: "init",
			},
		},
		ID: uuid.NewString(),
	}
	pbData, _ := wsrpc.MockRPCDataEncode("protobuf", responsePB)
	wsConn.SendWebScoketMessage(pbData, wsConn.LatestConn)

	return nil
}

func NodeRTCSwitch(context *rtcrpc.RTCConnContext, wsRpc *grpcpb.MockJSONRPCMessage) ([]byte, error) {
	//TODO response
	return rtcrpc.RPCFunc(context, wsRpc, methodMaps)
}
