package main

import (
	"log"
	"os"
	"tgbot/internal/database"
	"tgbot/internal/handlers"
	"time"

	"github.com/subosito/gotenv"
	"github.com/vitaliy-ukiru/fsm-telebot/v2"
	"github.com/vitaliy-ukiru/fsm-telebot/v2/pkg/storage/memory"
	"github.com/vitaliy-ukiru/telebot-filter/v2/dispatcher"
	tele "gopkg.in/telebot.v4"
)

func init() {
	if err := gotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {

	token, exists := os.LookupEnv("BOT_TOKEN")

	if !exists {
		log.Fatal("Error to read BOT_TOKEN")
		return
	}

	mongoUrl, exists := os.LookupEnv("MONGO_URL")

	mongoDbName, exists := os.LookupEnv("MONGO_DB")

	mongoDb, err := database.NewMongoDb(mongoUrl, mongoDbName)

	if err != nil {
		log.Fatalf("Error to connect mongo db: %v", err)
		return
	}

	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tele.NewBot(pref)

	if err != nil {
		log.Fatalf("Error to create bot: %v", err)
		return
	}

	commands := []tele.Command{
		{Text: "/start", Description: "Запуск бота"},
		{Text: "/help", Description: "Вывести справку"},
		{Text: "/new_report", Description: "Отправить заявку в поддержку"},
		{Text: "/my_report", Description: "Посмотреть оставленные заявки"},
		{Text: "/info_report", Description: "Посмотерть статус заявки-"},
	}

	if err = bot.SetCommands(commands); err != nil {
		log.Fatalf("Error to set commands: %v", err)
	}

	storage := memory.NewStorage()
	fsm := fsm.New(storage)

	g := bot.Group()
	g.Use(fsm.WrapContext)
	dp := dispatcher.NewDispatcher(g)

	h := handlers.New(bot, fsm, dp, mongoDb)
	h.Register()

	bot.Start()
}
