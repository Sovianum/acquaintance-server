package dao

import (
	"database/sql"
	"github.com/Sovianum/acquaintance-server/model"
	"fmt"
)

const (
	countPendingRequests = `
		SELECT count(*) FROM MeetRequest
		WHERE requesterId = $1 AND requestedId = $2 AND status = 'PENDING'
	`
	getIncomePendingRequests = `
		SELECT mr.id, mr.requesterId, u1.login, mr.requestedId, u2.login, mr.status, mr.time FROM MeetRequest mr
			JOIN Users u1 ON mr.requesterId = u1.id
			JOIN Users u2 ON mr.requestedId = u2.id
		WHERE mr.requestedId = $1 AND status = 'PENDING'
	`
	getOutcomePendingRequests = `
		SELECT mr.id, mr.requesterId, u1.login, mr.requestedId, u2.login, mr.status, mr.time FROM MeetRequest mr
			JOIN Users u1 ON mr.requesterId = u1.id
			JOIN Users u2 ON mr.requestedId = u2.id
		WHERE mr.requesterId = $1 AND status = 'PENDING'
	`
	getAllRequests = `
		SELECT mr.id, mr.requesterId, u1.login, mr.requestedId, u2.login, mr.status, mr.time FROM MeetRequest mr
			JOIN Users u1 ON mr.requesterId = u1.id
			JOIN Users u2 ON mr.requestedId = u2.id
		WHERE mr.requestedId = $1 OR mr.requesterId = $1
	`
	checkAccessibility = `
		SELECT ST_DistanceSphere(p1.point, p2.point) < $1 FROM
			(
				SELECT * FROM Position
				WHERE userId = $2 AND age(now(), time) < $4 * interval '1 minute'
				ORDER BY time DESC
				LIMIT 1
			) p1,
			(
				SELECT * FROM Position
				WHERE userId = $3 AND age(now(), time) < $4 * interval '1 minute'
				ORDER BY time DESC
				LIMIT 1
			) p2
	`
	createRequest = `
		INSERT INTO MeetRequest (requesterId, requestedId) VALUES ($1, $2)
	`
	updateRequestStatus = `
		UPDATE MeetRequest SET status = $1 WHERE id = $2 AND requestedId = $3
	`
	getRequestById = `
		SELECT id, requesterId, requestedId, status, time FROM MeetRequest WHERE id = $1
	`
	getLasRequestId = `
		SELECT max(id) FROM MeetRequest
	`
	declineAll = `
		UPDATE MeetRequest SET status = 'DECLINED' WHERE status = 'PENDING' AND age(now(), time) > $1 * interval '1 minute'
	`
)

const (
	ImpossibleID = -1 - iota
	RequestExists
	UserInaccessible

)

func IsInvalidId(id int) bool {
	return id < 0
}

func GetLogicalError(code int) error {
	switch code {
	case RequestExists:
		return fmt.Errorf("request already exists")
	case UserInaccessible:
		return fmt.Errorf("user inaccessible")
	default:
		return fmt.Errorf("unknown error with code %d", code)
	}
}

type MeetRequestDAO interface {
	CreateRequest(requesterId int, requestedId int, requestTimeoutMin int, maxDistance float64) (id int, dbErr error)
	GetAllRequests(userId int) ([]*model.MeetRequest, error)
	GetIncomePendingRequests(requestedId int) ([]*model.MeetRequest, error)
	GetOutcomePendingRequests(requesterId int) ([]*model.MeetRequest, error)
	GetRequestById(id int) (*model.MeetRequest, error)
	UpdateRequest(id int, requestedId int, status string) (int, error)
	DeclineAll(timeoutMin int) error
}

type meetRequestDAO struct {
	db *sql.DB
}

func NewMeetDAO(db *sql.DB) MeetRequestDAO {
	return &meetRequestDAO{
		db: db,
	}
}

