package rtcrpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/BANKA2017/mockrpc-exp/message/grpc/grpcpb"
	wsrpc "github.com/BANKA2017/mockrpc-exp/message/websocket"
	"github.com/BANKA2017/mockrpc-exp/shared"
	"github.com/pion/webrtc/v4"
	"google.golang.org/protobuf/proto"
)

func (rtcContext *RTCConnContext) CreatePeerChannel(channelName string) error {
	dataChannel, err := rtcContext.Peer.CreateDataChannel(channelName, nil)
	if err != nil {
		return err
	}
	rtcContext.Channel = dataChannel
	return nil
}

func (rtcContext *RTCConnContext) CreatePeerConnection() error {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				// TODO more ICE server
				URLs: []string{
					"stun:stun.miwifi.com",
					"stun:stun.cloudflare.com:3478",
					"stun:stun.l.google.com:19302",
				},
			},
		},
	}
	var err error
	rtcContext.Peer, err = webrtc.NewPeerConnection(config)

	return err
}

func (rtcContext *RTCConnContext) CreateOffer() (*RTCSignal, error) {
	offer, err := rtcContext.Peer.CreateOffer(nil)
	if err != nil {
		return nil, err
	}

	err = rtcContext.Peer.SetLocalDescription(offer)
	if err != nil {
		return nil, err
	}

	signal := RTCSignal{
		Type: "offer",
		Sdp:  offer.SDP,
	}

	return &signal, nil
}

func (rtcContext *RTCConnContext) CreateAnswer() (*RTCSignal, error) {
	answer, err := rtcContext.Peer.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	err = rtcContext.Peer.SetLocalDescription(answer)
	if err != nil {
		return nil, err
	}

	signal := RTCSignal{
		Type: "answer",
		Sdp:  answer.SDP,
	}

	return &signal, nil
}

func (rtcContext *RTCConnContext) SetRemoteDescription(signal *RTCSignal) error {
	if signal == nil {
		return errors.New("webrtc-rpc: nil signal")
	}

	sdpType := webrtc.SDPTypeOffer
	if signal.Type == "answer" {
		sdpType = webrtc.SDPTypeAnswer
	}

	return rtcContext.Peer.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  signal.Sdp,
		Type: sdpType,
	})
}

func (rtcContext *RTCConnContext) AddICECandidate(signal *RTCSignal) error {
	if signal == nil {
		return errors.New("webrtc-rpc: nil signal")
	}
	if signal.Type != "ice_candidate" {
		return errors.New("webrtc-rpc: invalid signal type")
	}

	return rtcContext.Peer.AddICECandidate(signal.ICECandidate)
}

func (rtcContext *RTCConnContext) Close() error {
	if rtcContext.IsClosed {
		return nil
	}
	rtcContext.IsClosed = true
	connKey := shared.AppendStrings(rtcContext.ConnType, ":", strconv.Itoa(int(rtcContext.ID)))
	rtcContext.Ext.WebRTCConnPool.Delete(connKey)
	rtcContext.Ext.ExportedConn = nil
	if err := rtcContext.Ext.OnDisConnected(rtcContext); err != nil {
		log.Println(err)
	}
	if rtcContext.Channel != nil {
		if err := rtcContext.Channel.Close(); err != nil {
			log.Println(err)
		}
	}
	if rtcContext.Peer != nil {
		if err := rtcContext.Peer.Close(); err != nil {
			log.Println(err)
		}
	}

	return nil
}

