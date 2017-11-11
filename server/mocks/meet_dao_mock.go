package mocks

import (
	"errors"
	"github.com/Sovianum/acquaintance-server/dao"
	"github.com/Sovianum/acquaintance-server/model"
	"time"
)

const (
	createErr   = "create error"
	getError    = "get error"
	acceptError = "accept error"

	RequesterId = 2
	RequestedId = 3
)

type createRequestFuncType func(requesterId int, requestedId int, requestTimeoutMin int, maxDistance float64) (int, error)
type getRequestsFuncType func(requestedId int) ([]*model.MeetRequest, error)
type updateRequestFuncType func(id int, requestedId int, status string) (int, error)
type getPendingRequestByIdFuncType func(id int) (*model.MeetRequest, error)

var createRequestSuccess createRequestFuncType = func(int, int, int, float64) (int, error) {
	return 1, nil
}
var createRequestConflict createRequestFuncType = func(int, int, int, float64) (int, error) {
	return dao.ImpossibleID, nil
}
var createRequestError createRequestFuncType = func(int, int, int, float64) (int, error) {
	return 0, errors.New(createErr)
}

var getRequestsSuccess getRequestsFuncType = func(requestedId int) ([]*model.MeetRequest, error) {
	return []*model.MeetRequest{
		{Id: 0, RequestedId: requestedId, RequesterId: 2, Status: model.StatusPending},
		{Id: 1, RequestedId: requestedId, RequesterId: 3, Status: model.StatusPending},
	}, nil
}
var getRequestsEmpty getRequestsFuncType = func(int) ([]*model.MeetRequest, error) {
	return []*model.MeetRequest{}, nil
}
var getRequestsError getRequestsFuncType = func(int) ([]*model.MeetRequest, error) {
	return nil, errors.New(getError)
}

var updateRequestSuccess updateRequestFuncType = func(int, int, string) (int, error) {
	return 1, nil
}
var updateRequestNoRequest updateRequestFuncType = func(int, int, string) (int, error) {
	return 0, nil
}
var updateRequestError updateRequestFuncType = func(int, int, string) (int, error) {
	return 0, errors.New(acceptError)
}

var getPendingRequestByIdSuccess getPendingRequestByIdFuncType = func(id int) (*model.MeetRequest, error) {
	return &model.MeetRequest{
		Time:        model.QuotedTime(time.Now()),
		Status:      model.StatusPending,
		RequestedId: RequestedId,
		RequesterId: RequesterId,
		Id:          id,
	}, nil
}
var getPendingRequestByIdNotFound getPendingRequestByIdFuncType = func(id int) (*model.MeetRequest, error) {
	return nil, errors.New("not found")
}

type MeetRequestDAOMockSuccess struct{}

func (*MeetRequestDAOMockSuccess) GetIncomePendingRequests(requestedId int) ([]*model.MeetRequest, error) {
	panic("implement me")
}

func (*MeetRequestDAOMockSuccess) GetOutcomePendingRequests(requesterId int) ([]*model.MeetRequest, error) {
	panic("implement me")
}

func (*MeetRequestDAOMockSuccess) CreateRequest(requesterId int, requestedId int, requestTimeoutMin int, maxDistance float64) (code int, dbErr error) {
	return createRequestSuccess(requesterId, requestedId, requestTimeoutMin, maxDistance)
}

func (*MeetRequestDAOMockSuccess) GetAllRequests(requestedId int) ([]*model.MeetRequest, error) {
	return getRequestsSuccess(requestedId)
}

func (*MeetRequestDAOMockSuccess) UpdateRequest(id int, requestedId int, status string) (int, error) {
	return updateRequestSuccess(id, requestedId, status)
}

func (*MeetRequestDAOMockSuccess) GetRequestById(id int) (*model.MeetRequest, error) {
	return getPendingRequestByIdSuccess(id)
}

func (*MeetRequestDAOMockSuccess) DeclineAll(timeoutMin int) error { return nil }

type MeetRequestDAOMockCreateConflict struct{}

func (*MeetRequestDAOMockCreateConflict) GetIncomePendingRequests(requestedId int) ([]*model.MeetRequest, error) {
	panic("implement me")
}

func (*MeetRequestDAOMockCreateConflict) GetOutcomePendingRequests(requesterId int) ([]*model.MeetRequest, error) {
	panic("implement me")
}

