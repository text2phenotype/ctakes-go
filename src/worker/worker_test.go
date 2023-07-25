package worker

import (
	"text2phenotype.com/fdl/logger"
	"text2phenotype.com/fdl/tasks"
	"github.com/streadway/amqp"
	"reflect"
	"testing"
)

type mockedClientsConfig struct {
	rmqMockConfig
	redisMockConfig
	s3MockConfig
	pipelineMockConfig
}

type mockedClients struct {
	redis    *redisMock
	rmq      *rmqMock
	s3       *s3Mock
	pipeline *pipelineMock
}

type methodsCalls struct {
	redis    redisMockCalls
	rmq      rmqMockCalls
	s3       s3MockCalls
	pipeline pipelineCall
}

func testConfiguration(t *testing.T, config mockedClientsConfig, expectedCalls methodsCalls) {
	worker, mocks := configureWorker(config)
	worker.processMessage(&amqp.Delivery{
		Body: []byte("{}"),
	})
	calls := methodsCalls{
		redis:    mocks.redis.calls,
		rmq:      mocks.rmq.calls,
		s3:       mocks.s3.calls,
		pipeline: mocks.pipeline.calls,
	}
	if !reflect.DeepEqual(calls, expectedCalls) {
		t.Errorf("Got unexpected called methods set.\nExpected:\n%+v\nGot:\n%+v", expectedCalls, calls)
	}
}

func configureWorker(config mockedClientsConfig) (*Worker, *mockedClients) {
	redis := &redisMock{config: config.redisMockConfig}
	s3 := &s3Mock{config: config.s3MockConfig}
	rmq := &rmqMock{config: config.rmqMockConfig}
	pplnMock := getPipelineMock(config.pipelineMockConfig)

	fdlLogger := logger.NewLogger("Test Worker")

	return &Worker{
			config:    Config{3},
			redis:     redis,
			s3:        s3,
			rmq:       rmq,
			fdlLogger: &fdlLogger,
			ppln:      pplnMock.ppln,
		}, &mockedClients{
			redis:    redis,
			rmq:      rmq,
			s3:       s3,
			pipeline: pplnMock,
		}
}

func TestWorker(t *testing.T) {
	t.Run("Successful", testSuccessfulTask)
	t.Run("Successful with job_task.stop_docs_on_failure == True", testSuccessfulTaskWithDocCheck)
	t.Run("Failed to get Chunk task", testGetChunkTaskFailed)
	t.Run("Failed to get Job task", testGetJobTaskFailed)
	t.Run("Failed to get Doc task", testGetDocTaskFailed)
	t.Run("Already complete with success", testAlreadyCompletedSuccessfully)
	t.Run("Already complete with failure", testAlreadyCompletedWithFailure)
	t.Run("User cancelled", testUserCancelled)
	t.Run("Exceeded attempts", testExceededAttempts)
	t.Run("Cancelled because other worker already failed", testCancelledBecauseOfOtherWorkerFailure)
	t.Run("Failed to update task in onTaskStarted", testFailedToUpdateOnTaskStarted)
	t.Run("Failed to load data from S3", testFailedToFetchFromS3)
	t.Run("Failed due to pipeline error", testPipelineError)
	t.Run("Failed to update task in onTaskFailedWithError", testFailedToUpdateOnTaskFailedWithError)
	t.Run("Failed to update task in onTaskComplete", testFailedToUpdateOnTaskComplete)
	t.Run("Failed to save result to S3", testFailedToSaveToS3)
	t.Run("Failed to acknowledge delivery", testFailedAckDelivery)
	t.Run("Failed to ping sequencer", testFailedPingSequencer)
}

func testSuccessfulTask(t *testing.T) {
	testConfiguration(
		t,
		mockedClientsConfig{},
		methodsCalls{
			redis: redisMockCalls{
				getChunkTask: true, getJobTask: true, onTaskStarted: true, onTaskComplete: true,
			},
			rmq: rmqMockCalls{pingSequencer: true, acknowledgeDelivery: true},
			s3: s3MockCalls{
				getProcessedData: true,
				saveResultsFile:  true,
			},
			pipeline: pipelineCall{true},
		},
	)
}

func testSuccessfulTaskWithDocCheck(t *testing.T) {
	testConfiguration(
		t,
		mockedClientsConfig{
			redisMockConfig: redisMockConfig{
				getJobTask: withValue{returnedValue: tasks.JobTask{StopDocumentsOnFailure: true}},
			},
		},
		methodsCalls{
			redis: redisMockCalls{
				getChunkTask: true, getJobTask: true, getDocTask: true, onTaskStarted: true, onTaskComplete: true,
			},
			rmq: rmqMockCalls{pingSequencer: true, acknowledgeDelivery: true},
			s3: s3MockCalls{
				getProcessedData: true,
				saveResultsFile:  true,
			},
			pipeline: pipelineCall{true},
		},
	)
}

