package server

import (
	"github.com/gorilla/mux"
	"net/http"
)

func GetRouter(env *Env) *mux.Router {
	var router = mux.NewRouter()
	router.HandleFunc("/api/v1/auth/register", env.UserRegisterPost).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/auth/login", env.UserSignInPost).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/user/self", env.UserGetSelfInfo).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/user/position/neighbours", env.UserGetNeighboursGet).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/user/position/save", env.UserSavePositionPost).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/user/position/neighbour/{id}", env.UserGetPositionById).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/user/request/create", env.CreateRequest).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/user/request/all", env.GetRequests).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/user/request/update", env.UpdateRequest).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/user/request/new", env.GetNewRequestsEvents).Methods(http.MethodGet)

	return router
}
