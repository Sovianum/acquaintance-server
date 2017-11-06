package dao

import (
	"errors"
	"github.com/Sovianum/acquaintance-server/model"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"testing"
	"time"
)

func TestDbPositionDAO_Save_Success(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.
		ExpectExec("INSERT INTO Position").
		WithArgs(100, 10., 20.).
		WillReturnResult(sqlmock.NewResult(1, 1))

	var position = &model.Position{UserId: 100, Point: model.Point{X: 10., Y: 20.}}

	var positionDAO = NewDBPositionDAO(db)
	var saveErr = positionDAO.Save(position)

	assert.Nil(t, saveErr)
}

func TestDbPositionDAO_Save_DuplicateLogin(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.
		ExpectExec("INSERT INTO Position").
		WithArgs(100, 10., 20.).
		WillReturnError(errors.New("Duplicate id"))

	var position = &model.Position{UserId: 100, Point: model.Point{X: 10., Y: 20.}}

	var positionDAO = NewDBPositionDAO(db)
	var saveErr = positionDAO.Save(position)

	assert.NotNil(t, saveErr)
	assert.Equal(t, "Duplicate id", saveErr.Error())
}

func TestDbPositionDAO_Get_Success(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var date = time.Date(2003, 10, 17, 0, 0, 0, 0, time.UTC)
	var rows = sqlmock.NewRows([]string{"id", "userId", "x", "y", "time"}).
		AddRow(1, 100, 10., 20., date)

	mock.
		ExpectQuery("SELECT id, userId").
		WithArgs(1).
		WillReturnRows(rows)

	var position = &model.Position{Id: 1, UserId: 100, Point: model.Point{X: 10., Y: 20.}, Time: model.QuotedTime(date)}

	var positionDAO = NewDBPositionDAO(db)
	var dbPosition, userErr = positionDAO.GetUserPosition(1)

	assert.Nil(t, userErr)
	assert.Equal(t, position, dbPosition)
}

func TestDbPositionDAO_Get_NotFound(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.
		ExpectQuery("SELECT").
		WithArgs(1).
		WillReturnError(errors.New("position not found"))

	var positionDAO = NewDBPositionDAO(db)
	var _, positionErr = positionDAO.GetUserPosition(1)

	assert.NotNil(t, positionErr)
	assert.Equal(t, "position not found", positionErr.Error())
}