func (*MeetRequestDAOMockCreateConflict) CreateRequest(requesterId int, requestedId int, requestTimeoutMin int, maxDistance float64) (code int, dbErr error) {
	return createRequestConflict(requesterId, requestedId, requestTimeoutMin, maxDistance)
}

func (*MeetRequestDAOMockCreateConflict) GetAllRequests(requestedId int) ([]*model.MeetRequest, error) {
	return getRequestsSuccess(requestedId)
}

func (*MeetRequestDAOMockCreateConflict) UpdateRequest(id int, requestedId int, status string) (int, error) {
	return updateRequestSuccess(id, requestedId, status)
}

func (*MeetRequestDAOMockCreateConflict) GetRequestById(id int) (*model.MeetRequest, error) {
	return getPendingRequestByIdSuccess(id)
}

func (*MeetRequestDAOMockCreateConflict) DeclineAll(timeoutMin int) error { return nil }

type MeetRequestDAOMockCreateError struct{}

func (*MeetRequestDAOMockCreateError) GetIncomePendingRequests(requestedId int) ([]*model.MeetRequest, error) {
	panic("implement me")
}

func (*MeetRequestDAOMockCreateError) GetOutcomePendingRequests(requesterId int) ([]*model.MeetRequest, error) {
	panic("implement me")
}

func (*MeetRequestDAOMockCreateError) CreateRequest(requesterId int, requestedId int, requestTimeoutMin int, maxDistance float64) (code int, dbErr error) {
	return createRequestError(requesterId, requestedId, requestTimeoutMin, maxDistance)
}

func (*MeetRequestDAOMockCreateError) GetAllRequests(requestedId int) ([]*model.MeetRequest, error) {
	return getRequestsSuccess(requestedId)
}

func (*MeetRequestDAOMockCreateError) UpdateRequest(id int, requestedId int, status string) (int, error) {
	return updateRequestSuccess(id, requestedId, status)
}

func (*MeetRequestDAOMockCreateError) GetRequestById(id int) (*model.MeetRequest, error) {
	return getPendingRequestByIdSuccess(id)
}

func (*MeetRequestDAOMockCreateError) DeclineAll(timeoutMin int) error { return nil }

type MeetRequestDAOMockGetRequestsEmpty struct{}

func (*MeetRequestDAOMockGetRequestsEmpty) GetIncomePendingRequests(requestedId int) ([]*model.MeetRequest, error) {
	panic("implement me")
}

func (*MeetRequestDAOMockGetRequestsEmpty) GetOutcomePendingRequests(requesterId int) ([]*model.MeetRequest, error) {
	panic("implement me")
}

func (*MeetRequestDAOMockGetRequestsEmpty) CreateRequest(requesterId int, requestedId int, requestTimeoutMin int, maxDistance float64) (code int, dbErr error) {
	return createRequestSuccess(requesterId, requestedId, requestTimeoutMin, maxDistance)
}

func (*MeetRequestDAOMockGetRequestsEmpty) GetAllRequests(requestedId int) ([]*model.MeetRequest, error) {
	return getRequestsEmpty(requestedId)
}

func (*MeetRequestDAOMockGetRequestsEmpty) UpdateRequest(id int, requestedId int, status string) (int, error) {
	return updateRequestSuccess(id, requestedId, status)
}

func (*MeetRequestDAOMockGetRequestsEmpty) GetRequestById(id int) (*model.MeetRequest, error) {
	return getPendingRequestByIdSuccess(id)
}

func (*MeetRequestDAOMockGetRequestsEmpty) DeclineAll(timeoutMin int) error { return nil }

type MeetRequestDAOMockGetRequestsError struct{}

func (*MeetRequestDAOMockGetRequestsError) GetIncomePendingRequests(requestedId int) ([]*model.MeetRequest, error) {
	panic("implement me")
}

func (*MeetRequestDAOMockGetRequestsError) GetOutcomePendingRequests(requesterId int) ([]*model.MeetRequest, error) {
	panic("implement me")
}

func (*MeetRequestDAOMockGetRequestsError) CreateRequest(requesterId int, requestedId int, requestTimeoutMin int, maxDistance float64) (code int, dbErr error) {
	return createRequestSuccess(requesterId, requestedId, requestTimeoutMin, maxDistance)
}

func (*MeetRequestDAOMockGetRequestsError) GetAllRequests(requestedId int) ([]*model.MeetRequest, error) {
	return getRequestsError(requestedId)
}

func (*MeetRequestDAOMockGetRequestsError) UpdateRequest(id int, requestedId int, status string) (int, error) {
	return updateRequestSuccess(id, requestedId, status)
}

