package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sandertv/gophertunnel/query"
	"math/rand"
)

type Command func(session *discordgo.Session, msg *discordgo.Message, args []string)

func GetCommands() map[string]Command {
	return map[string]Command{
		"query": QueryServer,
	}
}

func QueryServer(session *discordgo.Session, msg *discordgo.Message, args []string) {
	var address string
	if len(args) > 1 {
		address = args[0] + ":" + args[1]
	} else {
		address = args[0]
	}

	go func() {
		info, err := query.Do(address)

		if err != nil {
			_, _ = session.ChannelMessageSend(msg.ChannelID, err.Error())
		} else {
			embed := discordgo.MessageEmbed{
				Type:        discordgo.EmbedTypeRich,
				Title:       "Server Query Result",
				Description: info["hostname"],
				Color:       rand.Intn(0xFFFFFF),
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "IP: ",
						Value:  info["hostip"],
						Inline: true,
					},
					{
						Name:   "Port: ",
						Value:  info["hostport"],
						Inline: true,
					},
					{
						Name:   "PlayerCount: ",
						Value:  info["numplayers"],
						Inline: true,
					},
					{
						Name:   "MaxPlayers: ",
						Value:  info["maxplayers"],
						Inline: true,
					},
					{
						Name:   "Game Version: ",
						Value:  info["version"],
						Inline: true,
					},
				},
			}

			_, _ = session.ChannelMessageSendEmbed(msg.ChannelID, &embed)
		}
	}()
}
