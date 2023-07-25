package worker

import (
	"text2phenotype.com/fdl/tasks"
	"fmt"
)

type redisTransactions interface {
	getChunkTask(redisKey string) (*tasks.ChunkTask, error)
	getJobTask(task *Task) (*tasks.JobTask, error)
	getDocTask(task *Task) (*tasks.DocumentTaskCached, error)
	onTaskStarted(task *Task) error
	onTaskCancelled(task *Task, errorMessages ...string) error
	onTaskExceededRetries(task *Task, maxRetries int) error
	onTaskFailedWithError(task *Task, err error) error
	onTaskComplete(task *Task) error
	close()
}

type redisClientWrapper struct {
	tasksClient *tasks.Client
}

func (wrapper *redisClientWrapper) close() {
	wrapper.tasksClient.Close()
}

func (wrapper *redisClientWrapper) onTaskStarted(task *Task) error {
	err := wrapper.tasksClient.Chunks.Update(task.redisKey, func(task *tasks.ChunkTask) {
		task.TaskStatuses.FDL.Status = tasks.TaskStatusStarted
		task.TaskStatuses.FDL.Attempts += 1
		task.TaskStatuses.FDL.StartedAt = getFormattedNow()
		task.TaskStatuses.FDL.CompletedAt = nil
	})
	return err
}

func (wrapper *redisClientWrapper) onTaskCancelled(task *Task, errorMessages ...string) error {
	err := wrapper.tasksClient.Chunks.Update(task.redisKey, func(chunkTask *tasks.ChunkTask) {
		chunkTask.TaskStatuses.FDL.Status = tasks.TaskStatusCanceled
		chunkTask.TaskStatuses.FDL.StartedAt = getFormattedNow()
		chunkTask.TaskStatuses.FDL.CompletedAt = getFormattedNow()
		chunkTask.TaskStatuses.FDL.Attempts += 1
		chunkTask.TaskStatuses.FDL.ErrorMessages = append(
			chunkTask.TaskStatuses.FDL.ErrorMessages,
			errorMessages...,
		)
	})
	return err
}

func (wrapper *redisClientWrapper) onTaskExceededRetries(task *Task, maxRetries int) error {
	err := wrapper.tasksClient.Documents.Update(task.chunkTask.DocID, func(docTask *tasks.DocumentTask) {
		docTask.FailedTasks = append(docTask.FailedTasks, "fdl")
		docTask.FailedChunks[task.redisKey] = append(docTask.FailedChunks[task.redisKey], "fdl")
	})
	if err != nil {
		return err
	}
	err = wrapper.tasksClient.Chunks.Update(task.redisKey, func(chunkTask *tasks.ChunkTask) {
		chunkTask.TaskStatuses.FDL.Status = tasks.TaskStatusCompletedFailure
		chunkTask.TaskStatuses.FDL.StartedAt = getFormattedNow()
		chunkTask.TaskStatuses.FDL.CompletedAt = getFormattedNow()
		chunkTask.TaskStatuses.FDL.Attempts += 1
		chunkTask.TaskStatuses.FDL.ErrorMessages = append(
			chunkTask.TaskStatuses.FDL.ErrorMessages,
			fmt.Sprintf(
				"Task has exceeded retries. (Attempts: %d, max retries: %d )",
				chunkTask.TaskStatuses.FDL.Attempts,
				maxRetries,
			),
		)
	})
	return err
}

func (wrapper *redisClientWrapper) onTaskFailedWithError(task *Task, err error) error {
	return wrapper.tasksClient.Chunks.Update(task.redisKey, func(chunkTask *tasks.ChunkTask) {
		chunkTask.TaskStatuses.FDL.Status = tasks.TaskStatusFailed
		chunkTask.TaskStatuses.FDL.CompletedAt = getFormattedNow()
		chunkTask.TaskStatuses.FDL.ErrorMessages = append(chunkTask.TaskStatuses.FDL.ErrorMessages, err.Error())
	})
}

func (wrapper *redisClientWrapper) onTaskComplete(task *Task) error {
	return wrapper.tasksClient.Chunks.Update(task.redisKey, func(chunkTask *tasks.ChunkTask) {
		if !chunkTask.TaskStatuses.FDL.Status.Complete() {
			chunkTask.TaskStatuses.FDL.Status = tasks.TaskStatusCompletedSuccess
		}
		chunkTask.TaskStatuses.FDL.CompletedAt = getFormattedNow()
		chunkTask.TaskStatuses.FDL.ResultsFileKey = getResultsFileKey(task)
	})
}

func (wrapper *redisClientWrapper) getChunkTask(redisKey string) (*tasks.ChunkTask, error) {
	return wrapper.tasksClient.Chunks.Get(redisKey)
}

func (wrapper *redisClientWrapper) getJobTask(task *Task) (*tasks.JobTask, error) {
	return wrapper.tasksClient.Jobs.GetCached(task.chunkTask.JobID)
}

func (wrapper *redisClientWrapper) getDocTask(task *Task) (*tasks.DocumentTaskCached, error) {
	return wrapper.tasksClient.Documents.GetCached(task.chunkTask.DocID)
}
