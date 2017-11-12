package server

import (
	"encoding/json"
	"fmt"
	"github.com/Sovianum/acquaintance-server/config"
	"github.com/Sovianum/acquaintance-server/model"
	"github.com/Sovianum/acquaintance-server/mylog"
	"github.com/Sovianum/acquaintance-server/server/mocks"
	"github.com/dgrijalva/jwt-go"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	defaultExpiration = 5
	defaultCleanup    = 10
)

func TestEnv_CreateRequest_Success(t *testing.T) {
	var meetRequest = model.MeetRequest{RequesterId: mocks.RequesterId, RequestedId: mocks.RequestedId}
	var requestMsg, jsonErr = json.Marshal(meetRequest)
	assert.Nil(t, jsonErr)

	var env = &Env{
		conf:             getTotalConf(),
		meetRequestDAO:   &mocks.MeetRequestDAOMockSuccess{},
		meetRequestCache: cache.New(time.Second*defaultExpiration, time.Second*defaultCleanup),
		logger:           mylog.NewLogger(ioutil.Discard),
	}
	var tokenStr, _ = env.generateTokenString(mocks.RequesterId, "login")

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.CreateRequest,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestEnv_CreateRequest_NoIdInToken(t *testing.T) {
	var env = &Env{
		conf:           getTotalConf(),
		meetRequestDAO: &mocks.MeetRequestDAOMockSuccess{},
		logger:         mylog.NewLogger(ioutil.Discard),
	}
	var tokenStr, _ = getIncompleteToken(env)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.CreateRequest,
		strings.NewReader(""),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEnv_CreateRequest_BadToken(t *testing.T) {
	var env = &Env{conf: getTotalConf(), meetRequestDAO: &mocks.MeetRequestDAOMockSuccess{}, logger: mylog.NewLogger(ioutil.Discard)}
	var tokenStr = "Bad token"

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.CreateRequest,
		strings.NewReader(""),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEnv_CreateRequest_Conflict(t *testing.T) {
	var meetRequest = model.MeetRequest{RequesterId: 1, RequestedId: 2}
	var requestMsg, jsonErr = json.Marshal(meetRequest)
	assert.Nil(t, jsonErr)

	var env = &Env{
		conf:           getTotalConf(),
		meetRequestDAO: &mocks.MeetRequestDAOMockCreateConflict{},
		logger:         mylog.NewLogger(ioutil.Discard),
	}

	var tokenStr, _ = env.generateTokenString(1, "login")

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.CreateRequest,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestEnv_CreateRequest_Error(t *testing.T) {
	var meetRequest = model.MeetRequest{RequesterId: 1, RequestedId: 2}
	var requestMsg, jsonErr = json.Marshal(meetRequest)
	assert.Nil(t, jsonErr)

	var env = &Env{
		conf:           getTotalConf(),
		meetRequestDAO: &mocks.MeetRequestDAOMockCreateError{},
		logger:         mylog.NewLogger(ioutil.Discard),
	}

	var tokenStr, _ = env.generateTokenString(1, "login")

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.CreateRequest,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestEnv_GetRequests_Success(t *testing.T) {
	var env = &Env{conf: getTotalConf(), meetRequestDAO: &mocks.MeetRequestDAOMockSuccess{}, logger: mylog.NewLogger(ioutil.Discard)}
	var tokenStr, _ = env.generateTokenString(1, "login")

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodGet,
		env.GetRequests,
		strings.NewReader(""),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusOK, rec.Code)

	var requests, _ = env.meetRequestDAO.GetAllRequests(1)
	var gotRequests = make(map[string][]model.MeetRequest)
	var jsonErr = json.Unmarshal(rec.Body.Bytes(), &gotRequests)

	assert.Nil(t, jsonErr)
	assert.Equal(t, len(requests), len(gotRequests["data"]))
	for i := range requests {
		assert.Equal(t, *requests[i], gotRequests["data"][i])
	}
}

func TestEnv_GetRequests_Empty(t *testing.T) {
	var env = &Env{conf: getTotalConf(), meetRequestDAO: &mocks.MeetRequestDAOMockGetRequestsEmpty{}, logger: mylog.NewLogger(ioutil.Discard)}
	var tokenStr, _ = env.generateTokenString(1, "login")

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodGet,
		env.GetRequests,
		strings.NewReader(""),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusOK, rec.Code)

	var requests, _ = env.meetRequestDAO.GetAllRequests(1)
	var gotRequests = make(map[string][]model.MeetRequest)
	var jsonErr = json.Unmarshal(rec.Body.Bytes(), &gotRequests)

	assert.Nil(t, jsonErr)
	assert.Equal(t, len(requests), len(gotRequests["data"]))
	for i := range requests {
		assert.Equal(t, *requests[i], gotRequests["data"][i])
	}
}

func TestEnv_GetRequests_NoIdInToken(t *testing.T) {
	var env = &Env{conf: getTotalConf(), meetRequestDAO: &mocks.MeetRequestDAOMockSuccess{}, logger: mylog.NewLogger(ioutil.Discard)}
	var tokenStr, _ = getIncompleteToken(env)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodGet,
		env.GetRequests,
		strings.NewReader(""),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEnv_GetRequests_BadToken(t *testing.T) {
	var env = &Env{conf: getTotalConf(), meetRequestDAO: &mocks.MeetRequestDAOMockSuccess{}, logger: mylog.NewLogger(ioutil.Discard)}
	var tokenStr = "Bad token"

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodGet,
		env.GetRequests,
		strings.NewReader(""),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEnv_GetRequests_Error(t *testing.T) {
	var env = &Env{conf: getTotalConf(), meetRequestDAO: &mocks.MeetRequestDAOMockGetRequestsError{}, logger: mylog.NewLogger(ioutil.Discard)}
	var tokenStr, _ = env.generateTokenString(1, "login")

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodGet,
		env.GetRequests,
		strings.NewReader(""),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestEnv_UpdateRequest_NoIdInToken(t *testing.T) {
	var env = &Env{conf: getTotalConf(), meetRequestDAO: &mocks.MeetRequestDAOMockSuccess{}, logger: mylog.NewLogger(ioutil.Discard)}
	var tokenStr, _ = getIncompleteToken(env)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UpdateRequest,
		strings.NewReader(""),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEnv_UpdateRequest_BadToken(t *testing.T) {
	var env = &Env{conf: getTotalConf(), meetRequestDAO: &mocks.MeetRequestDAOMockSuccess{}, logger: mylog.NewLogger(ioutil.Discard)}
	var tokenStr = "bad string"

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UpdateRequest,
		strings.NewReader(""),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEnv_UpdateRequest_NoRequest(t *testing.T) {
	var env = &Env{
		conf:             getTotalConf(),
		meetRequestDAO:   &mocks.MeetRequestDAOMockUpdateNoRequest{},
		meetRequestCache: cache.New(time.Second*defaultExpiration, time.Second*defaultCleanup),
		logger:           mylog.NewLogger(ioutil.Discard),
	}
	var tokenStr, _ = env.generateTokenString(1, "login")

	var update = model.MeetRequestUpdate{Id: 1, Status: model.StatusAccepted}
	var requestMsg, err = json.Marshal(update)
	assert.Nil(t, err)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UpdateRequest,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestEnv_UpdateRequest_BadStatus(t *testing.T) {
	var env = &Env{conf: getTotalConf(), meetRequestDAO: &mocks.MeetRequestDAOMockUpdateNoRequest{}, logger: mylog.NewLogger(ioutil.Discard)}
	var tokenStr, _ = env.generateTokenString(1, "login")

	var update = model.MeetRequestUpdate{Id: 1, Status: "BAD"}
	var requestMsg, err = json.Marshal(update)
	assert.Nil(t, err)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UpdateRequest,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEnv_UpdateRequest_AcceptSuccess(t *testing.T) {
	var env = &Env{
		conf:             getTotalConf(),
		meetRequestDAO:   &mocks.MeetRequestDAOMockSuccess{},
		meetRequestCache: cache.New(time.Second*defaultExpiration, time.Second*defaultCleanup),
		logger:           mylog.NewLogger(ioutil.Discard),
	}
	var tokenStr, _ = env.generateTokenString(mocks.RequestedId, "login")

	var update = model.MeetRequestUpdate{Id: mocks.RequestedId, Status: model.StatusAccepted}
	var requestMsg, err = json.Marshal(update)
	assert.Nil(t, err)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UpdateRequest,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusOK, rec.Code)
}

//func TestEnv_UpdateRequest_AcceptNotFound(t *testing.T) {
//	var env = &Env{
//		conf:             getTotalConf(),
//		meetRequestDAO:   &mocks.MeetRequestDAOMockGetRequestByIdNotFound{},
//		meetRequestCache: cache.New(time.Second*defaultExpiration, time.Second*defaultCleanup),
//		logger:           mylog.NewLogger(ioutil.Discard),
//	}
//	var tokenStr, _ = env.generateTokenString(mocks.RequestedId, "login")
//
//	var update = model.MeetRequestUpdate{Id: mocks.RequestedId, Status: model.StatusAccepted}
//	var requestMsg, err = json.Marshal(update)
//	assert.Nil(t, err)
//
//	var rec, recErr = getRecorder(
//		urlSample,
//		http.MethodPost,
//		env.UpdateRequest,
//		strings.NewReader(string(requestMsg)),
//		headerPair{"Content-Type", "application/json"},
//		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
//	)
//
//	assert.Nil(t, recErr)
//	assert.Equal(t, http.StatusNotFound, rec.Code)
//}

func TestEnv_UpdateRequest_AcceptLocked(t *testing.T) {
	var env = &Env{
		conf:             getTotalConf(),
		meetRequestDAO:   &mocks.MeetRequestDAOMockSuccess{},
		meetRequestCache: cache.New(time.Second*defaultExpiration, time.Second*defaultCleanup),
		logger:           mylog.NewLogger(ioutil.Discard),
	}
	var request, _ = env.meetRequestDAO.GetRequestById(1)
	env.handleRequestAccept(request.Id, request.RequestedId)

	var tokenStr, _ = env.generateTokenString(request.RequestedId, "login")

	var update = model.MeetRequestUpdate{Id: request.RequestedId, Status: model.StatusAccepted}
	var requestMsg, err = json.Marshal(update)
	assert.Nil(t, err)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UpdateRequest,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusUnavailableForLegalReasons, rec.Code)
}

func TestEnv_UpdateRequest_DeclineSuccess(t *testing.T) {
	var env = &Env{
		conf:             getTotalConf(),
		meetRequestDAO:   &mocks.MeetRequestDAOMockSuccess{},
		meetRequestCache: cache.New(time.Second*defaultExpiration, time.Second*defaultCleanup),
		logger:           mylog.NewLogger(ioutil.Discard),
	}
	var tokenStr, _ = env.generateTokenString(mocks.RequestedId, "login")

	var update = model.MeetRequestUpdate{Id: mocks.RequestedId, Status: model.StatusDeclined}
	var requestMsg, err = json.Marshal(update)
	assert.Nil(t, err)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UpdateRequest,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestEnv_UpdateRequest_Error(t *testing.T) {
	var env = &Env{
		conf:             getTotalConf(),
		meetRequestDAO:   &mocks.MeetRequestDAOMockUpdateError{},
		meetRequestCache: cache.New(time.Second*defaultExpiration, time.Second*defaultCleanup),
		logger:           mylog.NewLogger(ioutil.Discard),
	}
	var tokenStr, _ = env.generateTokenString(mocks.RequestedId, "login")

	var update = model.MeetRequestUpdate{Id: 1, Status: model.StatusAccepted}
	var requestMsg, err = json.Marshal(update)
	assert.Nil(t, err)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UpdateRequest,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestEnv_GetNewRequests_Success(t *testing.T) {
	var env = &Env{
		conf:             getTotalConf(),
		meetRequestDAO:   &mocks.MeetRequestDAOMockSuccess{},
		meetRequestCache: cache.New(time.Second*defaultExpiration, time.Second*defaultCleanup),
		logger:           mylog.NewLogger(ioutil.Discard),
	}
	var tokenStr, _ = env.generateTokenString(mocks.RequestedId, "login")

	env.handleRequestPending(10, mocks.RequesterId)
	env.handleRequestPending(20, mocks.RequesterId)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodGet,
		env.GetNewRequestsEvents,
		strings.NewReader(""),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusOK, rec.Code)

	var gotRequests = make(map[string][]*model.MeetRequest)
	var jsonErr = json.Unmarshal(rec.Body.Bytes(), &gotRequests)

	assert.Nil(t, jsonErr)
	assert.Equal(t, 2, len(gotRequests["data"]))
	for _, request := range gotRequests["data"] {
		request.Status = model.StatusPending
	}
}

func TestEnv_GetNewRequests_NoIdInToken(t *testing.T) {
	var env = &Env{
		conf:             getTotalConf(),
		meetRequestDAO:   &mocks.MeetRequestDAOMockSuccess{},
		meetRequestCache: cache.New(time.Second*defaultExpiration, time.Second*defaultCleanup),
		logger:           mylog.NewLogger(ioutil.Discard),
	}
	var tokenStr, _ = getIncompleteToken(env)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodGet,
		env.GetNewRequestsEvents,
		strings.NewReader(""),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEnv_GetNewRequests_BadToken(t *testing.T) {
	var env = &Env{
		conf:             getTotalConf(),
		meetRequestDAO:   &mocks.MeetRequestDAOMockSuccess{},
		meetRequestCache: cache.New(time.Second*defaultExpiration, time.Second*defaultCleanup),
		logger:           mylog.NewLogger(ioutil.Discard),
	}
	var tokenStr = "Bad token"

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodGet,
		env.GetNewRequestsEvents,
		strings.NewReader(""),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEnv_GetNewRequests_Empty(t *testing.T) {
	var env = &Env{
		conf:             getTotalConf(),
		meetRequestDAO:   &mocks.MeetRequestDAOMockSuccess{},
		meetRequestCache: cache.New(time.Second*defaultExpiration, time.Second*defaultCleanup),
		logger:           mylog.NewLogger(ioutil.Discard),
	}
	var tokenStr, _ = env.generateTokenString(mocks.RequestedId, "login")

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodGet,
		env.GetNewRequestsEvents,
		strings.NewReader(""),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusOK, rec.Code)

	var gotRequests = make(map[string][]*model.MeetRequest)
	var jsonErr = json.Unmarshal(rec.Body.Bytes(), &gotRequests)

	assert.Nil(t, jsonErr)
	assert.Equal(t, 0, len(gotRequests["data"]))
}

func getIncompleteToken(env *Env) (string, error) {
	var token = jwt.New(jwt.SigningMethodHS256)
	var claims = token.Claims.(jwt.MapClaims)

	claims[loginStr] = "login"
	claims[expStr] = time.Now().Add(time.Hour * 24 * time.Duration(env.conf.Auth.ExpireDays)).Unix()

	var tokenKey = env.conf.Auth.GetTokenKey()
	return token.SignedString(tokenKey)
}

func getTotalConf() config.Conf {
	return config.Conf{
		Auth: config.AuthConfig{
			ExpireDays: expireDays,
			TokenKey:   tokenKey,
		},
		Logic: config.LogicConfig{
			OnlineTimeout: onlineTimeout,
			Distance:      distance,
		},
	}
}
