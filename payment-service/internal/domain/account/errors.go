package account

import "errors"

var (
	ErrInsufficientFunds = errors.New("недостаточно средств на счёте")
	ErrAccountNotFound   = errors.New("счёт пользователя не найден")
)
