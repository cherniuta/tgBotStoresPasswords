package e

import "fmt"

//принимает текст ошибки и саму ошибку

func Wrap(msg string, err error) error {
	return fmt.Errorf("%s: %w", msg, err)
}

// ошибка может быть нулевая, а наша фу-ия возвращает ненулевую ошибку
// поэтому проверяем с помощью еще одной фу-ии нулевая ли ошбика
func WrapIfErr(msg string, err error) error {
	if err == nil {
		return nil
	}
	return Wrap(msg, err)

}
