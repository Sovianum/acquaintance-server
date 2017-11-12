package server

import (
	"github.com/Sovianum/acquaintance-server/model"
	"github.com/Sovianum/acquaintance-server/mylog"
	"github.com/go-errors/errors"
	"sync"
	"time"
)

const (
	userHasAlreadyAcceptedRequest = "user has already accepted request"
)

type requestMapType map[int]*model.MeetRequest

func NewMailBox(logger *mylog.Logger) MailBox {
	return &mailBox{
		logger:       logger,
		syncChan:     make(chan int, 1),
		accepted:     false,
		requestsLock: sync.RWMutex{},
		acceptedLock: sync.RWMutex{},
		requestMap:   make(requestMapType),
	}
}

type MailBox interface {
	AddAccept(request *model.MeetRequest) error
	AddDecline(request *model.MeetRequest)
	AddPending(request *model.MeetRequest)
	Remove(requestId int)
	GetAll(seconds int) []*model.MeetRequest
}

type mailBox struct {
	logger       *mylog.Logger
	syncChan     chan int
	accepted     bool
	acceptedLock sync.RWMutex
	requestMap   requestMapType
	requestsLock sync.RWMutex
}

func (box *mailBox) AddAccept(request *model.MeetRequest) error {
	box.acceptedLock.Lock()
	if box.accepted {
		return errors.New(userHasAlreadyAcceptedRequest)
	} else {
		box.accepted = true
	}
	box.acceptedLock.Unlock()

	box.requestsLock.Lock()
	var requestCopy = new(model.MeetRequest)
	requestCopy.Status = model.StatusAccepted
	*requestCopy = *request

	box.requestMap[request.Id] = requestCopy
	box.requestsLock.Unlock()

	select {
	case box.syncChan <- 1:
		box.logger.Infof("pushed to sync chan of box")
	default:
		box.logger.Infof("sync chan of box already full")
	}

	return nil
}

func (box *mailBox) AddDecline(request *model.MeetRequest) {
	box.addNonAccept(request, model.StatusDeclined)
}

func (box *mailBox) AddPending(request *model.MeetRequest) {
	box.addNonAccept(request, model.StatusPending)
}

func (box *mailBox) Remove(requestId int) {
	delete(box.requestMap, requestId)
}

func (box *mailBox) GetAll(seconds int) []*model.MeetRequest {
	var result = make([]*model.MeetRequest, 0)

	select {
	case <-box.syncChan:
		box.requestsLock.Lock()
		for _, request := range box.requestMap {
			result = append(result, request)
		}
		box.requestMap = make(requestMapType)
		box.requestsLock.Unlock()
	case <-time.After(time.Second * time.Duration(seconds)):
	}

	box.acceptedLock.Lock()
	box.accepted = false
	box.acceptedLock.Unlock()
	return result
}

func (box *mailBox) addNonAccept(request *model.MeetRequest, status string) {
	box.requestsLock.Lock()
	var requestCopy = new(model.MeetRequest)
	*requestCopy = *request
	requestCopy.Status = status

	box.requestMap[request.Id] = requestCopy
	box.requestsLock.Unlock()

	select {
	case box.syncChan <- 1:
	default:
	}
}
