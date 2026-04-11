package keyboards

import tele "gopkg.in/telebot.v4"

var (
	MainMenu = &tele.ReplyMarkup{ResizeKeyboard: true}

	BtnBuyVpn = MainMenu.Text("Купить vpn")
	BtnMyVpn  = MainMenu.Text("Мой vpn")
)

func init() {
	MainMenu.Reply(
		MainMenu.Row(BtnBuyVpn, BtnMyVpn),
	)
}
