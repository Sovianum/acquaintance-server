package dao

import (
	"errors"
	"github.com/Sovianum/acquaintance-server/model"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"testing"
)

func TestDbUserDAO_ExistsByID_UserFound(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var rows = sqlmock.NewRows([]string{"cnt"}).
		AddRow(1)

	mock.
		ExpectQuery("SELECT count").
		WithArgs(10).
		WillReturnRows(rows)

	var userDAO = NewDBUserDAO(db)
	var exists, dbErr = userDAO.ExistsById(10)

	assert.Nil(t, dbErr)
	assert.True(t, exists, "Failed to find existing user")
}

func TestDbUserDAO_ExistsByID_UserNotFound(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var rows = sqlmock.NewRows([]string{"cnt"}).
		AddRow(0)

	mock.
		ExpectQuery("SELECT count").
		WithArgs(10).
		WillReturnRows(rows)

	var userDAO = NewDBUserDAO(db)
	var exists, dbErr = userDAO.ExistsById(10)

	assert.Nil(t, dbErr)
	assert.False(t, exists, "Succeeded to find non existing user")
}

func TestDbUserDAO_ExistsById_DBFailed(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.
		ExpectQuery("SELECT count").
		WithArgs(10).
		WillReturnError(errors.New("Failed to check"))

	var userDAO = NewDBUserDAO(db)
	var _, dbErr = userDAO.ExistsById(10)

	assert.NotNil(t, dbErr)
	assert.Equal(t, "Failed to check", dbErr.Error())
}

func TestDbUserDAO_ExistsByLogin_UserFound(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var rows = sqlmock.NewRows([]string{"cnt"}).
		AddRow(1)

	mock.
		ExpectQuery("SELECT count").
		WithArgs("login").
		WillReturnRows(rows)

	var userDAO = NewDBUserDAO(db)
	var exists, dbErr = userDAO.ExistsByLogin("login")

	assert.Nil(t, dbErr)
	assert.True(t, exists, "Failed to find existing user")
}

func TestDbUserDAO_ExistsByLogin_UserNotFound(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var rows = sqlmock.NewRows([]string{"cnt"}).
		AddRow(0)

	mock.
		ExpectQuery("SELECT count").
		WithArgs("login").
		WillReturnRows(rows)

	var userDAO = NewDBUserDAO(db)
	var exists, dbErr = userDAO.ExistsByLogin("login")

	assert.Nil(t, dbErr)
	assert.False(t, exists, "Succeeded to find non existing user")
}

func TestDbUserDAO_ExistsByLogin_DBFailed(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.
		ExpectQuery("SELECT count").
		WithArgs("login").
		WillReturnError(errors.New("Failed to check"))

	var userDAO = NewDBUserDAO(db)
	var _, dbErr = userDAO.ExistsByLogin("login")

	assert.NotNil(t, dbErr)
	assert.Equal(t, "Failed to check", dbErr.Error())
}

func TestDbUserDAO_Save_Success(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var pass = "password"
	var hash, _ = bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	mock.
		ExpectExec("INSERT INTO").
		WithArgs("login", string(hash), 100, model.FEMALE, "").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.
		ExpectQuery("SELECT").
		WithArgs("login").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	var user = &model.User{Login: "login", Password: string(hash), Sex: model.FEMALE, Age: 100}

	var userDAO = NewDBUserDAO(db)
	var id, saveErr = userDAO.Save(user)

	assert.Nil(t, saveErr)
	assert.Equal(t, 1, id)
}

func TestDbUserDAO_Save_DuplicateLogin(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.
		ExpectExec("INSERT INTO").
		WithArgs("login", "pass", 100, model.FEMALE, "").
		WillReturnError(errors.New("Duplicate id"))

	var user = &model.User{Login: "login", Password: "pass", Sex: model.FEMALE, Age: 100}

	var userDAO = NewDBUserDAO(db)
	var _, saveErr = userDAO.Save(user)

	assert.NotNil(t, saveErr)
	assert.Equal(t, "Duplicate id", saveErr.Error())
}

func TestDbUserDAO_GetUserById_Success(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var rows = sqlmock.NewRows([]string{"id", "login", "password", "age", "sex", "about"}).
		AddRow(1, "login", "pass", 100, model.MALE, "about")

	mock.
		ExpectQuery("SELECT").
		WithArgs(1).
		WillReturnRows(rows)

	var user = &model.User{Id: 1, Login: "login", Password: "pass", Sex: model.MALE, Age: 100, About: "about"}

	var userDAO = NewDBUserDAO(db)
	var dbUser, userErr = userDAO.GetUserById(1)

	assert.Nil(t, userErr)
	assert.Equal(t, user, dbUser)
}

