package main

import (
	// "encoding/json"
	"fmt"
	"time"

	// "io/ioutil"
	"os"
	"os/signal"

	// "strings"
	"log"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
)

//define game struct to hold game data from scrape
type Game struct {
	Name        string
	Price       string
	ReleaseDate string
}


func generateGenres() map[string]string {
	// create map of genres to their tag IDs
	genreMap := make(map[string]string)
	genreMap["action"] = "19"
	genreMap["indie"] = "492"
	genreMap["singleplayer"] = "4182"
	genreMap["adventure"] = "21"
	genreMap["simulation"] = "599"
	genreMap["casual"] = "597"
	genreMap["rpg"] = "122"
	genreMap["strategy"] = "9"
	genreMap["2d"] = "3871"
	genreMap["multiplayer"] = "3859"

	return genreMap
}

// function to trim first character from a string
func trimFirstChar(s string) string {
    for i := range s {
        if i > 0 {
            return s[i:]
        }
    }
    return ""
}

//function to format the games into string form for printing as a message
func formatGames(games []Game) string {
	formattedGames := ""
	for _, game := range games {
		formattedGames += "\nTitle: *" + game.Name + "* \nPrice:(original vs discounted) *" + game.Price + "* \nRelease: *" + game.ReleaseDate + "*\n"
	}
	return formattedGames
}

func getEnvVariable(key string) string {
	//method to retrieve bot token from .env
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Problem loading .env file.")
	}
	return os.Getenv(key)
}

func scrapeSteam(url string) []Game {
	//set up colly collector instance
	c := colly.NewCollector()
	//set request timeout limit
	c.SetRequestTimeout(120 * time.Second)

	//set up a slice of games
	games := make([]Game, 0)

	//scrape selectors for name, releasedate, and discounted prices.
	c.OnHTML("a.search_result_row", func(e *colly.HTMLElement) {
		e.ForEach("div.responsive_search_name_combined", func(i int, h *colly.HTMLElement) {
			//only grab games that are discounted
			if h.ChildText("div.discounted") != "" {
				newGame := Game{}
				newGame.Name = h.ChildText("span.title")
				newGame.ReleaseDate = h.ChildText("div.search_released")
				newGame.Price = h.ChildText("div.discounted")
				games = append(games, newGame)
			}
		})
	})

	//callbacks for logging in terminal for troubleshooting
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Got a response from", r.Request.URL)
	})

	c.OnError(func(r *colly.Response, e error) {
		fmt.Println("Received error:", e)
	})

	//visit to begin scraping 
	c.Visit(url)
	
	return games
}

func createMessage(session *discordgo.Session, message *discordgo.MessageCreate) {
	//func for creating messages from the bot to send to channel
	genreList := generateGenres()
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
	if message.Content == "!genres" {
		genreNames := ""
		for key := range genreList {
			genreNames += (key + "\n")
		}
		_, err := session.ChannelMessageSend(message.ChannelID, genreNames)
		if err != nil {
			fmt.Println(err)
		}
	}
	if genreList[trimFirstChar(message.Content)] != "" {
		results := scrapeSteam("https://store.steampowered.com/search/?filter=topsellers&tags=" + genreList[trimFirstChar(message.Content)])
		_, err := session.ChannelMessageSend(message.ChannelID, formatGames(results))
		if err != nil {
			fmt.Println(err)
		}
	} else {
		_, err := session.ChannelMessageSend(message.ChannelID, "genre not currently supported. Try !genres for a list of all supported genres")
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