func testAlreadyCompletedSuccessfully(t *testing.T) {
	testConfiguration(
		t,
		mockedClientsConfig{
			redisMockConfig: redisMockConfig{
				getChunkTask: withValue{
					returnedValue: tasks.ChunkTask{
						TaskStatuses: tasks.ChunkTaskStatuses{FDL: tasks.ChunkTaskInfo{Status: tasks.TaskStatusCompletedSuccess}},
					},
				},
			},
		},
		methodsCalls{
			redis: redisMockCalls{getChunkTask: true},
			rmq:   rmqMockCalls{pingSequencer: true, acknowledgeDelivery: true},
		},
	)
}

func testAlreadyCompletedWithFailure(t *testing.T) {
	testConfiguration(
		t,
		mockedClientsConfig{
			redisMockConfig: redisMockConfig{
				getChunkTask: withValue{
					returnedValue: tasks.ChunkTask{
						TaskStatuses: tasks.ChunkTaskStatuses{FDL: tasks.ChunkTaskInfo{Status: tasks.TaskStatusCompletedFailure}},
					},
				},
			},
		},
		methodsCalls{
			redis: redisMockCalls{getChunkTask: true},
			rmq:   rmqMockCalls{pingSequencer: true, acknowledgeDelivery: true},
		},
	)
}

func testUserCancelled(t *testing.T) {
	testConfiguration(
		t,
		mockedClientsConfig{
			redisMockConfig: redisMockConfig{
				getJobTask: withValue{returnedValue: tasks.JobTask{UserCanceled: true}},
			},
		},
		methodsCalls{
			redis: redisMockCalls{getChunkTask: true, getJobTask: true, onTaskCancelled: true},
			rmq:   rmqMockCalls{pingSequencer: true, acknowledgeDelivery: true},
		},
	)
}

func testExceededAttempts(t *testing.T) {
	testConfiguration(
		t,
		mockedClientsConfig{
			redisMockConfig: redisMockConfig{
				getChunkTask: withValue{
					returnedValue: tasks.ChunkTask{
						TaskStatuses: tasks.ChunkTaskStatuses{FDL: tasks.ChunkTaskInfo{Attempts: 3}},
					},
				},
			},
		},
		methodsCalls{
			redis: redisMockCalls{getChunkTask: true, getJobTask: true, onTaskExceededRetries: true},
			rmq:   rmqMockCalls{pingSequencer: true, acknowledgeDelivery: true},
		},
	)
}

func testCancelledBecauseOfOtherWorkerFailure(t *testing.T) {
	testConfiguration(
		t,
		mockedClientsConfig{
			redisMockConfig: redisMockConfig{
				getJobTask: withValue{
					returnedValue: tasks.JobTask{
						StopDocumentsOnFailure: true,
					},
				},
				getDocTask: withValue{
					returnedValue: tasks.DocumentTaskCached{
						FailedTasks: []string{"some other task"},
					},
				},
			},
		},
		methodsCalls{
			redis: redisMockCalls{getChunkTask: true, getJobTask: true, getDocTask: true, onTaskCancelled: true},
			rmq:   rmqMockCalls{pingSequencer: true, acknowledgeDelivery: true},
		},
	)
}

func testFailedToUpdateOnTaskStarted(t *testing.T) {
	testConfiguration(
		t,
		mockedClientsConfig{
			redisMockConfig: redisMockConfig{onTaskStarted: failingMethod{fail: true}},
		},
		methodsCalls{
			redis: redisMockCalls{
				getChunkTask: true, getJobTask: true, onTaskStarted: true,
			},
			rmq: rmqMockCalls{rejectDelivery: true},
		},
	)
}

func testFailedToUpdateOnTaskComplete(t *testing.T) {
	testConfiguration(
		t,
		mockedClientsConfig{
			redisMockConfig: redisMockConfig{onTaskComplete: failingMethod{fail: true}},
		},
		methodsCalls{
			redis: redisMockCalls{
				getChunkTask: true, getJobTask: true, onTaskStarted: true, onTaskComplete: true,
			},
			rmq: rmqMockCalls{rejectDelivery: true},
			s3: s3MockCalls{
				getProcessedData: true,
				saveResultsFile:  true,
			},
			pipeline: pipelineCall{pipeline: true},
		},
	)
}

func testFailedToFetchFromS3(t *testing.T) {
	testConfiguration(
		t,
		mockedClientsConfig{
			s3MockConfig: s3MockConfig{getProcessedData: withValue{fail: true}},
		},
		methodsCalls{
			redis: redisMockCalls{
				getChunkTask: true, getJobTask: true, onTaskStarted: true, onTaskFailedWithError: true,
			},
			rmq: rmqMockCalls{pingSequencer: true, acknowledgeDelivery: true},
			s3: s3MockCalls{
				getProcessedData: true,
			},
		},
	)
}

