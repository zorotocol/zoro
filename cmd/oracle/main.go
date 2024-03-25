package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/zorotocol/zoro/pkg/db"
	"github.com/zorotocol/zoro/pkg/mailer"
	"github.com/zorotocol/zoro/pkg/multirun"
	"github.com/zorotocol/zoro/pkg/oracle"
	"log"
	"net/url"
	"os"
	"time"
)

func main() {
	globalCtx := multirun.Signal(os.Kill, os.Interrupt)
	sqlDB := must(sql.Open("postgres", os.Getenv("DB")))
	defer sqlDB.Close()
	database := &db.DB{
		PG: sqlDB,
	}
	smtpURI := must(url.Parse(os.Getenv("SMTP")))
	mailerInstance := &mailer.Mailer{
		DB:     database,
		Delay:  time.Second * 3,
		Sender: SMTPSender(smtpURI),
	}
	ora := oracle.Oracle{
		EthClient: must(ethclient.Dial(os.Getenv("NODE"))),
		DB:        database,
		Contracts: []common.Address{common.HexToAddress(os.Getenv("CONTRACT"))},
		Salt:      []byte(os.Getenv("SALT")),
		Mailer:    mailerInstance,
		Finality:  1,
	}
	log.Println("start")
	err := multirun.Run(globalCtx,
		func(ctx context.Context) error {
			for {
				if err := ora.ProcessNextBlock(ctx); err != nil {
					if errors.Is(err, oracle.ErrVaporBlock) {
						time.Sleep(time.Millisecond * 1500) //1.5s
						continue
					} else if oracle.IsErrDuplicateInsert(err) {
						log.Println("another working oracle detected")
						continue
					}
					return err
				}
			}
		},
		func(ctx context.Context) error {
			return mailerInstance.Start(ctx)
		},
	)
	log.Fatalln(err)
}
