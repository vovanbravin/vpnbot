package middleware

import (
	"context"
	"tgbot/internal/database/models"
	"tgbot/internal/database/repositories"
	"time"

	"gopkg.in/telebot.v4"
)

func AdminOnly(userRepo *repositories.UserRepository) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			user, err := userRepo.GetUserByTelegramId(ctx, c.Sender().ID)

			if err != nil || user == nil || user.Role != models.RoleAdmin {
				return c.Send("У вас нет прав на эту команду")
			}

			return next(c)
		}
	}
}
