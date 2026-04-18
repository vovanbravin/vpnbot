package fsm

import (
	"context"
	"strconv"
	"strings"
	"tgbot/internal/database/models"
	"tgbot/internal/database/repositories"
	"tgbot/internal/keyboards"
	"tgbot/internal/logger"
	"tgbot/internal/message"
	"tgbot/internal/utils"
	"time"

	"github.com/vitaliy-ukiru/fsm-telebot/v2"
	"github.com/vitaliy-ukiru/fsm-telebot/v2/fsmopt"
	"github.com/vitaliy-ukiru/telebot-filter/v2/dispatcher"
	tf "github.com/vitaliy-ukiru/telebot-filter/v2/telefilter"
	"go.mongodb.org/mongo-driver/v2/bson"
	tele "gopkg.in/telebot.v4"
)

const (
	StateReportCategory = "report_category"
	StateReportSubject  = "report_subject"
	StateReportMessage  = "report_message"
	StateReportConfirm  = "report_confirm"
	StateReportNumber   = "report_number"
)

func (f *FSM) SetupSupport(dispatcher *dispatcher.Dispatcher) {
	logger.Debug("Настройка Support FSM", "module", "support")

	dispatcher.Handle("/info_report", tf.RawHandler{Callback: f.Manager.Adapt(f.StartInfoReport)})

	f.Manager.Bind(
		dispatcher,
		fsmopt.On(tele.OnText),
		fsmopt.OnStates(StateReportNumber),
		fsmopt.Do(f.GetInfoReport))

	dispatcher.Handle("/new_report", tf.RawHandler{Callback: f.Manager.Adapt(f.StartReport)})

	f.Manager.Bind(
		dispatcher,
		fsmopt.On(tele.OnCallback),
		fsmopt.OnStates(StateReportCategory),
		fsmopt.Do(f.ProcessReportCategory),
	)

	f.Manager.Bind(
		dispatcher,
		fsmopt.On(tele.OnText),
		fsmopt.OnStates(StateReportSubject),
		fsmopt.Do(f.ProcessReportSubject),
	)

	f.Manager.Bind(
		dispatcher,
		fsmopt.On(tele.OnText),
		fsmopt.OnStates(StateReportMessage),
		fsmopt.Do(f.ProcessReportMessage),
	)

	f.Manager.Bind(
		dispatcher,
		fsmopt.On(tele.OnCallback),
		fsmopt.OnStates(StateReportConfirm),
		fsmopt.Do(f.ProcessReportConfirm))

	logger.Info("Support FSM настроен", "states", []string{
		StateReportCategory, StateReportSubject, StateReportMessage, StateReportConfirm, StateReportNumber,
	})
}

func (f *FSM) StartReport(c tele.Context, state fsm.Context) error {
	logger.Debug("StartReport вызван",
		"user_id", c.Sender().ID,
		"username", c.Sender().Username,
	)

	if err := state.SetState(context.Background(), StateReportCategory); err != nil {
		logger.Error("Ошибка установки состояния StateReportCategory",
			"error", err,
			"user_id", c.Sender().ID,
		)
		return c.Send(message.SomethingWrong)
	}

	logger.Info("Пользователь начал создание заявки",
		"user_id", c.Sender().ID,
		"username", c.Sender().Username,
	)

	return c.Send(message.ReportCategoryPrompt, keyboards.GetReportCategoriesKeyboard())
}

func (f *FSM) ProcessReportCategory(c tele.Context, state fsm.Context) error {
	logger.Debug("ProcessReportCategory вызван",
		"user_id", c.Sender().ID,
		"callback_data", c.Callback().Data,
	)

	text := strings.TrimSpace(c.Callback().Data)
	category := strings.TrimPrefix(text, "report_category_")

	logger.Debug("Выбрана категория",
		"user_id", c.Sender().ID,
		"category", category,
	)

	data := ReportData{
		Category: category,
	}

	if err := state.Update(context.Background(), "data", data); err != nil {
		logger.Error("Ошибка сохранения категории в состояние",
			"error", err,
			"user_id", c.Sender().ID,
			"category", category,
		)
		return c.Send(message.SomethingWrong)
	}

	if err := state.SetState(context.Background(), StateReportSubject); err != nil {
		logger.Error("Ошибка установки состояния StateReportSubject",
			"error", err,
			"user_id", c.Sender().ID,
		)
		return c.Send(message.SomethingWrong)
	}

	return c.EditOrSend(message.ReportSubjectPrompt)
}

