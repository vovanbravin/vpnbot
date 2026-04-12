package keyboards

import (
	"tgbot/internal/message"

	tele "gopkg.in/telebot.v4"
)

var (
	MainMenu = &tele.ReplyMarkup{ResizeKeyboard: true}

	BtnBuyVpn = MainMenu.Text(message.ButtonBuyVpn)
	BtnMyVpn  = MainMenu.Text(message.ButtonMyTariff)
)

func init() {
	MainMenu.Reply(
		MainMenu.Row(BtnBuyVpn, BtnMyVpn),
	)
}
