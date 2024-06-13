package main

import (
	"context"
	"log"
	"taskbot/internal/delivery"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	BotToken   string = "XXX"
	WebhookURL string = "XXX"
)

func startTaskBot(ctx context.Context, httpListenAddr string) error {
	// сюда писать код
	/*
		в этом месте вы стартуете бота,
		стартуете хттп сервер который будет обслуживать этого бота
		инициализируете ваше приложение
		и потом будете обрабатывать входящие сообщения
	*/
	tgBot := delivery.NewDeliveryTelegram(httpListenAddr, BotToken, WebhookURL)
	log.Printf("Created tgBot: %#v\n", tgBot)
	if err := tgBot.StartDeliveryTelegram(ctx); err != nil {
		log.Printf("Error started tgBot: %s\n", err)
		return err
	}

	return nil
}

func main() {
	err := startTaskBot(context.Background(), ":8081")
	if err != nil {
		log.Fatalln(err)
	}
}

// это заглушка чтобы импорт сохранился
func __dummy() {
	tgbotapi.APIEndpoint = "_dummy"
}
