package model

const (
	StatusPending     = "PENDING"
	StatusAccepted    = "ACCEPTED"
	StatusDeclined    = "DECLINED"
	StatusInterrupted = "INTERRUPTED"
)

type MeetRequest struct {
	Id             int        `json:"id"`
	RequesterId    int        `json:"requester_id"`
	RequesterLogin string     `json:"requester_login"`
	RequesterAbout string     `json:"requester_about"`
	RequestedId    int        `json:"requested_id"`
	RequestedLogin string     `json:"requested_login"`
	RequestedAbout string     `json:"requested_about"`
	Time           QuotedTime `json:"time"`
	Status         string     `json:"status"`
}
