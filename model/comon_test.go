package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckPresence_ParseFail(t *testing.T) {
	var jsonStr = "{\"some\":\"foo\", \"any\":\"bar\""
	var err = checkPresence([]byte(jsonStr), []string{"some"}, []string{"some"})
	assert.NotNil(t, err)
}

func TestCheckPresence_PresenceFail(t *testing.T) {
	var jsonStr = "{\"some\":\"foo\", \"any\":\"bar\"}"
	var err = checkPresence([]byte(jsonStr), []string{"som"}, []string{"some"})
	assert.NotNil(t, err)
	assert.Equal(t, "some", err.Error())
}

func TestCheckPresence_Success(t *testing.T) {
	var jsonStr = "{\"some\":\"foo\", \"any\":\"bar\"}"
	var err = checkPresence([]byte(jsonStr), []string{"some"}, []string{"some"})
	assert.Nil(t, err)
}

func TestCheckPresence_UnequalLength(t *testing.T) {
	var jsonStr = "{\"some\":\"foo\", \"any\":\"bar\"}"
	var err = checkPresence([]byte(jsonStr), []string{"some"}, []string{})
	assert.NotNil(t, err)
}
