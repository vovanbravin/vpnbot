package fsm

type ReportData struct {
	Category string
	Subject  string
	Message  string
}

type AdminAnswer struct {
	Current int
	Answer  string
}
