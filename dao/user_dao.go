package dao

import (
	"database/sql"
	"github.com/Sovianum/acquaintance-server/model"
)

const (
	saveUser          = `INSERT INTO Users (login, password, age, sex, about) VALUES ($1, $2, $3, $4, $5)`
	getUserById       = `SELECT id, login, password, age, sex, about FROM Users WHERE id = $1`
	getUserByLogin    = `SELECT id, login, password, age, sex, about FROM Users WHERE login = $1`
	getIdByLogin      = `SELECT id FROM Users WHERE login = $1`
	getNeighbourUsers = `SELECT DISTINCT ON (u2.id) u2.id, u2.login, u2.age, u2.sex, u2.about
						 FROM Users u1
						 	JOIN Users u2 ON u2.id != u1.id
						 	JOIN Position p1 ON u1.id = p1.userId
						 	JOIN Position p2 ON u2.id = p2.userId
						 WHERE u1.id = $1
							AND ST_DistanceSphere(p1.point, p2.point) <= $2
							AND age(current_timestamp, p2.time) < $3 * interval '1 minute'`
	checkUserById    = `SELECT count(*) cnt FROM Users u WHERE u.id = $1`
	checkUserByLogin = `SELECT count(*) cnt FROM Users u WHERE u.login = $1`
)

type UserDAO interface {
	Save(user *model.User) (int, error)
	GetUserById(id int) (*model.User, error)
	GetUserByLogin(login string) (*model.User, error)
	GetNeighbourUsers(id int, distance float64, onlineTimeoutMin int) ([]*model.User, error)
	GetIdByLogin(login string) (int, error)
	ExistsById(id int) (bool, error)
	ExistsByLogin(login string) (bool, error)
}

type dbUserDAO struct {
	db *sql.DB
}

func NewDBUserDAO(db *sql.DB) UserDAO {
	var result = new(dbUserDAO)
	result.db = db
	return result
}

func (dao *dbUserDAO) Save(user *model.User) (int, error) {
	_, saveErr := dao.db.Exec(saveUser, user.Login, user.Password, user.Age, user.Sex, user.About)
	if saveErr != nil {
		return 0, saveErr
	}

	var id, getErr = dao.getIdByLogin(user.Login)
	if getErr != nil {
		return 0, getErr
	}
	// TODO add handling of the case when user saved but not extracted

	return id, nil
}

func (dao *dbUserDAO) GetIdByLogin(login string) (int, error) {
	return dao.getIdByLogin(login)
}

func (dao *dbUserDAO) GetUserById(id int) (*model.User, error) {
	var user = new(model.User)
	var err = dao.db.QueryRow(getUserById, id).Scan(&user.Id, &user.Login, &user.Password, &user.Age, &user.Sex, &user.About)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (dao *dbUserDAO) GetUserByLogin(login string) (*model.User, error) {
	var user = new(model.User)
	var err = dao.db.QueryRow(getUserByLogin, login).Scan(&user.Id, &user.Login, &user.Password, &user.Age, &user.Sex, &user.About)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (dao *dbUserDAO) GetNeighbourUsers(id int, distance float64, onlineTimeoutMin int) ([]*model.User, error) {
	var rows, err = dao.db.Query(getNeighbourUsers, id, distance, onlineTimeoutMin)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result = make([]*model.User, 0)
	for rows.Next() {
		var user = new(model.User)
		err = rows.Scan(&user.Id, &user.Login, &user.Age, &user.Sex, &user.About)
		if err != nil {
			return nil, err
		}

		result = append(result, user)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (dao *dbUserDAO) ExistsById(id int) (bool, error) {
	var cnt int
	var err = dao.db.QueryRow(checkUserById, id).Scan(&cnt)
	if err != nil {
		return false, err
	}

	return cnt > 0, nil
}

func (dao *dbUserDAO) ExistsByLogin(login string) (bool, error) {
	var cnt int
	var err = dao.db.QueryRow(checkUserByLogin, login).Scan(&cnt)
	if err != nil {
		return false, err
	}

	return cnt > 0, nil
}

func (dao *dbUserDAO) getIdByLogin(login string) (int, error) {
	var id int
	var getErr = dao.db.QueryRow(getIdByLogin, login).Scan(&id)
	return id, getErr
}
