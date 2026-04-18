package keyboards

import (
	"fmt"
	"tgbot/internal/message"

	tele "gopkg.in/telebot.v4"
)

var (
	SupportMenu = &tele.ReplyMarkup{}

	BtnNewReport = SupportMenu.Data(message.ButtonNewReport, "report_new")
	BtnMyReports = SupportMenu.Data(message.ButtonMyReport, "report_my")
)

func init() {
	SupportMenu.Inline(
		SupportMenu.Row(BtnNewReport, BtnMyReports),
	)
}

func GetReportCategoriesKeyboard() *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	btnTechnical := markup.Data(message.ButtonTechnical, "report_category_technical")
	btnPayment := markup.Data(message.ButtonPayment, "report_category_payment")
	btnConnection := markup.Data(message.ButtonConnection, "report_category_connection")
	btnFeature := markup.Data(message.ButtonFeature, "report_category_feature")
	btnOther := markup.Data(message.ButtonOther, "report_category_other")

	markup.Inline(
		markup.Row(btnTechnical),
		markup.Row(btnPayment),
		markup.Row(btnConnection),
		markup.Row(btnFeature),
		markup.Row(btnOther),
	)

	return markup
}

func GetReportConfirmKeyboard() *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	btnApprov := markup.Data(message.ButtonApprov, "report_approv")
	btnRestart := markup.Data(message.ButtonRestart, "report_restart")

	markup.Inline(
		markup.Row(btnApprov),
		markup.Row(btnRestart),
	)

	return markup
}

func GetUserReportMenu() *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	btnActive := markup.Data(message.ButtonActiveReports, "support_active_reports")
	btnClosed := markup.Data(message.ButtonClosedReports, "support_closed_reports")

	markup.Inline(
		markup.Row(btnActive),
		markup.Row(btnClosed),
	)

	return markup
}

func GetUserNavigationButtonsReport(active bool, current, total int) *tele.ReplyMarkup {

	markup := &tele.ReplyMarkup{}

	var rows []tele.Row

	var status string

	if active {
		status = "active"
	} else {
		status = "closed"
	}

	var navRow []tele.Btn
	if current > 1 {
		btnPrev := markup.Data(message.ButtonPrev, fmt.Sprintf("support_%s_report_%d", status, current-1))
		navRow = append(navRow, btnPrev)
	}
	if current < total {
		btnNext := markup.Data(message.ButtonNext, fmt.Sprintf("support_%s_report_%d", status, current+1))
		navRow = append(navRow, btnNext)
	}

	if len(navRow) > 0 {
		rows = append(rows, navRow)
	}

	rows = append(rows, []tele.Btn{
		markup.Data(message.ButtonMenu, "support_menu"),
	})

	markup.Inline(rows...)

	return markup
}
