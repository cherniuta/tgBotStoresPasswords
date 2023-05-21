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

	//пропишем логи для отслеживания того,кто нашему боту что пишет
	log.Printf("got new command '%s' from '%s'", text, username)

	switch text {
	case StartCmd:
		return p.sendHello(chatID, username)
	case HelpCmd:
		return p.sendHelp(chatID)
	}

	switch text {
	case GetCmd:
		ok, err := p.storage.IsUsersDataEmpty(context.Background(), username)
		if err != nil {
			return err
		}

		if ok == true {
			p.SetCommand(chatID, username, "not")
			p.SetService(chatID, username, "not")
			return p.tg.SendMessage(chatID, msgNoSaved)
		}

		err = p.tg.SendMessage(chatID, msgSetService)
		if err != nil {
			return err
		}

		return p.SetCommand(chatID, username, text)
	case SetCmd:
		err := p.tg.SendMessage(chatID, msgSetService)
		if err != nil {
			return err
		}
		return p.SetCommand(chatID, username, text)
	case DelCmd:
		ok, err := p.storage.IsUsersDataEmpty(context.Background(), username)
		if err != nil {
			return err
		}

		if ok == true {
			p.SetCommand(chatID, username, "not")
			p.SetService(chatID, username, "not")
			return p.tg.SendMessage(chatID, msgNoSaved)
		}
		err = p.tg.SendMessage(chatID, msgSetService)
		if err != nil {
			return err
		}
		return p.SetCommand(chatID, username, text)
	}

	if command, _ := p.storage.GetCommand(context.Background(), username); command.Text != "not" {
		if service, _ := p.storage.GetService(context.Background(), username); service.Text == "not" {

			err := p.SetService(chatID, username, text)
			if err != nil {
				return err
			}

			switch command.Text {
			case GetCmd:
				return p.sendData(chatID, username)
			case DelCmd:
				return p.deleteData(chatID, username)
			case SetCmd:
				return p.tg.SendMessage(chatID, msgSetData)
			default:
				return p.tg.SendMessage(chatID, msgUnknownCommand)

			}
		} else {

			switch command.Text {
			case SetCmd:
				if isAffCmd(text) {
					return p.saveData(chatID, text, username)
				}
				return p.tg.SendMessage(chatID, msgWrongFormat)
			case GetCmd:
				err := p.SetService(chatID, username, text)
				if err != nil {
					return err
				}
				return p.sendData(chatID, username)
			case DelCmd:
				err := p.SetService(chatID, username, text)
				if err != nil {
					return err
				}
				return p.deleteData(chatID, username)
			default:
				return p.tg.SendMessage(chatID, msgUnknownCommand)

			}
		}

	}
	return p.tg.SendMessage(chatID, msgUnknownCommand)

}

func (p *Processor) saveData(chatID int, pageURL string, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: save page", err) }()

	page := &storage.Page{
		Text:     pageURL,
		UserName: username,
	}
	service, err := p.storage.GetService(context.Background(), username)
	if err != nil {
		return err
	}

	if err := p.storage.Save(context.Background(), service.Text, page); err != nil {
		return err
	}

	var com string = "not"

	if err := p.storage.CreateCommand(context.Background(), com, username); err != nil {
		return err
	}

	if err := p.tg.SendMessage(chatID, msgSaved); err != nil {
		return err
	}

	return nil
}

func (p *Processor) sendData(chatID int, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: can't send random", err) }()

	service, err := p.storage.GetService(context.Background(), username)
	if err != nil {
		return err
	}

	page, err := p.storage.PickPage(context.Background(), service.Text, username)

	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}

	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tg.SendMessage(chatID, msgNoSavedService)
	}

	if err := p.tg.SendMessage(chatID, page.Text); err != nil {
		return err
	}

	var com string = "not"

	if err := p.storage.CreateCommand(context.Background(), com, username); err != nil {
		return err
	}

	return nil
}

func (p *Processor) deleteData(chatID int, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: can't delete data", err) }()

	service, err := p.storage.GetService(context.Background(), username)
	if err != nil {
		return err
	}

	page, err := p.storage.PickPage(context.Background(), service.Text, username)

	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}

	if errors.Is(err, storage.ErrNoSavedPages) {

		return p.tg.SendMessage(chatID, msgNoSavedService)
	}

	if err := p.storage.Remove(context.Background(), service.Text, page); err != nil {
		return err
	}

	if err := p.tg.SendMessage(chatID, msgDelete); err != nil {
		return err
	}

	var com string = "not"

	if err := p.storage.CreateCommand(context.Background(), com, username); err != nil {
		return err
	}

	return nil

}

func (p *Processor) SetCommand(chatID int, username string, command string) error {
	if err := p.storage.CreateCommand(context.Background(), command, username); err != nil {
		return err
	}

	return nil
}
func (p *Processor) SetService(chatID int, username string, service string) error {
	if err := p.storage.CreateService(context.Background(), service, username); err != nil {
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
