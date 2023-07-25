package rmq

import (
	"text2phenotype.com/fdl/logger"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

type Config struct {
	Host                    string `envconfig:"MDL_COMN_RMQ_HOST" required:"true"`
	Port                    string `envconfig:"MDL_COMN_RMQ_PORT" required:"true"`
	Username                string `envconfig:"MDL_COMN_RMQ_USERNAME" required:"true"`
	Password                string `envconfig:"MDL_COMN_RMQ_PASSWORD" required:"true"`
	Exchange                string `envconfig:"MDL_COMN_RMQ_DEFAULT_EXCHANGE" default:"text2phenotype-default-exchange"`
	MaxParallelRequestCount int    `envconfig:"FDL_MQ_MAX_PARALLEL_REQUESTS" default:"5"`
	FDLTaskQueue            string `envconfig:"MDL_COMN_FDL_TASK_QUEUE" required:"true"`
	SequencerTaskQueue      string `envconfig:"MDL_COMN_SEQUENCER_TASK_QUEUE" required:"true"`
}

type Client struct {
	Deliveries     <-chan amqp.Delivery
	ReqChanErrors  <-chan *amqp.Error
	RespChanErrors <-chan *amqp.Error
	config         Config
	reqConn        *amqp.Connection
	respConn       *amqp.Connection
	respChannel    *amqp.Channel
	fdlLogger      *zerolog.Logger
}

func NewClient() (*Client, error) {
	fdlLogger := logger.NewLogger("RMQ client")
	var err error
	var config Config
	if err = envconfig.Process("", &config); err != nil {
		fdlLogger.Error().Err(err).Msg("Could not read env config")
		return nil, err
	}

	url := getURL(config)
	respConn, respChannel, err := setup(url)
	if err != nil {
		return nil, fmt.Errorf("failed connection: %s", err)
	}
	reqConn, reqChannel, err := setup(url)
	if err != nil {
		return nil, fmt.Errorf("failed connection: %s", err)
	}

	q, err := reqChannel.QueueDeclarePassive(
		config.FDLTaskQueue, // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return nil, err
	}
	if err := reqChannel.QueueBind(
		config.FDLTaskQueue,
		config.FDLTaskQueue,
		config.Exchange,
		false,
		nil); err != nil {
		return nil, err
	}
	if err := reqChannel.Qos(config.MaxParallelRequestCount, 0, false); err != nil {
		return nil, fmt.Errorf("qos: %s", err)
	}

	deliveries, err := reqChannel.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("consume deliveries: %s", err)
	}
	reqChanErrors := reqChannel.NotifyClose(make(chan *amqp.Error))
	respChanErrors := respChannel.NotifyClose(make(chan *amqp.Error))

	return &Client{
		Deliveries:     deliveries,
		ReqChanErrors:  reqChanErrors,
		RespChanErrors: respChanErrors,
		config:         config,
		reqConn:        reqConn,
		respConn:       respConn,
		respChannel:    respChannel,
		fdlLogger:      &fdlLogger,
	}, nil
}

func (c *Client) SendMessageToSequencer(msg amqp.Publishing) error {
	return c.respChannel.Publish(
		c.config.Exchange,
		c.config.SequencerTaskQueue,
		false,
		false,
		msg)
}

func (c *Client) Close() {
	_ = c.reqConn.Close()
	_ = c.respConn.Close()
}

func getURL(config Config) string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s", config.Username, config.Password, config.Host, config.Port)
}

func setup(url string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}
	return conn, ch, nil
}
