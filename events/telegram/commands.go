package telegram

import (
	"context"
	"errors"
	"log"
	"regexp"
	"tgBotStoresPasswords/lib/e"
	"tgBotStoresPasswords/storage"
)

const (
	GetCmd   = "/get"
	SetCmd   = "/set"
	DelCmd   = "/del"
	StartCmd = "/start"
	HelpCmd  = "/help"
)

// все команды, которые сможет отправлять бот
// будем смотреть на текст сообщения и будем понимать что это за команда
func (p *Processor) doCmd(text string, chatID int, username string) error {
	//удалим из тектса сообщения лишние пробелы
	//text = strings.TrimSpace(text)
	//пропишем логи для отслеживания того,кто нашему боту что пишет
	log.Printf("got new command '%s' from '%s'", text, username)

	switch text {
	case StartCmd:
		return p.sendHello(chatID, username)
	case HelpCmd:
		return p.sendHelp(chatID)
	}

	if command, _ := p.storage.GetCommand(context.Background(), username); command.URL == "not" {
		switch text {
		case GetCmd:
			return p.SetCommand(chatID, username, text)
		case SetCmd:
			return p.SetCommand(chatID, username, text)
		case DelCmd:
			return p.SetCommand(chatID, username, text)
		default:
			return p.tg.SendMessage(chatID, msgUnknownCommand)
		}

	} else {
		switch command.URL {
		case GetCmd:
			return p.sendData(chatID, username)
		case SetCmd:
			if isAffCmd(text) {
				return p.saveData(chatID, text, username)
			}
			return p.tg.SendMessage(chatID, msgUnknownCommand)
		case DelCmd:
			return p.deleteData(chatID, username)
		default:
			return p.tg.SendMessage(chatID, msgUnknownCommand)

		}
	}

}

func (p *Processor) saveData(chatID int, pageURL string, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: save page", err) }()

	//подготовим станицу, которую хотим сохранить
	page := &storage.Page{
		URL:      pageURL,
		UserName: username,
	}

	//пытаемся сохранить страницу
	if err := p.storage.Save(context.Background(), page); err != nil {
		return err
	}

	var com string = "not"

	if err := p.storage.CreateCommand(context.Background(), com, username); err != nil {
		return err
	}

	//если страница корректнго сохранилась, то сообщаем об этом пользователю
	if err := p.tg.SendMessage(chatID, msgSaved); err != nil {
		return err
	}

	return nil
}

func (p *Processor) sendData(chatID int, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: can't send random", err) }()

	//ищем случайную статью
	page, err := p.storage.PickRandom(context.Background(), username)

	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}

	//особый тип ошибок, когда нет сохраненых страниц
	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tg.SendMessage(chatID, msgNoSavedPages)
	}

	//если же мф что-то нашли, отправляем эту ссылку пользователю
	if err := p.tg.SendMessage(chatID, page.URL); err != nil {
		return err
	}

	var com string = "not"

	if err := p.storage.CreateCommand(context.Background(), com, username); err != nil {
		return err
	}

	//если мф нашли и отправили ссылку, то нужно обязательно ее удалить
	return nil
}

func (p *Processor) deleteData(chatID int, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: can't delete data", err) }()

	//ищем случайную статью
	page, err := p.storage.PickRandom(context.Background(), username)

	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}

	//особый тип ошибок, когда нет сохраненых страниц
	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tg.SendMessage(chatID, msgNoSavedPages)
	}

	if err := p.storage.Remove(context.Background(), page); err != nil {
		return err
	}

	if err := p.tg.SendMessage(chatID, msgDelete); err != nil {
		return err
	}

	var com string = "not"

	if err := p.storage.CreateCommand(context.Background(), com, username); err != nil {
		return err
	}

	//если мф нашли и отправили ссылку, то нужно обязательно ее удалить
	return nil

}

func (p *Processor) SetCommand(chatID int, username string, command string) error {
	if err := p.storage.CreateCommand(context.Background(), command, username); err != nil {
		return err
	}

	return nil
}

func (p *Processor) sendHelp(chatID int) error {
	return p.tg.SendMessage(chatID, msgHelp)
}

func (p *Processor) sendHello(chatID int, userName string) error {
	p.SetCommand(chatID, userName, "not")
	return p.tg.SendMessage(chatID, msgHello)
}

func isAffCmd(text string) bool {
	return isTrueData(text)
}

func isTrueData(text string) bool {
	match, err := regexp.MatchString(`^\S+\s\S+$`, text)

	return err == nil && match
}
