package core

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/kademlia"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
)

var PortP2P uint16 = 7000

const (
	newBlock       = iota // Body will be: Block
	newTransaction        // Body will be: Transaction
	thisIsMyChain         // Body will be: []Block
	needChain             // Body will be: nil
)

// Stores a type of message and a body.
type NodeMessage struct {
	MessageType int         // Can be: newBlock, newTransaction, thisIsMyChain, or needChain
	Body        interface{} // The actual payload (it can be many types)
}

func (m NodeMessage) Marshal() []byte {
	var buf bytes.Buffer

	gob.Register([]Block(nil))
	gob.Register(Block{})
	gob.Register(Transaction{})
	gob.Register([]Transaction(nil))

	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(m); err != nil {
		log.Fatal(err)
	}

	return buf.Bytes()
}
func unMarshalNodeMessage(input []byte) (NodeMessage, error) {
	var msg NodeMessage
	buf := bytes.NewBuffer(input)
	dec := gob.NewDecoder(buf)

	if err := dec.Decode(&msg); err != nil {
		log.Fatal(err)
	}

	return msg, nil
}

// check panics if err is not nil.
func check(err error) {
	if err != nil {
		log.Error(err)
	}
}

// sendMessageToPeer sends a message to a peer directly through their address.
func (l *LocalNode) sendMessageToPeer(message NodeMessage, address string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	err := l.node.SendMessage(ctx, address, message)
	cancel()

	return err
}

// broadcast sends a message to all peers.
func (l *LocalNode) broadcast(message NodeMessage) {
	for _, id := range l.kademliaProtocol.Table().Peers() {
		err := l.sendMessageToPeer(message, id.Address)

		if err != nil {
			log.Warnf("Failed to send message to %s. Skipping... [error: %s]", id.Address, err)
			continue
		}
	}
}

// GetPeerConsensus sends a message to all peers requesting their chain, then we choose 5 of them and run consensus on them.
func (l *LocalNode) GetPeerConsensus() {
	// Create a channel
	l.incomingChains = make(chan []Block, l.MinimumChainsForConsensus)

	// Ask all peers for their chain
	l.broadcast(NodeMessage{
		MessageType: needChain,
		Body:        nil,
	})

	// Get the minimum amount of chains we need for consensus in slice
	var chains = make([][]Block, 0)
	for incomingChain := range l.incomingChains {
		chains = append(chains, incomingChain)
		if len(chains) == l.MinimumChainsForConsensus {
			close(l.incomingChains)
		}
	}

	// Run our consensus function
	l.Consensus(chains...)
}

// BroadcastBlock sends a block to all of our peers.
func (l *LocalNode) BroadcastBlock(b Block) {
	l.broadcast(NodeMessage{
		MessageType: newBlock,
		Body:        b,
	})

	log.Info("Sent peer(s) a block!")
}

// BroadcastTransaction sends a transaction to all of our peers.
func (l *LocalNode) BroadcastTransaction(t Transaction) {
	l.broadcast(NodeMessage{
		MessageType: newTransaction,
		Body:        t,
	})

	log.Info("Sent peer(s) a transaction!")
}

// SendPeerOurChain sends a specific peer our chain.
func (l *LocalNode) SendPeerOurChain(address string) {
	err := l.sendMessageToPeer(NodeMessage{
		MessageType: thisIsMyChain,
		Body:        l.Chain,
	}, address)

	log.Info("Sent peer our chain!")

	if err != nil {
		log.Errorf("Failed to send chain to %s", address)
	}
}

