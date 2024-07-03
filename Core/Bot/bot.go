package Bot

import (
	"BoostTool/Core/Utils"
	"flag"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

var s *discordgo.Session
var config, _ = Utils.LoadConfig()

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "boost",
			Description: "Boost a Server",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "invite",
					Description: "Invite Code of Server",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "amount",
					Description: "Amount of Boosts (Even Number Only)",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
				{
					Name:        "months",
					Description: "Number of Months (1 or 3)",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
			},
		},
		{
			Name:        "stock",
			Description: "Boost Bot Stock",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "type",
					Description: "Type of Tokens (1, 3, or All Stock)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "1 Month",
							Value: "1",
						},
						{
							Name:  "3 Months",
							Value: "3",
						},
						{
							Name:  "All Stock",
							Value: "All",
						},
					},
				},
			},
		},
		{
			Name:        "restock",
			Description: "Add tokens to the tokens file",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "type",
					Description: "Type of Tokens (1 or 3)",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
				{
					Name:        "code",
					Description: "Use Code From Paste.ee URl (Ex. https://paste.ee/p/xxxxx)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
				},
				{
					Name:        "file",
					Description: "Upload a file containing tokens",
					Type:        discordgo.ApplicationCommandOptionAttachment,
					Required:    false,
				},
			},
		},
		{
			Name:        "send",
			Description: "Send tokens from the token list",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "type",
					Description: "Type of Tokens (1 or 3)",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
				{
					Name:        "amount",
					Description: "The amount of tokens",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
				{
					Name:        "recipient",
					Description: "Who are you Sending the Tokesn to?",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    true,
				},
			},
		},
		{
			Name:        "captcha-status",
			Description: "Get Current Captcha Status",
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"boost":          PreSteps,
		"stock":          stockCommand,
		"restock":        restockCommand,
		"send":           sendCommand,
		"captcha-status": CaptchaStat,
	}
)

func init() {
	var err error
	s, err = discordgo.New("Bot " + config.DiscordSettings.Token)
	if err != nil {
		Utils.LogError("Invalid Bot Parameters", "Error", err.Error())
		return
	}

	if config.DiscordSettings.BotStatus != "" {
		s.Identify.Presence.Game = discordgo.Activity{
			Name: config.DiscordSettings.BotActivity,
			Type: getActivityType(),
		}
	}

}

func getActivityType() discordgo.ActivityType {
	switch config.DiscordSettings.BotStatus {
	case "playing":
		return discordgo.ActivityTypeGame
	case "watching":
		return discordgo.ActivityTypeWatching
	case "listening":
		return discordgo.ActivityTypeListening
	case "competing":
		return discordgo.ActivityTypeCompeting
	default:
		return discordgo.ActivityTypeWatching
	}
}

func StartBot() {
	var err error
	RemoveCommands := flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		botd := s.State.User.Username + "#" + s.State.User.Discriminator
		Utils.LogInfo("Successfully Logged In", "Bot", botd)
	})
	err = s.Open()
	if err != nil {
		Utils.LogError(err.Error(), "", "")
	}
	Utils.LogInfo("Merging Commands to Guild", "Guild", config.DiscordSettings.GuildID)
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err1 := s.ApplicationCommandCreate(s.State.User.ID, config.DiscordSettings.GuildID, v)
		if err1 != nil {
			Utils.LogError(err1.Error(), "", "")
		}
		registeredCommands[i] = cmd
	}
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	Utils.LogInfo("Press Ctrl+C to Exit & Remove Commands", "", "")
	<-stop

	if *RemoveCommands {
		Utils.LogInfo("Removing Commands & Shutting Down...", "", "")

		for _, v := range registeredCommands {
			_ = s.ApplicationCommandDelete(s.State.User.ID, config.DiscordSettings.GuildID, v.ID)
		}
	}
}
