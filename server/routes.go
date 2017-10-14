package server

import (
	"github.com/gorilla/mux"
	"net/http"
)

func GetRouter(env *Env) *mux.Router {
	var router = mux.NewRouter()
	router.HandleFunc("/api/v1/auth/sign_up/", env.UserRegisterPost).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/auth/sign_in/", env.UserSignInPost).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/pos/get_neighbours/", env.UserGetNeighboursGet).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/pos/save_position/", env.UserSavePositionPost).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/comm/create_request/", env.CreateRequest).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/comm/get_requests/", env.GetRequests).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/comm/update_request/", env.UpdateRequest).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/comm/get_new_requests/", env.GetNewRequestsEvents).Methods(http.MethodGet)

	return router
}
