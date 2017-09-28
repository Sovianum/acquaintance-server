package server

import (
	"encoding/json"
	"fmt"
	"github.com/Sovianum/acquaintanceServer/config"
	"github.com/Sovianum/acquaintanceServer/dao"
	"github.com/Sovianum/acquaintanceServer/model"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	distance      = 0.5
	onlineTimeout = 5
)

//getNeighbourUsers = `SELECT u2.id, u2.login, u2.age, u2.sex, u2.about
//						 FROM Users u1
//						 	JOIN Users u2 ON u2.id != u1.id
//						 	JOIN Position p1 ON u1.id = p1.id
//						 	JOIN Position p2 ON u2.id = p2.id
//						 WHERE u1.id = $1 AND ST_DistanceSphere(p1.geom, p2.geom) <= $2 AND age(current_timestamp, p2.time) < '$3 minutes'`

func TestEnv_UserGetNeighboursGet_Success(t *testing.T) {
	var db, mock, dbErr = sqlmock.New()

	if dbErr != nil {
		t.Fatal(dbErr)
	}
	defer db.Close()

	// mock user extraction
	mock.
		ExpectQuery("SELECT u2").
		WithArgs(1, distance, onlineTimeout).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "login", "age", "sex", "about"}).
				AddRow(1, "login1", 100, model.MALE, "about1").
				AddRow(2, "login2", 200, model.MALE, "about2"),
		)

	var env = getEnv(db)
	env.conf = getLogicConf()

	var tokenStr, _ = env.generateTokenString(1, "login")
	var rec, recErr = getRecorder(
		urlSample,
		http.MethodGet,
		env.UserGetNeighboursGet,
		strings.NewReader(""),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	fmt.Println(rec.Body.String())
}

