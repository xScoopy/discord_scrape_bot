package main

import (
	"fmt"
	"time"

	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
)

//define game struct to hold game data from scrape
type Game struct {
	Name        string
	Original 	string
	Discount    string
	ReleaseDate string
	Link 		string 
}

type MessageSection struct {
	Message string
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

//take scraped prices and separate them into 2 different price strings for data and presentation purposes. 
func separatePrices(prices string) []string {
	//trim first dollar sign
	newPrices := trimFirstChar(prices)
	original, discounted, _ := strings.Cut(newPrices, "$")
	priceSlice := make([]string, 2)
	priceSlice[0], priceSlice[1] = original, discounted
	return priceSlice
}

//function to format the games into string form for printing as a message
func formatGames(games []Game) []MessageSection {
	formattedGames := make([]MessageSection, 0)
	counter := 0
	newMessage := ""
	for _, game := range games {
		if counter < 6 {
			newMessage += "\nTitle: **" + game.Name + "** \nPrice: ~~$" + game.Original + "~~" + " **$" + game.Discount + "** \nRelease: " + game.ReleaseDate + "\nLink: <" + game.Link + ">\n"
			counter ++
		} else {
			newMessageSection := MessageSection{Message: newMessage}
			formattedGames = append(formattedGames, newMessageSection)
			counter = 0
			newMessage = ""
			newMessage += "\nTitle: **" + game.Name + "** \nPrice: ~~$" + game.Original + "~~" + " **$" + game.Discount + "** \nRelease: " + game.ReleaseDate + "\nLink: <" + game.Link + ">\n"
			counter ++
		}
	}
	return formattedGames
}

func getEnvVariable(key string) string {
	//method to retrieve bot token from .env
	err := godotenv.Load(".env")
	
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
				newGame.Link = e.Attr("href")
				newGame.Name = h.ChildText("span.title")
				newGame.ReleaseDate = h.ChildText("div.search_released")
				prices := separatePrices(h.ChildText("div.discounted"))
				newGame.Original, newGame.Discount = prices[0], prices[1]
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
	
	//ignore all messages sent by bot itself
	if message.Author.ID == session.State.User.ID {
		return
	}
	if message.Content[0] != '!' {
		return
	}
	//if a valid message is sent, generate genre list
	genreList := generateGenres()

	if message.Content == "!genres" {
		genreNames := ""
		for key := range genreList {
			genreNames += (key + "\n")
		}
		_, err := session.ChannelMessageSend(message.ChannelID, genreNames)
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	if genreList[trimFirstChar(message.Content)] != "" {
		results := scrapeSteam("https://store.steampowered.com/search/?filter=topsellers&tags=" + genreList[trimFirstChar(message.Content)])
		messages := formatGames(results)
		for i := range messages {
			_, err := session.ChannelMessageSend(message.ChannelID, messages[i].Message)
			if err != nil {
				fmt.Println(err)
			}
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