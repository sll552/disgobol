package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sll552/disgobol"
)

// You should get these either from a config file or via arguments/environment
var (
	BotToken    = os.Getenv("BOT_TOKEN")
	BotClientID = os.Getenv("BOT_CLIENTID")
	BotPrefix   = "!"
)

func main() {
	var err error
	bot := disgobol.Bot{
		Token:         BotToken,
		CommandPrefix: BotPrefix,
		ClientID:      BotClientID}
	invurl, err := bot.GetInviteURL()

	if err != nil {
		fmt.Println("Error getting invite URL: " + err.Error())
	} else {
		fmt.Println("Invite URL: " + invurl)
	}

	err = bot.AddCommand(
		disgobol.BotCommand{
			Name:        "ping",
			Description: "Responds with pong and time taken",
			Function: func(msg disgobol.MsgContext) string {
				tbef := time.Now()
				resp, err := msg.RespondSimple("Pong")
				if err != nil {
					fmt.Println("Error while sending message: " + err.Error())
				}
				taft := time.Now()
				_, err = resp.EditSimple(resp.Content + ", time taken: " + taft.Sub(tbef).String())
				if err != nil {
					fmt.Println("Error while editing message: " + err.Error())
				}
				return ""
			}})
	if err != nil {
		fmt.Println("Error while adding command: " + err.Error())
	}

	err = bot.Run()
	fmt.Println("Error while executing bot: " + err.Error())
}
