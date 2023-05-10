package event_consumer

import (
	"log"
	"tgBotStoresPasswords/events"
	"time"
)

type Consumer struct {
	fetcher   events.Fetcher
	processor events.Processor
	//размер пачки-сколько событий мы будемобрабатывать за раз
	batchSize int
}

func New(fetcher events.Fetcher, processor events.Processor, batchSize int) Consumer {
	return Consumer{
		fetcher:   fetcher,
		processor: processor,
		batchSize: batchSize,
	}
}

// метод старт
func (c Consumer) Start() error {
	//вечный цикл,который постоянно ждет новые события и обрабатывает их
	for {
		//получаем событие
		gotEvents, err := c.fetcher.Fetch(c.batchSize)
		if err != nil {
			//пишем в лог что случилось
			log.Printf("[ERR] consumer: %s", err.Error())
			//пропускаем данную итерацию
			continue
		}

		//проверям сколько мы получили событий ,если 0,то пропускам ит
		if len(gotEvents) == 0 {
			//ждем 1 секунду
			time.Sleep(1 * time.Second)

			continue
		}

		//если что-то нашли-обрабатываем
		if err := c.handleEvents(gotEvents); err != nil {
			log.Print(err)

			continue
		}
	}
}

// обрабатываем события с помощью функции,тк их может быть несколько
func (c *Consumer) handleEvents(events []events.Event) error {
	//перебираем события
	for _, event := range events {
		//пишем сообщение о том, что получили новое событие
		log.Printf("got new event: %s", event.Text)

		//для обработки событий у нас уже есть процессор
		if err := c.processor.Process(event); err != nil {
			//если что-то пойдет не так-выведет сообщение и перейдем к след ит
			log.Printf("can't handle event: %s", err.Error())

			continue
		}
	}

	return nil
}