func testPipelineError(t *testing.T) {
	testConfiguration(
		t,
		mockedClientsConfig{
			pipelineMockConfig: pipelineMockConfig{fail: true},
		},
		methodsCalls{
			redis: redisMockCalls{
				getChunkTask: true, getJobTask: true, onTaskStarted: true, onTaskFailedWithError: true,
			},
			rmq: rmqMockCalls{pingSequencer: true, acknowledgeDelivery: true},
			s3: s3MockCalls{
				getProcessedData: true,
			},
			pipeline: pipelineCall{true},
		},
	)
}

func testFailedToUpdateOnTaskFailedWithError(t *testing.T) {
	testConfiguration(
		t,
		mockedClientsConfig{
			pipelineMockConfig: pipelineMockConfig{fail: true},
			redisMockConfig:    redisMockConfig{onTaskFailedWithError: failingMethod{fail: true}},
		},
		methodsCalls{
			redis: redisMockCalls{
				getChunkTask: true, getJobTask: true, onTaskStarted: true, onTaskFailedWithError: true,
			},
			rmq: rmqMockCalls{rejectDelivery: true},
			s3: s3MockCalls{
				getProcessedData: true,
			},
			pipeline: pipelineCall{true},
		},
	)
}

func testFailedToSaveToS3(t *testing.T) {
	testConfiguration(
		t,
		mockedClientsConfig{
			s3MockConfig: s3MockConfig{saveResultsFile: failingMethod{fail: true}},
		},
		methodsCalls{
			redis: redisMockCalls{
				getChunkTask: true, getJobTask: true, onTaskStarted: true, onTaskFailedWithError: true,
			},
			rmq: rmqMockCalls{pingSequencer: true, acknowledgeDelivery: true},
			s3: s3MockCalls{
				getProcessedData: true,
				saveResultsFile:  true,
			},
			pipeline: pipelineCall{true},
		},
	)
}

func testFailedAckDelivery(t *testing.T) {
	testConfiguration(
		t,
		mockedClientsConfig{
			rmqMockConfig: rmqMockConfig{acknowledgeDelivery: failingMethod{fail: true}},
		},
		methodsCalls{
			redis: redisMockCalls{
				getChunkTask: true, getJobTask: true, onTaskStarted: true, onTaskComplete: true,
			},
			rmq: rmqMockCalls{pingSequencer: true, acknowledgeDelivery: true},
			s3: s3MockCalls{
				getProcessedData: true,
				saveResultsFile:  true,
			},
			pipeline: pipelineCall{true},
		},
	)
}

func testFailedPingSequencer(t *testing.T) {
	testConfiguration(
		t,
		mockedClientsConfig{
			rmqMockConfig: rmqMockConfig{pingSequencer: failingMethod{fail: true}},
		},
		methodsCalls{
			redis: redisMockCalls{
				getChunkTask: true, getJobTask: true, onTaskStarted: true, onTaskComplete: true,
			},
			rmq: rmqMockCalls{pingSequencer: true, rejectDelivery: true},
			s3: s3MockCalls{
				getProcessedData: true,
				saveResultsFile:  true,
			},
			pipeline: pipelineCall{true},
		},
	)
}

func testGetChunkTaskFailed(t *testing.T) {
	testConfiguration(
		t,
		mockedClientsConfig{
			redisMockConfig: redisMockConfig{getChunkTask: withValue{fail: true}},
		},
		methodsCalls{
			redis: redisMockCalls{
				getChunkTask: true,
			},
			rmq: rmqMockCalls{rejectDelivery: true},
		},
	)
}

func testGetJobTaskFailed(t *testing.T) {
	testConfiguration(
		t,
		mockedClientsConfig{
			redisMockConfig: redisMockConfig{getJobTask: withValue{fail: true}},
		},
		methodsCalls{
			redis: redisMockCalls{
				getChunkTask: true, getJobTask: true,
			},
			rmq: rmqMockCalls{rejectDelivery: true},
		},
	)
}

func testGetDocTaskFailed(t *testing.T) {
	testConfiguration(
		t,
		mockedClientsConfig{
			redisMockConfig: redisMockConfig{
				getJobTask: withValue{returnedValue: tasks.JobTask{StopDocumentsOnFailure: true}},
				getDocTask: withValue{fail: true},
			},
		},
		methodsCalls{
			redis: redisMockCalls{
				getChunkTask: true, getJobTask: true, getDocTask: true,
			},
			rmq: rmqMockCalls{rejectDelivery: true},
		},
	)
}
