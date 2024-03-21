package main

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/joho/godotenv/autoload"
	"github.com/zorotocol/oracle"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

func main() {

	ora := oracle.Oracle{
		EthClient:  oracle.Must(ethclient.Dial(os.Getenv("NODE"))),
		Collection: oracle.Must(mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGO")))).Database(os.Getenv("DB")).Collection("blocks"),
		Contracts:  []common.Address{common.HexToAddress(os.Getenv("CONTRACT"))},
		Salt:       []byte(os.Getenv("SALT")),
		Finality:   1,
	}
	for {
		if err := ora.ProcessNextBlock(context.TODO()); err != nil {
			log.Println(err)
			time.Sleep(time.Second * 3)
		}
	}
}
