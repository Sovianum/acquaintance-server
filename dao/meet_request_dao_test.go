package dao

import (
	"database/sql/driver"
	"github.com/Sovianum/acquaintance-server/model"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"strconv"
	"testing"
	"time"
)

func TestMeetRequestDAO_GetRequestById_Success(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var date = time.Date(2003, 10, 17, 0, 0, 0, 0, time.UTC)

	mock.
		ExpectQuery("SELECT").
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "requesterId", "requestedId", "status", "time"}).
				AddRow(1, 2, 3, model.StatusPending, date),
		)

	var request = &model.MeetRequest{
		Id:          1,
		RequesterId: 2,
		RequestedId: 3,
		Time:        model.QuotedTime(date),
		Status:      model.StatusPending,
	}

	var meetRequestDAO = NewMeetDAO(db)
	var dbRequest, dbErr = meetRequestDAO.GetRequestById(1)

	assert.Nil(t, dbErr)
	assert.Equal(t, *request, *dbRequest)
}

func TestMeetRequestDAO_GetRequestById_NotFound(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.
		ExpectQuery("SELECT").
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "requesterId", "requestedId", "status", "time"}),
		)

	var meetRequestDAO = NewMeetDAO(db)
	var _, dbErr = meetRequestDAO.GetRequestById(1)

	assert.NotNil(t, dbErr)
}

func TestMeetRequestDAO_GetRequests_Success(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var date = time.Date(2003, 10, 17, 0, 0, 0, 0, time.UTC)

	mock.
		ExpectQuery("SELECT").
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "requesterId", "requestedId", "status", "time"}).
				AddRow(1, 2, 3, model.StatusPending, date),
		)

	var request = &model.MeetRequest{
		Id:          1,
		RequesterId: 2,
		RequestedId: 3,
		Time:        model.QuotedTime(date),
		Status:      model.StatusPending,
	}

	var meetRequestDAO = NewMeetDAO(db)
	var dbRequestRequests, dbErr = meetRequestDAO.GetRequests(1)

	assert.Nil(t, dbErr)
	assert.Equal(t, 1, len(dbRequestRequests))
	assert.Equal(t, *request, *dbRequestRequests[0])
}

func TestMeetRequestDAO_GetRequests_Empty(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.
		ExpectQuery("SELECT").
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "requesterId", "requestedId", "status", "time"}),
		)

	var meetRequestDAO = NewMeetDAO(db)
	var dbRequestRequests, dbErr = meetRequestDAO.GetRequests(1)

	assert.Nil(t, dbErr)
	assert.Equal(t, 0, len(dbRequestRequests))
}

func TestMeetRequestDAO_GetRequests_DBErr(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.
		ExpectQuery("SELECT").
		WithArgs(1).
		WillReturnError(errors.New("fail"))

	var meetRequestDAO = NewMeetDAO(db)
	var _, dbErr = meetRequestDAO.GetRequests(1)

	assert.NotNil(t, dbErr)
	assert.Equal(t, "fail", dbErr.Error())
}

