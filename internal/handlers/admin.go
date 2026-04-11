package handlers

import (
	"context"
	"strconv"
	"strings"
	"tgbot/internal/database/middleware"
	"tgbot/internal/database/models"
	"tgbot/internal/database/repositories"
	"tgbot/internal/keyboards"
	"time"

	tele "gopkg.in/telebot.v4"
)

func (h *Handler) HandleAdminCallback(c tele.Context) error {
	data := strings.TrimSpace(c.Callback().Data)

	switch {
	case strings.HasPrefix(data, "admin_report_new_"):
		current, _ := strconv.Atoi(strings.TrimPrefix(data, "admin_report_new_"))
		return h.ReportShowAdmin(c, models.ReportStatusNew, current)
	}

	switch data {
	case "admin_report_new":
		return h.ReportAdminList(c, models.ReportStatusNew)
	default:
		return c.Send("Неизвестная команда")
	}
}

func (h *Handler) HandleAdmin() {

	userRepo := repositories.NewUserRepository(h.mongoDb)

	adminGroup := h.bot.Group()
	adminGroup.Use(middleware.AdminOnly(userRepo))

	adminGroup.Handle("/report_admin_menu", h.ReportAdminMenu)
}

func (h *Handler) IsAdmin(c tele.Context) error {
	return c.Send("Вы админ")
}

func (h *Handler) ReportAdminMenu(c tele.Context) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	reportRepo := repositories.NewReportRepository(h.mongoDb)

	newCount, err := reportRepo.CountByStatus(ctx, models.ReportStatusNew)
	if err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	inProgressCount, err := reportRepo.CountByStatus(ctx, models.ReportStatusInProgress)

	if err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	resolvedCount, err := reportRepo.CountByStatus(ctx, models.ReportStatusResolved)

	if err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	return c.Send("🈸 Заявки", keyboards.GetReportAdminMenu(newCount, inProgressCount, resolvedCount))
}

func (h *Handler) ReportShowAdmin(c tele.Context, status models.ReportStatus, current int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reportRepo := repositories.NewReportRepository(h.mongoDb)

	reports, err := reportRepo.GetAllByStatus(ctx, status)

	if err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	return c.Edit(reports[current-1].ShortInfo(), keyboards.GetNavigationButtonsReport(status, current, len(reports)))
}

func (h *Handler) ReportAdminList(c tele.Context, status models.ReportStatus) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reportRepo := repositories.NewReportRepository(h.mongoDb)

	reports, err := reportRepo.GetAllByStatus(ctx, status)

	if err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	if len(reports) == 0 {
		return c.Send("Не найдено заявок")
	}

	n := len(reports)

	return c.Edit(reports[0].ShortInfo(), keyboards.GetNavigationButtonsReport(status, 1, n))
}
