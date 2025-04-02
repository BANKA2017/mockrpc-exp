package functions

import (
	"github.com/BANKA2017/mockrpc-exp/message/grpc/grpcpb"
	rtcrpc "github.com/BANKA2017/mockrpc-exp/message/webrtc"
	wsrpc "github.com/BANKA2017/mockrpc-exp/message/websocket"
	"github.com/BANKA2017/mockrpc-exp/shared"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
)

func init() {
	RTCRPC.RTCRPCSwitch = CenterRTCSwitch
	RTCRPC.OnConnected = RTCOnConnected
	RTCRPC.OnDisConnected = RTCOnDisConnected
	RTCRPC.OnICECandidate = RTCOnICECandidate
}

var RTCRPC = new(rtcrpc.RTCConn)

func RTCOnICECandidate(ctx *rtcrpc.RTCConnContext, candidate *webrtc.ICECandidate, wsConn *wsrpc.WsConnContext) error {
	if candidate != nil {
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

func RTCOnConnected(rtcConn *rtcrpc.RTCConnContext) error {
	return nil
}
func RTCOnDisConnected(rtcConn *rtcrpc.RTCConnContext) error {
	return nil
}

func CenterRTCSwitch(context *rtcrpc.RTCConnContext, wsRpc *grpcpb.MockJSONRPCMessage) ([]byte, error) {
	//TODO response
	switch context.ConnType {
	case "node":
		return rtcrpc.RPCFunc(context, wsRpc, nodeMethodMaps)
	default:
		return rtcrpc.RPCFunc(context, wsRpc, map[string]wsrpc.TypeRPCFunc{})
	}
}
