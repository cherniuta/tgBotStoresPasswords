package telegram

// в getUpdate будет еще прочая информация кроме update
// поэтому нужно находить ok и поле result с апдейтами
type UpdatesResponse struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}
type Update struct {
	ID int `json:"update_id"` //теги json для того,чтобы правильно парсить и находит нужный кусок
	//message отдельный тип
	//тк сообщение может отсутсвовать, здесь может быть nil, поэтому указываем ссылку на стрктуру
	Message *IncomingMessage `json:"message"`
}

// входящее сообщение от пользователя с его структурой (text ,from(от кого),chat(какой чат))
type IncomingMessage struct {
	Text string `json:"text"`
	//from и chat также отдельные типы
	From From `json:"from"`
	Chat Chat `json:"chat"`
}

type From struct {
	Username string `json:"username"`
}

type Chat struct {
	ID int `json:"id"`
}
