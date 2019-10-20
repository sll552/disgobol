package main

import (
	"fmt"
	"os"

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
			Name:        "horn",
			Description: "Plays airhorn sound in the channel of the requesting user",
			Function: func(msg disgobol.MsgContext) string {

				return ""
			}})
	if err != nil {
		fmt.Println("Error while adding command: " + err.Error())
	}

	err = bot.Run()
	fmt.Println("Error while executing bot: " + err.Error())
}
