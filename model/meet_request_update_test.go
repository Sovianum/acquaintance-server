package model

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMeetRequestUpdate_Unmarshal_ParseError(t *testing.T) {
	var position = &Position{}
	var data = []byte("{")
	var err = json.Unmarshal(data, position)
	assert.NotNil(t, err)
}

func TestMeetRequestUpdate_Unmarshal_Success(t *testing.T) {
	var update = &MeetRequestUpdate{}
	var data = []byte("{\"id\": 100, \"status\": \"PENDING\"}")
	var err = json.Unmarshal(data, &update)

	assert.Nil(t, err)
	assert.Equal(t, 100, update.Id)
	assert.Equal(t, StatusPending, update.Status)
}

func TestMeetRequestUpdate_Unmarshal_Incomplete(t *testing.T) {
	var update = &MeetRequestUpdate{}
	var data = []byte("{\"id\": 100}")
	var err = json.Unmarshal(data, &update)

	assert.NotNil(t, err)
	assert.Equal(t, MeetRequestUpdateRequiredStatus, err.Error())
}

func TestMeetRequestUpdate_Unmarshal_BadStatus(t *testing.T) {
	var update = &MeetRequestUpdate{}
	var data = []byte("{\"id\": 100, \"status\": \"BAD\"}")
	var err = json.Unmarshal(data, &update)

	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("got invalid status %s", update.Status), err.Error())
}
