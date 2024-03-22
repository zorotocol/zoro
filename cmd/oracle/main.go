package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/joho/godotenv/autoload"
	"github.com/zorotocol/oracle"
	"github.com/zorotocol/oracle/pkg/multirun"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"net/url"
	"os"
	"time"
)

func main() {
	globalCtx := multirun.Signal(os.Kill, os.Interrupt)
	db := must(mongo.Connect(globalCtx, options.Client().ApplyURI(os.Getenv("MONGO")))).Database(os.Getenv("DB"))
	smtpURI := must(url.Parse(os.Getenv("SMTP")))
	mailerInstance := &oracle.Mailer{
		Collection: db.Collection("mailer"),
		Delay:      time.Hour,
		Sender:     oracle.SMTPSender(smtpURI),
	}
	ora := oracle.Oracle{
		EthClient:  must(ethclient.Dial(os.Getenv("NODE"))),
		Collection: db.Collection("blocks"),
		Contracts:  []common.Address{common.HexToAddress(os.Getenv("CONTRACT"))},
		Salt:       []byte(os.Getenv("SALT")),
		Mailer:     mailerInstance,
		Finality:   1,
	}
	fmt.Println("start", time.Now().Format(http.TimeFormat))
	err := multirun.Run(globalCtx,
		func(ctx context.Context) error {
			for {
				if err := ora.ProcessNextBlock(ctx); err != nil {
					if errors.Is(err, oracle.ErrVaporBlock) {
						time.Sleep(time.Second * 2)
					}
					return err
				}
			}
		},
		func(ctx context.Context) error {
			return mailerInstance.Start(ctx)
		},
	)
	fmt.Println(err)
}
