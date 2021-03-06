package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sovianum/acquaintance-server/common"
	"github.com/Sovianum/acquaintance-server/dao"
	"github.com/Sovianum/acquaintance-server/model"
	"github.com/Sovianum/acquaintance-server/mylog"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"net/http"
	"strconv"
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
		common.WriteWithLogging(r, w, common.GetErrorJson(parseErr), env.logger)
		return
	}
	var userId, tokenCode, tokenErr = env.getIdFromRequest(r)
	if tokenErr != nil {
		env.logger.LogRequestError(r, tokenErr)
		w.WriteHeader(tokenCode)
		common.WriteWithLogging(r, w, common.GetErrorJson(tokenErr), env.logger)
		return
	}
	meetRequest.RequesterId = userId

	var requestId, dbErr = env.meetRequestDAO.CreateRequest(
		meetRequest.RequesterId, meetRequest.RequestedId, env.conf.Logic.RequestExpiration, env.conf.Logic.Distance,
	)
	if dbErr != nil {
		env.logger.LogRequestError(r, dbErr)
		w.WriteHeader(http.StatusInternalServerError)
		common.WriteWithLogging(r, w, common.GetErrorJson(dbErr), env.logger)
		return
	}
	if dao.IsInvalidId(requestId) {
		env.logger.LogRequestError(r, requestCreationError(requestId))
		w.WriteHeader(requestCreationCode(requestId))
		common.WriteWithLogging(r, w, common.GetErrorJson(requestCreationError(requestId)), env.logger)
		return
	}
	var code, err = env.handleRequestPending(requestId, userId)
	if err != nil {
		env.logger.LogRequestError(r, err)
		w.WriteHeader(code)
		common.WriteWithLogging(r, w, common.GetErrorJson(err), env.logger)
		return
	}

	env.logger.LogRequestSuccess(r)
	common.WriteWithLogging(r, w, common.GetEmptyJson(), env.logger)
}

func (env *Env) GetOutcomePendingRequests(w http.ResponseWriter, r *http.Request) {
	env.getRequestsTemplate(func(userId int, dao dao.MeetRequestDAO) ([]*model.MeetRequest, error) {
		return dao.GetOutcomePendingRequests(userId)
	}, w, r)
}

func (env *Env) GetIncomePendingRequests(w http.ResponseWriter, r *http.Request) {
	env.getRequestsTemplate(func(userId int, dao dao.MeetRequestDAO) ([]*model.MeetRequest, error) {
		return dao.GetIncomePendingRequests(userId)
	}, w, r)
}

func (env *Env) GetRequests(w http.ResponseWriter, r *http.Request) {
	env.getRequestsTemplate(func(userId int, dao dao.MeetRequestDAO) ([]*model.MeetRequest, error) {
		return dao.GetAllRequests(userId)
	}, w, r)
}

func (env *Env) UpdateRequest(w http.ResponseWriter, r *http.Request) {
	env.logger.LogRequestStart(r)
	var update, parseCode, parseErr = parseRequestUpdate(r)
	if parseErr != nil {
		env.logger.LogRequestError(r, parseErr)
		w.WriteHeader(parseCode)
		common.WriteWithLogging(r, w, common.GetErrorJson(parseErr), env.logger)
		return
	}

	var userId, tokenCode, tokenErr = env.getIdFromRequest(r)
	if tokenErr != nil {
		env.logger.LogRequestError(r, tokenErr)
		w.WriteHeader(tokenCode)
		common.WriteWithLogging(r, w, common.GetErrorJson(tokenErr), env.logger)
		return
	}

	var dbRequest, err = env.meetRequestDAO.GetRequestById(update.Id)
	if err != nil {
		env.logger.LogRequestError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
		common.WriteWithLogging(r, w, common.GetErrorJson(err), env.logger)
		return
	}

	var validStatusCombination = dbRequest.Status == model.StatusPending && update.Status != model.StatusInterrupted
	validStatusCombination = validStatusCombination || dbRequest.Status == model.StatusAccepted && update.Status == model.StatusInterrupted
	if !validStatusCombination {
		var err = fmt.Errorf(
			"trying to update request with status \"%s\" to status \"%s\"",
			dbRequest.Status,
			update.Status,
		)
		env.logger.LogRequestError(r, err)
		common.WriteWithLogging(r, w, common.GetErrorJson(err), env.logger)
		return
	}

	var rowsAffected, dbErr = env.meetRequestDAO.UpdateRequest(update.Id, userId, update.Status)
	if dbErr != nil {
		env.logger.LogRequestError(r, dbErr)
		env.rollBackCache(update.Id, userId)
		w.WriteHeader(http.StatusInternalServerError)
		common.WriteWithLogging(r, w, common.GetErrorJson(dbErr), env.logger)
		return
	}

	if rowsAffected == 0 {
		var err = errors.New(requestNotFound)
		env.logger.LogRequestError(r, err)
		env.rollBackCache(update.Id, userId)
		w.WriteHeader(http.StatusNotFound)
		common.WriteWithLogging(r, w, common.GetErrorJson(err), env.logger)
		return
	}

	var handler func(int, int) (int, error) = nil
	switch update.Status {
	case model.StatusAccepted:
		handler = env.handleRequestAccept
	case model.StatusDeclined:
		handler = env.handleRequestDecline
	case model.StatusInterrupted:
		handler = env.handleRequestInterrupt
	}

	if handler != nil {
		var code, err = handler(update.Id, userId)
		env.logger.Infof("finish request update to status %s", update.Status)
		if err != nil {
			env.logger.LogRequestError(r, err)
			w.WriteHeader(code)
			common.WriteWithLogging(r, w, common.GetErrorJson(err), env.logger)
			return
		}
	}

	dbRequest.Status = update.Status
	env.logger.LogRequestSuccess(r)
	common.WriteWithLogging(r, w, common.GetDataJson(dbRequest), env.logger)
}

