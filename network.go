package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog/log"
)

type Node struct {
	ListenHost, RendezvousString, ProtocolID, NodeType string
	ListenPort                                         int

	// writeChannel chan interface{}
	// readChannel  chan interface{}
}

func (node *Node) InitalizeNode() {

	log.Debug().Msg(fmt.Sprintf("[*] Listening on: %s with port: %d\n", node.ListenHost, node.ListenPort))

	ctx := context.Background()
	r := rand.Reader

	// Creates a new RSA key pair for this host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", node.ListenHost, node.ListenPort))

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	host, err := libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
	if err != nil {
		panic(err)
	}

	log.Debug().Msg(fmt.Sprintf("\n[*] Your Multiaddress Is: /ip4/%s/tcp/%v/p2p/%s\n", node.ListenHost, node.ListenPort, host.ID().Pretty()))

	if node.NodeType == "master" {
		host.SetStreamHandler(protocol.ID(node.ProtocolID), handleStream)
	}

	peerChan := initMDNS(host, node.RendezvousString)

	for {
		peer := <-peerChan // will block until we discover a peer
		log.Debug().Msg(fmt.Sprintf("Founde Peer %s", peer))

		if node.NodeType == "master" {
			continue
		}

		if err := host.Connect(ctx, peer); err != nil {
			fmt.Println("Connection failed:", err)
			panic(err)
		}

		// open a stream, this stream will be handled by handleStream other end
		stream, err := host.NewStream(ctx, peer.ID, protocol.ID(node.ProtocolID))

		if err != nil {
			fmt.Println("Stream open failed", err)
			panic(err)
		} else {
			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

			go writeData(rw)
			go readData(rw)
			log.Debug().Msg(fmt.Sprintf("Connected to Peer %s", peer))
		}
	}
}

func handleStream(stream network.Stream) {
	fmt.Println("Got a new stream!")

	// Create a buffer stream for non-blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw)
	go writeData(rw)

	// 'stream' will stay open until you close it (or the other side closes it).
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer")
			panic(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			panic(err)
		}
	}
}
