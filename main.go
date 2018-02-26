package main

import (
	"flag"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
)

const ua = "manspreading"
const ver = "1.0.0"

var emptyHash = common.Hash{}
var realCurrentBlock = emptyHash
var realTD = big.NewInt(0)

// statusData is the network packet for the status message.
type statusData struct {
	ProtocolVersion uint32
	NetworkId       uint64
	TD              *big.Int
	CurrentBlock    common.Hash
	GenesisBlock    common.Hash
}

type conn struct {
	p  *p2p.Peer
	rw p2p.MsgReadWriter
}

type proxy struct {
	autopilot      uint32
	upstreamNode   *discover.Node
	upstreamConn   *conn
	downstreamConn *conn
	upstreamState  statusData
	srv            *p2p.Server
}

var pxy *proxy

var upstreamUrl = flag.String("upstream", "", "upstream enode url to connect to")
var listenAddr = flag.String("listenaddr", "127.0.0.1:36666", "listening addr")

func init() {
	flag.Parse()
}

func main() {
	nodekey, _ := crypto.GenerateKey()
	fmt.Println("Node Key Generated")

	node, _ := discover.ParseNode(*upstreamUrl)
	pxy = &proxy{
		autopilot:    1,
		upstreamNode: node,
	}

	config := p2p.Config{
		PrivateKey:     nodekey,
		MaxPeers:       2,
		NoDiscovery:    true,
		DiscoveryV5:    false,
		Name:           common.MakeName(fmt.Sprintf("%s/%s", ua, node.ID.String()), ver),
		BootstrapNodes: []*discover.Node{node},
		StaticNodes:    []*discover.Node{node},
		TrustedNodes:   []*discover.Node{node},

		Protocols: []p2p.Protocol{manspreadingProtocol()},

		ListenAddr: *listenAddr,
		Logger:     log.New(),
	}
	config.Logger.SetHandler(log.StdoutHandler)

	pxy.srv = &p2p.Server{Config: config}

	// Wait forever
	var wg sync.WaitGroup
	wg.Add(2)
	err := pxy.srv.Start()
	wg.Done()
	if err != nil {
		fmt.Println(err)
	}

	ticker := time.Tick(5 * time.Second)
	for {
		select {
		case <-ticker:
			fmt.Println("peers: ", pxy.srv.Peers())
		default:
		}
	}

	wg.Wait()
}
