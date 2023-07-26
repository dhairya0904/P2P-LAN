package main

import (
	"fmt"

	mapstructure "github.com/mitchellh/mapstructure"
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
	node.Serve()

	log.Debug().Msg("Connection initialized")

	a := Tmp{
		Hello: "I am being printed now",
	}
	rc, rw := node.GetNodeChannels()
	rw <- a
	data := <-rc

	var result Tmp
	err := mapstructure.Decode(data, &result)
	if err != nil {
		panic(err)
	}

	log.Debug().Msg(fmt.Sprintf("finally I got the data %+v", result))

	for {
	}
}
