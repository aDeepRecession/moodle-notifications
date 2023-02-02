package telegramnotifyer

import (
	"context"

	"github.com/aDeepRecession/moodle-scrapper/pkg/notifyer/telegram/telegram"
)

type TelegramNotifyer struct {
	tg *telegram.Telegram
}

func NewTelegramNotifyer(botid string, chatid int) TelegramNotifyer {
	tel, _ := telegram.New(botid)
	tel.AddReceivers(int64(chatid))

	return TelegramNotifyer{tel}
}

func (tn TelegramNotifyer) Send(msg string) error {
	ctx := context.Background()
	err := tn.tg.Send(ctx, msg)
	return err
}
