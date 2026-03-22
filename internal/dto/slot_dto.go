package dto

type SlotResponse struct {
	CoachID  int        `json:"coach_id"`
	Date     string     `json:"date"`
	Timezone string     `json:"time_zone"`
	Slots    []TimeSlot `json:"slots"`
}

type TimeSlot struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}
