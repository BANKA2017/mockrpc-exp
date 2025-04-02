package rtcrpc

import (
	"sync"

	"github.com/BANKA2017/mockrpc-exp/message/grpc/grpcpb"
	wsrpc "github.com/BANKA2017/mockrpc-exp/message/websocket"
	"github.com/pion/webrtc/v4"
)

type RTCConnContext struct {
	Peer     *webrtc.PeerConnection
	Channel  *webrtc.DataChannel
	IsClosed bool
	Export   bool
	Addr     string
	ID       int32
	ConnType string

	Ext *RTCConn
	// Ctx    context.Context
	// Cancel context.CancelFunc
}
type TypeRPCFunc func(*RTCConnContext, *grpcpb.MockJSONRPCMessage) ([]byte, error)

type WsRPCFunc func(*RTCConnContext, *grpcpb.MockJSONRPCMessage) ([]byte, error)

type RTCConn struct {
	WebRTCConnPool sync.Map
	ExportedConn   *RTCConnContext

	// funcs
	RTCRPCSwitch   TypeRPCFunc
	OnICECandidate func(*RTCConnContext, *webrtc.ICECandidate, *wsrpc.WsConnContext) error
	OnConnected    func(*RTCConnContext) error
	OnDisConnected func(*RTCConnContext) error
}

type RTCSignal struct {
	Type         string                  `json:"type"`
	Sdp          string                  `json:"sdp"`
	ICECandidate webrtc.ICECandidateInit `json:"ice_candidate,omitempty"`
}
