package modules

import (
	"context"
	"evolve/db/connection"
	"evolve/util"
	"fmt"
	"time"
)

func UserRuns(ctx context.Context, userID string, logger *util.Logger) ([]map[string]string, error) {
	db, err := connection.PoolConn(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("UserRuns: %s", err.Error()))
		return nil, fmt.Errorf("something went wrong")
	}

	var runIDs []string
	rows, err := db.Query(ctx, "SELECT runID FROM access WHERE userID = $1", userID)
	if err != nil {
		logger.Error(fmt.Sprintf("UserRuns.db.Query: %s", err.Error()))
		return nil, fmt.Errorf("something went wrong")
	}

	for rows.Next() {
		var runID string
		err = rows.Scan(&runID)
		if err != nil {
			logger.Error(fmt.Sprintf("UserRuns.rows.Scan: %s", err.Error()))
			return nil, fmt.Errorf("something went wrong")
		}
		runIDs = append(runIDs, runID)
	}

	if len(runIDs) == 0 {
		return make([]map[string]string, 0), nil
	}

	// logger.Info(fmt.Sprintf("RunIDs: %s", runIDs))

	rows, err = db.Query(ctx, "SELECT * FROM run WHERE id = ANY($1)", runIDs)
	if err != nil {
		logger.Error(fmt.Sprintf("UserRuns.db.Query: %s", err.Error()))
		return nil, fmt.Errorf("something went wrong")
	}

	runs := []map[string]string{}
	for rows.Next() {
		var id string
		var name string
		var description string
		var status string
		var runType string
		var command string
		var createdBy string
		var createdAt time.Time
		var updatedAt time.Time

		err := rows.Scan(&id, &name, &description, &status, &runType, &command, &createdBy, &createdAt, &updatedAt)
		if err != nil {
			logger.Error(fmt.Sprintf("UserRuns.rows.Scan: %s", err.Error()))
			return nil, fmt.Errorf("something went wrong")
		}

		run := map[string]string{
			"id":          id,
			"name":        name,
			"description": description,
			"status":      status,
			"type":        runType,
			"command":     command,
			"createdAt":   createdAt.Local().String(),
			"updatedAt":   updatedAt.Local().String(),
		}

		if createdBy != userID {
			run["isShared"] = "true"
			run["sharedBy"] = createdBy
		} else {
			run["isShared"] = "false"
			run["createdBy"] = createdBy
		}

		runs = append(runs, run)
	}

	// logger.Info(fmt.Sprintf("Runs: %s", runs))

	return runs, nil
}
