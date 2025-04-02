package main

import (
	"flag"

	httpapi "github.com/BANKA2017/mockrpc-exp/center/http_api"
	"github.com/BANKA2017/mockrpc-exp/shared"
)

func main() {
	flag.StringVar(&shared.Addr, "addr", "", "Address of center") // 0.0.0.0:11111
	flag.BoolVar(&shared.TestMode, "dev", false, "Test mode")

	flag.Parse()

	httpapi.Api(shared.Addr)
}
