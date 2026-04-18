package message

const (
	ReportCategoryPrompt = "📂 Выберите категорию проблемы:"
	ReportSubjectPrompt  = "✏️ Напишите тему проблемы"
	ReportMessagePrompt  = "📝 Опишите проблему подробнее"
	ReportCreated        = "✅ Заявление успешно создано!"
	ReportConfirmHeader  = "📋 *Проверьте данные заявки:*\n\n"
	ReportConfirmLabel   = "📂 Категория: %s\n📝 Тема: %s\n📄 Описание: %s"
	ReportNumberPrompt   = "🔢 Введите номер заявления"
	ReportNotFound       = "❌ Заявка не найдена. Убедитесь, что вы правильно ввели номер"
	NoActiveReports      = "📭 У вас пока нет активных заявок"
	NoClosedReports      = "📭 У вас пока нет закрытых заявок"
	UnknownOption        = "❌ Неизвестная опция"
	SupportMenuPrompt    = "🔧 Выберите опцию:"
	AnswerOnReport       = "📨 На вашу заявку с номером #%d ответили!\n\n👀 Посмотреть ответ вы можете в меню заявок /my_report"
	MyReport             = "📋 Мои заявки"
)

func ReportConfirm(category, subject, message string) string {
	return ReportConfirmLabel + "\n\n✅ Подтверждаете создание заявки?"
}

func ReportConfirmText(category, subject, message string) string {
	return "📂 Категория: " + category + "\n" +
		"📝 Тема: " + subject + "\n" +
		"📄 Описание: " + message
}
