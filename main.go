package main

import (
	"encoding/json"
	"fmt"
	"time"

	kafka "github.com/tarmalonchik/golibs/kafkawrapper"
)

func main() {
	cli, err := kafka.NewClient(kafka.Config{
		KafkaPassword:          "password",
		KafkaUser:              "user",
		KafkaPort:              "9094",
		KafkaControllersCount:  1,
		KafkaBrokerURLTemplate: "127.0.%d.1",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	con, err := cli.NewSyncProducer("test12")
	if err != nil {
		fmt.Println(err)
		return
	}

	u := userData{
		IPAddress: "127.0.0.1",
		Tag:       "vless",
		Inbound:   234534,
		Outbound:  345346,
	}

	uRaw, _ := json.Marshal(u)

	fmt.Println(con.SendMessage(uRaw))
}

type userData struct {
	Tag       string    `json:"id"`
	Time      time.Time `json:"time"`
	Inbound   int64     `json:"inbound"`
	Outbound  int64     `json:"outbound"`
	IPAddress string    `json:"ip_address"`
}
