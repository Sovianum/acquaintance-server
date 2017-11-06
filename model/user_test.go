package model

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUser_Unmarshal_ParseError(t *testing.T) {
	var user = &User{}
	var data = []byte("{\"age\": 10, \"sex\": \"F\"")
	var err = json.Unmarshal(data, user)
	assert.NotNil(t, err)
}

func TestUser_Unmarshal_InvalidUser(t *testing.T) {
	var user = &User{}
	var data = []byte("{\"login\": \"login\", \"password\": \"pass\", \"sex\": \"p\"}")
	var err = json.Unmarshal(data, &user)

	assert.NotNil(t, err)
	assert.Equal(t, RegistrationInvalidSex, err.Error())
}

func TestUser_Unmarshal_Success(t *testing.T) {
	var user = User{}
	var data = []byte("{\"id\": 10, \"login\": \"login\", \"password\": \"password\"}")
	var err = json.Unmarshal(data, &user)

	assert.Nil(t, err)
	assert.Equal(t, "login", user.Login)
	assert.Equal(t, 10, user.Id)
}