func (dao *meetRequestDAO) GetRequestById(id int) (*model.MeetRequest, error) {
	var r = new(model.MeetRequest)
	var err = dao.db.QueryRow(getRequestById, id).Scan(&r.Id, &r.RequesterId, &r.RequestedId, &r.Status, &r.Time)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (dao *meetRequestDAO) GetAllRequests(userId int) ([]*model.MeetRequest, error) {
	return dao.getRequestsTemplate(getAllRequests, userId)
}

func (dao *meetRequestDAO) GetIncomePendingRequests(userId int) ([]*model.MeetRequest, error) {
	return dao.getRequestsTemplate(getIncomePendingRequests, userId)
}

func (dao *meetRequestDAO) GetOutcomePendingRequests(userId int) ([]*model.MeetRequest, error) {
	return dao.getRequestsTemplate(getOutcomePendingRequests, userId)
}

func (dao *meetRequestDAO) CreateRequest(requesterId int, requestedId int, requestTimeoutMin int, maxDistance float64) (int, error) {
	var requestCnt, countErr = dao.countPendingRequests(requesterId, requestedId)
	if countErr != nil {
		return ImpossibleID, countErr
	}

	if requestCnt > 0 {
		return RequestExists, nil
	}

	var accessible, accessErr = dao.isAccessible(requesterId, requestedId, maxDistance, requestTimeoutMin)
	if accessErr != nil {
		return ImpossibleID, accessErr
	}

	if !accessible {
		return UserInaccessible, nil
	}

	var tx, txError = dao.db.Begin()
	if txError != nil {
		tx.Rollback()
		return ImpossibleID, txError
	}

	var _, createErr = tx.Exec(createRequest, requesterId, requestedId)
	if createErr != nil {
		tx.Rollback()
		return ImpossibleID, createErr
	}

	var lastId int
	var lastIdErr = tx.QueryRow(getLasRequestId).Scan(&lastId)
	if lastIdErr != nil {
		tx.Rollback()
		return ImpossibleID, lastIdErr
	}

	tx.Commit()
	return lastId, nil
}

func (dao *meetRequestDAO) UpdateRequest(id int, requestedId int, status string) (int, error) {
	var result, err = dao.db.Exec(updateRequestStatus, status, id, requestedId)
	if err != nil {
		return 0, err
	}

	var rowsAffected, rowsErr = result.RowsAffected()
	if rowsErr != nil {
		return 0, rowsErr
	}

	return int(rowsAffected), nil
}

func (dao *meetRequestDAO) DeclineAll(timeoutMin int) error {
	var _, err = dao.db.Exec(declineAll, timeoutMin)
	return err
}

func (dao *meetRequestDAO) isAccessible(id1 int, id2 int, maxDistance float64, timeoutMin int) (bool, error) {
	var rows, err = dao.db.Query(checkAccessibility, maxDistance, id1, id2, timeoutMin)
	if err != nil {
		return false, err
	}

	var accessible = false
	for rows.Next() {
		err = rows.Scan(&accessible)
		if err != nil {
			return false, err
		}
	}
	err = rows.Err()
	if err != nil {
		return false, err
	}

	return accessible, nil
}

func (dao *meetRequestDAO) createRequest(requesterId int, requestedId int) error {
	var _, err = dao.db.Exec(createRequest, requesterId, requestedId)
	return err
}

func (dao *meetRequestDAO) getRequestsTemplate(sql string, userId int) ([]*model.MeetRequest, error) {
	var rows, err = dao.db.Query(sql, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result = make([]*model.MeetRequest, 0)
	for rows.Next() {
		var request = new(model.MeetRequest)
		err = rows.Scan(
			&request.Id,
			&request.RequesterId,
			&request.RequesterLogin,
			&request.RequestedId,
			&request.RequestedLogin,
			&request.Status,
			&request.Time,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, request)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (dao *meetRequestDAO) countPendingRequests(requesterId int, requestedId int) (int, error) {
	var cnt int
	var err = dao.db.QueryRow(countPendingRequests, requesterId, requestedId).Scan(&cnt)
	return cnt, err
}
