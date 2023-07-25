package worker

import (
	"text2phenotype.com/fdl/rmq"
	"encoding/json"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

type rmqTransactions interface {
	pingSequencer(task *Task, message Message) error
	acknowledgeDelivery(delivery *amqp.Delivery) error
	rejectDelivery(delivery *amqp.Delivery, fdlLogger *zerolog.Logger)
	getDeliveriesCh() <-chan amqp.Delivery
	getReqChanErrorsCh() <-chan *amqp.Error
	getRespChanErrorsCh() <-chan *amqp.Error
	close()
}

type rmqClientWrapper struct {
	rmqClient *rmq.Client
}

func (wrapper *rmqClientWrapper) close() {
	wrapper.rmqClient.Close()
}

func (wrapper *rmqClientWrapper) getDeliveriesCh() <-chan amqp.Delivery {
	return wrapper.rmqClient.Deliveries
}

func (wrapper *rmqClientWrapper) getReqChanErrorsCh() <-chan *amqp.Error {
	return wrapper.rmqClient.ReqChanErrors
}

func (wrapper *rmqClientWrapper) getRespChanErrorsCh() <-chan *amqp.Error {
	return wrapper.rmqClient.RespChanErrors
}

func (wrapper *rmqClientWrapper) pingSequencer(task *Task, message Message) error {
	message.Sender = "fdl"
	b, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return wrapper.rmqClient.SendMessageToSequencer(
		amqp.Publishing{
			ContentType: task.delivery.ContentType,
			Body:        b,
		},
	)
}

func (wrapper *rmqClientWrapper) acknowledgeDelivery(delivery *amqp.Delivery) error {
	return delivery.Ack(false)
}

func (wrapper *rmqClientWrapper) rejectDelivery(delivery *amqp.Delivery, fdlLogger *zerolog.Logger) {
	if delivery.Redelivered {
		fdlLogger.Info().Msg("Rejecting delivery as it already has been redelivered")
		err := delivery.Reject(false)
		if err != nil {
			fdlLogger.Err(err).Msg("Failed to reject delivery")
		}
		return
	}
	fdlLogger.Info().Msg("Requeuing delivery as it has not been redelivered yet")
	err := delivery.Reject(true)
	if err != nil {
		fdlLogger.Err(err).Msg("Failed to requeue delivery")
	}
}
