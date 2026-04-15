package fsm

import (
	"context"
	"fmt"
	"strings"
	"tgbot/internal/database/models"
	"tgbot/internal/database/repositories"
	"tgbot/internal/keyboards"
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
}

func (f *FSM) ReportAdminAnswer(c tele.Context, state fsm.Context) error {

	text := strings.TrimSpace(c.Text())

	var current int

	if err := state.Data(context.Background(), "current", &current); err != nil {
		return c.Send(message.ErrorText + err.Error())
	}

	if err := state.SetState(context.Background(), StateAnswerConfirm); err != nil {
		return c.Send(message.ErrorText + err.Error())
	}

	answer := AdminAnswer{
		Current: current,
		Answer:  text,
	}

	if err := state.Update(context.Background(), "answer", answer); err != nil {
		return c.Send(message.ErrorText + err.Error())
	}

	res := fmt.Sprintf(message.CheckCorrectAnswer, text)

	return c.Send(res, keyboards.GetConfirmButtonAnswer())
}

func (f *FSM) ReportAdminAnswerConfirm(c tele.Context, state fsm.Context) error {

	data := strings.TrimSpace(c.Callback().Data)

	var answer AdminAnswer

	if err := state.Data(context.Background(), "answer", &answer); err != nil {
		return c.Send(message.ErrorText + err.Error())
	}

	switch data {
	case "admin_answer_confirm":

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		reportRepo := repositories.NewReportRepository(f.mongoDb)

		reports, err := reportRepo.GetAllByStatus(ctx, models.ReportStatusInProgress)

		if err != nil {
			return c.Send(message.ErrorText + err.Error())
		}

		report := reports[answer.Current-1]

		report.Answer = answer.Answer
		report.Status = models.ReportStatusResolved

		err = reportRepo.Update(ctx, report)

		if err != nil {
			return c.Send(message.ErrorText + err.Error())
		}

		user := &tele.User{ID: report.UserID}
		if _, err = c.Bot().Send(user, fmt.Sprintf(message.AnswerOnReport, report.Number)); err != nil {
			return c.Send(message.ErrorText + err.Error())
		}

		if err = state.Finish(context.Background(), true); err != nil {
			return c.Send(message.ErrorText + err.Error())
		}

		return c.EditOrSend(fmt.Sprintf(message.SendAnswerOnReport, report.Number))

	case "admin_answer_restart":

		if err := state.SetState(context.Background(), StateAnswer); err != nil {
			return c.Send(message.ErrorText + err.Error())
		}

		return c.EditOrSend(message.EnterAnswer)
	}

	return c.Send(message.UnknownCommand)
}
