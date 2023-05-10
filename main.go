package main

import (
	"context"
	"flag"
	"log"
	"tgBotStoresPasswords/clients/telegram"
	event_consumer "tgBotStoresPasswords/consumer/event-consumer"
	telegram2 "tgBotStoresPasswords/events/telegram"
	"tgBotStoresPasswords/storage/sqlite"
)

// но лучше сделать так жк с флагом,как и с токеном
const (
	tgBotHost         = "api.telegram.org"
	sqliteStoragePath = "data/sqlite/storage.db"
	batchSize         = 100
)

func main() {
	s, err := sqlite.New(sqliteStoragePath)
	if err != nil {
		log.Fatalf("can't connect to storage: %w", err)
	}

	if err := s.Init(context.TODO()); err != nil {
		log.Fatal("can't init storage:", err)
	}

	eventsProcessor := telegram2.New(
		telegram.New(tgBotHost, mustToken()),
		s)
	//сообщение, что сервер запущен
	log.Print("service started")
	//запускаем консьюмера
	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)
	if err := consumer.Start(); err != nil {
		//ошибка может быть если консюмер по какой-то причине аварийно остановился
		//запишет сообщение об ошибке и остановит программу
		log.Fatal("service is stopped", err)
	}

}

// фу-ия аварийно завершает прогамму, если токен оказался пустым (must)
func mustToken() string {
	//токен передаем из командной сторки при запуске программы
	//(имя флага,значение по умолчанию,подсказака для данного флага)
	//bot -tg-bot-token 'my token'
	token := flag.String(
		"tg-bot-token",
		"",
		"token for access to telegram bot",
	)
	//значение попадает во время вызова метода парс
	flag.Parse()

	if *token == "" {
		log.Fatal("token is not specified") //аварийно завершаем
	}

	return *token

}
