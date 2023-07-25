package tasks

import (
	"text2phenotype.com/fdl/redis"
	"text2phenotype.com/fdl/utils/maps"
)

const JobsDB redis.DB = 1

type JobTask struct {
	maps.BaseDocument
	UserCanceled           bool `json:"user_canceled"`
	StopDocumentsOnFailure bool `json:"stop_documents_on_failure"`
}

type JobTasks struct {
	client redis.Client
}

func (tasks JobTasks) GetCached(redisKey string) (*JobTask, error) {
	var task JobTask
	key := cachedPropertiesKey(redisKey)
	err := tasks.client.GetPartialDocument(key, &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}
