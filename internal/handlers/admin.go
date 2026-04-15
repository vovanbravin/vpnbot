package handlers

import (
	"context"
	"strconv"
	"strings"
	"tgbot/internal/database/middleware"
	"tgbot/internal/database/models"
	"tgbot/internal/database/repositories"
	"tgbot/internal/fsm"
	"tgbot/internal/keyboards"
	"tgbot/internal/message"
	"time"

	tele "gopkg.in/telebot.v4"
)

func (h *Handler) HandleAdminCallback(c tele.Context) error {
	data := strings.TrimSpace(c.Callback().Data)

	switch {
	case strings.HasPrefix(data, "admin_report_new_"):
		current, _ := strconv.Atoi(strings.TrimPrefix(data, "admin_report_new_"))
		return h.ReportShowAdmin(c, models.ReportStatusNew, current)
	case strings.HasPrefix(data, "admin_report_in_progress_"):
		current, _ := strconv.Atoi(strings.TrimPrefix(data, "admin_report_in_progress_"))
		return h.ReportShowAdmin(c, models.ReportStatusInProgress, current)
	case strings.HasPrefix(data, "admin_report_resolved_"):
		current, _ := strconv.Atoi(strings.TrimPrefix(data, "admin_report_resolved_"))
		return h.ReportShowAdmin(c, models.ReportStatusResolved, current)
	case strings.HasPrefix(data, "admin_report_hire_"):
		current, _ := strconv.Atoi(strings.TrimPrefix(data, "admin_report_hire_"))
		return h.ReportAdminHire(c, current)
	case strings.HasPrefix(data, "admin_report_answer_"):
		current, _ := strconv.Atoi(strings.TrimPrefix(data, "admin_report_answer_"))

		state, _ := h.fsm.NewContext(c)
		if err := state.SetState(context.Background(), fsm.StateAnswer); err != nil {
			return c.Send(message.ErrorText + err.Error())
		}

		if err := state.Update(context.Background(), "current", current); err != nil {
			return c.Send(message.ErrorText + err.Error())
		}

		return c.EditOrSend(message.EnterAnswer)
	}

	switch data {
	case "admin_report_new":
		return h.ReportAdminList(c, models.ReportStatusNew)
	case "admin_report_in_progress":
		return h.ReportAdminList(c, models.ReportStatusInProgress)
	case "admin_report_resolved":
		return h.ReportAdminList(c, models.ReportStatusResolved)
	case "admin_report_menu":
		return h.ReportAdminMenu(c)

	default:
		return c.Bot().Respond(c.Callback(), &tele.CallbackResponse{Text: message.UnknownCommand})
	}
}

func (h *Handler) HandleAdmin() {

	userRepo := repositories.NewUserRepository(h.mongoDb)

	adminGroup := h.bot.Group()
	adminGroup.Use(middleware.AdminOnly(userRepo))

	adminGroup.Handle("/report_admin_menu", h.ReportAdminMenu)
}

func (h *Handler) IsAdmin(c tele.Context) error {
	return c.Send(message.IsAdmin)
}

func (h *Handler) ReportAdminMenu(c tele.Context) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	reportRepo := repositories.NewReportRepository(h.mongoDb)

	newCount, err := reportRepo.CountByStatus(ctx, models.ReportStatusNew)
	if err != nil {
		return c.Send(message.ErrorText + err.Error())
	}

	inProgressCount, err := reportRepo.CountByStatus(ctx, models.ReportStatusInProgress)

	if err != nil {
		return c.Send(message.ErrorText + err.Error())
	}

	resolvedCount, err := reportRepo.CountByStatus(ctx, models.ReportStatusResolved)

	if err != nil {
		return c.Send(message.ErrorText + err.Error())
	}

	return c.EditOrSend(message.TitleMenuReports, keyboards.GetReportAdminMenu(newCount, inProgressCount, resolvedCount))
}

func (h *Handler) ReportShowAdmin(c tele.Context, status models.ReportStatus, current int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reportRepo := repositories.NewReportRepository(h.mongoDb)

	reports, err := reportRepo.GetAllByStatus(ctx, status)

	if err != nil {
		return c.Send(message.ErrorText + err.Error())
	}

	return c.Edit(reports[current-1].DetailInfo(), keyboards.GetNavigationButtonsReport(status, current, len(reports)))
}

func (h *Handler) ReportAdminList(c tele.Context, status models.ReportStatus) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reportRepo := repositories.NewReportRepository(h.mongoDb)

	reports, err := reportRepo.GetAllByStatus(ctx, status)

	if err != nil {
		return c.Send(message.ErrorText + err.Error())
	}

	if len(reports) == 0 {
		return c.Bot().Respond(c.Callback(), &tele.CallbackResponse{Text: "Не найдено заявок"})
	}

	n := len(reports)

	return c.Edit(reports[0].DetailInfo(), keyboards.GetNavigationButtonsReport(status, 1, n))
}

func (h *Handler) ReportAdminHire(c tele.Context, current int) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	reportRepo := repositories.NewReportRepository(h.mongoDb)

	reports, err := reportRepo.GetAllByStatus(ctx, models.ReportStatusNew)

	if err != nil {
		return c.Send(message.ErrorText + err.Error())
	}

	report := reports[current-1]

	report.Status = models.ReportStatusInProgress

	err = reportRepo.Update(ctx, report)

	if err != nil {
		return c.Send(message.ErrorText + err.Error())
	}

	if len(reports) == 1 {
		return h.ReportAdminMenu(c)
	} else if current == 1 {
		return h.ReportShowAdmin(c, models.ReportStatusNew, current+1)
	}

	return h.ReportShowAdmin(c, models.ReportStatusNew, current-1)
}
