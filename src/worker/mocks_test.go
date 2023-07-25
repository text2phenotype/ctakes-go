package worker

import (
	"text2phenotype.com/fdl/pipeline"
	"text2phenotype.com/fdl/tasks"
	"errors"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

type failingMethod struct {
	fail bool
}

type withValue struct {
	fail          bool
	returnedValue interface{}
}

type pipelineMock struct {
	ppln   pipeline.Pipeline
	config pipelineMockConfig
	calls  pipelineCall
}

type pipelineMockConfig struct {
	fail   bool
	result string
}

type pipelineCall struct {
	pipeline bool
}

type redisMock struct {
	config redisMockConfig
	calls  redisMockCalls
}

type redisMockConfig struct {
	getChunkTask          withValue
	getJobTask            withValue
	getDocTask            withValue
	onTaskCancelled       failingMethod
	onTaskStarted         failingMethod
	onTaskExceededRetries failingMethod
	onTaskFailedWithError failingMethod
	onTaskComplete        failingMethod
}

type redisMockCalls struct {
	getChunkTask          bool
	getJobTask            bool
	getDocTask            bool
	onTaskCancelled       bool
	onTaskStarted         bool
	onTaskExceededRetries bool
	onTaskFailedWithError bool
	onTaskComplete        bool
}

type rmqMock struct {
	config rmqMockConfig
	calls  rmqMockCalls
}

type rmqMockConfig struct {
	pingSequencer       failingMethod
	acknowledgeDelivery failingMethod
}

type rmqMockCalls struct {
	pingSequencer       bool
	acknowledgeDelivery bool
	rejectDelivery      bool
}

type s3Mock struct {
	config s3MockConfig
	calls  s3MockCalls
}

type s3MockConfig struct {
	getProcessedData withValue
	saveResultsFile  failingMethod
}

type s3MockCalls struct {
	getProcessedData bool
	saveResultsFile  bool
}

func (mock s3Mock) close() {}

func (mock *rmqMock) close() {}

func (mock *redisMock) close() {}

func getPipelineMock(config pipelineMockConfig) *pipelineMock {
	var mock pipelineMock
	if config.fail {
		mock.ppln = func(request pipeline.Request) <-chan string {
			mock.calls.pipeline = true
			ch := make(chan string)
			close(ch)
			return ch
		}
	} else {
		mock.ppln = func(request pipeline.Request) <-chan string {
			mock.calls.pipeline = true
			ch := make(chan string, 1)
			ch <- mock.config.result
			close(ch)
			return ch
		}
	}
	return &mock
}

func (mock *redisMock) getChunkTask(redisKey string) (*tasks.ChunkTask, error) {
	mock.calls.getChunkTask = true
	if mock.config.getChunkTask.fail {
		return nil, errors.New("failed to get chunk task")
	}
	switch mock.config.getChunkTask.returnedValue.(type) {
	case tasks.ChunkTask:
		task := mock.config.getChunkTask.returnedValue.(tasks.ChunkTask)
		return &task, nil
	default:
		return &tasks.ChunkTask{}, nil
	}

}

func (mock *redisMock) getJobTask(task *Task) (*tasks.JobTask, error) {
	mock.calls.getJobTask = true
	if mock.config.getJobTask.fail {
		return nil, errors.New("failed to get job task")
	}
	switch mock.config.getJobTask.returnedValue.(type) {
	case tasks.JobTask:
		jobTask := mock.config.getJobTask.returnedValue.(tasks.JobTask)
		return &jobTask, nil
	default:
		return &tasks.JobTask{}, nil
	}
}

func (mock *redisMock) getDocTask(task *Task) (*tasks.DocumentTaskCached, error) {
	mock.calls.getDocTask = true
	if mock.config.getDocTask.fail {
		return nil, errors.New("failed to get doc task")
	}
	switch mock.config.getDocTask.returnedValue.(type) {
	case tasks.DocumentTaskCached:
		documentTaskCached := mock.config.getDocTask.returnedValue.(tasks.DocumentTaskCached)
		return &documentTaskCached, nil
	default:
		return &tasks.DocumentTaskCached{}, nil
	}
}

func (mock *redisMock) onTaskStarted(task *Task) error {
	mock.calls.onTaskStarted = true
	if mock.config.onTaskStarted.fail {
		return errors.New("failed to update chunk task on start")
	}
	return nil
}

func (mock *redisMock) onTaskCancelled(task *Task, errorMessages ...string) error {
	mock.calls.onTaskCancelled = true
	if mock.config.onTaskCancelled.fail {
		return errors.New("failed to update chunk task on cancel")
	}
	return nil
}

func (mock *redisMock) onTaskExceededRetries(task *Task, maxRetries int) error {
	mock.calls.onTaskExceededRetries = true
	if mock.config.onTaskExceededRetries.fail {
		return errors.New("failed to update chunk task on exceeded retries")
	}
	return nil
}

func (mock *redisMock) onTaskFailedWithError(task *Task, err error) error {
	mock.calls.onTaskFailedWithError = true
	if mock.config.onTaskFailedWithError.fail {
		return errors.New("failed to update chunk task on fail with error")
	}
	return nil
}

func (mock *redisMock) onTaskComplete(task *Task) error {
	mock.calls.onTaskComplete = true
	if mock.config.onTaskComplete.fail {
		return errors.New("failed to update chunk task on complete")
	}
	return nil
}
func (mock *rmqMock) rejectDelivery(delivery *amqp.Delivery, fdlLogger *zerolog.Logger) {
	mock.calls.rejectDelivery = true
}
func (mock *rmqMock) getDeliveriesCh() <-chan amqp.Delivery {
	return nil
}

func (mock *rmqMock) getReqChanErrorsCh() <-chan *amqp.Error {
	return nil
}

func (mock *rmqMock) getRespChanErrorsCh() <-chan *amqp.Error {
	return nil
}

func (mock *rmqMock) pingSequencer(task *Task, message Message) error {
	mock.calls.pingSequencer = true
	if mock.config.pingSequencer.fail {
		return errors.New("failed to ping sequencer")
	}
	return nil
}

func (mock *rmqMock) acknowledgeDelivery(delivery *amqp.Delivery) error {
	mock.calls.acknowledgeDelivery = true
	if mock.config.acknowledgeDelivery.fail {
		return errors.New("failed to acknowledge delivery")
	}
	return nil
}

func (mock *s3Mock) getProcessedData(task *Task) ([]byte, error) {
	mock.calls.getProcessedData = true
	if mock.config.getProcessedData.fail {
		return nil, errors.New("mock: failed to load from s3")
	}
	switch mock.config.getProcessedData.returnedValue.(type) {
	case []byte:
		return mock.config.getProcessedData.returnedValue.([]byte), nil
	default:
		return []byte("some input"), nil
	}
}

func (mock *s3Mock) saveResultsFile(task *Task, result string) error {
	mock.calls.saveResultsFile = true
	if mock.config.saveResultsFile.fail {
		return errors.New("failed to upload results")
	}
	return nil
}