func (f *FSM) ProcessReportSubject(c tele.Context, state fsm.Context) error {
	logger.Debug("ProcessReportSubject вызван",
		"user_id", c.Sender().ID,
		"subject", c.Text(),
	)

	subject := strings.TrimSpace(c.Text())

	if subject == "" {
		logger.Warn("Пустая тема заявки",
			"user_id", c.Sender().ID,
		)
	}

	data := ReportData{}

	if err := state.Data(context.Background(), "data", &data); err != nil {
		logger.Error("Ошибка получения данных из состояния",
			"error", err,
			"user_id", c.Sender().ID,
		)
		return c.Send(message.SomethingWrong)
	}

	data.Subject = subject

	if err := state.Update(context.Background(), "data", data); err != nil {
		logger.Error("Ошибка сохранения темы в состояние",
			"error", err,
			"user_id", c.Sender().ID,
			"subject", subject,
		)
		return c.Send(message.SomethingWrong)
	}

	if err := state.SetState(context.Background(), StateReportMessage); err != nil {
		logger.Error("Ошибка установки состояния StateReportMessage",
			"error", err,
			"user_id", c.Sender().ID,
		)
		return c.Send(message.SomethingWrong)
	}

	return c.EditOrSend(message.ReportMessagePrompt)
}

func (f *FSM) ProcessReportMessage(c tele.Context, state fsm.Context) error {
	logger.Debug("ProcessReportMessage вызван",
		"user_id", c.Sender().ID,
		"message_length", len(c.Text()),
	)

	messag := strings.TrimSpace(c.Text())

	if messag == "" {
		logger.Warn("Пустое сообщение в заявке",
			"user_id", c.Sender().ID,
		)
	}

	data := ReportData{}

	if err := state.Data(context.Background(), "data", &data); err != nil {
		logger.Error("Ошибка получения данных из состояния",
			"error", err,
			"user_id", c.Sender().ID,
		)
		return c.Send(message.SomethingWrong)
	}

	data.Message = messag

	if err := state.Update(context.Background(), "data", data); err != nil {
		logger.Error("Ошибка сохранения сообщения в состояние",
			"error", err,
			"user_id", c.Sender().ID,
		)
		return c.Send(message.SomethingWrong)
	}

	if err := state.SetState(context.Background(), StateReportConfirm); err != nil {
		logger.Error("Ошибка установки состояния StateReportConfirm",
			"error", err,
			"user_id", c.Sender().ID,
		)
		return c.Send(message.SomethingWrong)
	}

	logger.Info("Пользователь заполнил заявку, ожидает подтверждения",
		"user_id", c.Sender().ID,
		"category", data.Category,
		"subject_length", len(data.Subject),
		"message_length", len(data.Message),
	)

	return c.EditOrSend(message.ReportConfirmText(data.Category, data.Subject, data.Message), keyboards.GetReportConfirmKeyboard())
}

func (f *FSM) ProcessReportApprov(c tele.Context, state fsm.Context) error {
	logger.Debug("ProcessReportApprov вызван",
		"user_id", c.Sender().ID,
	)

	data := ReportData{}

	if err := state.Data(context.Background(), "data", &data); err != nil {
		logger.Error("Ошибка получения данных из состояния",
			"error", err,
			"user_id", c.Sender().ID,
		)
		return c.Send(message.SomethingWrong)
	}

	logger.Info("Создание заявки",
		"user_id", c.Sender().ID,
		"username", c.Sender().Username,
		"category", data.Category,
		"subject", data.Subject,
	)

	reportRepo := repositories.NewReportRepository(f.mongoDb)
	counterRepo := repositories.NewCounterRepository(f.mongoDb)

	category := models.ReportCategory(data.Category)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	number, err := counterRepo.GetNextNumber(ctx, "reports")
	if err != nil {
		logger.Error("Ошибка получения номера заявки",
			"error", err,
			"user_id", c.Sender().ID,
		)
		return c.Send(message.SomethingWrong)
	}

	report := models.Report{
		ID:        bson.NewObjectID(),
		Number:    number,
		UserID:    c.Sender().ID,
		Username:  c.Sender().Username,
		Categoty:  category,
		Priority:  utils.GetPriorityByCategory(category),
		Status:    models.ReportStatusNew,
		Subject:   data.Subject,
		Message:   data.Message,
		Answer:    "",
		CreatedAt: time.Now(),
	}

	if err = reportRepo.Insert(ctx, &report); err != nil {
		logger.Error("Ошибка сохранения заявки в БД",
			"error", err,
			"user_id", c.Sender().ID,
			"report_number", number,
		)
		return c.Send(message.SomethingWrong)
	}

	if err = state.Finish(context.Background(), true); err != nil {
		logger.Error("Ошибка завершения FSM",
			"error", err,
			"user_id", c.Sender().ID,
		)
		return c.Send(message.SomethingWrong)
	}

	logger.Info("Заявка успешно создана",
		"user_id", c.Sender().ID,
		"username", c.Sender().Username,
		"report_number", number,
		"category", data.Category,
	)

	return c.EditOrSend(message.ReportCreated)
}

