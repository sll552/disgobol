package main

import (
	"fmt"
	"time"

	"github.com/sll552/disgobol"
)

// You should get these either from a config file or via arguments/env
const (
	BotToken    = "123123"
	BotClientID = "123123"
	BotPrefix   = "!"
)

func main() {
	var err error
	bot := disgobol.Bot{
		Token:         BotToken,
		CommandPrefix: BotPrefix,
		ClientID:      BotClientID}
	invurl, _ := bot.GetInviteURL()

	fmt.Println("Invite URL: " + invurl)

	err = bot.AddCommand("ping", "Responds with pong and time taken", func(msg disgobol.MsgContext) string {
		tbef := time.Now()
		resp, err := msg.RespondSimple("Pong")
		if err != nil {
			fmt.Println("Error while sending message: " + err.Error())
		}
		taft := time.Now()

		_, err = resp.EditSimple(resp.Content + " time taken: " + taft.Sub(tbef).String())
		if err != nil {
			fmt.Println("Error while editing message: " + err.Error())
		}

		return ""
	})
	if err != nil {
		fmt.Println("Error while adding command: " + err.Error())
	}

	err = bot.Run()
	fmt.Println("Error while executing bot: " + err.Error())
}
