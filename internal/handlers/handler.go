package handlers

import (
	"strings"
	"tgbot/internal/database"
	"tgbot/internal/message"

	"github.com/vitaliy-ukiru/fsm-telebot/v2"
	"github.com/vitaliy-ukiru/telebot-filter/v2/dispatcher"
	tf "github.com/vitaliy-ukiru/telebot-filter/v2/telefilter"
	tele "gopkg.in/telebot.v4"
)

type Handler struct {
	bot        *tele.Bot
	fsm        *fsm.Manager
	dispatcher *dispatcher.Dispatcher
	mongoDb    *database.MongoDB
}

func New(bot *tele.Bot, fsm *fsm.Manager, dispatcher *dispatcher.Dispatcher, mongoDb *database.MongoDB) *Handler {
	return &Handler{
		bot:        bot,
		dispatcher: dispatcher,
		fsm:        fsm,
		mongoDb:    mongoDb,
	}
}

func (h *Handler) Register() {
	h.RegisterUsers()
	h.HandleSupport()
	h.HandleAdmin()
	h.dispatcher.Handle(tele.OnText, tf.RawHandler{Callback: h.HandleText})
	h.dispatcher.Handle(tele.OnCallback, tf.RawHandler{Callback: h.HandleCallbacks})
}

func (h *Handler) HandleText(c tele.Context) error {
	text := strings.TrimSpace(c.Text())

	switch text {
	default:
		return nil
	}

}

func (h *Handler) HandleCallbacks(c tele.Context) error {
	data := strings.TrimSpace(c.Callback().Data)

	switch {
	case strings.HasPrefix(data, "admin_"):
		return h.HandleAdminCallback(c)
	default:
		return c.Bot().Respond(c.Callback(), &tele.CallbackResponse{Text: message.UnknownCommand})
	}
}
