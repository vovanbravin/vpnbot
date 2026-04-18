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
	"tgbot/internal/logger"
	"tgbot/internal/message"
	"time"

	tele "gopkg.in/telebot.v4"
)

func (h *Handler) HandleAdminCallback(c tele.Context) error {
	data := strings.TrimSpace(c.Callback().Data)

	logger.Debug("HandleAdminCallback вызван",
		"admin_id", c.Sender().ID,
		"callback_data", data,
	)

	switch {
	case strings.HasPrefix(data, "admin_report_new_"):
		current, _ := strconv.Atoi(strings.TrimPrefix(data, "admin_report_new_"))
		logger.Debug("Просмотр новых заявок", "admin_id", c.Sender().ID, "current", current)
		return h.ReportShowAdmin(c, models.ReportStatusNew, current)

	case strings.HasPrefix(data, "admin_report_in_progress_"):
		current, _ := strconv.Atoi(strings.TrimPrefix(data, "admin_report_in_progress_"))
		logger.Debug("Просмотр заявок в работе", "admin_id", c.Sender().ID, "current", current)
		return h.ReportShowAdmin(c, models.ReportStatusInProgress, current)

	case strings.HasPrefix(data, "admin_report_resolved_"):
		current, _ := strconv.Atoi(strings.TrimPrefix(data, "admin_report_resolved_"))
		logger.Debug("Просмотр решённых заявок", "admin_id", c.Sender().ID, "current", current)
		return h.ReportShowAdmin(c, models.ReportStatusResolved, current)

	case strings.HasPrefix(data, "admin_report_hire_"):
		current, _ := strconv.Atoi(strings.TrimPrefix(data, "admin_report_hire_"))
		logger.Debug("Назначение заявки администратору", "admin_id", c.Sender().ID, "current", current)
		return h.ReportAdminHire(c, current)

	case strings.HasPrefix(data, "admin_report_answer_"):
		current, _ := strconv.Atoi(strings.TrimPrefix(data, "admin_report_answer_"))

		logger.Info("Админ начал отвечать на заявку",
			"admin_id", c.Sender().ID,
			"username", c.Sender().Username,
			"report_index", current,
		)

		state, ok := h.fsm.NewContext(c)
		if !ok {
			logger.Error("Ошибка создания FSM контекста", "admin_id", c.Sender().ID)
			return c.Send(message.SomethingWrong)
		}

		if err := state.SetState(context.Background(), fsm.StateAnswer); err != nil {
			logger.Error("Ошибка установки состояния", "error", err, "admin_id", c.Sender().ID)
			return c.Send(message.ErrorText + err.Error())
		}

		if err := state.Update(context.Background(), "current", current); err != nil {
			logger.Error("Ошибка сохранения current в состояние", "error", err, "admin_id", c.Sender().ID)
			return c.Send(message.ErrorText + err.Error())
		}

		return c.EditOrSend(message.EnterAnswer)
	}

	switch data {
	case "admin_report_new":
		logger.Debug("Запрос списка новых заявок", "admin_id", c.Sender().ID)
		return h.ReportAdminList(c, models.ReportStatusNew)

	case "admin_report_in_progress":
		logger.Debug("Запрос списка заявок в работе", "admin_id", c.Sender().ID)
		return h.ReportAdminList(c, models.ReportStatusInProgress)

	case "admin_report_resolved":
		logger.Debug("Запрос списка решённых заявок", "admin_id", c.Sender().ID)
		return h.ReportAdminList(c, models.ReportStatusResolved)

	case "admin_report_menu":
		logger.Debug("Запрос меню администратора", "admin_id", c.Sender().ID)
		return h.ReportAdminMenu(c)

	default:
		logger.Warn("Неизвестная команда в админском callback",
			"admin_id", c.Sender().ID,
			"command", data,
		)
		return c.Bot().Respond(c.Callback(), &tele.CallbackResponse{Text: message.UnknownCommand})
	}
}

func (h *Handler) HandleAdmin() {
	logger.Debug("Настройка админских хендлеров")

	userRepo := repositories.NewUserRepository(h.mongoDb)

	adminGroup := h.bot.Group()
	adminGroup.Use(middleware.AdminOnly(userRepo))

	adminGroup.Handle("/report_admin_menu", h.ReportAdminMenu)

	logger.Info("Админские хендлеры настроены", "path", "/report_admin_menu")
}

func (h *Handler) IsAdmin(c tele.Context) error {
	logger.Debug("Проверка прав администратора", "user_id", c.Sender().ID)
	return c.Send(message.IsAdmin)
}

