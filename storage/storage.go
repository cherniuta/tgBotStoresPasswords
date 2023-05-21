package storage

import (
	"context"
	"errors"
)

//здесь как и в случае с event будет храниться только интерфейс
//поэтому мы сможем работать с любым сервисом
//будем работать тут с файлами

type Storage interface {
	Save(ctx context.Context, service string, p *Page) error

	PickPage(ctx context.Context, service, userName string) (*Page, error)

	Remove(ctx context.Context, service string, p *Page) error

	CreateCommand(ctx context.Context, command string, userName string) error

	GetCommand(ctx context.Context, userName string) (*Page, error)

	CreateService(ctx context.Context, service string, userName string) error

	GetService(ctx context.Context, userName string) (*Page, error)

	IsUsersDataEmpty(ctx context.Context, userName string) (bool, error)
}

var (
	ErrNoSavedPages   = errors.New("no saved page")
	ErrNoSavedCommand = errors.New("no saved command")
)

type Page struct {
	Text     string
	UserName string
}


