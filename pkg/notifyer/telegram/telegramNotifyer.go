package telegram

import (
	"context"

	"github.com/aDeepRecession/moodle-scrapper/pkg/notifyer/telegram/telegram"
)

type TelegramService struct {
	tg *telegram.Telegram
}

func NewTelegramService(botid string, chatid int) TelegramService {
	tel, _ := telegram.New(botid)
	tel.AddReceivers(int64(chatid))

	return TelegramService{tel}
}

func (tn TelegramService) Send(msg string) error {
	ctx := context.Background()
	err := tn.tg.Send(ctx, msg)

	return err
}
