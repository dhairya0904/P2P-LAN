package main

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Tmp struct {
	Hello string `json:"hello"`
}

func main() {

	cfg := parseFlags()
	if cfg.logLevel == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Debug().Msg(fmt.Sprintf("[*] Listening on: %s with port: %d\n", cfg.listenHost, cfg.listenPort))

	node := Node{
		RendezvousString: cfg.RendezvousString,
		ListenHost:       cfg.listenHost,
		ListenPort:       cfg.listenPort,
		ProtocolID:       cfg.ProtocolID,
		NodeType:         cfg.node,
	}
	node.InitializeNode()
	host := node.CreateHost()

	peerChan := initMDNS(host, node.RendezvousString)

	peer := <-peerChan
	fmt.Println("Found peer", peer)

	ctx := context.Background()

	if err := host.Connect(ctx, peer); err != nil {
		fmt.Println("Connection failed:", err)
		panic(err)
	}

	fmt.Println(len(host.Network().Peers()))
	fmt.Println(host.Network().Peers())
	for {
	}
	// 	node.Serve()

	// 	log.Debug().Msg("Connection initialized")

	// 	a := Tmp{
	// 		Hello: "I am being printed now",
	// 	}
	// 	rc, rw := node.GetNodeChannels()
	// 	rw <- a
	// 	data := <-rc

	// 	var result Tmp
	// 	err := mapstructure.Decode(data, &result)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	log.Debug().Msg(fmt.Sprintf("finally I got the data %+v", result))

	// for {
	// }
}
