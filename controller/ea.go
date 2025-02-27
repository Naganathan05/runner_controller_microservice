package controller

import (
	"evolve/util"
	"net/http"
)

func CreateEA(res http.ResponseWriter, req *http.Request) {
	var logger = util.NewLogger()
	logger.Info("CreateEA API called.")

	data, err := util.Body(req)
	if err != nil {
		util.JSONResponse(res, http.StatusBadRequest, err.Error(), nil)
		return
	}

	util.JSONResponse(res, http.StatusOK, "It works! 👍🏻", data)
}
