package fsm

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
}

func (f *FSM) StartReport(c tele.Context, state fsm.Context) error {
	if err := state.SetState(context.Background(), StateReportCategory); err != nil {
		return c.Send(message.ReportCategoryPrompt)
	}

	return c.Send(message.ReportCategoryPrompt, keyboards.GetReportCategoriesKeyboard())
}

func (f *FSM) ProcessReportCategory(c tele.Context, state fsm.Context) error {
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

func (f *FSM) ProcessReportSubject(c tele.Context, state fsm.Context) error {

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

func (f *FSM) ProcessReportMessage(c tele.Context, state fsm.Context) error {

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

func (f *FSM) ProcessReportApprov(c tele.Context, state fsm.Context) error {

	data := ReportData{}

	if err := state.Data(context.Background(), "data", &data); err != nil {
		return c.Send(message.ReportCreated)
	}

	reportRepo := repositories.NewReportRepository(f.mongoDb)

	counterRepo := repositories.NewCounterRepository(f.mongoDb)

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

func (f *FSM) ProcessReportConfirm(c tele.Context, state fsm.Context) error {

	text := strings.TrimSpace(c.Callback().Data)
	answer := strings.TrimPrefix(text, "report_")

	switch answer {
	case "approv":
		return f.ProcessReportApprov(c, state)
	case "restart":
		return f.StartReport(c, state)
	default:
		return c.Send(message.UnknownOption)
	}
}

func (f *FSM) Support(c tele.Context) error {
	return c.Send(message.SupportMenuPrompt, keyboards.SupportMenu)
}

func (f *FSM) StartInfoReport(c tele.Context, state fsm.Context) error {

	if err := state.SetState(context.Background(), StateReportNumber); err != nil {
		return c.Send(message.ReportNumberPrompt)
	}

	return c.Send(message.ReportNumberPrompt)
}

func (f *FSM) GetInfoReport(c tele.Context, state fsm.Context) error {
	strNumber := c.Text()

	number, err := strconv.Atoi(strNumber)

	if err != nil {
		return f.StartInfoReport(c, state)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reportRep := repositories.NewReportRepository(f.mongoDb)

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