func (env *Env) GetNewRequestsEvents(w http.ResponseWriter, r *http.Request) {
	env.logger.LogRequestStart(r)
	var userId, tokenCode, tokenErr = env.getIdFromRequest(r)
	if tokenErr != nil {
		env.logger.LogRequestError(r, tokenErr)
		w.WriteHeader(tokenCode)
		common.WriteWithLogging(r, w, common.GetErrorJson(tokenErr), env.logger)
		return
	}

	var box, boxErr = env.getMailBox(userId)
	if boxErr != nil {
		env.logger.LogRequestError(r, boxErr)
		w.WriteHeader(http.StatusInternalServerError)
		common.WriteWithLogging(r, w, common.GetErrorJson(boxErr), env.logger)
		return
	}

	var newRequestData = box.GetAll(env.conf.Logic.PollSeconds)

	env.logger.LogRequestSuccess(r)
	common.WriteWithLogging(r, w, common.GetDataJson(newRequestData), env.logger)
}

func (env *Env) getRequestsTemplate(
	daoFunc func(userId int, dao dao.MeetRequestDAO) ([]*model.MeetRequest, error),
	w http.ResponseWriter,
	r *http.Request,
) {
	env.logger.LogRequestStart(r)
	var userId, tokenCode, tokenErr = env.getIdFromRequest(r)
	if tokenErr != nil {
		env.logger.LogRequestError(r, tokenErr)
		w.WriteHeader(tokenCode)
		common.WriteWithLogging(r, w, common.GetErrorJson(tokenErr), env.logger)
		return
	}
	var requests, requestsErr = daoFunc(userId, env.meetRequestDAO)
	if requestsErr != nil {
		env.logger.LogRequestError(r, requestsErr)
		w.WriteHeader(http.StatusInternalServerError)
		common.WriteWithLogging(r, w, common.GetErrorJson(requestsErr), env.logger)
		return
	}

	env.logger.LogRequestSuccess(r)
	common.WriteWithLogging(r, w, common.GetDataJson(requests), env.logger)
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
		env.logger.Logger.Infof("add accepted request to mail box")
		if err := box.AddAccept(request); err != nil {
			return http.StatusUnavailableForLegalReasons, errors.New(alreadyAccepted)
		}
		return http.StatusOK, nil
	}
	var rightsCheckFunc = func(request *model.MeetRequest, userId int) bool {
		env.logger.Logger.Infof(
			"check accept request to add to mailbox: requested_id (%d) == userId (%d): %v",
			request.RequestedId,
			userId,
			request.RequestedId == userId,
		)
		return request.RequestedId == userId
	}
	var boxExtractFunc = func(userId int, request *model.MeetRequest) (MailBox, error) {
		// here we extract requester's mail box cos the one who initiated the requested should be
		// informed about request accept
		return env.getMailBox(request.RequesterId)
	}
	return env.dispatchRequest(boxFunc, boxExtractFunc, rightsCheckFunc, requestId, userId)
}

func (env *Env) handleRequestDecline(requestId int, userId int) (int, error) {
	var boxFunc = func(box MailBox, request *model.MeetRequest) (int, error) {
		env.logger.Logger.Infof("add declined request to mail box")
		box.AddDecline(request)
		return http.StatusOK, nil
	}
	var rightsCheckFunc = func(request *model.MeetRequest, userId int) bool {
		env.logger.Logger.Infof(
			"check decline request to add to mailbox: requested_id (%d) == userId (%d): %v",
			request.RequestedId,
			userId,
			request.RequestedId == userId,
		)
		return request.RequestedId == userId
	}
	var boxExtractFunc = func(userId int, request *model.MeetRequest) (MailBox, error) {
		// here we extract requester's mail box cos the one who initiated the requested should be
		// informed about request decline
		return env.getMailBox(request.RequesterId)
	}
	return env.dispatchRequest(boxFunc, boxExtractFunc, rightsCheckFunc, requestId, userId)
}

