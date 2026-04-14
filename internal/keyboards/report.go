package keyboards

import (
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
