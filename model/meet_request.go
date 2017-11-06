package model

const (
	StatusPending  = "PENDING"
	StatusAccepted = "ACCEPTED"
	StatusDeclined = "DECLINED"
)

type MeetRequest struct {
	Id             int        `json:"id"`
	RequesterId    int        `json:"requester_id"`
	RequesterLogin string     `json:"requester_login"`
	RequestedId    int        `json:"requested_id"`
	RequestedLogin string     `json:"requested_login"`
	Time           QuotedTime `json:"time"`
	Status         string     `json:"status"`
}
