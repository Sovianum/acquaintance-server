package model

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPosition_Unmarshal_ParseError(t *testing.T) {
	var user = &Position{}
	var data = []byte("{")
	var err = json.Unmarshal(data, user)
	assert.NotNil(t, err)
}

func TestPosition_Unmarshal_Success(t *testing.T) {
	var pos = Position{}
	var data = []byte("{\"user_id\": 100, \"time\": \"2006-01-02T15:04:05Z\", \"point\": {\"x\": 100, \"y\": 200}}")
	var err = json.Unmarshal(data, &pos)

	assert.Nil(t, err)
	assert.Equal(t, 100, pos.UserId)
	assert.Equal(t, Point{100, 200}, pos.Point)

	var timeStamp, timeErr = time.Parse("2006-01-02T15:04:05", "2006-01-02T15:04:05")
	assert.Nil(t, timeErr)
	assert.Equal(t, QuotedTime(timeStamp), pos.Time)
}
