package main

import (
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/rlp"
)

func newManspreadingProtocol() p2p.Protocol {
	return p2p.Protocol{
		Name:    eth.ProtocolName,
		Version: eth.ProtocolVersions[0],
		Length:  eth.ProtocolLengths[0],
		Run:     handle,
		NodeInfo: func() interface{} {
			fmt.Println("Noop: NodeInfo called")
			return nil
		},
		PeerInfo: func(id discover.NodeID) interface{} {
			fmt.Println("Noop: PeerInfo called")
			return nil
		},
	}
}

func handle(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	fmt.Println("Run called")

	for {
		fmt.Println("Waiting for msg...")
		msg, err := rw.ReadMsg()
		fmt.Println("Got a msg from: ", fromWhom(p.ID().String()))
		if err != nil {
			fmt.Println("readMsg err: ", err)

			if err == io.EOF {
				fmt.Println(fromWhom(p.ID().String()), " has dropped its connection...")
				pxy.lock.Lock()
				if p.ID() == pxy.upstreamNode.ID {
					pxy.upstreamConn = nil
				} else {
					pxy.downstreamConn = nil
				}
				pxy.lock.Unlock()
			}

			return err
		}
		fmt.Println("msg.Code: ", msg.Code)

		if msg.Code == eth.StatusMsg { // handshake
			var myMessage statusData
			err = msg.Decode(&myMessage)
			if err != nil {
				fmt.Println("decode statusData err: ", err)
				return err
			}

			fmt.Println("ProtocolVersion: ", myMessage.ProtocolVersion)
			fmt.Println("NetworkId:       ", myMessage.NetworkId)
			fmt.Println("TD:              ", myMessage.TD)
			fmt.Println("CurrentBlock:    ", myMessage.CurrentBlock.Hex())
			fmt.Println("GenesisBlock:    ", myMessage.GenesisBlock.Hex())

			pxy.lock.Lock()
			if p.ID() == pxy.upstreamNode.ID {
				pxy.upstreamState = myMessage
				pxy.upstreamConn = &conn{p, rw}
			} else {
				pxy.downstreamConn = &conn{p, rw}
			}
			pxy.lock.Unlock()

			err = p2p.Send(rw, eth.StatusMsg, &statusData{
				ProtocolVersion: myMessage.ProtocolVersion,
				NetworkId:       myMessage.NetworkId,
				TD:              pxy.upstreamState.TD,
				CurrentBlock:    pxy.upstreamState.CurrentBlock,
				GenesisBlock:    myMessage.GenesisBlock,
			})

			if err != nil {
				fmt.Println("handshake err: ", err)
				return err
			}
		} else if msg.Code == eth.NewBlockMsg {
			var myMessage newBlockData
			err = msg.Decode(&myMessage)
			if err != nil {
				fmt.Println("decode newBlockMsg err: ", err)
			}

			pxy.lock.Lock()
			if p.ID() == pxy.upstreamNode.ID {
				pxy.upstreamState.CurrentBlock = myMessage.Block.Hash()
				pxy.upstreamState.TD = myMessage.TD
			} //TODO: handle newBlock from downstream
			pxy.lock.Unlock()

			// need to re-encode msg
			size, r, err := rlp.EncodeToReader(myMessage)
			if err != nil {
				fmt.Println("encoding newBlockMsg err: ", err)
			}
			relay(p, p2p.Msg{Code: eth.NewBlockMsg, Size: uint32(size), Payload: r})
		} else {
			relay(p, msg)
		}
	}

	return nil
}

func relay(p *p2p.Peer, msg p2p.Msg) {
	var err error
	pxy.lock.RLock()
	defer pxy.lock.RUnlock()
	if p.ID() != pxy.upstreamNode.ID && pxy.upstreamConn != nil {
		err = pxy.upstreamConn.rw.WriteMsg(msg)
	} else if p.ID() == pxy.upstreamNode.ID && pxy.downstreamConn != nil {
		err = pxy.downstreamConn.rw.WriteMsg(msg)
	} else {
		fmt.Println("One of upstream/downstream isn't alive: ", pxy.srv.Peers())
	}

	if err != nil {
		fmt.Println("relaying err: ", err)
	}
}

func (pxy *proxy) upstreamAlive() bool {
	for _, peer := range pxy.srv.Peers() {
		if peer.ID() == pxy.upstreamNode.ID {
			return true
		}
	}
	return false
}

func fromWhom(nodeId string) string {
	if nodeId == pxy.upstreamNode.ID.String() {
		return "upstream"
	} else {
		return "downstream"
	}
}
