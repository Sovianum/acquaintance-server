package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Validator interface {
	Validate() error
}

// Function checks whether jsonData contains all fields from fields slice.
// errMessages slice contains messages which are used if some field is not found.
// Resulting error message consists of all corresponding errMessages, joined with ";\n"
func checkPresence(jsonData []byte, fields []string, errMessages []string) error {
	if len(fields) != len(errMessages) {
		return errors.New(
			fmt.Sprintf("Fields slice must have the same length (%v) as errMessages (%v)", len(fields), len(errMessages)),
		)
	}
	var m = make(map[string]interface{})
	var err = json.Unmarshal(jsonData, &m)

	if err != nil {
		return err
	}

	var messages = make([]string, 0)
	for i, field := range fields {
		_, ok := m[field]
		if !ok {
			messages = append(messages, errMessages[i])
		}
	}

	if len(messages) != 0 {
		return errors.New(strings.Join(messages, ";\n"))
	}
	return nil
}
