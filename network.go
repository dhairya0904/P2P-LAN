package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"

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

	writeChannel chan interface{}
	readChannel  chan interface{}
}

func (node *Node) InitializeNode() {
	node.readChannel = make(chan interface{})
	node.writeChannel = make(chan interface{})
}

func (node *Node) GetNodeChannels() (chan interface{}, chan interface{}) {
	return node.readChannel, node.writeChannel
}

func (node *Node) Serve() {

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
		host.SetStreamHandler(protocol.ID(node.ProtocolID), node.handleStream)
		initMDNS(host, node.RendezvousString)
		for {
			if host.Peerstore().Peers().Len() > 0 {
				return
			}
		}
	}

	peerChan := initMDNS(host, node.RendezvousString)

	peer := <-peerChan // will block until we discover a peer
	log.Debug().Msg(fmt.Sprintf("Founde Peer %s", peer))

	if err := host.Connect(ctx, peer); err != nil {
		fmt.Println("Connection failed:", err)
		panic(err)
	}

	if node.NodeType == "peer" { //// no need to create stream from both the nodes
		// open a stream, this stream will be handled by handleStream other end
		stream, err := host.NewStream(ctx, peer.ID, protocol.ID(node.ProtocolID))

		if err != nil {
			fmt.Println("Stream open failed", err)
			panic(err)
		} else {
			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

			go writeData(rw, node.writeChannel)
			go readData(rw, node.readChannel)
			log.Debug().Msg(fmt.Sprintf("Connected to Peer %s", peer))
		}
	}
}

func (node *Node) handleStream(stream network.Stream) {
	log.Debug().Msg("Got a new stream!")

	// Create a buffer stream for non-blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw, node.readChannel)
	go writeData(rw, node.writeChannel)

	// 'stream' will stay open until you close it (or the other side closes it).
}

func readData(rw *bufio.ReadWriter, readChannel chan<- interface{}) {
	for {
		receivedData := readJSON(rw)
		log.Debug().Msg(fmt.Sprintf("Received data %+v", receivedData))
		readChannel <- receivedData
		rw.Flush()
	}
}

func writeData(rw *bufio.ReadWriter, writeChannel <-chan interface{}) {

	for {
		data := <-writeChannel
		dataBytes, err := json.Marshal(data)

		if err != nil {
			panic(err)
		}

		_, err = rw.Write(dataBytes)
		if err != nil {
			// fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			// fmt.Println("Error flushing buffer")
			panic(err)
		}
	}
}

func readJSON(rw *bufio.ReadWriter) interface{} {

	var receivedData interface{}

	decoder := json.NewDecoder(rw.Reader)
	err := decoder.Decode(&receivedData)
	if err != nil {
		if err != io.EOF {
			panic(err)
		}
	}

	fmt.Printf("lolol %+v", receivedData)
	return receivedData
}
