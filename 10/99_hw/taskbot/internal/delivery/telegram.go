package delivery

import (
	"context"
	"log"
	"net/http"
	"taskbot/internal/models"
	"taskbot/internal/router"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type DeliveryTelegram struct {
	HttpListenAddr string
	WebhookURL     string
	BotToken       string
	Router         *router.Router
}

func NewDeliveryTelegram(httpListenAddr, botToken, webHook string) *DeliveryTelegram {
	return &DeliveryTelegram{
		HttpListenAddr: httpListenAddr,
		WebhookURL:     webHook,  //"https://d66b-46-0-92-51.ngrok-free.app",
		BotToken:       botToken, //"7228614571:AAFoUE_gM81EPgiXSSvlBoF4iAUIRFH_gaY",
		Router:         router.NewRouter(),
	}
}

func (dt *DeliveryTelegram) StartDeliveryTelegram(ctx context.Context) error {

	bot, err := tgbotapi.NewBotAPI(dt.BotToken)
	if err != nil {
		log.Fatalf("Error create bot API: %s\n", err)
	}

	log.Printf("Authorized on account %s\n", bot.Self.UserName)

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(dt.WebhookURL))
	if err != nil {
		log.Fatalf("Error set webhook: %s\n", err)
	}

	updates := bot.ListenForWebhook("/")

	go http.ListenAndServe(dt.HttpListenAddr, nil)
	log.Printf("start listen %s\n", dt.HttpListenAddr)

	// получаем все обновления из канала updates
	for update := range updates {

		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s\n", update.Message.From.UserName, update.Message.Text)

		// if update.Message.IsCommand() {
		user := &models.User{
			ID:       update.Message.From.ID,
			Username: update.Message.From.UserName,
			ChatID:   update.Message.Chat.ID,
		}
		cmd := update.Message.Text
		resp, err := dt.Router.Route(cmd, user)
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
			bot.Send(msg)
		} else {
			for chatID, message := range resp {
				msg := tgbotapi.NewMessage(chatID, message)
				bot.Send(msg)
			}
		}
		// } else {
		// 	log.Printf("User \"%#v\" enter not command: %s\n", update.Message.From, update.Message.Text)
		// }

	}
	return nil
}
