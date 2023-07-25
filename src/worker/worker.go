package worker

import (
	"text2phenotype.com/fdl/logger"
	"text2phenotype.com/fdl/pipeline"
	"text2phenotype.com/fdl/rmq"
	"text2phenotype.com/fdl/s3client"
	"text2phenotype.com/fdl/tasks"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
)

type Config struct {
	TaskMaxRetries int `envconfig:"MDL_COMN_RETRY_TASK_COUNT_MAX" default:"3"`
}

type Worker struct {
	config    Config
	redis     redisTransactions
	s3        s3Transactions
	rmq       rmqTransactions
	fdlLogger *zerolog.Logger
	ppln      pipeline.Pipeline
}

func New(ppln pipeline.Pipeline) (*Worker, error) {
	fdlLogger := logger.NewLogger("Worker")

	var config Config
	if err := envconfig.Process("", &config); err != nil {
		fdlLogger.Error().Err(err).Msg("Could not read config")
		return nil, err
	}

	worker := Worker{
		config:    config,
		fdlLogger: &fdlLogger,
		ppln:      ppln,
	}
	if err := worker.refreshRMQClient(); err != nil {
		fdlLogger.Error().Err(err).Msg("Could not create RMQ client")
		return nil, err
	}
	if err := worker.refreshS3Client(); err != nil {
		fdlLogger.Error().Err(err).Msg("Could not create S3 client")
		return nil, err
	}
	if err := worker.refreshRedisClients(); err != nil {
		fdlLogger.Error().Err(err).Msg("Could not create Redis client")
		return nil, err
	}
	return &worker, nil
}

func (worker *Worker) StartWorker() error {
	defer worker.Close()
	for {
		select {
		case delivery, ok := <-worker.rmq.getDeliveriesCh():
			if ok {
				go worker.processMessage(&delivery)
				continue
			}
			worker.fdlLogger.Error().Msg("Deliveries channel closed, trying to refresh RMQ client")
			if err := worker.refreshRMQClient(); err != nil {
				return fmt.Errorf(
					"rmq deliveries channel has been closed and refresh returned error: %w",
					err,
				)
			}
		case rmqErr := <-worker.rmq.getRespChanErrorsCh():
			if rmqErr == nil {
				continue
			}
			worker.fdlLogger.Err(rmqErr).Msg("Response connection received error, trying to refresh RMQ client")
			if err := worker.refreshRMQClient(); err != nil {
				return fmt.Errorf(
					"response connection received error and refresh failed with: %w",
					err,
				)
			}
		case rmqErr := <-worker.rmq.getReqChanErrorsCh():
			if rmqErr == nil {
				continue
			}
			worker.fdlLogger.Err(rmqErr).Msg("Request connection received error, trying to refresh RMQ client")
			if err := worker.refreshRMQClient(); err != nil {
				return fmt.Errorf(
					"request connection received error and refresh failed with: %w",
					err,
				)
			}
		}
	}
}

func (worker *Worker) Close() {
	worker.redis.close()
	worker.s3.close()
	worker.rmq.close()
}

func (worker *Worker) refreshRedisClients() error {
	worker.fdlLogger.Info().Msg("Refreshing Redis client")
	if oldClient := worker.redis; oldClient != nil {
		defer oldClient.close()
	}
	tasksClient, err := tasks.NewClient()
	if err != nil {
		worker.fdlLogger.Err(err).Msg("Failed to refresh Redis client")
		return err
	}
	worker.redis = &redisClientWrapper{&tasksClient}
	worker.fdlLogger.Info().Msg("Refreshed Redis client")
	return nil
}

func (worker *Worker) refreshRMQClient() error {
	worker.fdlLogger.Info().Msg("Refreshing RMQ client")
	if oldClient := worker.rmq; oldClient != nil {
		defer oldClient.close()
	}
	rmqClient, err := rmq.NewClient()
	if err != nil {
		worker.fdlLogger.Err(err).Msg("Failed to refresh RMQ client")
		return err
	}
	worker.rmq = &rmqClientWrapper{rmqClient}
	worker.fdlLogger.Info().Msg("Refreshed RMQ client")
	return nil
}

func (worker *Worker) refreshS3Client() error {
	worker.fdlLogger.Info().Msg("Refreshing S3 client")
	if oldClient := worker.s3; oldClient != nil {
		defer oldClient.close()
	}
	s3Client, err := s3client.New()
	if err != nil {
		worker.fdlLogger.Err(err).Msg("Failed to refresh S3 client")
		return err
	}
	worker.s3 = &s3ClientWrapper{s3Client}
	worker.fdlLogger.Info().Msg("Refreshed S3 client")
	return nil
}
