package worker

import (
	"fmt"
	"path"
	"time"
)

func getResultsFileKey(task *Task) string {
	return path.Join(
		"processed",
		"documents",
		task.chunkTask.DocID,
		"chunks",
		task.redisKey,
		fmt.Sprintf("%s.fdl_results.json", task.redisKey),
	)
}

const RFC3339Micro = "2006-01-02T15:04:05.000000-07:00"

func getFormattedNow() *string {
	now := time.Now().UTC().Format(RFC3339Micro)
	return &now
}
