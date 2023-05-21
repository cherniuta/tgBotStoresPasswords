package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"tgBotStoresPasswords/storage"
)

type Storage struct {
	//общий интерфейс для взаимодействия со всеми бд
	db *sql.DB
}

func New(path string) (*Storage, error) {
	//уточняем с какой бд будем работать и передаем путь до файла
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("can't open database: %w", err)
	}

	//проверяем удалось ли нам установить соединение с файлом
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("can't connect to database: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Save(ctx context.Context, service string, p *storage.Page) error {
	q := `SELECT COUNT(*) FROM login_details WHERE user_name=? AND service=?`

	var count int

	if err := s.db.QueryRowContext(ctx, q, p.UserName, service).Scan(&count); err != nil {
		return fmt.Errorf("can't check if page exists: %w", err)
	}

	if count == 0 {
		q = `INSERT INTO login_details(service,information,user_name) VALUES(?,?,?)`

		if _, err := s.db.ExecContext(ctx, q, service, p.Text, p.UserName); err != nil {
			return fmt.Errorf("can't save data: %w", err)
		}
	} else {
		q = `UPDATE login_details SET information=? WHERE user_name=? AND service=?`

		if _, err := s.db.ExecContext(ctx, q, p.Text, p.UserName, service); err != nil {
			return fmt.Errorf("can't update command: %w", err)
		}
	}

	return nil
}

func (s *Storage) PickPage(ctx context.Context, service, userName string) (*storage.Page, error) {

	q := `SELECT information FROM login_details WHERE user_name=? AND service=?`

	var information string

	err := s.db.QueryRowContext(ctx, q, userName, service).Scan(&information)

	if err == sql.ErrNoRows {
		return nil, storage.ErrNoSavedPages
	}
	if err != nil {
		return nil, fmt.Errorf("can't pick page: %w", err)

	}

	return &storage.Page{
		Text:     information,
		UserName: userName,
	}, nil
}

func (s *Storage) Remove(ctx context.Context, service string, page *storage.Page) error {
	q := `SELECT information FROM login_details WHERE user_name=? AND service=?`

	var information string

	err := s.db.QueryRowContext(ctx, q, page.UserName, service).Scan(&information)

	if err == sql.ErrNoRows {
		return storage.ErrNoSavedPages
	}
	if err != nil {
		return fmt.Errorf("can't pick page and remove page: %w", err)

	}

	q = `DELETE FROM login_details WHERE service=? AND user_name=?`

	_, err = s.db.ExecContext(ctx, q, service, page.UserName)

	if err != nil {
		return fmt.Errorf("can't remove page: %w", err)
	}

	return nil
}

func (s *Storage) CreateService(ctx context.Context, service string, userName string) error {
	q := `UPDATE commands SET service=? WHERE user_name=?`

	if _, err := s.db.ExecContext(ctx, q, service, userName); err != nil {
		return fmt.Errorf("can't update service: %w", err)
	}

	return nil
}

func (s *Storage) CreateCommand(ctx context.Context, command string, userName string) error {
	q := `SELECT COUNT(*) FROM commands WHERE user_name=?`

	var count int

	if err := s.db.QueryRowContext(ctx, q, userName).Scan(&count); err != nil {
		return fmt.Errorf("can't check if page exists: %w", err)
	}

	if count == 0 {
		q = `INSERT INTO commands(service,command,user_name) VALUES(?,?,?)`

		if _, err := s.db.ExecContext(ctx, q, "not", command, userName); err != nil {
			return fmt.Errorf("can't create command: %w", err)
		}
	} else {
		q = `UPDATE commands SET command=?,service=? WHERE user_name=?`

		if _, err := s.db.ExecContext(ctx, q, command, "not", userName); err != nil {
			return fmt.Errorf("can't update command: %w", err)
		}
	}

	return nil
}
func (s *Storage) IsUsersDataEmpty(ctx context.Context, userName string) (bool, error) {
	q := `SELECT COUNT(*) FROM login_details WHERE user_name=?`

	var count int

	if err := s.db.QueryRowContext(ctx, q, userName).Scan(&count); err != nil {
		return true, fmt.Errorf("can't check if page exists: %w", err)
	}

	return count == 0, nil
}

func (s *Storage) GetCommand(ctx context.Context, userName string) (*storage.Page, error) {
	q := `SELECT command FROM commands WHERE user_name=? `

	var com string

	err := s.db.QueryRowContext(ctx, q, userName).Scan(&com)

	if err == sql.ErrNoRows {
		return nil, storage.ErrNoSavedCommand
	}
	if err != nil {
		return nil, fmt.Errorf("can't pick random page: %w", err)

	}

	return &storage.Page{
		Text:     com,
		UserName: userName,
	}, nil

}

func (s *Storage) GetService(ctx context.Context, userName string) (*storage.Page, error) {
	g := `SELECT service FROM commands WHERE user_name=? `

	var service string

	err := s.db.QueryRowContext(ctx, g, userName).Scan(&service)
	if err != nil {
		return nil, fmt.Errorf("can't pick service from tabl commands: %w", err)

	}

	return &storage.Page{
		Text:     service,
		UserName: userName,
	}, nil
}

// инициализируем нашу базу
func (s *Storage) Init(ctx context.Context) error {
	//создать таблицу, если она еще не существует
	q := `CREATE TABLE IF NOT EXISTS login_details (service TEXT,information TEXT,user_name TEXT)`
	t := `CREATE TABLE IF NOT EXISTS commands (command TEXT,service TEXT,user_name TEXT)`

	_, err := s.db.ExecContext(ctx, q)
	if err != nil {
		return fmt.Errorf("can't create table: %w", err)
	}
	_, err = s.db.ExecContext(ctx, t)
	if err != nil {
		return fmt.Errorf("can't create table: %w", err)
	}

	return nil
}
