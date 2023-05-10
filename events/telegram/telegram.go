package telegram

import (
	"errors"
	"tgBotStoresPasswords/clients/telegram"
	"tgBotStoresPasswords/events"
	"tgBotStoresPasswords/lib/e"
	"tgBotStoresPasswords/storage"
)

// реализовывать оба интерфейса будет один тип данных(EventProcessor)
type Processor struct {
	//телеграм клиент
	tg     *telegram.Client
	offset int
	//используем абстрактный интерфейс , а не конкретную его реализацию
	storage storage.Storage
}

// определим тип мета, который будет относится только к телеграмм
type Meta struct {
	ChatID   int
	Username string
}

var (
	ErrUnknownEventType = errors.New("unknown event type")
	ErrUnknownMetaType  = errors.New("unknown meta type")
)

// слздаем экземпляр процессора
func New(client *telegram.Client, storage storage.Storage) *Processor {
	return &Processor{
		tg:      client,
		storage: storage,
	}
}

func (p *Processor) Fetch(limit int) ([]events.Event, error) {
	//с помощью клиента получаем все апдейты
	updates, err := p.tg.Updates(p.offset, limit)
	if err != nil {
		return nil, e.Wrap("can't get events", err)
	}
	//если апдейтов мы не нашли, то возвращаем нулевой результат
	if len(updates) == 0 {
		return nil, nil
	}

	//выделяем память под результат
	res := make([]events.Event, 0, len(updates))

	//обходим все апдейты и преобразовываем в ивенты
	for _, u := range updates {
		res = append(res, event(u))
	}

	//обновляем позицию offset, передвигая его на 1
	//берем последний использованный нами в этой иттерации апдейт и увеличиваем на 1
	// получается в следующий раз м уже будем брать апдейты больше данного
	p.offset = updates[len(updates)-1].ID + 1

	return res, nil
}

// будет выполнять различные действия , в зависимости от типа эвента
func (p *Processor) Process(event events.Event) error {
	switch event.Type {
	case events.Message:
		return p.processMessage(event)
	default:
		return e.Wrap("can't process message", ErrUnknownEventType)

	}
}

// фу-ия по работе с сообщениями
func (p *Processor) processMessage(event events.Event) error {
	//нужно получить мету из эвента
	meta, err := meta(event)
	if err != nil {
		return e.Wrap("can't process message", err)
	}

	//вызываем функцию по определению команд
	if err := p.doCmd(event.Text, meta.ChatID, meta.Username); err != nil {
		return e.Wrap("can't process message", err)
	}

	return nil

}

// получение меты из эвента
func meta(event events.Event) (Meta, error) {
	//тайпасершен
	//если здесь будет что-то другое, то вернется false
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, e.Wrap("can't get meta", ErrUnknownMetaType)
	}
	//если же все ок, возвращвем результат
	return res, nil
}

// фу-ия преобразования апдейтов в ивенты
func event(upd telegram.Update) events.Event {
	//тип собфтия вынесли в отдельную переменную
	updType := fetchType(upd)
	res := events.Event{
		Type: updType,
		Text: fetchText(upd),
	}

	//если это сообщение, то добавляем параметр meta
	if updType == events.Message {
		res.Meta = Meta{
			//тк тип сообщение, мы точно занем, что сообщение будет не нулевое
			ChatID:   upd.Message.Chat.ID,
			Username: upd.Message.From.Username,
		}
	}

	return res
}

// фу-ия для определения типа события
func fetchType(upd telegram.Update) events.Type {
	//если сообщение нулевое,то тип неизвестен
	if upd.Message == nil {
		return events.Unknown
	}
	//если не нудевым, то событие является сообщением
	return events.Message

}

// фу-ия определения тектса события
func fetchText(upd telegram.Update) string {
	//поле сообщение может быть нудевым, поэтому обрабатываем этот момент
	if upd.Message == nil {
		return ""
	}
	return upd.Message.Text
}
