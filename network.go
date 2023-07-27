package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
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

func (node *Node) CreateHost() host.Host {
	log.Debug().Msg(fmt.Sprintf("[*] Listening on: %s with port: %d\n", node.ListenHost, node.ListenPort))

	r := rand.Reader
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", node.ListenHost, node.ListenPort))

	// Creates a new RSA key pair for this host.
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
	return host
}

func (node *Node) connectWithPeer(host host.Host, peerAddress string) {

	ctx := context.Background()
	addr, err := multiaddr.NewMultiaddr(peerAddress)

	if err != nil {
		panic(err)
	}

	peer, err := peerstore.AddrInfoFromP2pAddr(addr)
	if err != nil {
		panic(err)
	}

	if err := host.Connect(context.Background(), *peer); err != nil {
		fmt.Println("Connection failed:", err)
		panic(err)
	}

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

func (node *Node) startMaster(host host.Host) {
	log.Debug().Msg("Creating master node")
	host.SetStreamHandler(protocol.ID(node.ProtocolID), node.handleStream)

	peerInfo := peerstore.AddrInfo{
		ID:    host.ID(),
		Addrs: host.Addrs(),
	}
	addrs, _ := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	privateIp, err := getPrivateIP()

	if err != nil {
		panic(err)
	}

	connectionString := strings.Replace(addrs[0].String(), "127.0.0.1", privateIp, 1)
	fmt.Printf("Connect using this %s", connectionString)
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
		var data interface{} = <-writeChannel
		log.Debug().Msg(fmt.Sprintf("%+v chacha", data))
		dataBytes, err := json.Marshal(data)
		log.Debug().Msg(string(dataBytes))

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

	log.Debug().Msg(fmt.Sprintf("I can read %+v", receivedData))
	return receivedData
}

func getPrivateIP() (string, error) {
	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	// Iterate through the network interfaces
	for _, iface := range interfaces {
		// Ignore loopback and non-up interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Get all addresses for the current interface
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		// Iterate through the addresses of the current interface
		for _, addr := range addrs {
			// Check if the address is an IP address
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				// Check if the IP address is a private IP
				if ipnet.IP.To4() != nil && isPrivateIP(ipnet.IP) {
					return ipnet.IP.String(), nil
				}
			}
		}
	}

	return "", fmt.Errorf("private IP not found")
}

func isPrivateIP(ip net.IP) bool {
	// Check if the IP address belongs to private IP address ranges
	privateCIDRs := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}

	for _, cidr := range privateCIDRs {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err == nil && ipnet.Contains(ip) {
			return true
		}
	}

	return false
}
