package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sovianum/acquaintance-server/common"
	"github.com/Sovianum/acquaintance-server/dao"
	"github.com/Sovianum/acquaintance-server/model"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"net/http"
	"strconv"
	"github.com/Sovianum/acquaintance-server/mylog"
)

const (
	requestNotFound = "request not found"
	alreadyAccepted = "user has already accepted another request"
)

func (env *Env) CreateRequest(w http.ResponseWriter, r *http.Request) {
	env.logger.LogRequestStart(r)
	var meetRequest, parseCode, parseErr = parseRequest(r, env.logger)
	if parseErr != nil {
		env.logger.LogRequestError(r, parseErr)
		w.WriteHeader(parseCode)
		w.Write(common.GetErrorJson(parseErr))
		return
	}
	var userId, tokenCode, tokenErr = env.getIdFromRequest(r)
	if tokenErr != nil {
		env.logger.LogRequestError(r, tokenErr)
		w.WriteHeader(tokenCode)
		w.Write(common.GetErrorJson(tokenErr))
		return
	}
	meetRequest.RequesterId = userId

	var requestId, dbErr = env.meetRequestDAO.CreateRequest(
		meetRequest.RequesterId, meetRequest.RequestedId, env.conf.Logic.RequestExpiration, env.conf.Logic.Distance,
	)
	if dbErr != nil {
		env.logger.LogRequestError(r, dbErr)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(common.GetErrorJson(dbErr))
		return
	}
	if dao.IsInvalidId(requestId) {
		env.logger.LogRequestError(r, dao.GetLogicalError(requestId))
		w.WriteHeader(http.StatusForbidden)
		w.Write(common.GetErrorJson(dao.GetLogicalError(requestId)))
		return
	}
	var code, err = env.handleRequestPending(requestId, userId)
	if err != nil {
		env.logger.LogRequestError(r, err)
		w.WriteHeader(code)
		w.Write(common.GetErrorJson(err))
		return
	}

	env.logger.LogRequestSuccess(r)
	w.Write(common.GetEmptyJson())
}

func (env *Env) GetRequests(w http.ResponseWriter, r *http.Request) {
	env.logger.LogRequestStart(r)
	var userId, tokenCode, tokenErr = env.getIdFromRequest(r)
	if tokenErr != nil {
		env.logger.LogRequestError(r, tokenErr)
		w.WriteHeader(tokenCode)
		w.Write(common.GetErrorJson(tokenErr))
		return
	}
	var requests, requestsErr = env.meetRequestDAO.GetRequests(userId)
	if requestsErr != nil {
		env.logger.LogRequestError(r, requestsErr)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(common.GetErrorJson(requestsErr))
		return
	}

	env.logger.LogRequestSuccess(r)
	w.Write(common.GetDataJson(requests))
}

func (env *Env) UpdateRequest(w http.ResponseWriter, r *http.Request) {
	env.logger.LogRequestStart(r)
	var update, parseCode, parseErr = parseRequestUpdate(r)
	if parseErr != nil {
		env.logger.LogRequestError(r, parseErr)
		w.WriteHeader(parseCode)
		w.Write(common.GetErrorJson(parseErr))
		return
	}

	var userId, tokenCode, tokenErr = env.getIdFromRequest(r)
	if tokenErr != nil {
		env.logger.LogRequestError(r, tokenErr)
		w.WriteHeader(tokenCode)
		w.Write(common.GetErrorJson(tokenErr))
		return
	}

	var rowsAffected, dbErr = env.meetRequestDAO.UpdateRequest(update.Id, userId, update.Status)
	if dbErr != nil {
		env.logger.LogRequestError(r, dbErr)
		env.rollBackCache(update.Id, userId)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(common.GetErrorJson(dbErr))
		return
	}

	if rowsAffected == 0 {
		var err = errors.New(requestNotFound)
		env.logger.LogRequestError(r, err)
		env.rollBackCache(update.Id, userId)
		w.WriteHeader(http.StatusNotFound)
		w.Write(common.GetErrorJson(err))
		return
	}

	switch update.Status {
	case model.StatusAccepted:
		var code, err = env.handleRequestAccept(update.Id, userId)
		if err != nil {
			env.logger.LogRequestError(r, err)
			w.WriteHeader(code)
			w.Write(common.GetErrorJson(err))
			return
		}
	case model.StatusDeclined:
		var code, err = env.handleRequestDecline(update.Id, userId)
		if err != nil {
			env.logger.LogRequestError(r, err)
			w.WriteHeader(code)
			w.Write(common.GetErrorJson(err))
			return
		}
	}

	env.logger.LogRequestSuccess(r)
	w.Write(common.GetEmptyJson())
}