func (*MeetRequestDAOMockGetRequestsError) GetRequestById(id int) (*model.MeetRequest, error) {
	return getPendingRequestByIdSuccess(id)
}

func (*MeetRequestDAOMockGetRequestsError) DeclineAll(timeoutMin int) error { return nil }

type MeetRequestDAOMockUpdateNoRequest struct{}

func (*MeetRequestDAOMockUpdateNoRequest) GetIncomePendingRequests(requestedId int) ([]*model.MeetRequest, error) {
	panic("implement me")
}

func (*MeetRequestDAOMockUpdateNoRequest) GetOutcomePendingRequests(requesterId int) ([]*model.MeetRequest, error) {
	panic("implement me")
}

func (*MeetRequestDAOMockUpdateNoRequest) CreateRequest(requesterId int, requestedId int, requestTimeoutMin int, maxDistance float64) (code int, dbErr error) {
	return createRequestSuccess(requesterId, requestedId, requestTimeoutMin, maxDistance)
}

func (*MeetRequestDAOMockUpdateNoRequest) GetAllRequests(requestedId int) ([]*model.MeetRequest, error) {
	return getRequestsSuccess(requestedId)
}

func (*MeetRequestDAOMockUpdateNoRequest) UpdateRequest(id int, requestedId int, status string) (int, error) {
	return updateRequestNoRequest(id, requestedId, status)
}

func (*MeetRequestDAOMockUpdateNoRequest) GetRequestById(id int) (*model.MeetRequest, error) {
	return getPendingRequestByIdSuccess(id)
}

func (*MeetRequestDAOMockUpdateNoRequest) DeclineAll(timeoutMin int) error { return nil }

type MeetRequestDAOMockUpdateError struct{}

func (*MeetRequestDAOMockUpdateError) GetIncomePendingRequests(requestedId int) ([]*model.MeetRequest, error) {
	panic("implement me")
}

func (*MeetRequestDAOMockUpdateError) GetOutcomePendingRequests(requesterId int) ([]*model.MeetRequest, error) {
	panic("implement me")
}

func (*MeetRequestDAOMockUpdateError) CreateRequest(requesterId int, requestedId int, requestTimeoutMin int, maxDistance float64) (code int, dbErr error) {
	return createRequestSuccess(requesterId, requestedId, requestTimeoutMin, maxDistance)
}

func (*MeetRequestDAOMockUpdateError) GetAllRequests(requestedId int) ([]*model.MeetRequest, error) {
	return getRequestsSuccess(requestedId)
}

func (*MeetRequestDAOMockUpdateError) UpdateRequest(id int, requestedId int, status string) (int, error) {
	return updateRequestError(id, requestedId, status)
}

func (*MeetRequestDAOMockUpdateError) GetRequestById(id int) (*model.MeetRequest, error) {
	return getPendingRequestByIdSuccess(id)
}

func (*MeetRequestDAOMockUpdateError) DeclineAll(timeoutMin int) error { return nil }

type MeetRequestDAOMockGetRequestByIdNotFound struct{}

func (*MeetRequestDAOMockGetRequestByIdNotFound) GetIncomePendingRequests(requestedId int) ([]*model.MeetRequest, error) {
	panic("implement me")
}

func (*MeetRequestDAOMockGetRequestByIdNotFound) GetOutcomePendingRequests(requesterId int) ([]*model.MeetRequest, error) {
	panic("implement me")
}

func (*MeetRequestDAOMockGetRequestByIdNotFound) CreateRequest(requesterId int, requestedId int, requestTimeoutMin int, maxDistance float64) (code int, dbErr error) {
	return createRequestSuccess(requesterId, requestedId, requestTimeoutMin, maxDistance)
}

func (*MeetRequestDAOMockGetRequestByIdNotFound) GetAllRequests(requestedId int) ([]*model.MeetRequest, error) {
	return getRequestsSuccess(requestedId)
}

func (*MeetRequestDAOMockGetRequestByIdNotFound) UpdateRequest(id int, requestedId int, status string) (int, error) {
	return updateRequestSuccess(id, requestedId, status)
}

func (*MeetRequestDAOMockGetRequestByIdNotFound) GetRequestById(id int) (*model.MeetRequest, error) {
	return getPendingRequestByIdNotFound(id)
}

func (*MeetRequestDAOMockGetRequestByIdNotFound) DeclineAll(timeoutMin int) error { return nil }
