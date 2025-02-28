package controller

import (
	"evolve/modules"
	"evolve/util"
	"net/http"
	"os"
)

func CreateML(res http.ResponseWriter, req *http.Request) {
	var logger = util.NewLogger()
	logger.Info("CreateML API called.")

	// TODO: Add Auth.

	data, err := util.Body(req)
	if err != nil {
		util.JSONResponse(res, http.StatusBadRequest, err.Error(), nil)
		return
	}

	ml, err := modules.MLFromJSON(data)
	if err != nil {
		util.JSONResponse(res, http.StatusBadRequest, err.Error(), nil)
		return
	}

	code, err := ml.Code()
	if err != nil {
		util.JSONResponse(res, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// TODO: Write code for MinIO and remove this.
	os.Mkdir("code", 0755)
	os.WriteFile("code/ml.py", []byte(code), 0644)

	util.JSONResponse(res, http.StatusOK, "It works! 👍🏻", data)
}