func (env *Env) GetNewRequestsEvents(w http.ResponseWriter, r *http.Request) {
	env.logger.LogRequestStart(r)
	var userId, tokenCode, tokenErr = env.getIdFromRequest(r)
	if tokenErr != nil {
		env.logger.LogRequestError(r, tokenErr)
		w.WriteHeader(tokenCode)
		w.Write(common.GetErrorJson(tokenErr))
		return
	}

	var box, boxErr = env.getMailBox(userId)
	if boxErr != nil {
		env.logger.LogRequestError(r, boxErr)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(common.GetErrorJson(boxErr))
		return
	}

	var newRequestData = box.GetAll(env.conf.Logic.PollSeconds)

	env.logger.LogRequestSuccess(r)
	w.Write(common.GetDataJson(newRequestData))
}

func (env *Env) rollBackCache(requestId int, userId int) {
	var box, found = env.meetRequestCache.Get(strconv.Itoa(userId))
	if !found {
		return
	}

	box.(MailBox).Remove(requestId)
}

func (env *Env) handleRequestAccept(requestId int, userId int) (int, error) {
	var boxFunc = func(box MailBox, request *model.MeetRequest) (int, error) {
		if err := box.AddAccept(request); err != nil {
			return http.StatusUnavailableForLegalReasons, errors.New(alreadyAccepted)
		}
		return http.StatusOK, nil
	}
	var rightsCheckFunc = func(request *model.MeetRequest, userId int) bool { return request.RequestedId == userId }
	var boxExtractFunc = func(request *model.MeetRequest) (MailBox, error) {
		return env.getMailBox(request.RequesterId)
	}
	return env.dispatchRequest(boxFunc, boxExtractFunc, rightsCheckFunc, requestId, userId)
}

func (env *Env) handleRequestDecline(requestId int, userId int) (int, error) {
	var boxFunc = func(box MailBox, request *model.MeetRequest) (int, error) {
		box.AddDecline(request)
		return http.StatusOK, nil
	}
	var rightsCheckFunc = func(request *model.MeetRequest, userId int) bool { return request.RequestedId == userId }
	var boxExtractFunc = func(request *model.MeetRequest) (MailBox, error) {
		return env.getMailBox(request.RequesterId)
	}
	return env.dispatchRequest(boxFunc, boxExtractFunc, rightsCheckFunc, requestId, userId)
}

func (env *Env) handleRequestPending(requestId int, userId int) (int, error) {
	var boxFunc = func(box MailBox, request *model.MeetRequest) (int, error) {
		box.AddPending(request)
		return http.StatusOK, nil
	}
	var rightsCheckFunc = func(request *model.MeetRequest, userId int) bool {
		return request.RequesterId == userId
	}
	var boxExtractFunc = func(request *model.MeetRequest) (MailBox, error) {
		return env.getMailBox(request.RequestedId)
	}
	return env.dispatchRequest(boxFunc, boxExtractFunc, rightsCheckFunc, requestId, userId)
}

func (env *Env) dispatchRequest(
	boxFunc func(MailBox, *model.MeetRequest) (int, error),
	boxExtractFunc func(request *model.MeetRequest) (MailBox, error),
	rightsCheckFunc func(request *model.MeetRequest, userId int) bool,
	requestId int,
	userId int,
) (int, error) {
	var request, requestErr = env.meetRequestDAO.GetRequestById(requestId)
	if requestErr != nil {
		return http.StatusNotFound, requestErr
	}

	if !rightsCheckFunc(request, userId) {
		return http.StatusNotFound, errors.New(requestNotFound)
	}

	var box, boxErr = boxExtractFunc(request)
	if boxErr != nil {
		return http.StatusInternalServerError, boxErr
	}

	return boxFunc(box, request)
}

func (env *Env) getMailBox(id int) (MailBox, error) {
	var box, found = env.meetRequestCache.Get(strconv.Itoa(id))
	if !found {
		box = NewMailBox()
		env.meetRequestCache.Set(strconv.Itoa(id), box, cache.DefaultExpiration)
	}

	var casted, ok = box.(MailBox)
	if !ok {
		return nil, fmt.Errorf("failed to cast box. Type: %T", box)
	}
	return casted, nil
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

func parseRequest(r *http.Request, logger *mylog.Logger) (*model.MeetRequest, int, error) {
	var body, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	if err := r.Body.Close(); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	logger.LogRequestBody(r, string(body))

	var request = new(model.MeetRequest)

	if err := json.Unmarshal(body, &request); err != nil {
		return nil, http.StatusBadRequest, err
	}

	return request, http.StatusOK, nil
}
