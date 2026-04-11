package handlers

import (
	tf "github.com/vitaliy-ukiru/telebot-filter/v2/telefilter"
	tele "gopkg.in/telebot.v4"
)

func (h *Handler) RegisterUsers() {
	h.dispatcher.Handle("/start", tf.RawHandler{Callback: h.Start})
	h.dispatcher.Handle("/help", tf.RawHandler{Callback: h.Help})
}

func (h *Handler) Start(c tele.Context) error {
	return c.Send(c.Sender().ID)
}

func (h *Handler) Help(c tele.Context) error {
	return c.Send("This help")
}