func (env *Env) handleRequestInterrupt(requestId int, userId int) (int, error) {
	var boxFunc = func(box MailBox, request *model.MeetRequest) (int, error) {
		env.logger.Logger.Infof("interrupt request in mail box")
		err := box.Interrupt(request)

		if err != nil {
			return http.StatusConflict, err
		}
		return http.StatusOK, nil
	}
	var rightsCheckFunc = func(request *model.MeetRequest, userId int) bool {
		var hasRights = request.RequesterId == userId || request.RequestedId == userId
		env.logger.Logger.Infof(
			"check interrupt request to add to mailbox: "+
				"userId (%d) is either requester_id (%d) or requested_id (%d): %v",
			userId,
			request.RequesterId,
			request.RequestedId,
			hasRights,
		)
		return hasRights
	}
	var boxExtractFunc = func(userId int, request *model.MeetRequest) (MailBox, error) {
		var address int
		if userId == request.RequesterId {
			env.logger.Infof("chosen requested with id = %d", request.RequestedId)
			address = request.RequestedId
		} else {
			env.logger.Infof("chosen requester with id = %d", request.RequesterId)
			address = request.RequesterId
		}
		// here we extract mail box of the one who didn't interrupt the request
		return env.getMailBox(address)
	}
	return env.dispatchRequest(boxFunc, boxExtractFunc, rightsCheckFunc, requestId, userId)
}

func (env *Env) handleRequestPending(requestId int, userId int) (int, error) {
	var boxFunc = func(box MailBox, request *model.MeetRequest) (int, error) {
		env.logger.Logger.Infof("add pending request to mail box")
		box.AddPending(request)
		return http.StatusOK, nil
	}
	var rightsCheckFunc = func(request *model.MeetRequest, userId int) bool {
		env.logger.Logger.Infof(
			"check pending request to add to mailbox: requester_id (%d) == userId (%d): %v",
			request.RequesterId,
			userId,
			request.RequesterId == userId,
		)
		return request.RequesterId == userId
	}
	var boxExtractFunc = func(userId int, request *model.MeetRequest) (MailBox, error) {
		// here we extract requested's mail box cos the one whom the request was addressed should be
		// informed about new request
		return env.getMailBox(request.RequestedId)
	}
	return env.dispatchRequest(boxFunc, boxExtractFunc, rightsCheckFunc, requestId, userId)
}

func (env *Env) dispatchRequest(
	boxFunc func(MailBox, *model.MeetRequest) (int, error),
	boxExtractFunc func(userId int, request *model.MeetRequest) (MailBox, error),
	rightsCheckFunc func(request *model.MeetRequest, userId int) bool,
	requestId int,
	userId int,
) (int, error) {
	env.logger.Logger.Infof("entered dispatchRequest")

	var request, requestErr = env.meetRequestDAO.GetRequestById(requestId)
	if requestErr != nil {
		env.logger.Errorf("failed to extract request with id %d in dispatcher", requestId)
		return http.StatusNotFound, requestErr
	}

	if !rightsCheckFunc(request, userId) {
		env.logger.Errorf("request check failed")
		return http.StatusNotFound, errors.New(requestNotFound)
	}

	var box, boxErr = boxExtractFunc(userId, request)
	if boxErr != nil {
		env.logger.Logger.Errorf(
			"mail box for request with %d => %d not found",
			request.RequesterId,
			request.RequesterId,
		)
		return http.StatusInternalServerError, boxErr
	}
	return boxFunc(box, request)
}

func (env *Env) getMailBox(id int) (MailBox, error) {
	var box, found = env.meetRequestCache.Get(strconv.Itoa(id))
	if !found {
		box = NewMailBox(env.logger)
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

func requestCreationError(code int) error {
	switch code {
	case dao.RequestExists:
		return fmt.Errorf("request already exists")
	case dao.UserInaccessible:
		return fmt.Errorf("user inaccessible")
	default:
		return fmt.Errorf("unknown error with code %d", code)
	}
}

func requestCreationCode(code int) int {
	switch code {
	case dao.RequestExists:
		return http.StatusConflict
	case dao.UserInaccessible:
		return http.StatusFailedDependency
	default:
		return http.StatusForbidden
	}
}