// Starts all P2P functions. Takes a list of seedNodes.
func (l *LocalNode) Start(seedNodes []string) {

	// Create a new configured node.
	node, err := noise.NewNode(noise.WithNodeBindHost(GetOutboundIP()), noise.WithNodeAddress(fmt.Sprintf("%s:%d", GetOutboundIP().String(), PortP2P)), noise.WithNodeBindPort(PortP2P))
	check(err)

	defer func() { log.Warn("Closing node..."); node.Close() }()

	// Register the chatMessage Go type to the node with an associated unmarshal function.
	node.RegisterMessage(NodeMessage{}, unMarshalNodeMessage)

	// Register a message handler to the node.
	node.Handle(func(ctx noise.HandlerContext) error {
		if ctx.IsRequest() {
			return nil
		}

		obj, err := ctx.DecodeMessage()
		check(err)

		msg, ok := obj.(NodeMessage)
		if !ok {
			log.Error("NodeMessage was unable to be deserialized!")
			return nil
		}

		switch msg.MessageType {

		case needChain:
			log.Info("A peer just requested our chain!")

			// If a fellow node needs our chain, send it to them!
			l.SendPeerOurChain(ctx.ID().Address)

		case newBlock:
			log.Info("A peer just gave us a new block!")

			block, ok := msg.Body.(Block)
			if !ok {
				log.Error("Block was unable to be deserialized!")
				break
			}

			// If our peer's mined block was valid and added to chain:
			if l.AddMinedBlockToChain(block) == true {
				log.Info("We just got a new mined block from a peer and added it to the chain!")
			} else {
				log.Warn("The block we just got from a peer was not valid! It was not added to the chain and the UTXO was not updated!")
			}

		case newTransaction:
			log.Info("A peer just gave us a new transaction!")

			transaction, ok := msg.Body.(Transaction)
			if !ok {
				log.Error("Transaction was unable to be deserialized!")
				break
			}

			// If a peer has gotten a new transaction request, add it to our MemPool.
			l.AddTransactionToMemPool(transaction)

		case thisIsMyChain:
			chain, ok := msg.Body.([]Block)
			if !ok {
				log.Error("Chain was unable to be deserialized!")
				break
			}

			log.Infof("We just got a chain from one of our peers! Here is the amount of the first transaction of the genesis block: %d", chain[0].Transactions[0].Amount)

			// Add the incoming chain we requested to our channel
			l.incomingChains <- chain

		default:
			log.Panic("We got an invalid message type!")
		}

		return nil
	})

	// Instantiate Kademlia.
	events := kademlia.Events{
		OnPeerAdmitted: func(id noise.ID) {
			log.Infof("Learned about a new peer %s.\n", id.Address)
		},
		OnPeerEvicted: func(id noise.ID) {
			log.Infof("Forgotten a peer (as we pinged them and they didn't respond) %s.\n", id.Address)
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

	// Wait 3 seconds
	time.Sleep(3 * time.Second)

	log.Info("We've started P2P! Now trying to get peer consensus...")

	// Get peer consensus
	l.GetPeerConsensus()

	log.Info("Got peer consensus over P2P!")

	WaitForCtrlC()

}

// bootstrap pings and dials an array of network addresses which we may interact with and  discover peers from.
func bootstrap(node *noise.Node, addresses []string) {
	for _, addr := range addresses {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		vic, err := node.Ping(ctx, addr)
		cancel()

		if err != nil {
			log.Warnf("Failed to ping seed node %s. Skipping... [error: %s]\n", addr, err)
			continue
		} else {
			log.Info("We made a connection with one of our seed nodes: " + vic.ID().Address)
		}
	}
}

// discover uses Kademlia to discover new peers from nodes we already are aware of.
func discover(overlay *kademlia.Protocol) {
	ids := overlay.Discover()

	var str []string
	for _, id := range ids {
		str = append(str, id.Host.String())
	}

	if len(ids) > 0 {
		log.Info("Discovered ", len(ids), " peer(s): ", strings.Join(str, ", "))
	} else {
		log.Warn("Did not discover any peers.")
	}
}

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func WaitForCtrlC() {
	var endWaiter sync.WaitGroup
	endWaiter.Add(1)
	var signalChannel chan os.Signal
	signalChannel = make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	go func() {
		<-signalChannel
		endWaiter.Done()
	}()
	endWaiter.Wait()
}
