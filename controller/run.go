package controller

import (
	"evolve/modules"
	"evolve/util"
	"fmt"
	"net/http"
)

func UserRuns(res http.ResponseWriter, req *http.Request) {
	var logger = util.NewLogger()
	logger.Info("UserRuns API called.")

	user, err := modules.Auth(req)
	if err != nil {
		util.JSONResponse(res, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	// User has id, role, userName, email & fullName.
	logger.Info(fmt.Sprintf("User: %s", user))

	runs, err := modules.UserRuns(req.Context(), user["id"], logger)
	if err != nil {
		util.JSONResponse(res, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	util.JSONResponse(res, http.StatusOK, "User runs", runs)

}
