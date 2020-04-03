package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/kademlia"
	"github.com/transmissionsdev/cosmosis/core"
	"os"
	"os/signal"
	"strings"
	"time"
)

// It stores what could be 3 options: a block, a transaction or a bool that when true indicates a request for this node's chain/
type NodeMessage struct {
	// Broadcasts
	NewBlock       core.Block
	NewTransaction core.Transaction
	// Requests
	NeedChain bool
}

func (m NodeMessage) Marshal() []byte {
	b, err := json.Marshal(m)
	check(err)

	return b
}
func UnMarshalNodeMessage(buf []byte) (NodeMessage, bool) {
	var msg NodeMessage
	err := json.Unmarshal(buf, &msg)
	check(err)
	return msg, false
}

// check panics if err is not nil.
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Stores the local blockchain on this node as well as details about this node as a p2p node
type NoiseWrapper struct {
	LocalChain *core.Blockchain
	localNode  *noise.Node
}

// Starts all P2P functions. Takes a list of seedNodes
func (n *NoiseWrapper) Start(seedNodes []string) {
	// Create a new configured node.
	node, err := noise.NewNode()
	check(err)

	// Release resources associated to node at the end of the program.
	defer node.Close()

	// Register the chatMessage Go type to the node with an associated unmarshal function.
	node.RegisterMessage(NodeMessage{}, UnMarshalNodeMessage)

	// Register a message handler to the node.
	node.Handle(func(ctx noise.HandlerContext) error {
		if ctx.IsRequest() {
			return nil
		}

		obj, err := ctx.DecodeMessage()
		if err != nil {
			return nil
		}

		msg, ok := obj.(NodeMessage)
		if !ok {
			return nil
		}

		fmt.Printf("%s> %v\n", ctx.ID().Address, msg)

		return nil
	})

	// Instantiate Kademlia.
	events := kademlia.Events{
		OnPeerAdmitted: func(id noise.ID) {
			fmt.Printf("Learned about a new peer %s.\n", id.Address)
		},
		OnPeerEvicted: func(id noise.ID) {
			fmt.Printf("Forgotten a peer %s.\n", id.Address)
		},
	}

	overlay := kademlia.New(kademlia.WithProtocolEvents(events))

	// Bind Kademlia to the node.
	node.Bind(overlay.Protocol())

	// Have the node start listening for new peers.
	check(node.Listen())

	// Ping nodes to initially bootstrap and discover peers from.
	bootstrap(node, seedNodes)

	// Attempt to discover peers if we are bootstrapped to any nodes.
	discover(overlay)

	n.localNode = node

	// Wait until Ctrl+C or a termination call is done.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

// bootstrap pings and dials an array of network addresses which we may interact with and  discover peers from.
func bootstrap(node *noise.Node, addresses []string) {
	for _, addr := range addresses {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, err := node.Ping(ctx, addr)
		cancel()

		if err != nil {
			fmt.Printf("Failed to ping bootstrap node %s. Skipping... [error: %s]\n", addr, err)
			continue
		}
	}
}

// discover uses Kademlia to discover new peers from nodes we already are aware of.
func discover(overlay *kademlia.Protocol) {
	ids := overlay.Discover()

	var str []string
	for _, id := range ids {
		str = append(str, fmt.Sprintf("%s", id.Address))
	}

	if len(ids) > 0 {
		fmt.Printf("Discovered %d peer(s): [%v]\n", len(ids), strings.Join(str, ", "))
	} else {
		fmt.Printf("Did not discover any peers.\n")
	}
}
