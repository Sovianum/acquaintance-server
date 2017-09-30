package server

import (
	"net/http"
	"io/ioutil"
	"github.com/Sovianum/acquaintanceServer/model"
	"encoding/json"
	"errors"
	"github.com/Sovianum/acquaintanceServer/common"
	"strconv"
	"github.com/patrickmn/go-cache"
	"fmt"
)

const (
	requestNotFound = "request not found"
	alreadyAccepted = "user has already accepted another request"
)

func (env *Env) CreateRequest(w http.ResponseWriter, r *http.Request) {
	var meetRequest, parseCode, parseErr = parseRequest(r)
	if parseErr != nil {
		w.WriteHeader(parseCode)
		w.Write(common.GetErrorJson(parseErr))
		return
	}
	var userId, tokenCode, tokenErr = env.getIdFromRequest(r)
	if tokenErr != nil {
		w.WriteHeader(tokenCode)
		w.Write(common.GetErrorJson(tokenErr))
		return
	}
	if userId != meetRequest.RequesterId {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var rowsAffected, dbErr = env.meetRequestDAO.CreateRequest(
		meetRequest.RequesterId, meetRequest.RequestedId, env.conf.Logic.RequestExpiration, env.conf.Logic.Distance,
	)
	if dbErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(common.GetErrorJson(dbErr))
		return
	}
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusConflict)
		return
	}
}

func (env *Env) GetRequests(w http.ResponseWriter, r *http.Request) {
	var userId, tokenCode, tokenErr = env.getIdFromRequest(r)
	if tokenErr != nil {
		w.WriteHeader(tokenCode)
		w.Write(common.GetErrorJson(tokenErr))
		return
	}
	var requests, requestsErr = env.meetRequestDAO.GetRequests(userId)
	if requestsErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(common.GetErrorJson(requestsErr))
	}

	var msg, msgErr = json.Marshal(requests)
	if msgErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(common.GetErrorJson(msgErr))
	}
	w.Write(msg)
}

func (env *Env) UpdateRequest(w http.ResponseWriter, r *http.Request) {
	var update, parseCode, parseErr = parseRequestUpdate(r)
	if parseErr != nil {
		w.WriteHeader(parseCode)
		w.Write(common.GetErrorJson(parseErr))
		return
	}

	var userId, tokenCode, tokenErr = env.getIdFromRequest(r)
	if tokenErr != nil {
		w.WriteHeader(tokenCode)
		w.Write(common.GetErrorJson(tokenErr))
	}

	switch update.Status {
	case model.StatusAccepted:
		var code, err = env.handleRequestAccept(update.Id, userId)
		if err != nil {
			w.WriteHeader(code)
			w.Write(common.GetErrorJson(err))
			return
		}
	case model.StatusDeclined:
		var code, err = env.handleRequestDecline(update.Id, userId)
		if err != nil {
			w.WriteHeader(code)
			w.Write(common.GetErrorJson(err))
			return
		}
	}

	var rowsAffected, dbErr = env.meetRequestDAO.UpdateRequest(update.Id, userId, update.Status)
	if dbErr != nil {
		env.revertCache(update.Id, userId)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(common.GetErrorJson(dbErr))
	}

	if rowsAffected == 0 {
		env.revertCache(update.Id, userId)
		w.WriteHeader(http.StatusNotFound)
		w.Write(common.GetErrorJson(errors.New(requestNotFound)))
	}
}

func (env *Env) revertCache(requestId int, userId int) {
	var connect, found = env.meetRequestCache.Get(strconv.Itoa(userId))
	if !found {
		return
	}

	connect.(MeetConnection).Remove(requestId, userId)
}

func (env *Env) handleRequestAccept(requestId int, userId int) (int, error) {
	var f = func(connection MeetConnection, request *model.MeetRequest) (int, error) {
		if err := connection.AddAccept(request); err != nil {
			return http.StatusUnavailableForLegalReasons, errors.New(alreadyAccepted)
		}
		return http.StatusOK, nil
	}

	return env.handleRequestUpdate(f, requestId, userId)
}

func (env *Env) handleRequestDecline(requestId int, userId int) (int, error) {
	var f = func(connection MeetConnection, request *model.MeetRequest) (int, error) {
		connection.AddDecline(request)
		return http.StatusOK, nil
	}
	return env.handleRequestUpdate(f, requestId, userId)
}

func (env *Env) handleRequestUpdate(f func(MeetConnection, *model.MeetRequest) (int, error), requestId int, userId int) (int, error) {
	var request, requestErr = env.meetRequestDAO.GetRequestById(requestId)
	if requestErr != nil || request.Status != model.StatusPending {
		return http.StatusNotFound, requestErr
	}

	if request.RequestedId != userId {
		return http.StatusNotFound, errors.New(requestNotFound)
	}

	var connect, found = env.meetRequestCache.Get(strconv.Itoa(request.RequestedId))
	if !found {
		connect = NewMeetConnection()
		env.meetRequestCache.Set(strconv.Itoa(request.RequestedId), connect, cache.DefaultExpiration)
	}

	var casted, ok = connect.(MeetConnection)
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("failed to cast connect. Type: %T", connect)
	}

	return f(casted, request)
}

func parseRequestUpdate(r *http.Request) (*model.MeetRequestUpdate, int, error) {
	var body, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	if err := r.Body.Close(); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	var update = new(model.MeetRequestUpdate)
	if err := json.Unmarshal(body, &update); err != nil {
		return nil, http.StatusBadRequest, err
	}

	return update, http.StatusOK, nil
}

func parseRequest(r *http.Request) (*model.MeetRequest, int, error) {
	var body, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	if err := r.Body.Close(); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	var request = new(model.MeetRequest)
	request.RequestedId = -1
	request.RequesterId = -1

	if err := json.Unmarshal(body, &request); err != nil {
		return nil, http.StatusBadRequest, err
	}

	if request.RequesterId == -1 || request.RequestedId == -1 {
		return nil, http.StatusBadRequest, errors.New("Empty request")
	}

	return request, http.StatusOK, nil
}
