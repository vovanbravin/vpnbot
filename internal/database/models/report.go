package models

import (
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ReportStatus string

const (
	ReportStatusNew        ReportStatus = "new"
	ReportStatusInProgress ReportStatus = "in_progress"
	ReportStatusResolved   ReportStatus = "resolved"
)

type ReportPriority string

const (
	ReportPriorityLow    ReportPriority = "low"
	ReportPriorityMedium ReportPriority = "medium"
	ReportPriorityHigh   ReportPriority = "high"
)

type ReportCategory string

const (
	CategoryTechnical  ReportCategory = "technical"
	CategoryPayment    ReportCategory = "payment"
	CategoryConnection ReportCategory = "connection"
	CategoryFeature    ReportCategory = "feature"
	CategoryOther      ReportCategory = "other"
)

type Report struct {
	ID       bson.ObjectID `bson:"_id,omitempty"`
	Number   int64         `bson:"number"`
	UserID   int64         `bson:"user_id"`
	Username string        `bson:"username"`

	Categoty ReportCategory `bson:"categoty"`
	Priority ReportPriority `bson:"priority"`
	Status   ReportStatus   `bson:"status"`
	Subject  string         `bson:"subject"`
	Message  string         `bson:"message"`
	Answer   string         `bson:"answer"`

	CreatedAt time.Time `bson:"createdAt"`
}

func (r *Report) ShortInfo() string {
	statusEmoji := r.Status.Emoji()

	return fmt.Sprintf(
		"%s #%d  %s\n📝 %s\n📅 %s",
		statusEmoji,
		r.Number,
		r.Categoty.DisplayName(),
		Truncate(r.Subject, 40),
		r.CreatedAt.Format("02.01 15:04"),
	)
}

func (r *Report) DetailInfo() string {
	var sb strings.Builder

	sb.WriteString("📋 *ЗАЯВКА #" + fmt.Sprintf("%d", r.Number) + "*\n\n")

	sb.WriteString(fmt.Sprintf("📌 Статус: %s %s\n", r.Status.Emoji(), r.Status.DisplayName()))

	sb.WriteString(fmt.Sprintf("📂 Категория: %s\n", r.Categoty.DisplayName()))

	sb.WriteString(fmt.Sprintf("👤 Пользователь: @%s (ID: %d)\n", r.Username, r.UserID))

	sb.WriteString(fmt.Sprintf("📝 Тема: %s\n", r.Subject))

	sb.WriteString(fmt.Sprintf("📄 Описание:%s\n", r.Message))

	sb.WriteString(fmt.Sprintf("📅 Создана: %s", r.CreatedAt.Format("02.01.2006 15:04:05")))

	return sb.String()
}

func (s ReportStatus) Emoji() string {
	switch s {
	case ReportStatusNew:
		return "⚪"
	case ReportStatusInProgress:
		return "🟡"
	case ReportStatusResolved:
		return "🟢"
	default:
		return "⚪"
	}
}

func (s ReportStatus) DisplayName() string {
	switch s {
	case ReportStatusNew:
		return "Новая"
	case ReportStatusInProgress:
		return "В обработке"
	case ReportStatusResolved:
		return "Закрыта"
	default:
		return "Неизвестный статус"
	}
}
func (s ReportCategory) DisplayName() string {
	switch s {
	case CategoryTechnical:
		return "Технические проблемы"
	case CategoryConnection:
		return "Проблемы с соединение"
	case CategoryPayment:
		return "Проблемы с оплатой"
	case CategoryFeature:
		return "Предложения"
	case CategoryOther:
		return "Другое"
	default:
		return "Неизвестная"
	}
}

func Truncate(s string, maxLen int) string {
	runes := []rune(s)

	if len(runes) < maxLen {
		return s
	}

	return string(runes[:maxLen-3]) + "..."
}
