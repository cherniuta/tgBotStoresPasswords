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

func (s *Storage) Save(ctx context.Context, p *storage.Page) error {
	//пишем sql запрос,который сохраняет запись в бд
	q := `INSERT INTO pages(url,user_name) VALUES(?,?)`
	//выполянем этот запрос с помощью сущности db
	//использование конекста(общий таймаут на все вложенные и побочные вызовы) -хороший тон
	//(конекст,запрос,аргументы к запросу)
	//результат был бы интересен, если бы мы с этими данными дальше захотели бы что-то сделать
	if _, err := s.db.ExecContext(ctx, q, p.URL, p.UserName); err != nil {
		return fmt.Errorf("can't save page: %w", err)
	}

	return nil
}

func (s *Storage) PickRandom(ctx context.Context, userName string) (*storage.Page, error) {
	//с селект, тк получаем данные
	//получаем ссылку от данного пользователя отсортированные в случайно порядке и возьмем первую из них
	q := `SELECT url FROM pages WHERE user_name=? ORDER BY RANDOM() LIMIT 1`
	//переменная для ссылки
	var url string
	//выполянем запрос с помощью уже другой функции
	//тк данная функция возвращает row ,то нужно преобразовать ее с помощью scan
	err := s.db.QueryRowContext(ctx, q, userName).Scan(&url)
	//но может быть тип ошибки, когда в базе не нашлось данных по нашему запросу
	//для нас его нужно обработать по-другому
	if err == sql.ErrNoRows {
		return nil, storage.ErrNoSavedPages
	}
	if err != nil {
		return nil, fmt.Errorf("can't pick random page: %w", err)

	}

	return &storage.Page{
		URL:      url,
		UserName: userName,
	}, nil
}

func (s *Storage) Remove(ctx context.Context, page *storage.Page) error {
	q := `DELETE FROM pages WHERE url=? AND user_name=?`
	if _, err := s.db.ExecContext(ctx, q, page.URL, page.UserName); err != nil {
		return fmt.Errorf("can't remove page: %w", err)
	}

	return nil
}

func (s *Storage) CreateCommand(ctx context.Context, command string, userName string) error {
	q := `SELECT COUNT(*) FROM pages WHERE url=? AND user_name=?`

	var count int

	if err := s.db.QueryRowContext(ctx, q, command, userName).Scan(&count); err != nil {
		return fmt.Errorf("can't check if page exists: %w", err)
	}

	if count == 0 {
		q = `INSERT INTO commands(command,user_name) VALUES(?,?)`

		if _, err := s.db.ExecContext(ctx, q, command, userName); err != nil {
			return fmt.Errorf("can't create command: %w", err)
		}
	} else {
		q = `UPDATE commands SET command=? WHERE user_name=?`

		if _, err := s.db.ExecContext(ctx, q, command, userName); err != nil {
			return fmt.Errorf("can't update command: %w", err)
		}
	}

	return nil
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
		URL:      com,
		UserName: userName,
	}, nil

}

// инициализируем нашу базу
func (s *Storage) Init(ctx context.Context) error {
	//создать таблицу, если она еще не существует
	q := `CREATE TABLE IF NOT EXISTS pages (url TEXT,user_name TEXT)`
	t := `CREATE TABLE IF NOT EXISTS commands (command TEXT,user_name TEXT)`

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
