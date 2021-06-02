package bot

import (
	"errors"
	"fmt"
)

var (
	ErrNotEnoughArgs = errors.New("команда пишется так: !ban <первый фолловер> <последний фолловер>")
)

type ErrUserNotFound struct {
	UserName string
}

func (e *ErrUserNotFound) Error() string {
	return fmt.Sprintf("Фолловер %s в базе не найден", e.UserName)
}
