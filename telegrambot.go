package main

import (
	"goBot/goUnits/logger/logger"
	"goBot/pkg/bot"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/ijnkawakaze/telegram-bot-api"
	//tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// var bot.Bot, Err = tgbotapi.NewBotAPI(bot.GetToken("config.json"))

type Config struct {
	Token    string `json:"token"`
	Loglevel int    `json:"loglevel"`
}

func main() {
	bot.InitBot("config.json")
	if bot.Bot.Token == "" {
		logger.Debug("%s", bot.Bot.Token)
		logger.Error("Fail to read bot toke,please check config.json")
		time.Sleep(10 * time.Microsecond)
		return
	}
	logger.SetLogLevel(bot.GetLogLevel("config.json"))
	if bot.GetLogLevel("config.json") <= 5 {
		bot.Bot.Debug = true
	}
	if bot.Err != nil {
		logger.Error("%s", bot.Err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 600

	updates := bot.Bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil { // ignore not message update
			continue
		}
		// Check is private chat
		if update.Message.Chat.IsPrivate() && strings.HasPrefix(update.Message.Text, "/ban") {
			args := strings.Split(update.Message.Text, " ")
			if len(args) != 3 {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Usage: /ban <group_id> <user_id>")
				bot.Bot.Send(msg)

				continue
			}

			groupIDStr := args[1]
			userIDStr := args[2]
			groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid group_id format.")
				bot.Bot.Send(msg)
				continue
			}

			userID, err := strconv.ParseInt(userIDStr, 10, 64)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid user_id format.")
				bot.Bot.Send(msg)
				continue
			}
			chatConfig := tgbotapi.ChatInfoConfig{
				ChatConfig: tgbotapi.ChatConfig{ChatID: groupID},
			}
			groupNameSrt, _ := bot.Bot.GetChat(chatConfig)
			checkID := update.Message.Chat.ID
			logger.Info("CheckID:%d", checkID)
			if bot.VerifiedUser(checkID, groupID, groupNameSrt.Title) {

				_, err = bot.Bot.BanChatMember(groupID, userID)
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Failed to ban user: "+err.Error())
					bot.Bot.Send(msg)
					logger.Error("%s", err.Error())
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "User banned successfully!")
					bot.Bot.Send(msg)
				}
			} else {
				msg := tgbotapi.NewMessage((update.Message.Chat.ID), "You have no access to action")
				bot.Bot.Send(msg)
				logger.Error("Found access dinded ID:%d", update.Message.From.ID)
			}
		}
		if update.Message.Chat.IsPrivate() && strings.HasPrefix(update.Message.Text, "/help") {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Usage: /ban <group_id> <user_id>")
			bot.Bot.Send(msg)
		}

	}
}
