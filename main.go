package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/sandertv/gophertunnel/query"
	"io/ioutil"
	"napim/commands"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
)

type Config struct {
	Token               string `json:"token"`
	StatusServerName    string `json:"status_server_name"`
	StatusServerAddress string `json:"status_server_address"`
}

var token string
var config = loadConfig()

func init() {
	flag.StringVar(&token, "t", "", "Discord Bot Token")
	flag.Parse()

	if token == "" {
		token = os.Getenv("BOT-TOKEN")
	}
}

func loadConfig() Config {
	cfg := Config{}

	_, err := os.Stat("config.json")
	if os.IsNotExist(err) {
		bytes, _ := json.MarshalIndent(cfg, "", " ")
		_ = ioutil.WriteFile("config.json", bytes, 0644)
	} else {
		bytes, _ := ioutil.ReadFile("config.json")
		_ = json.Unmarshal(bytes, &cfg)
	}

	return cfg
}

func main() {
	if token == "" {
		token = config.Token
	}
	if token == "" {
		fmt.Println("Bot Token couldnt found, Enter token for authentication: ")
		_, _ = fmt.Scanln(&token)
	}

	dc, err := discordgo.New("Bot " + token)

	if err != nil {
		fmt.Println("Error while creating discord session, ", err)
		return
	}

	dc.AddHandler(func(session *discordgo.Session, msg *discordgo.MessageCreate) {
		if session.State.User.ID == msg.Author.ID { // ignore messages created by bot itself
			return
		}

		yesme := false

		for _, mentioned := range msg.Mentions {
			if session.State.User.ID == mentioned.ID { // if this bot is mentioned
				yesme = true
				break
			}
		}

		if yesme {
			regex, _ := regexp.Compile("<@(.*?)>")
			text := regex.ReplaceAllString(msg.Content, "")

			if strings.HasPrefix(text, " ") {
				text = text[1:]
			}

			if strings.HasSuffix(text, " ") {
				text = text[0 : len(text)-1]
			}

			RunCommand(session, msg.Message, text)
		}
	})
	dc.Identify.Intents = discordgo.IntentsGuildMessages

	err = dc.Open()
	if err != nil {
		fmt.Println("Error while opening websocket, ", err)
		return
	}

	go func() {
		for {
			if config.StatusServerName != "" {
				info, _ := query.Do(config.StatusServerAddress)
				numplayers, ok := info["numplayers"]
				if !ok {
					numplayers = "0"
				}

				_ = dc.UpdateGameStatus(0, numplayers+" players in "+config.StatusServerName)
			}

			time.Sleep(time.Second * 5)
		}
	}()

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-signalChannel

	_ = dc.Close()
}

func RunCommand(session *discordgo.Session, msg *discordgo.Message, text string) {
	parts := strings.Split(text, " ")
	name := parts[0]
	args := parts[1:]

	cmd, ok := commands.GetCommands()[name]

	if ok {
		cmd(session, msg, args)
	} else {
		_, _ = session.ChannelMessageSend(msg.ChannelID, "Unknown command: "+name)
	}
}
