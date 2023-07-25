package redis

import (
	"text2phenotype.com/fdl/utils/maps"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v8"
	"github.com/kelseyhightower/envconfig"
	"time"
)

type DB int
type ReleaseLock func() error
type Error error

type Client struct {
	client         redis.UniversalClient
	lockExpiration time.Duration
}

var ctx = context.Background()

type Config struct {
	LockExpirationSeconds   int     `envconfig:"MDL_COMN_REDIS_LOCK_EXPIRATION" default:"3"`
	Host                    string  `envconfig:"MDL_COMN_REDIS_HOST" required:"true"`
	Port                    string  `envconfig:"MDL_COMN_REDIS_PORT" required:"true"`
	HASentinelPort          string  `envconfig:"MDL_COMN_REDIS_HA_SENTINEL_PORT" default:"26379"`
	HASentinelMasterName    string  `envconfig:"MDL_COMN_REDIS_HA_MASTER_NAME" default:"mymaster"`
	Password                string  `envconfig:"MDL_COMN_REDIS_AUTH_PASSWORD" default:"0"`
	AuthRequired            bool    `envconfig:"MDL_COMN_REDIS_AUTH_REQUIRED" default:"false"`
	HAMode                  bool    `envconfig:"MDL_COMN_REDIS_HA_MODE" default:"false"`
	HASentinelSocketTimeout float32 `envconfig:"MDL_COMN_REDIS_SOCKET_TIMEOUT" default:"0.5"`
}

func NewClient(db DB) (Client, error) {
	cfg, err := readEnvironment()
	if err != nil {
		return Client{}, err
	}
	var client redis.UniversalClient
	if cfg.HAMode {
		client = CreateClusterClient(cfg, db)
	} else {
		client = CreateClient(cfg, db)
	}
	return Client{
		client:         client,
		lockExpiration: time.Duration(cfg.LockExpirationSeconds) * time.Second,
	}, nil
}

func CreateClusterClient(cfg *Config, db DB) *redis.ClusterClient {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.HASentinelPort)
	timeout := time.Duration(cfg.HASentinelSocketTimeout) * time.Second
	options := redis.FailoverOptions{
		SentinelAddrs: []string{addr},
		ReadTimeout:   timeout,
		WriteTimeout:  timeout,
		MaxRetries:    6,
		DB:            int(db),
		MasterName:    cfg.HASentinelMasterName,
	}
	if cfg.AuthRequired {
		options.Password = cfg.Password
	}
	return redis.NewFailoverClusterClient(&options)
}

func CreateClient(cfg *Config, db DB) *redis.Client {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	options := redis.Options{
		Addr:       addr,
		MaxRetries: 6,
		DB:         int(db),
	}
	if cfg.AuthRequired {
		options.Password = cfg.Password
	}
	return redis.NewClient(&options)
}

func (client *Client) GetPartialDocument(redisKey string, doc maps.PartialDocument) error {
	response := client.client.Get(ctx, redisKey)
	if response.Err() != nil {
		return response.Err().(Error)
	}
	b, err := response.Bytes()
	if err != nil {
		panic(err)
	}
	var raw map[string]interface{}
	err = json.Unmarshal(b, &raw)
	if err != nil {
		panic(err)
	}
	err = maps.FillFromMap(doc, &raw)
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) UpdatePartialDocument(
	redisKey string,
	doc maps.PartialDocument,
	updateFunc interface{}) (err error) {
	releaseLock, err := client.Lock(redisKey)
	if err != nil {
		return err
	}
	defer func() {
		err = releaseLock()
	}()
	err = client.GetPartialDocument(redisKey, doc)
	if err != nil {
		return err
	}
	err = maps.ApplyUpdates(doc, updateFunc)
	if err != nil {
		return err
	}
	err = client.SaveDoc(redisKey, doc)
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) Lock(redisKey string) (ReleaseLock, error) {
	lockCl := redislock.New(client.client)
	str := redislock.LimitRetry(redislock.LinearBackoff(time.Second), 20)
	lockKey := fmt.Sprintf("lock:%s", redisKey)
	lock, err := lockCl.Obtain(ctx, lockKey, client.lockExpiration, &redislock.Options{RetryStrategy: str})
	if err != nil {
		return nil, err
	}
	return func() error {
		return lock.Release(ctx)
	}, nil
}

func (client *Client) SaveDoc(redisKey string, document maps.PartialDocument) error {
	b, err := json.Marshal(document)
	if err != nil {
		return err
	}
	response := client.client.Set(ctx, redisKey, b, 0)
	if response.Err() != nil {
		return response.Err().(Error)
	}
	return nil
}

func (client *Client) Close() error {
	return client.client.Close()
}

func readEnvironment() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