func (f *FSM) ProcessReportConfirm(c tele.Context, state fsm.Context) error {
	logger.Debug("ProcessReportConfirm вызван",
		"user_id", c.Sender().ID,
		"callback_data", c.Callback().Data,
	)

	text := strings.TrimSpace(c.Callback().Data)
	answer := strings.TrimPrefix(text, "report_")

	switch answer {
	case "approv":
		logger.Debug("Пользователь подтвердил создание заявки",
			"user_id", c.Sender().ID,
		)
		return f.ProcessReportApprov(c, state)
	case "restart":
		logger.Debug("Пользователь решил перезаполнить заявку",
			"user_id", c.Sender().ID,
		)
		return f.StartReport(c, state)
	default:
		logger.Warn("Неизвестный ответ в подтверждении",
			"user_id", c.Sender().ID,
			"answer", answer,
		)
		return c.Send(message.UnknownOption)
	}
}

func (f *FSM) Support(c tele.Context) error {
	logger.Debug("Support меню вызвано",
		"user_id", c.Sender().ID,
		"username", c.Sender().Username,
	)
	return c.Send(message.SupportMenuPrompt, keyboards.SupportMenu)
}

func (f *FSM) StartInfoReport(c tele.Context, state fsm.Context) error {
	logger.Debug("StartInfoReport вызван",
		"user_id", c.Sender().ID,
	)

	if err := state.SetState(context.Background(), StateReportNumber); err != nil {
		logger.Error("Ошибка установки состояния StateReportNumber",
			"error", err,
			"user_id", c.Sender().ID,
		)
		return c.Send(message.SomethingWrong)
	}

	return c.Send(message.ReportNumberPrompt)
}

func (f *FSM) GetInfoReport(c tele.Context, state fsm.Context) error {
	logger.Debug("GetInfoReport вызван",
		"user_id", c.Sender().ID,
		"input", c.Text(),
	)

	strNumber := c.Text()

	number, err := strconv.Atoi(strNumber)
	if err != nil {
		logger.Warn("Некорректный ввод номера заявки",
			"user_id", c.Sender().ID,
			"input", strNumber,
			"error", err,
		)
		return f.StartInfoReport(c, state)
	}

	logger.Debug("Поиск заявки по номеру",
		"user_id", c.Sender().ID,
		"report_number", number,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reportRep := repositories.NewReportRepository(f.mongoDb)

	report, err := reportRep.GetReportByNumber(ctx, int64(number))

	if report == nil || report.UserID != c.Sender().ID || err != nil {
		logger.Warn("Заявка не найдена или доступ запрещён",
			"user_id", c.Sender().ID,
			"report_number", number,
			"report_found", report != nil,
			"error", err,
		)

		if err = state.Finish(context.Background(), true); err != nil {
			logger.Error("Ошибка завершения FSM",
				"error", err,
				"user_id", c.Sender().ID,
			)
			return c.Send(message.SomethingWrong)
		}
		return c.Send(message.ReportNotFound)
	}

	if err = state.Finish(context.Background(), true); err != nil {
		logger.Error("Ошибка завершения FSM",
			"error", err,
			"user_id", c.Sender().ID,
		)
		return c.Send(message.SomethingWrong)
	}

	logger.Info("Пользователь получил информацию о заявке",
		"user_id", c.Sender().ID,
		"report_number", number,
		"report_status", report.Status,
	)

	return c.Send(report.DetailInfo())
}
