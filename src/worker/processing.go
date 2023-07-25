package worker

import (
	"text2phenotype.com/fdl/pipeline"
	"text2phenotype.com/fdl/tasks"
	"text2phenotype.com/fdl/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

type Message struct {
	WorkType string `json:"work_type"`
	RedisKey string `json:"redis_key"`
	Sender   string `json:"sender"`
	Version  string `json:"version"`
}

type Task struct {
	delivery  *amqp.Delivery
	chunkTask *tasks.ChunkTask
	message   *Message
	redisKey  string
	fdlLogger *zerolog.Logger
}

func (worker *Worker) processMessage(delivery *amqp.Delivery) {
	task, err := worker.createTask(delivery)
	rejectLogger := worker.fdlLogger.With().Str("message_id", delivery.MessageId).Logger()
	if err != nil {
		worker.fdlLogger.Err(err).
			Str("message_id", delivery.MessageId).
			Str("tid", string(delivery.Body)).
			Msg("Failed to create task for delivery")
		worker.rmq.rejectDelivery(delivery, &rejectLogger)
		return
	}
	if err = worker.processTask(task); err != nil {
		worker.rmq.rejectDelivery(delivery, &rejectLogger)
		return
	}
	if err = worker.rmq.pingSequencer(task, *task.message); err != nil {
		task.fdlLogger.Err(err).Msg("Got error while sending message to sequencer queue")
		worker.rmq.rejectDelivery(delivery, &rejectLogger)
		return
	}
	if err = worker.rmq.acknowledgeDelivery(delivery); err != nil {
		task.fdlLogger.Err(err).Msg("Failed to acknowledge delivery")
	}
	task.fdlLogger.Info().Msg("Finished processing RMQ message")
}

func (worker *Worker) createTask(delivery *amqp.Delivery) (*Task, error) {
	var message Message
	err := json.Unmarshal(delivery.Body, &message)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal message, got error %w", err)
	}
	chunkTask, err := worker.redis.getChunkTask(message.RedisKey)
	if err != nil {
		return nil, fmt.Errorf("failed to query chunk task for message, got error %w", err)
	}
	taskLogger := worker.fdlLogger.With().Str("tid", message.RedisKey).Logger()
	task := Task{
		delivery:  delivery,
		chunkTask: chunkTask,
		redisKey:  message.RedisKey,
		message:   &message,
		fdlLogger: &taskLogger,
	}
	return &task, nil
}

func (worker *Worker) processTask(task *Task) error {
	shouldPerform, err := worker.shouldPerformTask(task)
	if err != nil {
		task.fdlLogger.Err(err).
			Msg("Got error while trying to decide whether to run task")
		return err
	}
	if !shouldPerform {
		return nil
	}
	if err = worker.redis.onTaskStarted(task); err != nil {
		task.fdlLogger.Err(err).Msg("Failed to update task info")
		return fmt.Errorf("failed to update TaskInfo: %w", err)
	}
	if err = worker.runPipeline(task); err != nil {
		task.fdlLogger.Err(err).Msg("Got error while running pipeline")
		if err = worker.redis.onTaskFailedWithError(task, err); err != nil {
			return err
		}
		return nil
	}
	task.fdlLogger.Info().Msg("Saved results, marking task as complete")
	if err = worker.redis.onTaskComplete(task); err != nil {
		task.fdlLogger.Err(err).Msg("Got error while trying to mark task as complete")
		return err
	}
	return nil
}

func (worker *Worker) runPipeline(task *Task) (err error) {
	defer utils.RecoverWithError(&err)
	task.fdlLogger.Info().Msgf("Processing message from RMQ, attempt # %d", task.chunkTask.TaskStatuses.FDL.Attempts)
	data, err := worker.s3.getProcessedData(task)
	if err != nil {
		task.fdlLogger.Err(err).Caller().Msg("Could not fetch text data from s3")
		return fmt.Errorf("failed fetch data from s3: %w", err)
	}
	request := pipeline.Request{
		Tid:  task.redisKey,
		Text: string(data),
	}
	result, ok := <-worker.ppln(request)
	if !ok {
		task.fdlLogger.Error().Msg("Pipeline channel was closed before returning anything")
		return errors.New("pipeline channel was closed before returning anything")
	}
	task.fdlLogger.Info().Msg("Finished pipeline, saving results to s3")
	if err = worker.s3.saveResultsFile(task, result); err != nil {
		task.fdlLogger.Err(err).Msg("Got error while trying to save results")
		return err
	}
	return nil
}
func (worker *Worker) shouldPerformTask(task *Task) (bool, error) {
	taskInfo := task.chunkTask.TaskStatuses.FDL
	taskLogger := task.fdlLogger

	if taskInfo.Status.Complete() {
		taskLogger.Info().Msg("Task is already done. (might indicate issue acking message with RMQ). Sending back to Sequencer.")
		return false, nil
	}
	taskJob, err := worker.redis.getJobTask(task)
	if err != nil {
		taskLogger.Err(err).Msg("Failed to query job task for chunk task")
		return false, err
	}
	if taskJob.UserCanceled {
		taskLogger.Info().Msg("Job was canceled, no need to perform this task. Sending back to Sequencer.")
		err := worker.redis.onTaskCancelled(task)
		return false, err
	}
	var docTask *tasks.DocumentTaskCached
	if taskJob.StopDocumentsOnFailure {
		docTask, err = worker.redis.getDocTask(task)
		if err != nil {
			return false, err
		}
		if docTask == nil {
			return false, fmt.Errorf("document task not found")
		}
	}
	if taskJob.StopDocumentsOnFailure && len(docTask.FailedTasks) > 0 {
		failedTask := docTask.FailedTasks[0]
		taskLogger.Info().Msgf("Task is not required because the \"%s\" already completed failure "+
			"and document won't be processed successfully. Sending back to Sequencer.", failedTask)
		err := worker.redis.onTaskCancelled(
			task,
			fmt.Sprintf(
				"Task was marked as \"%s\" because of the current document has failed "+
					"in the \"%s\" worker and won't be processed successfully.",
				tasks.TaskStatusCanceled,
				failedTask,
			),
		)
		return false, err
	}
	if taskInfo.Attempts >= worker.config.TaskMaxRetries {
		taskLogger.Info().Msg("FDL task has exceeded retries. Sending back to Sequencer.")
		err = worker.redis.onTaskExceededRetries(task, worker.config.TaskMaxRetries)
		return false, err
	}
	return true, nil
}
