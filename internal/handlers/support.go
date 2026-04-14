package handlers

import (
	"context"
	"strconv"
	"strings"
	"tgbot/internal/database/models"
	"tgbot/internal/database/repositories"
	"tgbot/internal/keyboards"
	"tgbot/internal/message"
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
		return c.Send(message.ReportCategoryPrompt)
	}

	return c.Send(message.ReportCategoryPrompt, keyboards.GetReportCategoriesKeyboard())
}

func (h *Handler) ProcessReportCategory(c tele.Context, state fsm.Context) error {
	text := strings.TrimSpace(c.Callback().Data)
	category := strings.TrimPrefix(text, "report_category_")

	data := ReportData{
		Category: category,
	}

	if err := state.Update(context.Background(), "data", data); err != nil {
		return c.Send(message.ReportSubjectPrompt)
	}

	if err := state.SetState(context.Background(), StateReportSubject); err != nil {
		return c.Send(message.ReportSubjectPrompt)
	}

	return c.EditOrSend(message.ReportSubjectPrompt)
}

func (h *Handler) ProcessReportSubject(c tele.Context, state fsm.Context) error {

	subject := strings.TrimSpace(c.Text())

	data := ReportData{}

	if err := state.Data(context.Background(), "data", &data); err != nil {
		return c.Send(message.ReportMessagePrompt)
	}

	data.Subject = subject

	if err := state.Update(context.Background(), "data", data); err != nil {
		return c.Send(message.ReportMessagePrompt)
	}

	if err := state.SetState(context.Background(), StateReportMessage); err != nil {
		return c.Send(message.ReportMessagePrompt)
	}

	return c.EditOrSend(message.ReportMessagePrompt)
}

func (h *Handler) ProcessReportMessage(c tele.Context, state fsm.Context) error {

	messag := strings.TrimSpace(c.Text())

	data := ReportData{}

	if err := state.Data(context.Background(), "data", &data); err != nil {
		return c.Send(message.ReportConfirmText(data.Category, data.Subject, data.Message))
	}

	data.Message = messag

	if err := state.Update(context.Background(), "data", data); err != nil {
		return c.Send(message.ReportConfirmText(data.Category, data.Subject, data.Message))
	}

	if err := state.SetState(context.Background(), StateReportConfirm); err != nil {
		return c.Send(message.ReportConfirmText(data.Category, data.Subject, data.Message))
	}

	return c.EditOrSend(message.ReportConfirmText(data.Category, data.Subject, data.Message), keyboards.GetReportConfirmKeyboard())
}

func (h *Handler) ProcessReportApprov(c tele.Context, state fsm.Context) error {

	data := ReportData{}

	if err := state.Data(context.Background(), "data", &data); err != nil {
		return c.Send(message.ReportCreated)
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
		return c.Send(message.ReportCreated)
	}

	if err := reportRepo.Insert(ctx, &report); err != nil {
		return c.Send(message.ReportCreated)
	}

	if err := state.Finish(context.Background(), true); err != nil {
		return c.Send(message.ReportCreated)
	}

	return c.EditOrSend(message.ReportCreated)
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
		return c.Send(message.UnknownOption)
	}
}

func (h *Handler) Support(c tele.Context) error {
	return c.Send(message.SupportMenuPrompt, keyboards.SupportMenu)
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

func (h *Handler) StartInfoReport(c tele.Context, state fsm.Context) error {

	if err := state.SetState(context.Background(), StateReportNumber); err != nil {
		return c.Send(message.ReportNumberPrompt)
	}

	return c.Send(message.ReportNumberPrompt)
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
			return c.Send(message.ReportNotFound)
		}
		return c.Send(message.ReportNotFound)
	}

	if err = state.Finish(context.Background(), true); err != nil {
		return c.Send(message.ReportNotFound)
	}
	return c.Send(report.DetailInfo())
}
