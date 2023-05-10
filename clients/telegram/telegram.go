package telegram

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"tgBotStoresPasswords/lib/e"
)

type Client struct {
	//хост API телеграма
	host string
	//базовый путь- префикс, с которого начинаются все запросы
	//tg-bot.com/bot<token>
	basePath string
	//протокол, по которому все это будет передаваться
	client http.Client
}

const (
	getUpdatesMethod  = "getUpdates"
	sendMessageMethod = "sendMessage"
)

// фу-ия будет создавать экземпляр структры client
func New(host string, token string) *Client {
	return &Client{
		host: host,
		//могли бы написать "bot"+token,но вынесли в отдельную фу-ию, чтобы если что менять только одну ф-ию,а не все места,где есть токен
		basePath: newBasePath(token),
		client:   http.Client{},
	}
}

func newBasePath(token string) string {
	return "bot" + token
}

// полчение обновлений(сообщений)
func (c *Client) Updates(offset int, limit int) ([]Update, error) {
	//формируем запрос где ключи и значения строки -параметры запроса HTTP для формир URL-дреса
	q := url.Values{}
	//добавлеяем параметры к запросу
	q.Add("offset", strconv.Itoa(offset))
	q.Add("limit", strconv.Itoa(limit))

	//отправляем запрос и получаем ответ
	data, err := c.doRequest(getUpdatesMethod, q)
	if err != nil {
		return nil, err
	}

	//ответ в виде json, поэтому нужно его распарсить
	var res UpdatesResponse

	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	return res.Result, nil

}

// отправка сообщений
// экспортируемые метода должны быть выше
func (c *Client) SendMessage(chatID int, text string) error {
	q := url.Values{}

	q.Add("chat_id", strconv.Itoa(chatID))
	q.Add("text", text)

	//Выполняем запрос,но тело ответа нам не понадобится
	_, err := c.doRequest(sendMessageMethod, q)
	if err != nil {
		return e.Wrap("can't send messege", err)
	}

	return nil

}

// фу-ия для отправки запросов клиента
func (c *Client) doRequest(method string, query url.Values) (data []byte, err error) {
	//фу-ия вызовется после нашей и проверит на ошибку
	defer func() { err = e.WrapIfErr("can't do request", err) }()
	//формируем url, на который будет отправляться запрос
	u := url.URL{
		Scheme: "https",
		Host:   c.host,
		//можно было сформировать пусть вот так c.basePath+method
		//но так делать неудобно, т к между ними могуть быть лишние / или их может не доставать
		//есть фу-я, которая все сама правильно склеивает
		Path: path.Join(c.basePath, method),
	}

	//формирует объект запросы с помощью метода HTTP метода ,нашего url и тела запроса(nil в нашем случае)
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	// передаем в объект наши парметры в аргументе
	//приводит параметры к такому виду, в котором мы сможем отправлять их на сервер
	req.URL.RawQuery = query.Encode()

	// отправляем запрос с помощью подготовленного клиента и его метода Do,в который передаем наш объект
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	//закрываем тело ответы после выполенеия нашей фу-ии, чтобы оно не осталось открыты в памяти
	//это нужно для экономии ресурсов и паямти
	defer func() { _ = resp.Body.Close() }()

	//только после этого получаем содержимое тела
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil

}
