package handlers

import (
	"context"
	"tgbot/internal/database/repositories"
	"tgbot/internal/message"
	"time"

	tf "github.com/vitaliy-ukiru/telebot-filter/v2/telefilter"
	tele "gopkg.in/telebot.v4"
)

func (h *Handler) HandleSupport() {
	h.dispatcher.Handle("/my_report", tf.RawHandler{Callback: h.MyReports})
}

func (h *Handler) MyReports(c tele.Context) error {
	user_id := c.Sender().ID

	reportRes := repositories.NewReportRepository(h.mongoDb)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reports, err := reportRes.GetAllReportByUserId(ctx, user_id)

	if err != nil {
		return c.Send(message.NoReports)
	}

	for i, report := range reports {
		if i == 0 {
			c.Send(report.ShortInfo())
		} else {
			c.Bot().Send(c.Chat(), report.ShortInfo())
		}
	}

	return nil
}
