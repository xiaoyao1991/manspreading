# Manspreading
Manspreading helps you greedily occupy a peer seat in a remote geth node. 

### Introduction
Sadly, due to the fact that many nodes on Ethereum network do not change the [default](https://github.com/ethereum/go-ethereum/wiki/Command-Line-Options) `--maxpeers` settings, most nodes are full and won't accept new peers. 
Although there are `static-nodes.json` and `trusted-nodes.json` where a node owner can hardcode peers that will always connect regardless of the restriction by `--maxpeers`, these options are buried deep in the document and codebase, so that few knows they exist.
Preserving a node peer seat at a peer becomes essential in the development/research on geth.  

Manspreading is a proxy server that can be run as daemon and occupies a "seat" at a remote geth peer, so the real geth instance behind the proxy can stop and restart anytime without worrying the seat at the remote peer been taken during the restart period.

### Prerequisite
- Golang (v1.8+)
- You need to install [geth](https://github.com/ethereum/go-ethereum/) as the dependency by running `go get github.com/ethereum/go-ethereum`     

### Usage
1. `go build .`
2. `./manspreading --upstream="<remote node enode url>" --listenaddr="127.0.0.1:36666"`  
    A log line will show what's the enode url of the manspreading proxy in the format of `enode://<nodekey>@<listenaddr>`
3. Start your real geth instance and add the manspreading enode url as a peer by running `admin.addPeer("<manspreading enode url>")`

