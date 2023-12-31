package main

import (
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
	node.InitializeNode()

	if len(cfg.peerAddress) == 0 {
		node.startMaster(host)
	} else {
		node.connectWithPeer(host, cfg.peerAddress)
	}

	a := Tmp{
		Hello: "fsd",
	}
	rc, wc := node.GetNodeChannels()
	wc <- a
	fmt.Println(<-rc)

	for {
	}
}
