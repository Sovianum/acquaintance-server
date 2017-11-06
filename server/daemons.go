package server

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

func (env *Env) RunDaemons() {
	go env.runDaemons()
}

func (env *Env) runDaemons() {
	for {
		select {
		case <-time.After(time.Duration(env.conf.Logic.CleanupInterval) * time.Minute):
			env.declineAll(env.conf.Logic.RequestExpiration)
		}
	}
}

func (env *Env) declineAll(timeoutMin int) error {
	if err := env.meetRequestDAO.DeclineAll(timeoutMin); err != nil {
		return err
	}

	var msgList = make([]string, 0)
	for userIdStr, item := range env.meetRequestCache.Items() {
		var userId, _ = strconv.Atoi(userIdStr)
		var box = item.Object.(MailBox)
		for _, meetRequest := range box.GetAll(60 * timeoutMin) {
			if _, err := env.handleRequestDecline(meetRequest.RequestedId, userId); err != nil {
				msgList = append(msgList, err.Error())
			}
		}
	}

	if len(msgList) != 0 {
		return errors.New(strings.Join(msgList, ","))
	}
	return nil
}
