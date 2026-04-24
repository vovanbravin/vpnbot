package main

import (
	"log"
	"os"
	"tgbot/internal/client"
	"tgbot/internal/database"
	fsm "tgbot/internal/fsm"
	"tgbot/internal/handlers"
	"tgbot/internal/logger"
	"tgbot/internal/utils"
	"time"

	"github.com/subosito/gotenv"
	"github.com/vitaliy-ukiru/telebot-filter/v2/dispatcher"
	tele "gopkg.in/telebot.v4"
)

func init() {
	if err := gotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {

	logLevel, _ := os.LookupEnv("LOG_LEVEl")

	maxSize, _ := utils.GetIntFromEnv("MAX_SIZE_FILE_LOG")
	maxBackups, _ := utils.GetIntFromEnv("MAX_BACKUPS_LOG")
	maxAge, _ := utils.GetIntFromEnv("MAX_AGE_FILE_LOG")

	config := logger.Config{
		FilePath:   "logs/bot.json",
		Level:      logLevel,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   true,
		Console:    true,
	}

	err := logger.Init(config)

	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	clientConfig, err := utils.Load3xuiClientConfig()

	if err != nil {
		logger.Error("Error load client config: " + err.Error())
	}

	client, err := client.NewClient(clientConfig)

	token, exists := os.LookupEnv("BOT_TOKEN")

	if !exists {
		logger.Error("Error to read BOT_TOKEN")
		return
	}

	mongoUrl, exists := os.LookupEnv("MONGO_URL")

	if !exists {
		logger.Error("Error to read MONGO_URL")
		return
	}

	mongoDbName, exists := os.LookupEnv("MONGO_DB")

	if !exists {
		logger.Error("Error to read MONGO_DB")
		return
	}

	mongoDb, err := database.NewMongoDb(mongoUrl, mongoDbName)

	if err != nil {
		logger.Error("Error to connect mongo db: %v", err)
		return
	}

	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tele.NewBot(pref)

	if err != nil {
		logger.Error("Error to create bot: %v", err)
		return
	}

	commands := []tele.Command{
		{Text: "/start", Description: "Запуск бота"},
		{Text: "/help", Description: "Вывести справку"},
		{Text: "/new_report", Description: "Отправить заявку в поддержку"},
		{Text: "/my_report", Description: "Посмотреть оставленные заявки"},
		{Text: "/info_report", Description: "Посмотреть статус заявки"},
	}

	if err = bot.SetCommands(commands); err != nil {
		logger.Warn("Error to set commands: %v", err)
	}

	logger.Debug("Комманды успешно установлены")

	FSM := fsm.NewFSM(mongoDb)

	g := bot.Group()
	g.Use(FSM.Manager.WrapContext)
	dp := dispatcher.NewDispatcher(g)

	FSM.Setup(dp)

	h := handlers.New(bot, FSM.Manager, dp, mongoDb, client)
	h.Register()

	bot.Start()
}
