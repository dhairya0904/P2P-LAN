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

	peerChan := initMDNS(host, cfg.RendezvousString)

	for {
		peer := <-peerChan
		fmt.Println(peer)
	}

	// node.InitializeNode()

	// if len(cfg.peerAddress) < 2 {
	// 	node.startMaster(host)
	// }

	// // print the node's PeerInfo in multiaddr format
	// peerInfo := peerstore.AddrInfo{
	// 	ID:    host.ID(),
	// 	Addrs: host.Addrs(),
	// }
	// addrs, _ := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	// fmt.Println("libp2p node address:", addrs[0])

	// if len(cfg.peerAddress) > 2 {
	// 	addr, err := multiaddr.NewMultiaddr(cfg.peerAddress)

	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	peer, err := peerstore.AddrInfoFromP2pAddr(addr)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	if err := host.Connect(context.Background(), *peer); err != nil {
	// 		fmt.Println("Connection failed:", err)
	// 		panic(err)
	// 	}

	// 	stream, err := host.NewStream(context.Background(), peer.ID, protocol.ID(node.ProtocolID))

	// 	if err != nil {
	// 		fmt.Println("Stream open failed", err)
	// 		panic(err)
	// 	} else {
	// 		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	// 		go writeData(rw, node.writeChannel)
	// 		go readData(rw, node.readChannel)
	// 		log.Debug().Msg(fmt.Sprintf("Connected to Peer %s", peer))
	// 	}
	// }

	// a := Tmp{
	// 	Hello: "fsd",
	// }
	// rc, wc := node.GetNodeChannels()
	// wc <- a
	// fmt.Println(<-rc)

	for {
	}
}
