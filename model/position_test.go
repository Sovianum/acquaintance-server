package model

import (
	"testing"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"time"
)

func TestPosition_ReadJsonIn_ParseError(t *testing.T) {
	var user = &Position{}
	var data = []byte("{")
	var err = json.Unmarshal(data, user)
	assert.NotNil(t, err)
}

func TestPosition_ReadJsonIn_IncompleteData(t *testing.T) {
	var user = User{}
	var data = []byte("{\"login\": \"login\"}")
	var err = json.Unmarshal(data, &user)

	assert.NotNil(t, err)
	assert.Equal(t, UserRequiredPassword, err.Error())
}

func TestPosition_Unmarshal_Success(t *testing.T) {
	var pos = Position{}
	var data = []byte("{\"user_id\": 100, \"time\": \"2006-01-02T15:04:05\", \"point\": {\"x\": 100, \"y\": 200}}")
	var err = json.Unmarshal(data, &pos)

	assert.Nil(t, err)
	assert.Equal(t, 100, pos.UserId)
	assert.Equal(t, Point{100, 200}, pos.Point)

	var timeStamp, timeErr = time.Parse("2006-01-02T15:04:05", "2006-01-02T15:04:05")
	assert.Nil(t, timeErr)
	assert.Equal(t, QuotedTime(timeStamp), pos.Time)
}