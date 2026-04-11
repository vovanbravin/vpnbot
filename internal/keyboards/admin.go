package keyboards

import (
	"fmt"
	"tgbot/internal/database/models"

	tele "gopkg.in/telebot.v4"
)

func GetReportAdminMenu(newCount, inProgressCount, resolvedCount int) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	btnNew := markup.Data(fmt.Sprintf("⚪ Новые заявки (%d)", newCount), "admin_report_new")
	btnInProgress := markup.Data(fmt.Sprintf("🟡 В обработке (%d)", inProgressCount), "admin_report_in_progress")
	btnResolved := markup.Data(fmt.Sprintf("🟢 Завершены (%d)", resolvedCount), "admin_report_resolved")

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
		btnPrev := markup.Data("Назад", fmt.Sprintf("admin_report_%s_%d", status, current-1))
		buttons = append(buttons, btnPrev)
	}

	if current < total {
		btnNext := markup.Data("Вперед", fmt.Sprintf("admin_report_%s_%d", status, current+1))
		buttons = append(buttons, btnNext)
	}

	btnMenu := markup.Data("Назад в меню", "admin_report_menu")

	markup.Inline(
		markup.Row(buttons...),
		markup.Row(btnMenu),
	)

	return markup
}
