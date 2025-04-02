package functions

import (
	"github.com/BANKA2017/mockrpc-exp/message/grpc/grpcpb"
)

type ConnTypeValueType int

const (
	ConnTypeValueWebsocket ConnTypeValueType = iota
	ConnTypeValueWebRTC
)

var TopConnectType = ConnTypeValueWebsocket

var CurrentConnectType = ConnTypeValueWebsocket

func GetCurrentConnStatus() bool {
	if CurrentConnectType == ConnTypeValueWebRTC {
		return RTCRPC.ExportedConn != nil
	}
	return WsRPC.LatestConn != nil
}

func SendRPCMessage(message *grpcpb.MockJSONRPCMessage) error {
	if CurrentConnectType == ConnTypeValueWebRTC {
		return RTCRPC.SendMockRPCMessage(message, RTCRPC.ExportedConn)
	}
	return WsRPC.SendMockRPCMessage(message, WsRPC.LatestConn)
}
