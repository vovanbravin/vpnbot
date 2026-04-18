package handlers

import (
	"context"
	"strconv"
	"strings"
	"tgbot/internal/database/models"
	"tgbot/internal/database/repositories"
	"tgbot/internal/keyboards"
	"tgbot/internal/message"
	"time"

	tf "github.com/vitaliy-ukiru/telebot-filter/v2/telefilter"
	tele "gopkg.in/telebot.v4"
)

func (h *Handler) HandleSupportCallbacks(c tele.Context) error {
	data := strings.TrimSpace(c.Callback().Data)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	reportRepo := repositories.NewReportRepository(h.mongoDb)

	switch {
	case strings.HasPrefix(data, "support_active_report_"):
		current, _ := strconv.Atoi(strings.TrimPrefix(data, "support_active_report_"))
		return h.UserShowReport(c, current, true)
	case strings.HasPrefix(data, "support_closed_report_"):
		current, _ := strconv.Atoi(strings.TrimPrefix(data, "support_closed_report_"))
		return h.UserShowReport(c, current, false)
	}

	switch data {
	case "support_active_reports":
		userId := c.Sender().ID
		reports, err := reportRepo.GetAllByStatusesAndUserId(ctx, userId, []models.ReportStatus{models.ReportStatusNew, models.ReportStatusInProgress})

		if err != nil {
			return c.Send(message.ErrorText + err.Error())
		}

		if len(reports) == 0 {
			return c.Bot().Respond(c.Callback(), &tele.CallbackResponse{Text: message.NoActiveReports})
		}

		return c.EditOrSend(reports[0].DetailInfo(), keyboards.GetUserNavigationButtonsReport(true, 1, len(reports)))

	case "support_closed_reports":
		userId := c.Sender().ID
		reports, err := reportRepo.GetAllByStatusesAndUserId(ctx, userId, []models.ReportStatus{models.ReportStatusResolved})

		if err != nil {
			return c.Send(message.ErrorText + err.Error())
		}

		if len(reports) == 0 {
			return c.Bot().Respond(c.Callback(), &tele.CallbackResponse{Text: message.NoClosedReports})
		}

		return c.EditOrSend(reports[0].DetailInfo(), keyboards.GetUserNavigationButtonsReport(false, 1, len(reports)))

	case "support_menu":
		return h.MyReports(c)
	}

	return c.Send(message.UnknownCommand)
}

func (h *Handler) HandleSupport() {
	h.dispatcher.Handle("/my_report", tf.RawHandler{Callback: h.MyReports})
}

func (h *Handler) MyReports(c tele.Context) error {
	return c.EditOrSend(message.MyReport, keyboards.GetUserReportMenu())
}

func (h *Handler) UserShowReport(c tele.Context, current int, active bool) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	reportRepo := repositories.NewReportRepository(h.mongoDb)

	var reports []models.Report
	var err error

	userId := c.Sender().ID

	if active {
		reports, err = reportRepo.GetAllByStatusesAndUserId(ctx, userId, []models.ReportStatus{models.ReportStatusNew, models.ReportStatusInProgress})
	} else {
		reports, err = reportRepo.GetAllByStatusesAndUserId(ctx, userId, []models.ReportStatus{models.ReportStatusResolved})
	}

	if err != nil {
		return c.Send(message.ErrorText + err.Error())
	}

	return c.EditOrSend(reports[current-1].DetailInfo(), keyboards.GetUserNavigationButtonsReport(active, current, len(reports)))
}
