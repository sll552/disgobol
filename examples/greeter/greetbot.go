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
		Name:          "Greetbot",
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
			Name:        "hello",
			Description: "Responds with greeting for the given name",
			Args:        []disgobol.CommandArg{{Name: "Who", Description: "Who to greet", Example: "test"}},
			Function: func(msg disgobol.MsgContext) string {
				return "hello " + msg.Args["Who"].(string)
			}})
	if err != nil {
		fmt.Println("Error while adding command: " + err.Error())
	}

	err = bot.Run()
	fmt.Println("Error while executing bot: " + err.Error())
}
