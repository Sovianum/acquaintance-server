package server

import (
	"encoding/json"
	"errors"
	"github.com/Sovianum/acquaintance-server/common"
	"github.com/Sovianum/acquaintance-server/model"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	idStr    = "id"
	loginStr = "login"
	expStr   = "exp"
)

func (env *Env) UserRegisterPost(w http.ResponseWriter, r *http.Request) {
	env.logger.LogRequestStart(r)
	var user, code, parseErr = parseUser(r)
	if parseErr != nil {
		env.logger.LogRequestError(r, parseErr)
		w.WriteHeader(code)
		common.WriteWithLogging(r, w, common.GetErrorJson(parseErr), env.logger)
		return
	}

	var exists, existsErr = env.userDAO.ExistsByLogin(user.Login)
	if existsErr != nil {
		env.logger.LogRequestError(r, existsErr)
		w.WriteHeader(http.StatusInternalServerError)
		common.WriteWithLogging(r, w, common.GetErrorJson(existsErr), env.logger)
		return
	}
	if exists {
		var err = errors.New("user already exists")
		env.logger.LogRequestError(r, err)
		w.WriteHeader(http.StatusConflict)
		common.WriteWithLogging(r, w, common.GetErrorJson(err), env.logger)
		return
	}

	var hash, err = env.hashFunc([]byte(user.Password))
	if err != nil {
		env.logger.LogRequestError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
		common.WriteWithLogging(r, w, common.GetErrorJson(err), env.logger)
		return
	}
	user.Password = string(hash)

	var userId, saveErr = env.userDAO.Save(user)
	if saveErr != nil {
		env.logger.LogRequestError(r, saveErr)
		w.WriteHeader(http.StatusInternalServerError)
		common.WriteWithLogging(r, w, common.GetErrorJson(saveErr), env.logger)
		return
	}

	var tokenString, tokenErr = env.generateTokenString(userId, user.Login)
	if tokenErr != nil {
		env.logger.LogRequestError(r, tokenErr)
		w.WriteHeader(http.StatusInternalServerError)
		common.WriteWithLogging(r, w, common.GetErrorJson(tokenErr), env.logger)
		// TODO add info that user has been successfully saved
		return
	}

	env.logger.LogRequestSuccess(r)
	common.WriteWithLogging(r, w, common.GetDataJson(tokenString), env.logger)
}

func (env *Env) UserSignInPost(w http.ResponseWriter, r *http.Request) {
	env.logger.LogRequestStart(r)
	var user, code, parseErr = parseUser(r)
	if parseErr != nil {
		env.logger.LogRequestError(r, parseErr)
		w.WriteHeader(code)
		common.WriteWithLogging(r, w, common.GetErrorJson(parseErr), env.logger)
		return
	}

	var exists, existsErr = env.userDAO.ExistsByLogin(user.Login)
	if existsErr != nil {
		env.logger.LogRequestError(r, existsErr)
		w.WriteHeader(http.StatusInternalServerError)
		common.WriteWithLogging(r, w, common.GetErrorJson(existsErr), env.logger)
		return
	}
	if !exists {
		var err = errors.New("not found")
		env.logger.LogRequestError(r, err)
		w.WriteHeader(http.StatusNotFound)
		common.WriteWithLogging(r, w, common.GetErrorJson(err), env.logger)
		return
	}

	var dbUser, dbErr = env.userDAO.GetUserByLogin(user.Login)
	if dbErr != nil {
		env.logger.LogRequestError(r, dbErr)
		w.WriteHeader(http.StatusInternalServerError)
		common.WriteWithLogging(r, w, common.GetErrorJson(dbErr), env.logger)
		return
	}

	if err := env.hashValidator([]byte(user.Password), []byte(dbUser.Password)); err != nil {
		env.logger.LogRequestError(r, err)
		w.WriteHeader(http.StatusNotFound)
		common.WriteWithLogging(r, w, common.GetErrorJson(err), env.logger)
		return
	}

	var tokenString, tokenErr = env.generateTokenString(dbUser.Id, dbUser.Login)
	if tokenErr != nil {
		env.logger.LogRequestError(r, tokenErr)
		w.WriteHeader(http.StatusInternalServerError)
		common.WriteWithLogging(r, w, common.GetErrorJson(tokenErr), env.logger)
		// TODO add info that user has been successfully saved
		return
	}

	env.logger.LogRequestSuccess(r)
	common.WriteWithLogging(r, w, common.GetDataJson(tokenString), env.logger)
}

func (env *Env) UserGetSelfInfo(w http.ResponseWriter, r *http.Request) {
	env.logger.LogRequestStart(r)
	var user, code, parseErr = parseUser(r)
	if parseErr != nil {
		env.logger.LogRequestError(r, parseErr)
		w.WriteHeader(code)
		common.WriteWithLogging(r, w, common.GetErrorJson(parseErr), env.logger)
		return
	}

	var exists, existsErr = env.userDAO.ExistsByLogin(user.Login)
	if existsErr != nil {
		env.logger.LogRequestError(r, existsErr)
		w.WriteHeader(http.StatusInternalServerError)
		common.WriteWithLogging(r, w, common.GetErrorJson(existsErr), env.logger)
		return
	}
	if !exists {
		var err = errors.New("not found")
		env.logger.LogRequestError(r, err)
		w.WriteHeader(http.StatusNotFound)
		common.WriteWithLogging(r, w, common.GetErrorJson(err), env.logger)
		return
	}

	var dbUser, dbErr = env.userDAO.GetUserByLogin(user.Login)
	if dbErr != nil {
		env.logger.LogRequestError(r, dbErr)
		w.WriteHeader(http.StatusInternalServerError)
		common.WriteWithLogging(r, w, common.GetErrorJson(dbErr), env.logger)
		return
	}

	env.logger.LogRequestSuccess(r)
	common.WriteWithLogging(r, w, common.GetDataJson(dbUser), env.logger)
}

func (env *Env) generateTokenString(id int, login string) (string, error) {
	var token = jwt.New(jwt.SigningMethodHS256)
	var claims = token.Claims.(jwt.MapClaims)

	claims[idStr] = id
	claims[loginStr] = login
	claims[expStr] = time.Now().Add(time.Hour * 24 * time.Duration(env.conf.Auth.ExpireDays)).Unix()

	var tokenKey = env.conf.Auth.GetTokenKey()
	return token.SignedString(tokenKey)
}

func parseUser(r *http.Request) (*model.User, int, error) {
	var body, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	if err := r.Body.Close(); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	var user = new(model.User)
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, http.StatusBadRequest, err
	}

	if user.Login == "" || user.Password == "" {
		return nil, http.StatusBadRequest, errors.New("Empty user")
	}

	return user, http.StatusOK, nil
}
