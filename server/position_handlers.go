package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sovianum/acquaintance-server/common"
	"github.com/Sovianum/acquaintance-server/model"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"net/http"
	"strings"
	"github.com/gorilla/mux"
	"strconv"
)

const (
	authorizationStr = "Authorization"
	id = "id"
)

func (env *Env) UserGetNeighboursGet(w http.ResponseWriter, r *http.Request) {
	env.logger.LogRequestStart(r)
	var userId, idCode, idErr = env.getIdFromRequest(r)
	if idErr != nil {
		env.logger.LogRequestError(r, idErr)
		w.WriteHeader(idCode)
		w.Write(common.GetErrorJson(idErr))
		return
	}

	var neighbours, nErr = env.userDAO.GetNeighbourUsers(userId, env.conf.Logic.Distance, env.conf.Logic.OnlineTimeout)
	if nErr != nil {
		env.logger.LogRequestError(r, nErr)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(common.GetErrorJson(nErr))
		return
	}

	env.logger.LogRequestSuccess(r)
	w.Write(common.GetDataJson(neighbours))
}

func (env *Env) UserSavePositionPost(w http.ResponseWriter, r *http.Request) {
	env.logger.LogRequestStart(r)
	var userId, idCode, idErr = env.getIdFromRequest(r)
	if idErr != nil {
		env.logger.LogRequestError(r, idErr)
		w.WriteHeader(idCode)
		w.Write(common.GetErrorJson(idErr))
		return
	}

	var position, code, parseErr = parsePosition(r)
	if parseErr != nil {
		env.logger.LogRequestError(r, parseErr)
		w.WriteHeader(code)
		w.Write(common.GetErrorJson(parseErr))
		return
	}
	position.UserId = userId

	var saveErr = env.positionDAO.Save(position)
	if saveErr != nil {
		env.logger.LogRequestError(r, saveErr)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(common.GetErrorJson(saveErr))
		return
	}

	env.logger.LogRequestSuccess(r)
	w.Write(common.GetEmptyJson())
}

// TODO add tests
func (env *Env) UserGetPositionById(w http.ResponseWriter, r *http.Request) {
	var vars = mux.Vars(r)
	var neighbourIdStr = vars[id]
	var neighbourId, neighbourIdErr = strconv.Atoi(neighbourIdStr)
	if neighbourIdErr != nil {
		env.logger.LogRequestError(r, neighbourIdErr)
		w.WriteHeader(http.StatusNotFound)
		w.Write(common.GetErrorJson(neighbourIdErr))
		return
	}

	env.logger.LogRequestStart(r)
	var _, idCode, userIdErr = env.getIdFromRequest(r)
	if userIdErr != nil {
		env.logger.LogRequestError(r, userIdErr)
		w.WriteHeader(idCode)
		w.Write(common.GetErrorJson(userIdErr))
		return
	}

	// todo check if current user has submitted request to requested user
	var neighbour, nErr = env.positionDAO.GetUserPositionById(neighbourId)
	if nErr != nil {
		env.logger.LogRequestError(r, nErr)	// TODO handle user not found case
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(common.GetErrorJson(nErr))
		return
	}

	env.logger.LogRequestSuccess(r)
	w.Write(common.GetDataJson(neighbour))
}

func parsePosition(r *http.Request) (*model.Position, int, error) {
	var body, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	if err := r.Body.Close(); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	var position = new(model.Position)
	if err := json.Unmarshal(body, &position); err != nil {
		return nil, http.StatusBadRequest, err
	}

	// TODO add position validation

	return position, http.StatusOK, nil
}

// TODO use some standard mechanisms instead of bicycles
func (env *Env) parseTokenString(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return env.conf.Auth.GetTokenKey(), nil
	})
}

func (env *Env) getIdFromRequest(r *http.Request) (id int, code int, err error) {
	var headers = r.Header
	var authHeaderList, ok = headers[authorizationStr]
	if !ok {
		return 0, http.StatusUnauthorized, errors.New("Header \"Authorization\" not set in request")
	}
	if len(authHeaderList) != 1 {
		return 0, http.StatusBadRequest, fmt.Errorf("You set too many (%d) \"Authorization\" headers", len(authHeaderList))
	}
	var authHeader = authHeaderList[0]

	var fields = strings.Fields(authHeader) // getting last word to remove Bearer word from header
	var tokenString = fields[len(fields)-1]

	var token, tokenErr = env.parseTokenString(tokenString)
	if tokenErr != nil {
		return 0, http.StatusBadRequest, errors.New("You sent unparseable token")
	}

	var userId, idErr = env.getIdFromTokenString(token)
	if idErr != nil {
		return 0, http.StatusBadRequest, errors.New("Your token does not contain your id")
	}

	return userId, http.StatusOK, nil
}

func (env *Env) getIdFromTokenString(token *jwt.Token) (int, error) {
	var claims, okClaims = token.Claims.(jwt.MapClaims)
	if !okClaims {
		return 0, errors.New("Failed to extract claims from token")
	}

	var idData, okId = claims[idStr]
	if !okId {
		return 0, errors.New("Failed to extract id from claims")
	}

	var id int
	switch idData.(type) {
	case int:
		id = idData.(int)
	case float64:
		var floatId = idData.(float64)
		id = round(floatId)
	default:
		return 0, errors.New("Failed to cast claims[id] to int")
	}

	return id, nil
}

func round(f float64) int {
	var floor = int(f)
	var ceil = floor + 1

	var floorDiff = f - float64(floor)
	var ceilDiff = float64(ceil) - f

	if floorDiff < ceilDiff {
		return floor
	} else {
		return ceil
	}
}