func (h *Handler) ReportAdminMenu(c tele.Context) error {
	logger.Debug("ReportAdminMenu вызван", "admin_id", c.Sender().ID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	reportRepo := repositories.NewReportRepository(h.mongoDb)

	newCount, err := reportRepo.CountByStatus(ctx, models.ReportStatusNew)
	if err != nil {
		logger.Error("Ошибка подсчёта новых заявок", "error", err, "admin_id", c.Sender().ID)
		return c.Send(message.ErrorText + err.Error())
	}

	inProgressCount, err := reportRepo.CountByStatus(ctx, models.ReportStatusInProgress)

	if err != nil {
		logger.Error("Ошибка подсчёта заявок в работе", "error", err, "admin_id", c.Sender().ID)
		return c.Send(message.ErrorText + err.Error())
	}

	resolvedCount, err := reportRepo.CountByStatus(ctx, models.ReportStatusResolved)

	if err != nil {
		logger.Error("Ошибка подсчёта решённых заявок", "error", err, "admin_id", c.Sender().ID)
		return c.Send(message.ErrorText + err.Error())
	}

	logger.Debug("Статистика заявок",
		"admin_id", c.Sender().ID,
		"new", newCount,
		"in_progress", inProgressCount,
		"resolved", resolvedCount,
	)

	return c.EditOrSend(message.TitleMenuReports, keyboards.GetReportAdminMenu(newCount, inProgressCount, resolvedCount))
}

func (h *Handler) ReportShowAdmin(c tele.Context, status models.ReportStatus, current int) error {
	logger.Debug("ReportShowAdmin вызван",
		"admin_id", c.Sender().ID,
		"status", status,
		"current", current,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reportRepo := repositories.NewReportRepository(h.mongoDb)

	reports, err := reportRepo.GetAllByStatus(ctx, status)

	if err != nil {
		logger.Error("Ошибка получения заявок", "error", err, "admin_id", c.Sender().ID, "status", status)
		return c.Send(message.ErrorText + err.Error())
	}

	if len(reports) == 0 {
		logger.Warn("Нет заявок с указанным статусом", "admin_id", c.Sender().ID, "status", status)
		return c.Send("Нет заявок")
	}

	if current-1 >= len(reports) {
		logger.Warn("Индекс заявки вне диапазона",
			"admin_id", c.Sender().ID,
			"current", current,
			"total", len(reports),
		)
		current = len(reports)
	}

	report := reports[current-1]

	logger.Debug("Показ заявки",
		"admin_id", c.Sender().ID,
		"report_number", report.Number,
		"status", status,
		"current", current,
		"total", len(reports),
	)

	return c.Edit(report.DetailInfo(), keyboards.GetAdminNavigationButtonsReport(status, current, len(reports)))
}

func (h *Handler) ReportAdminList(c tele.Context, status models.ReportStatus) error {
	logger.Debug("ReportAdminList вызван",
		"admin_id", c.Sender().ID,
		"status", status,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reportRepo := repositories.NewReportRepository(h.mongoDb)

	reports, err := reportRepo.GetAllByStatus(ctx, status)

	if err != nil {
		logger.Error("Ошибка получения заявок", "error", err, "admin_id", c.Sender().ID, "status", status)
		return c.Send(message.ErrorText + err.Error())
	}

	if len(reports) == 0 {
		logger.Info("Нет заявок с указанным статусом",
			"admin_id", c.Sender().ID,
			"status", status,
		)
		return c.Bot().Respond(c.Callback(), &tele.CallbackResponse{Text: "Не найдено заявок"})
	}

	n := len(reports)

	logger.Info("Показан список заявок",
		"admin_id", c.Sender().ID,
		"status", status,
		"total", n,
	)

	return c.Edit(reports[0].DetailInfo(), keyboards.GetAdminNavigationButtonsReport(status, 1, n))
}

func (h *Handler) ReportAdminHire(c tele.Context, current int) error {
	logger.Info("Назначение заявки администратору",
		"admin_id", c.Sender().ID,
		"username", c.Sender().Username,
		"report_index", current,
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	reportRepo := repositories.NewReportRepository(h.mongoDb)

	reports, err := reportRepo.GetAllByStatus(ctx, models.ReportStatusNew)

	if err != nil {
		logger.Error("Ошибка получения новых заявок", "error", err, "admin_id", c.Sender().ID)
		return c.Send(message.ErrorText + err.Error())
	}

	if current-1 >= len(reports) {
		logger.Error("Индекс заявки вне диапазона",
			"admin_id", c.Sender().ID,
			"current", current,
			"total", len(reports),
		)
		return c.Send("Заявка не найдена")
	}

	report := reports[current-1]

	report.Status = models.ReportStatusInProgress

	err = reportRepo.Update(ctx, report)

	if err != nil {
		logger.Error("Ошибка обновления статуса заявки",
			"error", err,
			"admin_id", c.Sender().ID,
			"report_number", report.Number,
		)
		return c.Send(message.ErrorText + err.Error())
	}

	logger.Info("Заявка назначена администратору",
		"admin_id", c.Sender().ID,
		"report_number", report.Number,
		"previous_status", models.ReportStatusNew,
		"new_status", models.ReportStatusInProgress,
	)

	if len(reports) == 1 {
		logger.Debug("Все заявки обработаны, возврат в меню", "admin_id", c.Sender().ID)
		return h.ReportAdminMenu(c)
	} else if current == 1 {
		return h.ReportShowAdmin(c, models.ReportStatusNew, current+1)
	}

	return h.ReportShowAdmin(c, models.ReportStatusNew, current-1)
}
