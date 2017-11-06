package server

import (
	"github.com/Sovianum/acquaintance-server/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMailBox_AddAccept(t *testing.T) {
	var box = NewMailBox()
	var request = new(model.MeetRequest)

	var err1 = box.AddAccept(request)
	assert.Nil(t, err1)

	var err2 = box.AddAccept(request)
	assert.NotNil(t, err2)
}

func TestMailBox_AddDecline(t *testing.T) {
	var box = NewMailBox()
	var request = new(model.MeetRequest)

	box.AddDecline(request)

	var err2 = box.AddAccept(request)
	assert.Nil(t, err2)
}

func TestMailBox_GetAll(t *testing.T) {
	var box = NewMailBox()
	var request = new(model.MeetRequest)

	box.AddDecline(request)

	var requests = box.GetAll(1)
	assert.Equal(t, 1, len(requests))

	requests = box.GetAll(1)
	assert.Equal(t, 0, len(requests))
}
