package handlers

import (
	"context"
	"strconv"
	"strings"
	"tgbot/internal/database/models"
	"tgbot/internal/database/repositories"
	"tgbot/internal/keyboards"
	"tgbot/internal/utils"
	"time"

	"github.com/vitaliy-ukiru/fsm-telebot/v2"
	"github.com/vitaliy-ukiru/fsm-telebot/v2/fsmopt"
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

type ReportData struct {
	Category string
	Subject  string
	Message  string
}

func (h *Handler) HandleSupport() {

	h.dispatcher.Handle("/info_report", tf.RawHandler{Callback: h.fsm.Adapt(h.StartInfoReport)})

	h.fsm.Bind(
		h.dispatcher,
		fsmopt.On(tele.OnText),
		fsmopt.OnStates(StateReportNumber),
		fsmopt.Do(h.GetInfoReport))

	h.dispatcher.Handle("/my_report", tf.RawHandler{Callback: h.MyReports})

	h.dispatcher.Handle("/new_report", tf.RawHandler{Callback: h.fsm.Adapt(h.StartReport)})

	h.fsm.Bind(
		h.dispatcher,
		fsmopt.On(tele.OnCallback),
		fsmopt.OnStates(StateReportCategory),
		fsmopt.Do(h.ProcessReportCategory),
	)

	h.fsm.Bind(
		h.dispatcher,
		fsmopt.On(tele.OnText),
		fsmopt.OnStates(StateReportSubject),
		fsmopt.Do(h.ProcessReportSubject),
	)

	h.fsm.Bind(
		h.dispatcher,
		fsmopt.On(tele.OnText),
		fsmopt.OnStates(StateReportMessage),
		fsmopt.Do(h.ProcessReportMessage),
	)

	h.fsm.Bind(
		h.dispatcher,
		fsmopt.On(tele.OnCallback),
		fsmopt.OnStates(StateReportConfirm),
		fsmopt.Do(h.ProcessReportConfirm))

}

func (h *Handler) StartReport(c tele.Context, state fsm.Context) error {
	if err := state.SetState(context.Background(), StateReportCategory); err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	return c.Send("Выберете категорию проблемы:", keyboards.GetReportCategoriesKeyboard())
}

func (h *Handler) ProcessReportCategory(c tele.Context, state fsm.Context) error {
	text := strings.TrimSpace(c.Callback().Data)
	category := strings.TrimPrefix(text, "report_category_")

	data := ReportData{
		Category: category,
	}

	if err := state.Update(context.Background(), "data", data); err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	if err := state.SetState(context.Background(), StateReportSubject); err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	return c.EditOrSend("Напишите тему проблемы")
}

func (h *Handler) ProcessReportSubject(c tele.Context, state fsm.Context) error {

	subject := strings.TrimSpace(c.Text())

	data := ReportData{}

	if err := state.Data(context.Background(), "data", &data); err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	data.Subject = subject

	if err := state.Update(context.Background(), "data", data); err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	if err := state.SetState(context.Background(), StateReportMessage); err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	return c.EditOrSend("Опишите проблему")
}

func (h *Handler) ProcessReportMessage(c tele.Context, state fsm.Context) error {

	message := strings.TrimSpace(c.Text())

	data := ReportData{}

	if err := state.Data(context.Background(), "data", &data); err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	data.Message = message

	if err := state.Update(context.Background(), "data", data); err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	if err := state.SetState(context.Background(), StateReportConfirm); err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	var result string

	category := utils.GetCategoryByName(data.Category)

	result += "📂 Категория: " + category + "\n"
	result += "📝 Тема: " + data.Subject + "\n"
	result += "📄 Описание: " + data.Message + "\n"

	return c.EditOrSend(result, keyboards.GetReportConfirmKeyboard())
}

func (h *Handler) ProcessReportApprov(c tele.Context, state fsm.Context) error {

	data := ReportData{}

	if err := state.Data(context.Background(), "data", &data); err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	reportRepo := repositories.NewReportRepository(h.mongoDb)

	counterRepo := repositories.NewCounterRepository(h.mongoDb)

	category := models.ReportCategory(data.Category)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	number, err := counterRepo.GetNextNumber(ctx, "reports")

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

	if err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	if err := reportRepo.Insert(ctx, &report); err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	if err := state.Finish(context.Background(), true); err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	return c.EditOrSend("Заявление создано")
}

func (h *Handler) ProcessReportConfirm(c tele.Context, state fsm.Context) error {

	text := strings.TrimSpace(c.Callback().Data)
	answer := strings.TrimPrefix(text, "report_")

	switch answer {
	case "approv":
		return h.ProcessReportApprov(c, state)
	case "restart":
		return h.StartReport(c, state)
	default:
		return c.Send("Неизвесная опция")
	}
}

func (h *Handler) Support(c tele.Context) error {
	return c.Send("Выберете опцию: ", keyboards.SupportMenu)
}

func (h *Handler) MyReports(c tele.Context) error {
	user_id := c.Sender().ID

	reportRes := repositories.NewReportRepository(h.mongoDb)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reports, err := reportRes.GetAllReportByUserId(ctx, user_id)

	if err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	result := ""

	for i, report := range reports {
		if i == 0 {
			c.Send(report.ShortInfo())
		} else {
			c.Bot().Send(c.Chat(), report.ShortInfo())
		}
	}

	return c.Send(result)
}

func (h *Handler) StartInfoReport(c tele.Context, state fsm.Context) error {

	if err := state.SetState(context.Background(), StateReportNumber); err != nil {
		return c.Send("Ошибка " + err.Error())
	}

	return c.Send("Введите номер заявления")
}

func (h *Handler) GetInfoReport(c tele.Context, state fsm.Context) error {
	strNumber := c.Text()

	number, err := strconv.Atoi(strNumber)

	if err != nil {
		return h.StartInfoReport(c, state)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reportRep := repositories.NewReportRepository(h.mongoDb)

	report, err := reportRep.GetReportByNumber(ctx, int64(number))

	if report == nil || report.UserID != c.Sender().ID || err != nil {
		if err = state.Finish(context.Background(), true); err != nil {
			return c.Send("Ошибка " + err.Error())
		}
		return c.Send("Заявка не найдена, убедитесь, что вы правильно ввели номер")
	}

	if err = state.Finish(context.Background(), true); err != nil {
		return c.Send("Ошибка " + err.Error())
	}
	return c.Send(report.DetailInfo())
}
