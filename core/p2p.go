package core

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/kademlia"
	"os"
	"os/signal"
	"strings"
	"time"
)

const (
	newBlock       = iota // Body will be: Block
	newTransaction        // Body will be: Transaction
	thisIsMyChain         // Body will be: []Block
	needChain             // Body will be: nil
)

// It stores what could be 4 options: a block, a transaction, a node's chain or a bool that when true indicates a request for this node's chain
type nodeMessage struct {
	messageType int         // Can be: newBlock, newTransaction, thisIsMyChain, or needChain
	body        interface{} // The actual payload (it can be many types)
}

func (m nodeMessage) Marshal() []byte {
	b, err := json.Marshal(m)
	check(err)

	return b
}
func unMarshalNodeMessage(buf []byte) (nodeMessage, bool) {
	var msg nodeMessage
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

func (l *LocalNode) sendMessageToPeer(message nodeMessage, address string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	err := l.node.SendMessage(ctx, address, message)
	cancel()

	return err
}

func (l *LocalNode) broadcast(message nodeMessage) {
	for _, id := range l.kademliaProtocol.Table().Peers() {
		err := l.sendMessageToPeer(message, id.Address)

		if err != nil {
			fmt.Printf("Failed to send message to %s. Skipping... [error: %s]\n",
				id.Address,
				err,
			)
			continue
		}
	}
}

func (l *LocalNode) GetPeerConsensus() {
	// Create a buffered channel of 5
	l.incomingChains = make(chan []Block)

	// Ask all peers for their chain
	l.broadcast(nodeMessage{
		messageType: needChain,
		body:        nil,
	})

	// Wait until we have 4 chains from our peers and then run our consensus function
	l.Consensus(<-l.incomingChains, <-l.incomingChains, <-l.incomingChains, <-l.incomingChains)
}

func (l *LocalNode) BroadcastBlock(b Block) {
	// Alert all peers of our new block
	l.broadcast(nodeMessage{
		messageType: newBlock,
		body:        b,
	})

}

func (l *LocalNode) BroadcastTransaction(t Transaction) {
	// Alert all peers of a transaction we have received
	l.broadcast(nodeMessage{
		messageType: newTransaction,
		body:        t,
	})
}

func (l *LocalNode) SendPeerOurChain(address string) {
	err := l.sendMessageToPeer(nodeMessage{
		messageType: thisIsMyChain,
		body:        l.Chain,
	}, address)

	if err != nil {
		fmt.Printf("Failed to send chain to %s\n",
			address,
		)
	}
}

// Starts all P2P functions. Takes a list of seedNodes
func (l *LocalNode) Start(seedNodes []string) {
	// Create a new configured node.
	node, err := noise.NewNode()
	check(err)

	// Release resources associated to node at the end of the program.
	defer node.Close()

	// Register the chatMessage Go type to the node with an associated unmarshal function.
	node.RegisterMessage(nodeMessage{}, unMarshalNodeMessage)

	// Register a message handler to the node.
	node.Handle(func(ctx noise.HandlerContext) error {
		if ctx.IsRequest() {
			return nil
		}

		obj, err := ctx.DecodeMessage()
		check(err)

		msg, ok := obj.(nodeMessage)
		if !ok {
			return nil
		}

		switch msg.messageType {

		case needChain:
			// If a fellow node needs our chain, send it to them!
			l.SendPeerOurChain(ctx.ID().Address)

		case newBlock:
			block, ok := msg.body.(Block)
			if !ok {
				return nil
			}

			// If our peer's mined block was valid and added to chain:
			if l.AddMinedBlockToChain(block) == true {
				fmt.Println("We just mined a new block and added it to the chain!")
			} else {
				fmt.Println("The block we just mined was not valid! It was not added to the chain and the UTXO was not updated!")
			}

		case newTransaction:
			transaction, ok := msg.body.(Transaction)
			if !ok {
				return nil
			}

			// If a peer has gotten a new transaction request, add it to our MemPool.
			l.AddTransactionToMemPool(transaction)

		case thisIsMyChain:
			chain, ok := msg.body.([]Block)
			if !ok {
				return nil
			}

			// Add the incoming chain we requested to our channel
			l.incomingChains <- chain

		default:
			panic("We got an invalid message type!")
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

	l.node = node
	l.kademliaProtocol = overlay

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
