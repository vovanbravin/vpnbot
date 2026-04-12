package keyboards

import (
	"fmt"
	"tgbot/internal/database/models"
	"tgbot/internal/message"

	tele "gopkg.in/telebot.v4"
)

func GetReportAdminMenu(newCount, inProgressCount, resolvedCount int) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	btnNew := markup.Data(fmt.Sprintf(message.ButtonNewReports, newCount), "admin_report_new")
	btnInProgress := markup.Data(fmt.Sprintf(message.ButtonInProgressReports, inProgressCount), "admin_report_in_progress")
	btnResolved := markup.Data(fmt.Sprintf(message.ButtonResolvedReports, resolvedCount), "admin_report_resolved")

	markup.Inline(
		markup.Row(btnNew),
		markup.Row(btnInProgress),
		markup.Row(btnResolved),
	)

	return markup
}

func GetNavigationButtonsReport(status models.ReportStatus, current, total int) *tele.ReplyMarkup {

	markup := &tele.ReplyMarkup{}

	var rows []tele.Row

	var navRow []tele.Btn
	if current > 1 {
		btnPrev := markup.Data(message.ButtonPrev, fmt.Sprintf("admin_report_%s_%d", status, current-1))
		navRow = append(navRow, btnPrev)
	}
	if current < total {
		btnNext := markup.Data(message.ButtonNext, fmt.Sprintf("admin_report_%s_%d", status, current+1))
		navRow = append(navRow, btnNext)
	}

	if len(navRow) > 0 {
		rows = append(rows, navRow)
	}

	switch status {
	case models.ReportStatusNew:
		rows = append(rows, []tele.Btn{
			markup.Data("Взять в работу", fmt.Sprintf("admin_report_hire_%d", current)),
		})
	case models.ReportStatusInProgress:
		rows = append(rows, []tele.Btn{
			markup.Data("Ответить", fmt.Sprintf("admir_report_answer_%d", current)),
		})
	}

	rows = append(rows, []tele.Btn{
		markup.Data(message.ButtonMenu, "admin_report_menu"),
	})

	markup.Inline(rows...)

	return markup
}
