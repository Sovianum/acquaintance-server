package model

import (
	"encoding/json"
	"fmt"
)

const (
	MeetRequestUpdateRequiredId     = "\"id\" field required"
	MeetRequestUpdateRequiredStatus = "\"status\" field required"
)

type MeetRequestUpdate struct {
	Id     int    `json:"id"`
	Status string `json:"status"`
}

func (update *MeetRequestUpdate) UnmarshalJSON(data []byte) error {
	var err = checkPresence(
		data,
		[]string{"id", "status"},
		[]string{MeetRequestUpdateRequiredId, MeetRequestUpdateRequiredStatus},
	)
	if err != nil {
		return err
	}

	type updateAlias MeetRequestUpdate
	var dest = (*updateAlias)(update)

	err = json.Unmarshal(data, dest)
	if err != nil {
		return err
	}

	err = update.Validate()

	return err
}

func (update *MeetRequestUpdate) Validate() error {
	var fail = update.Status != StatusPending
	fail = fail && update.Status != StatusAccepted
	fail = fail && update.Status != StatusDeclined
	fail = fail && update.Status != StatusInterrupted
	if fail {
		return fmt.Errorf("got invalid status %s", update.Status)
	}
	return nil
}
