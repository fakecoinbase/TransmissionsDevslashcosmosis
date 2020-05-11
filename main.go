package main

import (
	"flag"
	"fmt"
	"github.com/carlescere/scheduler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/transmissionsdev/cosmosis/core"
	"net/http"
	"os"
	"strings"
	"time"
)

var self core.LocalNode

const introMessage = `
_________                                    _____        
__  ____/___________________ ___________________(_)_______
_  /    _  __ \_  ___/_  __  __ \  __ \_  ___/_  /__  ___/
/ /___  / /_/ /(__  )_  / / / / / /_/ /(__  )_  / _(__  )
\____/  \____//____/ /_/ /_/ /_/\____//____/ /_/  /____/
`

func main() {
	fmt.Println(introMessage)

	var operatorPublicKey string
	flag.StringVar(&operatorPublicKey, "publicKey", "", "A valid public key where funds from mining can be sent to your account")
	var validationServerURL string
	flag.StringVar(&validationServerURL, "validationServer", "https://crows.sh/verifySignature", "A full url (with http://) that operates as a valid ECDSA SECP256k1 signature validation webserver. We recommend you run one locally. Go to: https://github.com/transmissionsdev/cosmosisUtils to find instructions to run one!")
	var seedNodeIPsRaw string
	flag.StringVar(&seedNodeIPsRaw, "seedNodes", "", "A list of addresses of other nodes separated by commas (Example: 75.82.156.254,25.92.256.254)")
	var minimumChainsForConsensus int
	flag.IntVar(&minimumChainsForConsensus, "minimumChainsForConsensus", 4, "How many chains you wish to get before making consensus.")
	var hostJSONEndpoints bool
	flag.BoolVar(&hostJSONEndpoints, "hostJSONEndpoints", false, "Include this flag if you would like a webserver to be hosted alongside the P2P protocol for communicating with wallets, etc.")

	flag.Parse()

	var seedNodeIPs = make([]string, 0)

	if seedNodeIPsRaw != "" {
		seedNodeIPs = strings.Split(seedNodeIPsRaw, ",")

		// Add port to each seed node
		for i, ip := range seedNodeIPs {
			seedNodeIPs[i] = fmt.Sprintf("%s:%d", ip, core.PortP2P)
		}
	}

	// ------[Validate Flags]----------
	if operatorPublicKey == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	client := http.Client{
		Timeout: 1 * time.Second,
	}
	_, err := client.Get(validationServerURL)
	if err != nil {
		flag.PrintDefaults()
		log.Fatal("Your validation server URL is unreachable! See instructions to run your own here: https://github.com/transmissionsdev/cosmosisUtils\n")
	}
	// --------------------------------

	self = core.LocalNode{Chain: []core.Block{core.GenesisBlock}, MemPool: make([]core.Transaction, 0), UTXO: make(core.UTXO), ValidationServerURL: validationServerURL, OperatorPublicKey: operatorPublicKey, MinimumChainsForConsensus: minimumChainsForConsensus}

	scheduler.Every(1).Minutes().NotImmediately().Run(func() {
		// Save all young transactions and filter out stale transactions.
		i := 0
		for _, transaction := range self.MemPool {
			// If transaction is younger than 24 hours
			if time.Now().Sub(time.Unix(transaction.Timestamp, 0)).Hours() <= 24 {
				self.MemPool[i] = transaction
				i++
			} else {
				log.Warnf("Removing a stale transaction from the MemPool.... (%+v)", transaction)
			}
		}
		self.MemPool = self.MemPool[:i]

		// Start mining if we have enough transactions
		if len(self.MemPool) > 0 && self.IsMining == false {
			log.Infof("Starting to mine a block with %d transactions...", len(self.MemPool))

			block := self.MineBlock(&self.IsMining)

			// If we finished mining and won the race!
			if block != nil {

				// If mined block was valid and added to chain:
				if self.AddMinedBlockToChain(*block) == true {

					// Alert all other nodes of our new valid block.
					self.BroadcastBlock(*block)

					log.Info("We just mined a new block and added it to the chain!")
				} else {
					log.Warn("The block we just mined was not valid! It was not added to the chain and the UTXO was not updated!")
				}
			} else {
				log.Info("We didn't mine a block in time, another node got the next block before us (or mining was canceled as there were no valid transactions).")
			}
		}
	})

	if hostJSONEndpoints {
		router := gin.Default()
		router.POST("/cosmosis/newTransaction", newTransaction)
		router.GET("/cosmosis/getChain", getChain)
		router.GET("/cosmosis/getUTXOs", getUTXOs)
		router.GET("/cosmosis/getMemPool", getMemPool)
		router.Use(cors.Default())
		go router.Run(":9000")
	}

	self.Start(seedNodeIPs)
}

func newTransaction(c *gin.Context) {
	var json core.Transaction
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	self.AddTransactionToMemPool(json)

	c.JSON(200, gin.H{
		"received": true,
	})
}

func getChain(c *gin.Context) {
	c.JSON(200, self.Chain)
}

func getUTXOs(c *gin.Context) {
	c.JSON(200, self.UTXO)
}

func getMemPool(c *gin.Context) {
	c.JSON(200, self.MemPool)
}
