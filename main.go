package main

import (
	"crypto/ecdsa"
	"flag"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
)

const ua = "manspreading"
const ver = "1.0.0"

// statusData is the network packet for the status message.
type statusData struct {
	ProtocolVersion uint32
	NetworkId       uint64
	TD              *big.Int
	CurrentBlock    common.Hash
	GenesisBlock    common.Hash
}

// newBlockData is the network packet for the block propagation message.
type newBlockData struct {
	Block *types.Block
	TD    *big.Int
}

type conn struct {
	p  *p2p.Peer
	rw p2p.MsgReadWriter
}

type proxy struct {
	lock           sync.RWMutex
	upstreamNode   *discover.Node
	upstreamConn   *conn
	downstreamConn *conn
	upstreamState  statusData
	srv            *p2p.Server
}

var pxy *proxy

var upstreamUrl = flag.String("upstream", "", "upstream enode url to connect to")
var listenAddr = flag.String("listenaddr", "127.0.0.1:36666", "listening addr")
var privkey = flag.String("nodekey", "", "nodekey file")

func init() {
	flag.Parse()
}

func main() {
	var nodekey *ecdsa.PrivateKey
	if *privkey != "" {
		nodekey, _ = crypto.LoadECDSA(*privkey)
		fmt.Println("Node Key Loaded from ", *privkey)
	} else {
		nodekey, _ = crypto.GenerateKey()
		fmt.Println("Node Key Generated")
	}

	node, _ := discover.ParseNode(*upstreamUrl)
	pxy = &proxy{
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

		Protocols: []p2p.Protocol{newManspreadingProtocol()},

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
	wg.Wait()
}
