package tasks

import (
	"text2phenotype.com/fdl/redis"
	"text2phenotype.com/fdl/utils/maps"
)

const ChunksDB redis.DB = 2

type TaskStatus string

const (
	TaskStatusProcessing       TaskStatus = "processing"
	TaskStatusSubmitted        TaskStatus = "submitted"
	TaskStatusStarted          TaskStatus = "started"
	TaskStatusFailed           TaskStatus = "failed"
	TaskStatusCompletedSuccess TaskStatus = "completed - success"
	TaskStatusCompletedFailure TaskStatus = "completed - failure"
	TaskStatusCanceled         TaskStatus = "canceled"
)

func (s TaskStatus) Complete() bool {
	return s == TaskStatusCompletedSuccess || s == TaskStatusCompletedFailure || s == TaskStatusCanceled
}

func (s TaskStatus) Submitted() bool {
	return s == TaskStatusSubmitted || s == TaskStatusStarted || s == TaskStatusProcessing
}

type ChunkTask struct {
	maps.BaseDocument
	DocID        string            `json:"document_id"`
	JobID        string            `json:"job_id"`
	TextFileKey  string            `json:"text_file_key"`
	TaskStatuses ChunkTaskStatuses `json:"task_statuses"`
}

type ChunkTaskStatuses struct {
	FDL ChunkTaskInfo `json:"fdl"`
}

type ChunkTaskInfo struct {
	ResultsFileKey    string     `json:"results_file_key"`
	StartedAt         *string    `json:"started_at"`
	CompletedAt       *string    `json:"completed_at"`
	Attempts          int        `json:"attempts"`
	Status            TaskStatus `json:"status"`
	Dependencies      []string   `json:"dependencies"`
	ModelDependencies []float64  `json:"model_dependencies"`
	ErrorMessages     []string   `json:"error_messages"`
}

type ChunkTasks struct {
	client redis.Client
}

func (tasks ChunkTasks) Get(redisKey string) (*ChunkTask, error) {
	var task ChunkTask
	err := tasks.client.GetPartialDocument(redisKey, &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (tasks ChunkTasks) Update(redisKey string, updateFunc func(task *ChunkTask)) error {
	var task ChunkTask
	return tasks.client.UpdatePartialDocument(redisKey, &task, updateFunc)
}
