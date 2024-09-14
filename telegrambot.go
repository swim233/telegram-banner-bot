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

	// 注册指令
	b := bot.Bot.AddHandle()
	b.NewPrivateCommandProcessor("ban", banUserHandle)
	b.NewPrivateCommandProcessor("start", helpHandle)
	b.NewPrivateCommandProcessor("help", helpHandle)
	b.Run()
}

// 封禁用户
func banUserHandle(update tgbotapi.Update) error {
	args := strings.Split(update.Message.CommandArguments(), " ")
	if len(args) != 2 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "使用方法:/ban <群组id> <用户id>")
		bot.Bot.Send(msg)
		return fmt.Errorf("参数异常")
	}
	//获取群组和用户id
	groupIDStr := args[0]
	userIDStr := args[1]
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "无效的群组id格式")
		bot.Bot.Send(msg)
		return fmt.Errorf("无效的群组id格式")
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "无效的用户id格式")
		bot.Bot.Send(msg)
		return fmt.Errorf("无效的用户id格式")
	}

	//检查/记录id
	checkID := update.FromChat().ID
	logger.Info("CheckID:%d", checkID)

	//检查操作用户权限
	if bot.Bot.IsAdminWithPermissions(checkID, groupID, tgbotapi.AdminCanRestrictMembers) {
		_, err = bot.Bot.BanChatMember(groupID, userID)
		if err != nil {
			msg := tgbotapi.NewMessage(update.FromChat().ID, "无法封禁用户: "+err.Error())
			bot.Bot.Send(msg)
			user := bot.ListUserInfo(userID, groupID)
			Chatinfo := fmt.Sprintf("ChatInfo \n  id:%d \n name:%s \n groupid:%d \n groupname:%s \n status:%s", user.Id, user.Name, user.Groupid, user.Groupname, user.Status)
			msg1 := tgbotapi.NewMessage(update.FromChat().ID, Chatinfo)
			bot.Bot.Send(msg1)
			logger.Error("%s", err.Error())

		} else {
			msg := tgbotapi.NewMessage(update.FromChat().ID, "成功封禁用户!")
			bot.Bot.Send(msg)
			user := bot.ListUserInfo(userID, groupID)
			Chatinfo := fmt.Sprintf("ChatInfo id:%d \n name:%s \n groupid:%d \n groupname:%s \n status:%s", user.Id, user.Name, user.Groupid, user.Groupname, user.Status)
			msg1 := tgbotapi.NewMessage(update.FromChat().ID, Chatinfo)
			bot.Bot.Send(msg1)

			if bot.EnableActionMessage() {
				msg := tgbotapi.NewMessage(user.Groupid, "用户"+user.Name+"已被封禁!")
				bot.Bot.Send(msg)
			}
		}
	} else {
		msg := tgbotapi.NewMessage(update.FromChat().ID, "你没有权限操作！")
		bot.Bot.Send(msg)
		user := bot.ListUserInfo(userID, groupID)
		Chatinfo := fmt.Sprintf("ChatInfo id:%d \n name:%s \n groupid:%d \n groupname:%s \n status:%s", user.Id, user.Name, user.Groupid, user.Groupname, user.Status)
		msg1 := tgbotapi.NewMessage(update.FromChat().ID, Chatinfo)
		bot.Bot.Send(msg1)
		logger.Error("Found access denied ID:%d", update.FromChat().ID)
	}
	return nil
}

// 发送帮助
func helpHandle(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "使用方法:/ban <群组id> <用户id>")
	bot.Bot.Send(msg)
	return nil
}
