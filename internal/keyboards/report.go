package keyboards

import (
	tele "gopkg.in/telebot.v4"
)

var (
	SupportMenu = &tele.ReplyMarkup{}

	BtnNewReport = SupportMenu.Data("Новое заявление", "report_new")
	BtnMyReports = SupportMenu.Data("Мои заявления", "report_my")
)

func init() {
	SupportMenu.Inline(
		SupportMenu.Row(BtnNewReport, BtnMyReports),
	)
}

func GetReportCategoriesKeyboard() *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}

	btnTechnical := markup.Data("Техническая проблема", "report_category_technical")
	btnPayment := markup.Data("Проблемы с оплатой", "report_category_payment")
	btnConnection := markup.Data("Проблемы с подключением", "report_category_connection")
	btnFeature := markup.Data("Предложения", "report_category_feature")
	btnOther := markup.Data("Другое", "report_category_other")

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

	btnApprov := markup.Data("Все верно", "report_approv")
	btnRestart := markup.Data("Начать сначала", "report_restart")

	markup.Inline(
		markup.Row(btnApprov),
		markup.Row(btnRestart),
	)

	return markup
}