func (rtcConn *RTCConn) InitWebRTC(_ctx context.Context, c *wsrpc.WsConnContext, export bool, nodeID int32, connType string) *RTCConnContext {
	rtcContext := &RTCConnContext{
		ID:       nodeID,
		ConnType: connType,
		Addr:     c.Addr,
		Export:   export,
		Ext:      rtcConn,
	}

	// TODO errors
	err := rtcContext.CreatePeerConnection()
	if err != nil {
		log.Println("webrtc-cerate-peer-conn:", err)
		return nil
	}
	err = rtcContext.CreatePeerChannel(fmt.Sprintf("tmpro_rx_tx_%s:%d", connType, nodeID))
	if err != nil {
		log.Println("webrtc-cerate-peer-channel:", err)
		rtcContext.Close()
		return nil
	}

	if rtcContext.Export {
		rtcConn.ExportedConn = rtcContext
	}

	connKey := shared.AppendStrings(rtcContext.ConnType, ":", strconv.Itoa(int(rtcContext.ID)))

	// kick out exists connect
	times := 20
	for {
		if existsConn, ok := rtcConn.WebRTCConnPool.Load(connKey); ok && times > 0 {
			existsConn.(*RTCConnContext).Close()
			log.Println("webrtc: kick out exists conn", connKey)
			time.Sleep(time.Millisecond)
		} else if times <= 0 {
			log.Println("connect-to-webrtc: times exhausted")
			rtcContext.Close()
			if rtcContext.Export {
				rtcConn.ExportedConn = nil
			}
			return nil
		} else {
			break
		}
		times--
	}

	rtcConn.WebRTCConnPool.Store(connKey, rtcContext)

	rtcContext.Peer.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			rtcConn.OnICECandidate(rtcContext, candidate, c)
		}
	})

	rtcContext.Peer.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Printf("Peer ICE Connection State Changed: %s [%s]\n", state, rtcContext.Addr)
		switch state {
		case webrtc.ICEConnectionStateConnected:
			rtcContext.Ext.OnConnected(rtcContext)
		case webrtc.ICEConnectionStateDisconnected, webrtc.ICEConnectionStateClosed, webrtc.ICEConnectionStateFailed:
			rtcContext.Close()
		}
	})

	rtcContext.Peer.OnDataChannel(func(dc *webrtc.DataChannel) {
		log.Println("Peer: DataChannel created")

		dc.OnError(func(err error) {
			log.Println(err)
		})

		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			log.Println("rtc is_closed:", rtcContext.IsClosed)

			if !rtcContext.IsClosed {
				var response []byte
				var wsRpc = new(grpcpb.MockJSONRPCMessage)
				var err error
				switch rtcContext.ConnType {
				case "node", "client":
					err = proto.Unmarshal(msg.Data, wsRpc)
					if err != nil {
						log.Println(err)
						response, _ = wsrpc.MockRPCDataEncode("protobuf", &grpcpb.MockJSONRPCMessage{
							Error: &grpcpb.MockJSONRPCError{
								Code:    -32700,
								Message: "Parse error",
							},
							ID: wsRpc.ID,
						})
					} else {
						response, err = rtcContext.Ext.RTCRPCSwitch(rtcContext, wsRpc)
						log.Println(wsRpc)
					}
				}

				if err != nil {
					log.Println(err)
				}
				if len(response) > 0 {
					log.Println("res log:", response)
					if err = rtcContext.Channel.Send(response); err != nil {
						log.Println("res err:", err, response)
					}
				}
			}
		})
	})

	return rtcContext
}

func (rtcConn *RTCConn) SendMockRPCMessage(message *grpcpb.MockJSONRPCMessage, rtcContext *RTCConnContext) error {
	data, err := wsrpc.MockRPCDataEncode("protobuf", message)
	if err != nil {
		return err
	}
	return rtcConn.SendRTCMessage(data, rtcContext)
}

func (rtcConn *RTCConn) SendRTCMessage(data []byte, rtcContext *RTCConnContext) error {
	if rtcContext != nil && !rtcContext.IsClosed {
		rtcContext.Channel.Send(data)
	} else if rtcContext == nil {
		log.Println("webrtc-rpc: rtcconn is nil!", string(data))
	}
	return nil
}

func (rtcConn *RTCConn) BroadcastToRTC(data []byte, connTypeFilter string) {
	rtcConn.WebRTCConnPool.Range(func(k any, v any) bool {
		conn := (v).(*RTCConnContext)
		if connTypeFilter != "" && conn.ConnType != connTypeFilter {
			return true
		}
		log.Print(k.(string), data)
		rtcConn.SendRTCMessage(data, conn)
		return true
	})
}

func RPCFunc(rtcContext *RTCConnContext, wsRpc *grpcpb.MockJSONRPCMessage, methodMaps map[string]wsrpc.TypeRPCFunc) ([]byte, error) {
	if handler, exists := methodMaps[wsRpc.Method]; exists {

		if wsRpc.Method == "restart_center_signal" {
			log.Println("disconnect in 1s...")

			go func() {
				<-time.After(time.Second * 1)
				rtcContext.Close()
				log.Println("disconnecting...")
			}()
		}

		// TODO fix wrapper
		return handler(&wsrpc.WsConnContext{
			Addr:     rtcContext.Addr,
			ID:       rtcContext.ID,
			ConnType: rtcContext.ConnType,
		}, wsRpc)
	}

	return wsrpc.MockRPCDataEncode("protobuf", &grpcpb.MockJSONRPCMessage{
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
