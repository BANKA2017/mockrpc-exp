package main

import (
	"flag"
	"log"
	"math/rand"
	"runtime/debug"
	"strconv"
	"time"

	//_ "net/http/pprof"
	"runtime"

	"github.com/BANKA2017/mockrpc-exp/message/grpc/grpcpb"
	"github.com/BANKA2017/mockrpc-exp/node/functions"
	"github.com/BANKA2017/mockrpc-exp/shared"
)

// var err error
var isHTTPS bool
var httpProtocol = "http"
var wsProtocol = "ws"
var noRTC bool

// https://stackoverflow.com/questions/23577091/generating-random-numbers-over-a-range-in-go
func randRange(min, max int) int {
	return rand.Intn(max-min) + min
}

func main() {
	flag.StringVar(&shared.Addr, "wsurl", "", "Address of Center WebSocket Server")
	flag.StringVar(&shared.Key, "wspwd", "", "Password for Center WebSocket Server")

	flag.BoolVar(&shared.TestMode, "dev", false, "Test mode")
	flag.BoolVar(&isHTTPS, "https", false, "https mode")
	flag.BoolVar(&noRTC, "no_rtc", false, "use websocket only")
	flag.Parse()

	if isHTTPS {
		httpProtocol = "https"
		wsProtocol = "wss"
	}

	// connect to ws
	///if !shared.LocalMode {
	// connect
	if len(shared.Addr) <= 0 {
		log.Println("Invalid ws_url/ws_pwd, will not connect to the ws endpoint")
	} else {
		tmpAccessToken := functions.GetAccessToken(httpProtocol+"://"+shared.Addr+"/api/account/token", shared.Key)
		functions.ConnectToWebsocket(wsProtocol+"://"+shared.Addr+"/api/ws", tmpAccessToken)
		if functions.WsRPC.LatestConn == nil {
			log.Fatalln("websocket: failed!")
		}
		log.Println("websocket: init!")

		if !noRTC {
			err := functions.ConnectToRTC(functions.WsRPC)
			if err != nil {
				log.Println(err)
			}
			log.Println("wait for WebRTC #10s")
			<-time.After(time.Second * 10)
		}

	}

	log.Println("wait for tokens #5s")
	<-time.After(time.Second * 5)

	// GC
	debug.SetGCPercent(300)

	functions.InitTicker()
	log.Println("ticker: init!")
	defer functions.WsConnReconnect.Stop()
	defer functions.UpdateServerStatusTicker.Stop()
	updateStatusTicker := time.NewTicker(time.Millisecond * 100)
	defer updateStatusTicker.Stop()

	for {
		select {
		case <-updateStatusTicker.C:
			shared.ActiveCount = int64(randRange(1, 20))
			shared.DoneCount = shared.SendCount
			shared.SendCount += shared.ActiveCount
		case <-functions.WsConnReconnect.C:
			if !functions.GetCurrentConnStatus() && len(shared.Addr) > 0 {
				tmpAccessToken := functions.GetAccessToken(httpProtocol+"://"+shared.Addr+"/api/account/token", shared.Key)
				log.Println("node: trying reconnect to wsrpc, #", int(functions.WsConnReconnectTicker))
				err := functions.ConnectToWebsocket(wsProtocol+"://"+shared.Addr+"/api/ws", tmpAccessToken)
				if err != nil {
					log.Println(err)
					// backoff
					if functions.WsConnReconnectTicker < functions.WsConnReconnectMaxTicker {
						functions.WsConnReconnectTicker *= 2
						functions.WsConnReconnect.Reset(time.Second * functions.WsConnReconnectTicker)
					}
				} else {
					if !noRTC && functions.TopConnectType == functions.ConnTypeValueWebRTC {
						err = functions.ConnectToRTC(functions.WsRPC)
						if err != nil {
							log.Println(err)
						}
					}
				}
			}
		case <-functions.UpdateServerStatusTicker.C:
			go func() {
				serverStatus := &grpcpb.ServerStatusQuery{
					SendCount:      shared.SendCount,
					ActiveCount:    shared.ActiveCount,
					DoneCount:      shared.DoneCount,
					FailedCount:    shared.FailedCount,
					GoroutineCount: int64(runtime.NumGoroutine()),
					Interval:       shared.Interval,
					StartAt:        shared.StartAt.Unix(),
				}

				if functions.CurrentConnectType == functions.ConnTypeValueWebRTC && functions.GetCurrentConnStatus() {
					err := functions.SendRPCMessage(&grpcpb.MockJSONRPCMessage{
						//JsonRPC: "2.0",
						Method: "push_node_status",
						Data: &grpcpb.MockJSONRPCMessage_Server{
							Server: serverStatus,
						},
						ID: strconv.Itoa(rand.Int()),
					})

					if err != nil {
						log.Println("server status:", err)
					}
				}

				log.Println("Server status:", serverStatus)
			}()
		}
	}
}
