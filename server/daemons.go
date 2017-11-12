package server

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

const (
	minToSec = 60
)

func (env *Env) RunDaemons() {
	go env.runDaemons()
}

func (env *Env) runDaemons() {
	for {
		select {
		case <-time.After(time.Duration(env.conf.Logic.CleanupInterval) * time.Minute):
			env.logger.Infof("%v", time.Duration(env.conf.Logic.CleanupInterval)*time.Minute)
			err := env.declineAll(env.conf.Logic.RequestExpiration)
			if err != nil {
				env.logger.Errorf("failed decline all with error: %s", err.Error())
			} else {
				env.logger.Infof("decline all succeeded")
			}
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
		for _, meetRequest := range box.GetAll(minToSec * timeoutMin) {
			var requestAge = time.Now().Sub((time.Time)(meetRequest.Time))

			if requestAge.Minutes() > float64(timeoutMin) {
				if _, err := env.handleRequestDecline(meetRequest.RequestedId, userId); err != nil {
					msgList = append(msgList, err.Error())
				}
			}
		}
	}

	if len(msgList) != 0 {
		return errors.New(strings.Join(msgList, ","))
	}
	return nil
}
