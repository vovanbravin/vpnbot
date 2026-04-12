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

	var buttons []tele.Btn

	if current > 1 {
		btnPrev := markup.Data(message.ButtonPrev, fmt.Sprintf("admin_report_%s_%d", status, current-1))
		buttons = append(buttons, btnPrev)
	}

	if current < total {
		btnNext := markup.Data(message.ButtonNext, fmt.Sprintf("admin_report_%s_%d", status, current+1))
		buttons = append(buttons, btnNext)
	}

	btnMenu := markup.Data(message.ButtonMenu, "admin_report_menu")

	markup.Inline(
		markup.Row(buttons...),
		markup.Row(btnMenu),
	)

	return markup
}