func TestMeetRequestDAO_CreateRequest(t *testing.T) {
	var cases = []struct {
		requesterId       int
		requestedId       int
		requestTimeOutMin int
		maxDistance       float64

		countErrIsNil bool
		countErrMsg   string
		countRes      []driver.Value

		accessErrIsNil bool
		accessErrMsg   string
		accessRes      []driver.Value

		createErrIsNil bool
		createErrMsg   string

		expectedId int
	}{
		{
			requesterId:       1,
			requestedId:       2,
			requestTimeOutMin: 10,
			maxDistance:       10,
			countErrIsNil:     true,
			countRes:          []driver.Value{0},

			accessErrIsNil: true,
			accessRes:      []driver.Value{true},

			createErrIsNil: true,

			expectedId: 100,
		},
		{
			requesterId:       1,
			requestedId:       2,
			requestTimeOutMin: 10,
			maxDistance:       10,
			countErrIsNil:     true,
			countRes:          []driver.Value{1},

			accessErrIsNil: true,
			accessRes:      []driver.Value{true},

			createErrIsNil: true,

			expectedId: ImpossibleID,
		},
		{
			requesterId:       1,
			requestedId:       2,
			requestTimeOutMin: 10,
			maxDistance:       10,
			countErrIsNil:     false,
			countErrMsg:       "countErr",

			accessErrIsNil: true,
			accessRes:      []driver.Value{true},

			createErrIsNil: true,

			expectedId: ImpossibleID,
		},
		{
			requesterId:       1,
			requestedId:       2,
			requestTimeOutMin: 10,
			maxDistance:       10,
			countErrIsNil:     true,
			countRes:          []driver.Value{0},

			accessErrIsNil: true,
			accessRes:      []driver.Value{false},

			createErrIsNil: true,

			expectedId: ImpossibleID,
		},
		{
			requesterId:       1,
			requestedId:       2,
			requestTimeOutMin: 10,
			maxDistance:       10,
			countErrIsNil:     true,
			countRes:          []driver.Value{0},

			accessErrIsNil: false,
			accessErrMsg:   "accessErr",

			createErrIsNil: true,

			expectedId: ImpossibleID,
		},
		{
			requesterId:       1,
			requestedId:       2,
			requestTimeOutMin: 10,
			maxDistance:       10,
			countErrIsNil:     true,
			countRes:          []driver.Value{0},

			accessErrIsNil: true,
			accessRes:      []driver.Value{true},

			createErrIsNil: false,
			createErrMsg:   "createErr",

			expectedId: ImpossibleID,
		},
	}

	for i, testCase := range cases {
		var db, mock, err = sqlmock.New()

		if err != nil {
			t.Fatal(err)
		}

		if testCase.countErrIsNil {
			mock.
				ExpectQuery("SELECT count").
				WithArgs(testCase.requesterId, testCase.requestedId).
				WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(testCase.countRes...))
		} else {
			mock.
				ExpectQuery("SELECT count").
				WithArgs(testCase.requesterId, testCase.requestedId).
				WillReturnError(errors.New(testCase.countErrMsg))
		}

		if testCase.countErrIsNil {
			if testCase.accessErrIsNil {
				mock.
					ExpectQuery("SELECT").
					WithArgs(testCase.maxDistance, testCase.requesterId, testCase.requestedId, testCase.requestTimeOutMin).
					WillReturnRows(sqlmock.NewRows([]string{"accessible"}).AddRow(testCase.accessRes...))
			} else {
				mock.
					ExpectQuery("SELECT").
					WithArgs(testCase.maxDistance, testCase.requesterId, testCase.requestedId, testCase.requestTimeOutMin).
					WillReturnError(errors.New(testCase.accessErrMsg))
			}
		}

		mock.ExpectBegin()
		if testCase.countErrIsNil && testCase.accessErrIsNil {
			if testCase.createErrIsNil {
				mock.
					ExpectExec("INSERT").
					WithArgs(testCase.requesterId, testCase.requestedId).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.
					ExpectQuery("SELECT").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(100))
				mock.ExpectCommit()
			} else {
				mock.
					ExpectExec("INSERT").
					WithArgs(testCase.requesterId, testCase.requestedId).
					WillReturnError(errors.New(testCase.createErrMsg))
				mock.ExpectRollback()
			}
		}

		var meetRequestDAO = NewMeetDAO(db)

		var lastId, dbErr = meetRequestDAO.CreateRequest(testCase.requesterId, testCase.requestedId, testCase.requestTimeOutMin, testCase.maxDistance)

		if testCase.countErrIsNil && testCase.accessErrIsNil && testCase.createErrIsNil {
			assert.Nil(t, dbErr, strconv.Itoa(i))
		} else {
			var msg string
			if !testCase.countErrIsNil {
				msg = testCase.countErrMsg
			} else if !testCase.accessErrIsNil {
				msg = testCase.accessErrMsg
			} else {
				msg = testCase.createErrMsg
			}

			assert.Equal(t, msg, dbErr.Error(), strconv.Itoa(i))
		}
		assert.Equal(t, testCase.expectedId, lastId, strconv.Itoa(i))

		db.Close()
	}
}

func TestMeetRequestDAO_UpdateRequest(t *testing.T) {
	var cases = []struct {
		requestId    int
		status       string
		errIsNil     bool
		errMsg       string
		rowsAffected int64
	}{
		{
			requestId:    1,
			status:       model.StatusAccepted,
			errIsNil:     true,
			rowsAffected: 1,
		},
		{
			requestId:    1,
			status:       model.StatusAccepted,
			errIsNil:     true,
			rowsAffected: 0,
		},
		{
			requestId: 1,
			status:    model.StatusAccepted,
			errIsNil:  false,
			errMsg:    "err",
		},
	}

	for i, testCase := range cases {
		var db, mock, err = sqlmock.New()

		if err != nil {
			t.Fatal(err)
		}

		if testCase.errIsNil {
			mock.
				ExpectExec("UPDATE").
				WithArgs(model.StatusAccepted, testCase.requestId, 100).
				WillReturnResult(sqlmock.NewResult(1, testCase.rowsAffected))
		} else {
			mock.
				ExpectExec("UPDATE").
				WithArgs(model.StatusAccepted, testCase.requestId, 100).
				WillReturnError(errors.New(testCase.errMsg))
		}

		var meetRequestDAO = NewMeetDAO(db)
		var rowsAffected, dbErr = meetRequestDAO.UpdateRequest(testCase.requestId, 100, testCase.status)

		if !testCase.errIsNil {
			assert.NotNil(t, dbErr, strconv.Itoa(i))
			assert.Equal(t, testCase.errMsg, dbErr.Error(), strconv.Itoa(i))
		} else {
			assert.Nil(t, dbErr, strconv.Itoa(i))
			assert.Equal(t, int(testCase.rowsAffected), rowsAffected, strconv.Itoa(i))
		}

		db.Close()
	}
}
