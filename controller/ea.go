package controller

import (
	"evolve/modules"
	"evolve/util"
	"net/http"
	"os"
)

func CreateEA(res http.ResponseWriter, req *http.Request) {
	var logger = util.NewLogger()
	logger.Info("CreateEA API called.")

	// TODO: Add Auth.

	data, err := util.Body(req)
	if err != nil {
		util.JSONResponse(res, http.StatusBadRequest, err.Error(), nil)
		return
	}

	ea, err := modules.EAFromJSON(data)
	if err != nil {
		util.JSONResponse(res, http.StatusBadRequest, err.Error(), nil)
		return
	}

	code, err := ea.Code()
	if err != nil {
		util.JSONResponse(res, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// TODO: Write code for MinIO and remove this.
	os.Mkdir("code", 0755)
	os.WriteFile("code/ea.py", []byte(code), 0644)

	util.JSONResponse(res, http.StatusOK, "It works! 👍🏻", data)
}
