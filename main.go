package main

import (
	"flag"
	"fmt"
	"github.com/carlescere/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/transmissionsdev/cosmosis/core"
	"net/http"
	"os"
	"strconv"
	"time"
)

var chain core.Blockchain
var validationServerURL string

func main() {
	var operatorPublicKey string
	flag.StringVar(&operatorPublicKey, "publicKey", "", "A valid public key where funds from mining can be sent to your account")

	flag.StringVar(&validationServerURL, "validationServer", "http://0.0.0.0:1337/verifySignature", "A full url (with http://) that operates as a valid ECDSA SECP256k1 signature validation webserver. We recommend you run one locally. Go to: https://github.com/transmissionsdev/cosmosisUtils to find instructions to run one!")

	flag.Parse()

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
		fmt.Println("Your validation server URL is unreachable! See instructions to run your own here: https://github.com/transmissionsdev/cosmosisUtils\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	// --------------------------------

	chain = core.Blockchain{Chain: make([]core.Block, 0), MemPool: make([]core.Transaction, 0), UTXO: make(core.UTXO), ValidationServerURL: validationServerURL, OperatorPublicKey: core.UserPublicKey(operatorPublicKey)}

	scheduler.Every(1).Minutes().Run(func() {
		// Clear out stale transactions
		for index, transaction := range chain.MemPool {
			// If the transaction is older than 24 hours
			if time.Unix(transaction.Timestamp, 0).Sub(time.Now()).Hours() >= 24 {
				// Update the MemPool with the removed transaction gone
				chain.MemPool = core.RemoveFromTransactions(chain.MemPool, index)
			}
		}

		// Start mining if we have enough transactions
		if len(chain.MemPool) > 0 {
			go func() {

				block := chain.MineBlock()

				// If mined block was valid and added to chain
				if chain.AddMinedBlockToChain(block) == true {
					fmt.Println("We just mined a new block and added it to the chain!")
				} else {
					fmt.Println("The block we just mined was not valid! It was not added to the chain and the UTXO was not updated!")
				}
			}()
		}
	})

	router := gin.Default()
	router.GET("/validateSignature", validateSignature)

	router.Run(":8080")
}

func validateSignature(c *gin.Context) {
	sender := c.Query("sender")
	recipient := c.Query("recipient")
	amountStr := c.Query("amount")
	amount, _ := strconv.ParseFloat(amountStr, 64)
	timestampStr := c.Query("timestamp")
	timestamp, _ := strconv.Atoi(timestampStr)
	signature := c.Query("signature")

	c.JSON(200, gin.H{
		"valid": core.ValidateSignature(core.Transaction{Sender: core.UserPublicKey(sender), Recipient: core.UserPublicKey(recipient), Timestamp: int64(timestamp), Amount: amount, Signature: signature}, validationServerURL),
	})
}
