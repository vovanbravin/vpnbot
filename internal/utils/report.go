package utils

import "tgbot/internal/database/models"

func GetCategoryByName(name string) string {
	switch name {
	case "technical":
		return "Техническая"
	case "connection":
		return "Проблемы с соединением"
	case "payment":
		return "Проблемы с оплатой"
	case "feature":
		return "Предложения"
	case "other":
		return "Другое"
	default:
		return "Неизстная категория"
	}
}

func GetPriorityByCategory(category models.ReportCategory) models.ReportPriority {
	switch category {
	case models.CategoryTechnical, models.CategoryConnection, models.CategoryPayment:
		return models.ReportPriorityHigh
	case models.CategoryFeature:
		return models.ReportPriorityLow
	case models.CategoryOther:
		return models.ReportPriorityMedium
	default:
		return models.ReportPriorityLow
	}
}
