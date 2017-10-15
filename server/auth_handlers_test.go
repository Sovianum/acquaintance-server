package server

import (
	"encoding/json"
	"github.com/Sovianum/acquaintance-server/config"
	"github.com/Sovianum/acquaintance-server/dao"
	"github.com/Sovianum/acquaintance-server/model"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"errors"
	"crypto/sha256"
	"fmt"
	"database/sql"
	"github.com/Sovianum/acquaintance-server/mylog"
	"io/ioutil"
)

const (
	urlSample = "URL"
	expireDays = 100
	tokenKey = "token90"
)

type headerPair struct {
	key   string
	value string
}

func TestEnv_UserRegisterPost_Success(t *testing.T) {
	var db, mock, dbErr = sqlmock.New()

	if dbErr != nil {
		t.Fatal(dbErr)
	}
	defer db.Close()

	var user = &model.User{
		Login:    "login",
		Password: "password",
		About:    "about",
		Sex:      model.MALE,
		Age:      100,
	}

	var env = getEnv(db)
	var hash, _ = env.hashFunc([]byte(user.Password))

	// mock exists
	mock.
		ExpectQuery("SELECT count").
		WithArgs(user.Login).
		WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(0))

	// mock user insertion
	mock.
		ExpectExec("INSERT INTO Users").
		WithArgs(user.Login, string(hash), user.Age, user.Sex, user.About).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// mock id selection
	mock.
		ExpectQuery("SELECT id FROM").
		WithArgs(user.Login).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	var requestMsg, jsonErr = json.Marshal(user)
	assert.Nil(t, jsonErr)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserRegisterPost,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestEnv_UserRegisterPost_ParseFail(t *testing.T) {

	var env = &Env{
		conf: getAuthConf(),
		logger: mylog.NewLogger(ioutil.Discard),
	}

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserRegisterPost,
		strings.NewReader("Invalid json"),
		headerPair{"Content-Type", "application/json"},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEnv_UserRegisterPost_CheckFail(t *testing.T) {
	var db, mock, dbErr = sqlmock.New()

	if dbErr != nil {
		t.Fatal(dbErr)
	}
	defer db.Close()

	var user = &model.User{
		Login:    "login",
		Password: "password",
		About:    "about",
		Sex:      model.MALE,
		Age:      100,
	}

	// mock exists
	mock.
	ExpectQuery("SELECT count").
		WithArgs(user.Login).
		WillReturnError(errors.New("db fail"))

	var env = getEnv(db)

	var requestMsg, jsonErr = json.Marshal(user)
	assert.Nil(t, jsonErr)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserRegisterPost,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestEnv_UserRegisterPost_Conflict(t *testing.T) {
	var db, mock, dbErr = sqlmock.New()

	if dbErr != nil {
		t.Fatal(dbErr)
	}
	defer db.Close()

	var user = &model.User{
		Login:    "login",
		Password: "password",
		About:    "about",
		Sex:      model.MALE,
		Age:      100,
	}

	// mock exists
	mock.
	ExpectQuery("SELECT count").
		WithArgs(user.Login).
		WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(1))

	var env = getEnv(db)

	var requestMsg, jsonErr = json.Marshal(user)
	assert.Nil(t, jsonErr)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserRegisterPost,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestEnv_UserRegisterPost_SaveErr(t *testing.T) {
	var db, mock, dbErr = sqlmock.New()

	if dbErr != nil {
		t.Fatal(dbErr)
	}
	defer db.Close()

	var user = &model.User{
		Login:    "login",
		Password: "password",
		About:    "about",
		Sex:      model.MALE,
		Age:      100,
	}

	var env = getEnv(db)

	var hash, _ = env.hashFunc([]byte(user.Password))

	// mock exists
	mock.
	ExpectQuery("SELECT count").
		WithArgs(user.Login).
		WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(0))

	// mock user insertion
	mock.
	ExpectExec("INSERT INTO Users").
		WithArgs(user.Login, hash, user.Age, user.Sex, user.About).
		WillReturnError(errors.New("db fail"))

	// mock id selection
	mock.
	ExpectQuery("SELECT id FROM").
		WithArgs(user.Login).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	var requestMsg, jsonErr = json.Marshal(user)
	assert.Nil(t, jsonErr)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserRegisterPost,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestEnv_UserRegisterPost_IdExtraction(t *testing.T) {
	var db, mock, dbErr = sqlmock.New()

	if dbErr != nil {
		t.Fatal(dbErr)
	}
	defer db.Close()

	var user = &model.User{
		Login:    "login",
		Password: "password",
		About:    "about",
		Sex:      model.MALE,
		Age:      100,
	}

	var env = getEnv(db)

	var hash, _ = env.hashFunc([]byte(user.Password))

	// mock exists
	mock.
	ExpectQuery("SELECT count").
		WithArgs(user.Login).
		WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(0))

	// mock user insertion
	mock.
	ExpectExec("INSERT INTO Users").
		WithArgs(user.Login, hash, user.Age, user.Sex, user.About).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// mock id selection
	mock.
	ExpectQuery("SELECT id FROM").
		WithArgs(user.Login).
		WillReturnError(errors.New("db fail"))

	var requestMsg, jsonErr = json.Marshal(user)
	assert.Nil(t, jsonErr)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserRegisterPost,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestEnv_UserRegisterPost_NoLogin(t *testing.T) {
	var user = &model.User{
		Password: "password",
		About:    "about",
		Sex:      model.MALE,
		Age:      100,
	}

	var env = &Env{
		conf: getAuthConf(),
		logger:mylog.NewLogger(ioutil.Discard),
	}

	var requestMsg, jsonErr = json.Marshal(user)
	assert.Nil(t, jsonErr)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserRegisterPost,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEnv_UserRegisterPost_NoPassword(t *testing.T) {
	var user = &model.User{
		Login: "login",
		About:    "about",
		Sex:      model.MALE,
		Age:      100,
	}

	var env = &Env{
		conf: getAuthConf(),
		logger: mylog.NewLogger(ioutil.Discard),
	}

	var requestMsg, jsonErr = json.Marshal(user)
	assert.Nil(t, jsonErr)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserRegisterPost,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEnv_UserSignInPost_Success(t *testing.T) {
	var db, mock, dbErr = sqlmock.New()

	if dbErr != nil {
		t.Fatal(dbErr)
	}
	defer db.Close()

	var user = &model.User{
		Login:    "login",
		Password: "password",
		About:    "about",
		Sex:      model.MALE,
		Age:      100,
	}

	var env = getEnv(db)

	// mock exists
	mock.
	ExpectQuery("SELECT count").
		WithArgs(user.Login).
		WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(1))

	// mock user extraction
	var hash, _ = env.hashFunc([]byte(user.Password))
	mock.
	ExpectQuery("SELECT id").
		WithArgs(user.Login).
		WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "login", "password", "age", "sex", "about"}).
				AddRow(1, "login", hash, 100, model.MALE, "about"),
		)

	var requestMsg, jsonErr = json.Marshal(user)
	assert.Nil(t, jsonErr)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserSignInPost,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestEnv_UserSignInPost_WrongPassword(t *testing.T) {
	var db, mock, dbErr = sqlmock.New()

	if dbErr != nil {
		t.Fatal(dbErr)
	}
	defer db.Close()

	var user = &model.User{
		Login:    "login",
		Password: "password",
		About:    "about",
		Sex:      model.MALE,
		Age:      100,
	}

	// mock exists
	mock.
	ExpectQuery("SELECT count").
		WithArgs(user.Login).
		WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(1))

	// mock user extraction
	mock.
	ExpectQuery("SELECT id").
		WithArgs(user.Login).
		WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "login", "password", "age", "sex", "about"}).
			AddRow(1, "login", "pass", 100, model.MALE, "about"),
	)

	var env = getEnv(db)

	var requestMsg, jsonErr = json.Marshal(user)
	assert.Nil(t, jsonErr)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserSignInPost,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestEnv_UserSignInPost_ParseError(t *testing.T) {
	var env = &Env{
		conf: getAuthConf(),
		logger: mylog.NewLogger(ioutil.Discard),
	}

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserSignInPost,
		strings.NewReader("Invalid json"),
		headerPair{"Content-Type", "application/json"},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEnv_UserSignInPost_CheckDBFail(t *testing.T) {
	var db, mock, dbErr = sqlmock.New()

	if dbErr != nil {
		t.Fatal(dbErr)
	}
	defer db.Close()

	var user = &model.User{
		Login:    "login",
		Password: "password",
		About:    "about",
		Sex:      model.MALE,
		Age:      100,
	}

	// mock exists
	mock.
	ExpectQuery("SELECT count").
		WithArgs(user.Login).
		WillReturnError(errors.New("db fail"))

	var env = getEnv(db)

	var requestMsg, jsonErr = json.Marshal(user)
	assert.Nil(t, jsonErr)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserSignInPost,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestEnv_UserSignInPost_NotFound(t *testing.T) {
	var db, mock, dbErr = sqlmock.New()

	if dbErr != nil {
		t.Fatal(dbErr)
	}
	defer db.Close()

	var user = &model.User{
		Login:    "login",
		Password: "password",
		About:    "about",
		Sex:      model.MALE,
		Age:      100,
	}

	// mock exists
	mock.
	ExpectQuery("SELECT count").
		WithArgs(user.Login).
		WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(0))

	var env = getEnv(db)

	var requestMsg, jsonErr = json.Marshal(user)
	assert.Nil(t, jsonErr)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserSignInPost,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestEnv_UserSignInPost_IdExtractionFail(t *testing.T) {
	var db, mock, dbErr = sqlmock.New()

	if dbErr != nil {
		t.Fatal(dbErr)
	}
	defer db.Close()

	var user = &model.User{
		Login:    "login",
		Password: "password",
		About:    "about",
		Sex:      model.MALE,
		Age:      100,
	}

	// mock exists
	mock.
	ExpectQuery("SELECT count").
		WithArgs(user.Login).
		WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(1))

	// mock id selection
	mock.
	ExpectQuery("SELECT id FROM").
		WithArgs(user.Login).
		WillReturnError(errors.New("db fail"))

	var env = getEnv(db)

	var requestMsg, jsonErr = json.Marshal(user)
	assert.Nil(t, jsonErr)

	var rec, recErr = getRecorder(
		urlSample,
		http.MethodPost,
		env.UserSignInPost,
		strings.NewReader(string(requestMsg)),
		headerPair{"Content-Type", "application/json"},
	)

	assert.Nil(t, recErr)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func getAuthConf() config.Conf {
	return config.Conf{
		Auth:config.AuthConfig{
			ExpireDays: expireDays,
			TokenKey:   tokenKey,
		},
	}
}

func getRecorder(
	url string,
	method string,
	handlerFunc func(http.ResponseWriter, *http.Request),
	body io.Reader,
	headers ...headerPair,
) (*httptest.ResponseRecorder, error) {
	var req, err = http.NewRequest(
		method,
		url,
		body,
	)

	for _, hp := range headers {
		req.Header.Set(hp.key, hp.value)
	}

	if err != nil {
		return nil, err
	}

	var handler = http.HandlerFunc(handlerFunc)
	var rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec, nil
}

func getEnv(db *sql.DB) *Env {
	return &Env{
		userDAO:  dao.NewDBUserDAO(db),
		conf: getAuthConf(),
		hashFunc: func(password []byte) ([]byte, error) {
			var h = sha256.New()
			h.Write(password)
			return h.Sum(nil), nil
		},
		hashValidator: func(password []byte, hash []byte) error {
			var h = sha256.New()
			h.Write(password)
			var passHash = h.Sum(nil)

			if string(passHash) != string(hash) {
				return fmt.Errorf("hashes %s, %s do not match", string(passHash), string(hash))
			}
			return nil
		},
		logger: mylog.NewLogger(ioutil.Discard),
	}
}
