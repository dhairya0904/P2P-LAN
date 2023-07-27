package main

import (
	"flag"
)

type config struct {
	RendezvousString string
	ProtocolID       string
	listenHost       string
	listenPort       int
	logLevel         string
	node             string
	peerAddress      string
}

func parseFlags() *config {
	c := &config{}

	flag.StringVar(&c.RendezvousString, "rendezvous", "meetme", "Unique string to identify group of nodes. Share this with your friends to let them connect with you")
	flag.StringVar(&c.listenHost, "host", "192.168.100.1", "The bootstrap node host listen address\n")
	flag.StringVar(&c.ProtocolID, "pid", "/chat/1.1.0", "Sets a protocol id for stream headers")
	flag.StringVar(&c.logLevel, "logLevel", "info", "Sets lob level for debugging")
	flag.StringVar(&c.node, "node", "peer", "Sets peer master")
	flag.IntVar(&c.listenPort, "port", 4001, "node listen port")
	flag.StringVar(&c.peerAddress, "peer", "_", "Sets peer address")

	flag.Parse()
	return c
}
