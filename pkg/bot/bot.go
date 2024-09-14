package bot

import (
	"encoding/json"
	"goBot/goUnits/logger/logger"
	"os"
	"time"

	tgbotapi "github.com/ijnkawakaze/telegram-bot-api"
	// "github.com/spf13/viper"
)

var Bot *tgbotapi.BotAPI
var Err error

type Config struct {
	Token               string `json:"token"`
	Loglevel            int    `json:"loglevel"`
	EnableDebug         bool   `json:"enable_debug"`
	EnableActionMessage bool   `json:"enable_action_message"`
}

var Configptr *Config

// 检查配置文件
func CheckConfigFile() bool {
	cfginfo, err := os.Stat("config.json")
	if err != nil {
		// 如果文件不存在或无法读取，创建新的 config.json 文件
		logger.Error("Error in reading config! \n Reason:%s", err)
		logger.Debug("config info:%v \n Trying to create config", cfginfo)
		time.Sleep(2 * time.Second)
		// 创建新的配置
		newconfig := Config{
			Token:               "YOUR_BOT_TOKEN",
			Loglevel:            1,
			EnableDebug:         true,
			EnableActionMessage: false,
		}

		// 将配置转换为 JSON 格式
		data, err := json.MarshalIndent(newconfig, "", "  ")
		if err != nil {
			logger.Error("Error marshaling config: %s", err)
			return false
		}

		// 写入文件
		err = os.WriteFile("config.json", data, 0755)
		if err != nil {
			logger.Error("Error writing config file: %s", err)
			return false
		}
		logger.Info("Successfully created config.json\n Please edit your config")
		time.Sleep(10 * time.Second)
		os.Exit(1)
	} else {
		// 文件存在
		return true
	}
	return false
}

// 加载配置文件
func getConfig() {

	if CheckConfigFile() {
		configFile, err := os.Open("config.json")
		if err != nil {
			logger.Error("Error in reading config! \n Reason:%s", err)
		}
		defer configFile.Close()
		var Config Config
		Configptr = &Config
		decoder := json.NewDecoder(configFile)
		if err := decoder.Decode(&Configptr); err != nil {
			logger.Error("Error decoding config file: %s", err)
			return

		}
	}
}

// 从配置文件获取token
func GetToken(file string) (token string) {
	getConfig()
	token = Configptr.Token
	return token

}

// 初始化bot
func InitBot(file string) {
	token := GetToken(file)
	Bot, Err = tgbotapi.NewBotAPI(token)
	if Err != nil {
		logger.Error("Failed to create Telegram bot: %v", Err)
	}
	logger.Info("Authorized on account %s", Bot.Self.UserName)
}

// 获取日志等级
func GetLogLevel(file string) (LogLevel int, EnableDebug bool) {

	if CheckConfigFile() {
		configFile, err := os.Open(file)
		if err != nil {
			logger.Error("Error in reading config! \n Reason:%s", err)
		}
		defer configFile.Close()
		var config Config
		decoder := json.NewDecoder(configFile)
		if err := decoder.Decode(&config); err != nil {
			logger.Error("Error decoding config file: %s", err)

		}

		if err != nil {
			logger.Error("%s", err)
		}
		LogLevel = config.Loglevel
		EnableDebug = config.EnableDebug
		return LogLevel, EnableDebug
	} else {
		logger.Error("Fail to read config, please check it")
	}
	return
}

type UserInfo struct {
	Id        int64
	Name      string
	Groupid   int64
	Groupname string
	Status    string
}

// 列出用户信息
func ListUserInfo(uid int64, gid int64) UserInfo {
	var User UserInfo
	User.Id = uid
	User.Groupid = gid
	USerconfig := tgbotapi.ChatConfigWithUser{
		ChatID: gid,
		UserID: uid,
	}
	chatMenberConfig := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: USerconfig,
	}
	getChatMenberBot, err := Bot.GetChatMember(chatMenberConfig)
	if err != nil {
		logger.Error("Fail to get chat menber:%s", err)

	}
	User.Status = getChatMenberBot.Status

	ChatConfig := tgbotapi.ChatConfig{
		ChatID: chatMenberConfig.ChatID,
	}
	chatConfig := tgbotapi.ChatInfoConfig{
		ChatConfig: ChatConfig,
	}

	UserConfig := tgbotapi.ChatConfig{
		ChatID: chatMenberConfig.UserID,
	}
	userConfig := tgbotapi.ChatInfoConfig{
		ChatConfig: UserConfig,
	}
	chat, err := Bot.GetChat(chatConfig)
	logger.Warn("%s", err)

	userchat, err := Bot.GetChat(userConfig)
	logger.Warn("%s", err)

	User.Groupname = chat.Title
	User.Name = userchat.UserName
	return User
}

// 获取是否ban人后发送群组提醒
func EnableActionMessage() bool {
	getConfig()
	enableActionMessage := Configptr.EnableActionMessage
	return enableActionMessage
}
