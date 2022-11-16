package model

const (
	EventsBasedDisplayCategory = "event_based"
)

var KpiCustomEventsConfig = map[string]interface{}{
	"category":         EventCategory,
	"display_category": EventsBasedDisplayCategory,
}
