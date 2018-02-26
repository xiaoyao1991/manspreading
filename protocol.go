package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
)

func manspreadingProtocol() p2p.Protocol {
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
			return err
		}

		fmt.Println("msg.Code: ", msg.Code)
		if msg.Code == eth.StatusMsg {
			var myMessage statusData
			err = msg.Decode(&myMessage)
			if err != nil {
				fmt.Println("decode statusData err: ", err)
			}

			fmt.Println("ProtocolVersion", myMessage.ProtocolVersion)
			fmt.Println("NetworkId   ", myMessage.NetworkId)
			fmt.Println("TD          ", myMessage.TD)
			fmt.Println("CurrentBlock", myMessage.CurrentBlock)
			fmt.Println("GenesisBlock", myMessage.GenesisBlock)

			if p.ID() == pxy.upstreamNode.ID {
				realCurrentBlock = myMessage.CurrentBlock
				realTD = myMessage.TD

				pxy.upstreamConn = &conn{
					p:  p,
					rw: rw,
				}
			} else {
				pxy.downstreamConn = &conn{
					p:  p,
					rw: rw,
				}
			}

			err = p2p.Send(rw, eth.StatusMsg, &statusData{
				ProtocolVersion: myMessage.ProtocolVersion,
				NetworkId:       myMessage.NetworkId,
				TD:              realTD,
				CurrentBlock:    realCurrentBlock,
				GenesisBlock:    myMessage.GenesisBlock,
			})

			if err != nil {
				fmt.Println("handshake err: ", err)
				return err
			}
		} else {
			if p.ID() != pxy.upstreamNode.ID {
				err = pxy.upstreamConn.rw.WriteMsg(msg)
			} else {
				err = pxy.downstreamConn.rw.WriteMsg(msg)
			}

			if err != nil {
				fmt.Println("relaying err: ", err)
				return err
			}
		}
	}

	return nil
}

func fromWhom(nodeId string) string {
	if nodeId == pxy.upstreamNode.ID.String() {
		return "upstream"
	} else {
		return "downstream"
	}
}
