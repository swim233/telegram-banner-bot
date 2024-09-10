package main

import (
	"fmt"
	"goBot/goUnits/logger/logger"
	"goBot/pkg/bot"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/ijnkawakaze/telegram-bot-api"
)

type Config struct {
	Token       string `json:"token"`
	Loglevel    int    `json:"loglevel"`
	EnableDebug bool   `json:"enabledebug"`
}

func main() {
	//初始化配置文件
	bot.InitBot("config.json")
	if bot.Bot.Token == "" {
		logger.Debug("%s", bot.Bot.Token)
		logger.Error("Fail to read bot toke,please check config.json")
		time.Sleep(5 * time.Second)
		return
	}
	//设置日志等级
	loglevel, enabledebug := bot.GetLogLevel("config.json")
	logger.SetLogLevel(loglevel)
	if enabledebug {
		bot.Bot.Debug = true
	}
	if bot.Err != nil {
		logger.Error("%s", bot.Err)
	}

	//获取消息更新
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 600

	updates := bot.Bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil { // 忽略非更新信息
			continue
		}
		//检查是否私聊
		if update.Message.Chat.IsPrivate() && strings.HasPrefix(update.Message.Text, "/ban") {
			args := strings.Split(update.Message.Text, " ")
			if len(args) != 3 {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Usage: /ban <group_id> <user_id>")
				bot.Bot.Send(msg)

				continue
			}
			//获取群组和用户id
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

			//检查/记录id
			checkID := update.Message.Chat.ID
			logger.Info("CheckID:%d", checkID)

			//检查操作用户权限
			if bot.VerifiedUser(checkID, groupID, groupNameSrt.Title) {
				_, err = bot.Bot.BanChatMember(groupID, userID)
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Failed to ban user: "+err.Error())
					bot.Bot.Send(msg)
					user := bot.ListUserInfo(userID, groupID)
					Chatinfo := fmt.Sprintf("ChatInfo id:%d \n name:%s \n groupid:%d \n groupname:%s \n status:%s", user.Id, user.Name, user.Groupid, user.Groupname, user.Status)
					msg1 := tgbotapi.NewMessage((update.Message.Chat.ID), Chatinfo)
					bot.Bot.Send(msg1)
					logger.Error("%s", err.Error())
				} else {

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "User banned successfully!")
					bot.Bot.Send(msg)
					user := bot.ListUserInfo(userID, groupID)
					Chatinfo := fmt.Sprintf("ChatInfo id:%d \n name:%s \n groupid:%d \n groupname:%s \n status:%s", user.Id, user.Name, user.Groupid, user.Groupname, user.Status)
					msg1 := tgbotapi.NewMessage((update.Message.Chat.ID), Chatinfo)
					bot.Bot.Send(msg1)
				}
			} else {
				msg := tgbotapi.NewMessage((update.Message.Chat.ID), "You have no access to action")
				bot.Bot.Send(msg)
				user := bot.ListUserInfo(userID, groupID)
				Chatinfo := fmt.Sprintf("ChatInfo id:%d \n name:%s \n groupid:%d \n groupname:%s \n status:%s", user.Id, user.Name, user.Groupid, user.Groupname, user.Status)
				msg1 := tgbotapi.NewMessage((update.Message.Chat.ID), Chatinfo)
				bot.Bot.Send(msg1)
				logger.Error("Found access dinded ID:%d", update.Message.From.ID)
			}
		}
		//发送帮助信息
		if update.Message.Chat.IsPrivate() && strings.HasPrefix(update.Message.Text, "/help") {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Usage: /ban <group_id> <user_id>")
			bot.Bot.Send(msg)
		}

	}
}
