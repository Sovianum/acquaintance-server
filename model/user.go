package model

import (
	"errors"
	"strings"
	"encoding/json"
)

const (
	MALE   = "M"
	FEMALE = "F"

	UserRequiredLogin      = "\"login\" field required"
	UserRequiredPassword   = "\"password\" field required"
	RegistrationInvalidSex = "\"invalid sex: must be either M or F\""
)

type User struct {
	Id       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Age      int    `json:"age"`
	Sex      string `json:"sex"`
	About    string `json:"about"`
}

func (user *User) UnmarshalJSON(data []byte) error {
	var err = checkPresence(
		data,
		[]string{"login", "password"},
		[]string{UserRequiredLogin, UserRequiredPassword},
	)
	if err != nil {
		return err
	}

	type Alias User
	var dest = (*Alias)(user)

	err = json.Unmarshal(data, dest)
	if err != nil {
		return err
	}

	err = user.Validate()

	return err
}

func (user *User) Validate() error {
	var msgList = make([]string, 0)
	if user.Sex != "" && user.Sex != MALE && user.Sex != FEMALE {
		msgList = append(msgList, RegistrationInvalidSex)
	}

	if len(msgList) != 0 {
		return errors.New(strings.Join(msgList, ";\n"))
	}
	return nil
}
