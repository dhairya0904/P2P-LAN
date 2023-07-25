package main

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type tmp struct {
	hello string
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
	node.Serve()

	a := tmp{
		hello: "fsadfsadfsfsd",
	}
	rc, rw := node.GetNodeChannels()
	rw <- a
	data := <-rc

	user, _ := data.(tmp)
	log.Debug().Msg(fmt.Sprintf("I got the data %+v", user))

	for {
	}
}
