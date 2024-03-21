package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/joho/godotenv/autoload"
	"github.com/zorotocol/oracle"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/url"
	"os"
	"path"
	"time"
)

func main() {
	ora := oracle.Oracle{
		EthClient:  oracle.Must(ethclient.Dial(os.Getenv("NODE"))),
		Collection: oracle.Must(mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGO")))).Database(path.Base(oracle.Must(url.Parse(os.Getenv("MONGO"))).Path)).Collection("blocks"),
		Contracts:  []common.Address{common.HexToAddress(os.Getenv("CONTRACT"))},
		Salt:       []byte(os.Getenv("SALT")),
	}
	for {
		if err := ora.ProcessNextBlock(context.TODO()); err != nil {
			fmt.Println(err)
			fmt.Println("sleep")
			time.Sleep(time.Second * 3)
		} else {
			fmt.Println("done")
		}
	}
}
