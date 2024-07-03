package Bot

import (
	"BoostTool/Core/Discord"
	"BoostTool/Core/Utils"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var EmbedColor, _ = strconv.ParseInt(strings.Replace(config.DiscordSettings.EmbedColor, "#", "", -1), 16, 32)

func PreSteps(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var invite string
	var file string
	var err error

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	if !Utils.CheckPermissions(i.Member.User.ID) {
		content := "You do not have permissions to run this command!"
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				{
					Title:       "Boost Bot Error",
					Description: content,
					Color:       int(EmbedColor),
				},
			},
		})
		return
	}

	servercode := i.Interaction.ApplicationCommandData().Options[0].StringValue()
	amount := i.Interaction.ApplicationCommandData().Options[1].IntValue()
	duration := i.Interaction.ApplicationCommandData().Options[2].IntValue()

	inviteParts := strings.Split(servercode, "/")
	count := len(inviteParts)

	if count == 4 {
		invite = inviteParts[3]
	} else if count == 2 {
		invite = inviteParts[1]
	} else {
		invite = servercode
	}

	if duration == 1 {
		file = "1 Month Tokens.txt"

	} else if duration == 3 {
		file = "3 Month Tokens.txt"
	} else {
		err = errors.New("Choose Proper Duration Type")
	}

	if err != nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				{
					Title:       "Boost Bot Error",
					Description: fmt.Sprintf("We have failed to boost a server. The error is below.\n**__%v__**", err.Error()),
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "Order Information",
							Value:  fmt.Sprintf("```\nBoost Amount: %v\nServer Invite: %v\n```", amount, invite),
							Inline: false,
						},
					},
					Color: int(EmbedColor),
				},
			},
		})

		return
	}

	Utils.ClearScreen()
	Utils.PrintASCII()

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Title:       "Boost Bot",
				Description: "We have started boosting the server provided.",
				Color:       int(EmbedColor),
			},
		},
	})

	respo, err := Discord.BoostServer(invite, int(amount), file)
	if err != nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				{
					Title:       "Boost Bot Error",
					Description: fmt.Sprintf("We have failed to boost a server. The error is below.\n**__%v__**", err.Error()),
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "Order Information",
							Value:  fmt.Sprintf("```\nBoost Amount: %v\nServer Invite: %v\n```", amount, invite),
							Inline: false,
						},
					},
					Color: int(EmbedColor),
				},
			},
		})

		return
	}

	var success string
	if len(respo.SuccessTokens) != 0 {
		success = strings.Join(respo.SuccessTokens, "\n")
	} else {
		success = "No succeeded tokens."
	}

	var failed string
	if len(respo.FailedTokens) != 0 {
		failed = strings.Join(respo.FailedTokens, "\n")
	} else {
		failed = "No failed tokens."
	}
	descrip := fmt.Sprintf("``🔗`` Invite Link: **[%v](https://discord.gg/%v)**\n``💎`` Amount: **%v**\n``📆`` Duration: **%v Month**\n``✅`` Success: **%v**\n``❌`` Failed: **%v**\n``⏰`` Elapsed Time: **%v**", invite, invite, amount, duration, respo.Success, respo.Failed, respo.ElapsedTime)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Title:       "Boosts Finished (Order Info)",
				Description: descrip,
				Color:       int(EmbedColor),
			},
		},
	})
	embed := discordgo.MessageEmbed{
		Title:       "Boost Bot",
		Description: fmt.Sprintf("We have boosted a server successfully.\n**Success**: %v | **Failed**: %v\n**Elapsed Time**: %v", respo.Success, respo.Failed, respo.ElapsedTime),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Order Information",
				Value:  fmt.Sprintf("```\nBoost Amount: %v\nServer Invite: %v\n```", amount, invite),
				Inline: false,
			},
			{
				Name:   "``📍`` Succeeded Tokens",
				Value:  fmt.Sprintf("```\n%v\n```", success),
				Inline: false,
			},
			{
				Name:   "``❌`` Failed Tokens",
				Value:  fmt.Sprintf("```\n%v\n```", failed),
				Inline: false,
			},
		},
		Color: int(EmbedColor),
	}
	s.ChannelMessageSendEmbed(config.DiscordSettings.LogsChannel, &embed)

	return
}

func stockCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	typet := i.Interaction.ApplicationCommandData().Options[0].StringValue()

	// Responder con un mensaje de "pensando"
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	switch typet {
	case "All":
		go showCombinedStock(s, i)
	case "1":
		go sendResponse(s, i, "1 Month", Utils.Get1mTokens())
	case "3":
		go sendResponse(s, i, "3 Months", Utils.Get3MTokens())
	default:
		content := "Invalid type provided. Please specify 'All', '1', or '3'."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
	}
}

func showCombinedStock(s *discordgo.Session, i *discordgo.InteractionCreate) {
	value1 := Utils.Get1mTokens()
	value3 := Utils.Get3MTokens()

	boosts1 := value1 * 2
	boosts3 := value3 * 2

	content := fmt.Sprintf("**__1 Month Stock:__**\n`📦` Tokens → %d\n`🚀` Boosts → %d\n\n**__3 Months Stock:__**\n`📦` Tokens → %d\n`🚀` Boosts → %d", value1, boosts1, value3, boosts3)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Title:       "All Stock",
				Description: content,
				Color:       int(EmbedColor),
			},
		},
	})
}

func sendResponse(s *discordgo.Session, i *discordgo.InteractionCreate, typet string, value int) {
	boosts := value * 2
	content := fmt.Sprintf("**__%s Stock:__**\n`📦` Tokens → %d\n`🚀` Boosts → %d", typet, value, boosts)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Title:       "Boost Bot Stock",
				Description: content,
				Color:       int(EmbedColor),
			},
		},
	})
}

func restockCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	if !Utils.CheckPermissions(i.Member.User.ID) {
		content := "You do not have permissions to run this command!"
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				{
					Title:       "Boost Bot Error",
					Description: content,
					Color:       int(EmbedColor),
				}},
		})
		return
	}

	options := i.ApplicationCommandData().Options
	var lines []string
	var typet int64

	for _, option := range options {
		switch option.Name {
		case "type":
			typet = option.IntValue()
		case "code":
			urlInput := option.StringValue()
			if !strings.HasPrefix(urlInput, "https://") {
				// Si no es una URL completa, asumimos que es un ID corto
				urlInput = "https://paste.ee/d/" + urlInput
			} else if strings.Contains(urlInput, "/p/") {
				// Si es un enlace de vista, lo convertimos a enlace de descarga
				urlInput = strings.Replace(urlInput, "/p/", "/d/", 1)
			} else if !strings.Contains(urlInput, "/d/") {
				// Si no contiene ni /p/ ni /d/, añadimos /d/
				urlInput = strings.TrimSuffix(urlInput, "/") + "/d/"
			}
			lines = getContentFromURL(urlInput)
		case "file":
			// Manejo específico para archivos adjuntos
			attachmentID := option.Value.(string)
			attachment := i.ApplicationCommandData().Resolved.Attachments[attachmentID]
			if attachment != nil {
				lines = getContentFromURL(attachment.URL)
			}
		}
	}

	if len(lines) == 0 {
		content := "No tokens found to add"
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	var fileName string
	if typet == 1 {
		fileName = "1 Month Tokens.txt"
	} else if typet == 3 {
		fileName = "3 Month Tokens.txt"
	} else {
		content := "Invalid token type selected"
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	err := appendLinesToFile(lines, fileName)
	if err != nil {
		content := "Failed to add tokens to the file"
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	content := fmt.Sprintf("Added %d tokens to %s successfully", len(lines), fileName)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}

func getContentFromURL(url string) []string {
	resp, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return strings.Split(string(b), "\n")
}

func appendLinesToFile(lines []string, fileName string) error {
	file, err := os.OpenFile("./Data/"+fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			_, err := writer.WriteString(line + "\n")
			if err != nil {
				return err
			}
		}
	}
	return writer.Flush()
}

func sendCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var file string
	var tokens []string

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	if !Utils.CheckPermissions(i.Member.User.ID) {
		content := "You do not have permissions to run this command!"
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				{
					Title:       "Boost Bot Error",
					Description: content,
					Color:       int(EmbedColor),
				},
			},
		})
		return
	}
	typed := int(i.Interaction.ApplicationCommandData().Options[0].IntValue())
	amount := int(i.Interaction.ApplicationCommandData().Options[1].IntValue())
	member := i.Interaction.ApplicationCommandData().Options[2].UserValue(s).ID

	_, err := os.Create("./Data/tokens.txt")
	if err != nil {
		Utils.LogError(err.Error(), "", "")
		return
	}

	file123, err := os.Open("./Data/tokens.txt")
	if err != nil {
		Utils.LogError(err.Error(), "", "")
		return
	}

	if typed == 3 {
		file = "3 Month Tokens.txt"
	} else if typed == 1 {
		file = "1 Month Tokens.txt"
	}

	for i := 0; i < amount; i++ {
		token12 := Utils.SendToken(file)
		tokens = append(tokens, token12+"\n")
	}

	for _, tokens1 := range tokens {
		Utils.AppendTextToFile(tokens1, "tokens.txt")
	}

	fileToSend := &discordgo.File{
		Name:        "tokens.txt", // Name of the file when received by the user
		Reader:      file123,
		ContentType: "text/plain", // Adjust the content type as needed
	}

	channel, err := s.UserChannelCreate(member)
	channelid := channel.ID

	_, err = s.ChannelMessageSendComplex(channelid, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title:       "Boost Bot Tokens",
			Description: fmt.Sprintf("Here are **%v** Tokens!", amount),
			Color:       int(EmbedColor),
		},
		Files: []*discordgo.File{fileToSend},
	})

	if err != nil {
		Utils.LogError(err.Error(), "", "")
		content := "Failed Sending Tokens, User has DM's Disabled. Tokens Returned to File!"
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})

		for _, tokens1 := range tokens {
			Utils.AppendTextToFile(tokens1, file)
		}

	} else {
		content := "Successfully Sent Tokens!"
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})

	}

	_ = file123.Close()

	err = os.Remove("./Data/tokens.txt")
	if err != nil {
		Utils.LogError(err.Error(), "", "")
	}

	return
}

func CaptchaStat(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var Captcha Discord.CaptchaStruct
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	client := http.Client{Timeout: time.Second * 60}
	req, _ := http.NewRequest("GET", "https://massdm.guide/Getting+Started/Resources#Captcha+Solvers", nil)
	resp, _ := client.Do(req)
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &Captcha)
	content := fmt.Sprintf("**__Hcoptcha:__**\nState: **%v**\nWorking: **%v**\n\n**__Capsolver:__**\nState: **%v**\nWorking: **%v**\n\n```\n❌  State 1)  Everything is broken to the brim, captcha will not solved & will cause token locks.\n\n❗️ State 2)  Keys are still shit, however, there is a 50/50 chance of success, can cause token locks.\n\n❗️ State 3)  Keys are getting better, there is a high chance of success, low chance of token locks.\n\n✅  State 4)  Keys are 'as they should be' where there is garantueed success, no token locks.\n```", Captcha.CaptchaServices.Hcoptcha.State, Captcha.CaptchaServices.Hcoptcha.Working, Captcha.CaptchaServices.Capsolver.State, Captcha.CaptchaServices.Capsolver.Working)
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Title:       "Boost Bot Captcha Status",
				Description: content,
				Color:       int(EmbedColor),
			},
		},
	})
	if err != nil {
		Utils.LogError(err.Error(), "", "")
	}
	return
}
