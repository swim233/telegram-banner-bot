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
	Token               string `json:"token"`
	Loglevel            int    `json:"loglevel"`
	EnableDebug         bool   `json:"enable_debug"`
	EnableActionMessage bool   `json:"enable_action_message"`
}

func main() {
	//初始化配置文件
	bot.InitBot("config.json")
	if bot.Bot.Token == "" {
		logger.Debug("%s", bot.Bot.Token)
		logger.Error("无法获取bot的token，请检查config.json")
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
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "使用方法:/ban <群组id> <用户id>")
				bot.Bot.Send(msg)

				continue
			}
			//获取群组和用户id
			groupIDStr := args[1]
			userIDStr := args[2]
			groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "无效的群组id格式")
				bot.Bot.Send(msg)
				continue
			}

			userID, err := strconv.ParseInt(userIDStr, 10, 64)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "无效的用户id格式")
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
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "无法封禁用户: "+err.Error())
					bot.Bot.Send(msg)
					user := bot.ListUserInfo(userID, groupID)
					Chatinfo := fmt.Sprintf("ChatInfo \n  id:%d \n name:%s \n groupid:%d \n groupname:%s \n status:%s", user.Id, user.Name, user.Groupid, user.Groupname, user.Status)
					msg1 := tgbotapi.NewMessage((update.Message.Chat.ID), Chatinfo)
					bot.Bot.Send(msg1)
					logger.Error("%s", err.Error())

				} else {

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "成功封禁用户!")
					bot.Bot.Send(msg)
					user := bot.ListUserInfo(userID, groupID)
					Chatinfo := fmt.Sprintf("ChatInfo id:%d \n name:%s \n groupid:%d \n groupname:%s \n status:%s", user.Id, user.Name, user.Groupid, user.Groupname, user.Status)
					msg1 := tgbotapi.NewMessage((update.Message.Chat.ID), Chatinfo)
					bot.Bot.Send(msg1)

					if bot.EnableActionMessage() {
						msg := tgbotapi.NewMessage(user.Groupid, "用户"+user.Name+"已被封禁!")
						bot.Bot.Send(msg)
					}
				}
			} else {
				msg := tgbotapi.NewMessage((update.Message.Chat.ID), "你没有权限操作！")
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
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "使用方法:/ban <群组id> <用户id>")
			bot.Bot.Send(msg)
		}

	}
}