func TestDbUserDAO_GetUserById_NotFound(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.
		ExpectQuery("SELECT").
		WithArgs(1).
		WillReturnError(errors.New("user not found"))

	var userDAO = NewDBUserDAO(db)
	var _, userErr = userDAO.GetUserById(1)

	assert.NotNil(t, userErr)
	assert.Equal(t, "user not found", userErr.Error())
}

func TestDbUserDAO_GetUserByLogin_Success(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var rows = sqlmock.NewRows([]string{"id", "login", "password", "age", "sex", "about"}).
		AddRow(1, "login", "pass", 100, model.MALE, "about")

	mock.
		ExpectQuery("SELECT").
		WithArgs("login").
		WillReturnRows(rows)

	var user = &model.User{Id: 1, Login: "login", Password: "pass", Sex: model.MALE, Age: 100, About: "about"}

	var userDAO = NewDBUserDAO(db)
	var dbUser, userErr = userDAO.GetUserByLogin(user.Login)

	assert.Nil(t, userErr)
	assert.Equal(t, user, dbUser)
}

func TestDbUserDAO_GetUserByLogin_NotFound(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.
		ExpectQuery("SELECT").
		WithArgs("login").
		WillReturnError(errors.New("user not found"))

	var userDAO = NewDBUserDAO(db)
	var _, userErr = userDAO.GetUserByLogin("login")

	assert.NotNil(t, userErr)
	assert.Equal(t, "user not found", userErr.Error())
}

func TestDbUserDAO_GetIdByLogin_Success(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var rows = sqlmock.NewRows([]string{"id"}).AddRow(1)

	mock.
		ExpectQuery("SELECT").
		WithArgs("login").
		WillReturnRows(rows)

	var userDAO = NewDBUserDAO(db)
	var id, userErr = userDAO.GetIdByLogin("login")

	assert.Nil(t, userErr)
	assert.Equal(t, 1, id)
}

func TestDbUserDAO_GetIdByLogin_NotFound(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.
		ExpectQuery("SELECT").
		WithArgs("login").
		WillReturnError(errors.New("user not found"))

	var userDAO = NewDBUserDAO(db)
	var _, userErr = userDAO.GetUserByLogin("login")

	assert.Equal(t, "user not found", userErr.Error())
}

func TestDbUserDAO_GetNeighbour_Success(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var rows = sqlmock.NewRows([]string{"id", "login", "age", "sex", "about"}).
		AddRow(1, "login1", 101, model.MALE, "about1").
		AddRow(2, "login2", 102, model.FEMALE, "about2")

	mock.
		ExpectQuery("SELECT").
		WithArgs(0, float64(100), 1).
		WillReturnRows(rows)

	var users = []*model.User{
		{Id: 1, Login: "login1", Sex: model.MALE, Age: 101, About: "about1"},
		{Id: 2, Login: "login2", Sex: model.FEMALE, Age: 102, About: "about2"},
	}

	var userDAO = NewDBUserDAO(db)
	var dbUsers, userErr = userDAO.GetNeighbourUsers(0, float64(100), 1)

	assert.Nil(t, userErr)
	assert.Equal(t, len(users), len(dbUsers))

	for i := 0; i != len(users); i++ {
		assert.Equal(t, users[i], dbUsers[i], i)
	}
}

func TestDbUserDAO_GetNeighbour_Empty(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var rows = sqlmock.NewRows([]string{"id", "login", "age", "sex", "about"})

	mock.
		ExpectQuery("SELECT").
		WithArgs(0, float64(100), 1).
		WillReturnRows(rows)

	var userDAO = NewDBUserDAO(db)
	var dbUsers, userErr = userDAO.GetNeighbourUsers(0, float64(100), 1)

	assert.Nil(t, userErr)
	assert.Equal(t, 0, len(dbUsers))
}

func TestDbUserDAO_GetNeighbour_DBError(t *testing.T) {
	var db, mock, err = sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.
		ExpectQuery("SELECT").
		WithArgs(0, float64(100), 1).
		WillReturnError(errors.New("failed to get"))

	var userDAO = NewDBUserDAO(db)
	var _, userErr = userDAO.GetNeighbourUsers(0, float64(100), 1)

	assert.NotNil(t, userErr)
	assert.Equal(t, "failed to get", userErr.Error())
}
