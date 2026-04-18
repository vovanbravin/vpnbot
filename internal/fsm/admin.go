package fsm

import (
	"context"
	"fmt"
	"strings"
	"tgbot/internal/database/models"
	"tgbot/internal/database/repositories"
	"tgbot/internal/keyboards"
	"tgbot/internal/logger"
	"tgbot/internal/message"
	"time"

	"github.com/vitaliy-ukiru/fsm-telebot/v2"
	"github.com/vitaliy-ukiru/fsm-telebot/v2/fsmopt"
	"github.com/vitaliy-ukiru/telebot-filter/v2/dispatcher"
	tele "gopkg.in/telebot.v4"
)

const (
	StateAnswer        = "admin_answer"
	StateAnswerConfirm = "admin_answer_confirm"
)

func (f *FSM) SetupAdmin(dispatcher *dispatcher.Dispatcher) {
	logger.Debug("Настройка Admin FSM", "module", "admin")

	f.Manager.Bind(
		dispatcher,
		fsmopt.On(tele.OnText),
		fsmopt.OnStates(StateAnswer),
		fsmopt.Do(f.ReportAdminAnswer),
	)

	f.Manager.Bind(
		dispatcher,
		fsmopt.On(tele.OnCallback),
		fsmopt.OnStates(StateAnswerConfirm),
		fsmopt.Do(f.ReportAdminAnswerConfirm),
	)

	logger.Info("Admin FSM настроен", "states", []string{StateAnswer, StateAnswerConfirm})
}

func (f *FSM) ReportAdminAnswer(c tele.Context, state fsm.Context) error {
	logger.Debug("ReportAdminAnswer вызван",
		"user_id", c.Sender().ID,
		"username", c.Sender().Username,
		"text", c.Message().Text,
	)

	text := strings.TrimSpace(c.Text())

	if text == "" {
		logger.Warn("Пустой ответ от админа", "user_id", c.Sender().ID)
		return c.Send("❌ Ответ не может быть пустым. Попробуйте ещё раз:")
	}

	var current int

	if err := state.Data(context.Background(), "current", &current); err != nil {
		logger.Error("Не удалось получить current из состояния",
			"error", err,
			"user_id", c.Sender().ID,
		)
		return c.Send(message.ErrorText + err.Error())
	}

	logger.Debug("Получен номер заявки", "report_index", current, "user_id", c.Sender().ID)

	if err := state.SetState(context.Background(), StateAnswerConfirm); err != nil {
		logger.Error("Не удалось установить состояние подтверждения",
			"error", err,
			"user_id", c.Sender().ID,
		)
		return c.Send(message.ErrorText + err.Error())
	}

	answer := AdminAnswer{
		Current: current,
		Answer:  text,
	}

	if err := state.Update(context.Background(), "answer", answer); err != nil {
		logger.Error("Не удалось сохранить ответ в состояние",
			"error", err,
			"user_id", c.Sender().ID,
			"report_index", current,
		)
		return c.Send(message.ErrorText + err.Error())
	}

	logger.Info("Админ ввёл ответ на заявку",
		"admin_id", c.Sender().ID,
		"admin_username", c.Sender().Username,
		"report_index", current,
		"answer_length", len(text),
	)

	res := fmt.Sprintf(message.CheckCorrectAnswer, text)

	return c.Send(res, keyboards.GetConfirmButtonAnswer())
}

func (f *FSM) ReportAdminAnswerConfirm(c tele.Context, state fsm.Context) error {
	logger.Debug("ReportAdminAnswerConfirm вызван",
		"user_id", c.Sender().ID,
		"callback_data", c.Callback().Data,
	)

	data := strings.TrimSpace(c.Callback().Data)

	var answer AdminAnswer

	if err := state.Data(context.Background(), "answer", &answer); err != nil {
		logger.Error("Не удалось получить ответ из состояния",
			"error", err,
			"user_id", c.Sender().ID,
		)
		return c.Send(message.ErrorText + err.Error())
	}

	logger.Debug("Получены данные из состояния",
		"report_index", answer.Current,
		"answer_length", len(answer.Answer),
	)

	switch data {
	case "admin_answer_confirm":
		logger.Info("Админ подтвердил отправку ответа",
			"admin_id", c.Sender().ID,
			"report_index", answer.Current,
		)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		reportRepo := repositories.NewReportRepository(f.mongoDb)

		reports, err := reportRepo.GetAllByStatus(ctx, models.ReportStatusInProgress)
		if err != nil {
			logger.Error("Ошибка при получении заявок со статусом InProgress",
				"error", err,
				"admin_id", c.Sender().ID,
			)
			return c.Send(message.ErrorText + err.Error())
		}

		if answer.Current-1 >= len(reports) {
			logger.Error("Индекс заявки вне диапазона",
				"report_index", answer.Current,
				"reports_count", len(reports),
				"admin_id", c.Sender().ID,
			)
			return c.Send("❌ Заявка не найдена. Возможно, она уже была обработана.")
		}

		report := reports[answer.Current-1]

		logger.Debug("Найдена заявка",
			"report_number", report.Number,
			"report_id", report.ID,
			"user_id", report.UserID,
		)

		report.Answer = answer.Answer
		report.Status = models.ReportStatusResolved

		err = reportRepo.Update(ctx, report)
		if err != nil {
			logger.Error("Ошибка при обновлении заявки",
				"error", err,
				"report_number", report.Number,
				"admin_id", c.Sender().ID,
			)
			return c.Send(message.ErrorText + err.Error())
		}

		logger.Info("Заявка обновлена",
			"report_number", report.Number,
			"new_status", report.Status,
			"admin_id", c.Sender().ID,
		)

		// Отправляем ответ пользователю
		user := &tele.User{ID: report.UserID}
		if _, err = c.Bot().Send(user, fmt.Sprintf(message.AnswerOnReport, report.Number)); err != nil {
			logger.Error("Не удалось отправить ответ пользователю",
				"error", err,
				"user_id", report.UserID,
				"report_number", report.Number,
			)
			return c.Send(message.ErrorText + "Не удалось уведомить пользователя: " + err.Error())
		}

		logger.Info("Ответ отправлен пользователю",
			"user_id", report.UserID,
			"report_number", report.Number,
		)

		if err = state.Finish(context.Background(), true); err != nil {
			logger.Error("Не удалось завершить FSM состояние",
				"error", err,
				"admin_id", c.Sender().ID,
			)
			return c.Send(message.ErrorText + err.Error())
		}

		logger.Info("FSM успешно завершён",
			"admin_id", c.Sender().ID,
			"report_number", report.Number,
		)

		return c.EditOrSend(fmt.Sprintf(message.SendAnswerOnReport, report.Number))

	case "admin_answer_restart":
		logger.Info("Админ выбрал редактирование ответа",
			"admin_id", c.Sender().ID,
			"report_index", answer.Current,
		)

		if err := state.SetState(context.Background(), StateAnswer); err != nil {
			logger.Error("Не удалось вернуться в состояние ввода ответа",
				"error", err,
				"admin_id", c.Sender().ID,
			)
			return c.Send(message.ErrorText + err.Error())
		}

		return c.EditOrSend(message.EnterAnswer)
	}

	logger.Warn("Неизвестная команда в callback",
		"command", data,
		"user_id", c.Sender().ID,
	)

	return c.Send(message.UnknownCommand)
}