func TestEnv_UserGetNeighboursGet_BadToken(t *testing.T) {
	var db, mock, dbErr = sqlmock.New()

	if dbErr != nil {
		t.Fatal(dbErr)
	}
	defer db.Close()

	// mock user extraction
	mock.
	ExpectQuery("SELECT u2").
		WithArgs(1, distance, onlineTimeout).
		WillReturnRows(
		sqlmock.NewRows([]string{"id", "login", "age", "sex", "about"}),
	)

	var env = getEnv(db)
	env.conf = getLogicConf()

	var tokenStr = []byte("invalid_token")
	var rec, recErr = getRecorder(
		urlSample,
		http.MethodGet,
		env.UserGetNeighboursGet,
		strings.NewReader(""),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
}

func TestEnv_UserGetNeighboursGet_DBErr(t *testing.T) {
	var db, mock, dbErr = sqlmock.New()

	if dbErr != nil {
		t.Fatal(dbErr)
	}
	defer db.Close()

	// mock user extraction
	mock.
	ExpectQuery("SELECT u2").
		WithArgs(1, distance, onlineTimeout).
		WillReturnError(errors.New("err"))


	var env = getEnv(db)
	env.conf = getLogicConf()

	var tokenStr, _ = env.generateTokenString(1, "login")
	var rec, recErr = getRecorder(
		urlSample,
		http.MethodGet,
		env.UserGetNeighboursGet,
		strings.NewReader(""),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusInternalServerError, rec.Code, rec.Body.String())
}

func TestEnv_UserSavePositionPost_Success(t *testing.T) {
	var db, mock, dbErr = sqlmock.New()

	if dbErr != nil {
		t.Fatal(dbErr)
	}
	defer db.Close()

	var date = time.Date(2003, 10, 17, 0, 0, 0, 0, time.UTC)
	var pos = &model.Position{
		UserId: 1,
		Point:  model.Point{X: 100, Y: 200},
		Time:   model.QuotedTime(date),
	}

	// mock position insertion
	mock.
		ExpectExec("INSERT INTO Position").
		WithArgs(pos.UserId, pos.Point.X, pos.Point.Y).
		WillReturnResult(sqlmock.NewResult(1, 1))

	var env = &Env{
		positionDAO: dao.NewDBPositionDAO(db),
		conf:    getAuthConf(),
	}

	var requestMsg, jsonErr = json.Marshal(pos)
	assert.Nil(t, jsonErr)

	var tokenStr, _ = env.generateTokenString(1, "login")
	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserSavePositionPost,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
}

func TestEnv_UserSavePositionPost_BadFormat(t *testing.T) {
	var env = &Env{
		conf: getAuthConf(),
	}

	var requestMsg = "invalid json"

	var tokenStr, _ = env.generateTokenString(1, "login")
	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserSavePositionPost,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEnv_UserSavePositionPost_Unauthorized(t *testing.T) {
	var date = time.Date(2003, 10, 17, 0, 0, 0, 0, time.UTC)
	var pos = &model.Position{
		UserId: 1,
		Point:  model.Point{X: 100, Y: 200},
		Time:   model.QuotedTime(date),
	}

	var env = &Env{
		conf: getAuthConf(),
	}

	var requestMsg, jsonErr = json.Marshal(pos)
	assert.Nil(t, jsonErr)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserSavePositionPost,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestEnv_UserSavePositionPost_BadToken(t *testing.T) {
	var date = time.Date(2003, 10, 17, 0, 0, 0, 0, time.UTC)
	var pos = &model.Position{
		UserId: 1,
		Point:  model.Point{X: 100, Y: 200},
		Time:   model.QuotedTime(date),
	}

	var env = &Env{
		conf: getAuthConf(),
	}

	var requestMsg, jsonErr = json.Marshal(pos)
	assert.Nil(t, jsonErr)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserSavePositionPost,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer some_strange_token")},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEnv_UserSavePositionPost_WrongId(t *testing.T) {
	var date = time.Date(2003, 10, 17, 0, 0, 0, 0, time.UTC)
	var pos = &model.Position{
		UserId: 1,
		Point:  model.Point{X: 100, Y: 200},
		Time:   model.QuotedTime(date),
	}

	var env = &Env{
		conf: getAuthConf(),
	}

	var requestMsg, jsonErr = json.Marshal(pos)
	assert.Nil(t, jsonErr)

	var tokenStr, _ = env.generateTokenString(10, "login")
	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserSavePositionPost,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestEnv_UserSavePositionPost_SaveErr(t *testing.T) {
	var db, mock, dbErr = sqlmock.New()

	if dbErr != nil {
		t.Fatal(dbErr)
	}
	defer db.Close()

	var date = time.Date(2003, 10, 17, 0, 0, 0, 0, time.UTC)
	var pos = &model.Position{
		UserId: 1,
		Point:  model.Point{X: 100, Y: 200},
		Time:   model.QuotedTime(date),
	}

	// mock position insertion
	mock.
		ExpectExec("INSERT INTO Position").
		WithArgs(pos.UserId, pos.Point.X, pos.Point.Y, pos.Time.String()).
		WillReturnError(errors.New("Save error"))

	var env = &Env{
		positionDAO: dao.NewDBPositionDAO(db),
		conf:    getAuthConf(),
	}

	var requestMsg, jsonErr = json.Marshal(pos)
	assert.Nil(t, jsonErr)

	var tokenStr, _ = env.generateTokenString(1, "login")
	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserSavePositionPost,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
		headerPair{authorizationStr, fmt.Sprintf("Bearer %s", tokenStr)},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestEnv_getIdFromTokenString_Success(t *testing.T) {
	var env = &Env{
		conf: getAuthConf(),
	}
	var tokenStr, _ = env.generateTokenString(1, "login")
	var token, _ = env.parseTokenString(tokenStr)
	var id, err = env.getIdFromTokenString(token)

	assert.Nil(t, err)
	assert.Equal(t, 1, id)
}

func TestEnv_parseTokenString_Success(t *testing.T) {
	var env = &Env{
		conf: getAuthConf(),
	}
	var tokenStr, _ = env.generateTokenString(1, "login")
	var _, err = env.parseTokenString(tokenStr)

	assert.Nil(t, err)
}

func TestEnv_parseTokenString_Fail(t *testing.T) {
	var env = &Env{
		conf: getAuthConf(),
	}
	var _, err = env.parseTokenString("Some_strange_str")

	assert.NotNil(t, err)
}

func TestRound(t *testing.T) {
	var testData = []struct {
		input    float64
		expected int
	}{
		{0.999999, 1},
		{1, 1},
		{1.000000001, 1},
		{0.4999999, 0},
	}

	for i, item := range testData {
		assert.Equal(t, item.expected, round(item.input), i)
	}
}

func getLogicConf() config.Conf {
	return config.Conf{
		Logic: config.LogicConfig{
			OnlineTimeout: onlineTimeout,
			Distance:      distance,
		},
	}
}
