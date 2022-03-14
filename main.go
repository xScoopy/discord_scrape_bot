package main

import (
	// "encoding/json"
	// "flag"
	"fmt"
	// "io/ioutil"
	// "net/http"
	"os"
	"os/signal"

	// "strings"
	"log"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func getEnvVariable(key string) string {
	//method to retrieve bot token from .env
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Problem loading .env file.")
	}
	return os.Getenv(key)
}

func createMessage(session *discordgo.Session, message *discordgo.MessageCreate) {
	//func for creating messages from the bot to send to channel

	//ignore all messages sent by bot itself
	if message.Author.ID == session.State.User.ID {
		return
	}

	//simple hello message
	if message.Content == "!hello" {
		_, err := session.ChannelMessageSend(message.ChannelID, "Ahoy there steam scrapers")
		if err != nil {
			fmt.Println(err)
		}
	}
}

func main() {
	botToken := getEnvVariable("SECRET")
    // make new instance with our bot token
    discordInstance, err := discordgo.New("Bot " + botToken)
    if err != nil {
        fmt.Println("error creating Discord session,", err)
        return
    }

    // set callback for creating messages
    discordInstance.AddHandler(createMessage)

    // Bot only used for receiving messages
    discordInstance.Identify.Intents = discordgo.IntentsGuildMessages

    // connect to discord and listen for events
    err = discordInstance.Open()
    if err != nil {
        fmt.Println("error with opening discord session:,", err)
        return
    }

    // wait for kill command
    fmt.Println("Bot is now running. Press CTRL-C to exit.")
    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
    <-sc

    // Close out session.
    discordInstance.Close()
	
}