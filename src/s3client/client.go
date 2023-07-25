package s3client

import (
	"text2phenotype.com/fdl/logger"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"strings"
)

type Client struct {
	holder     *sessionHolder
	bucketName string
	region     string
	env        EnvironmentConfig
}

type sessionHolder struct {
	curr      *session.Session
	requestCh <-chan *session.Session
	errorCh   chan<- error
	closeCh   chan<- struct{}
}

var clientLogger = logger.NewLogger("S3Client")
var sdkLogger = logger.NewLogger("S3-SDK")

func New() (*Client, error) {
	errLogger := clientLogger.With().Caller().Logger()
	env, err := readEnvironment(&errLogger)
	if err != nil {
		clientLogger.Err(err).Msg("Failed to get proper variables from environment")
		return nil, err
	}
	client := Client{
		bucketName: env.BucketName,
		region:     env.Region,
		env:        env,
	}
	sessionCh := make(chan *session.Session)
	errorCh := make(chan error)
	closeCh := make(chan struct{}, 1)

	client.holder = &sessionHolder{
		requestCh: sessionCh,
		errorCh:   errorCh,
		closeCh:   closeCh,
	}
	if err := client.acquireNewSession(); err != nil {
		return nil, err
	}
	go keepSessionRefreshed(&client, sessionCh, errorCh, closeCh)
	return &client, nil
}

func (client Client) Upload(data string, key string) (*s3manager.UploadOutput, error) {
	file := strings.NewReader(data)
	params := &s3manager.UploadInput{
		Bucket: &client.bucketName,
		Key:    &key,
		Body:   file,
	}
	sess, err := client.session()
	if err != nil {
		return nil, err
	}
	output, err := client.upload(sess, params)
	if err == nil {
		return output, nil
	}
	sess, err = client.tryRefreshingSession(err)
	if err != nil {
		return nil, err
	}
	return client.upload(sess, params)
}

func (client Client) Download(key string) ([]byte, error) {
	params := &s3.GetObjectInput{
		Bucket: &client.bucketName,
		Key:    &key,
	}
	sess, err := client.session()
	if err != nil {
		return nil, err
	}
	res, err := client.download(sess, params)
	if err == nil {
		return res, nil
	}
	sess, err = client.tryRefreshingSession(err)
	if err != nil {
		return nil, err
	}
	return client.download(sess, params)
}

func (client Client) Close() {
	client.holder.closeCh <- struct{}{}
}

func (client Client) upload(sess *session.Session, params *s3manager.UploadInput) (*s3manager.UploadOutput, error) {
	fdlLogger := clientLogger.With().
		Str("key", *params.Key).
		Str("bucket", *params.Bucket).Logger()

	sdkLog := sdkLogger.With().
		Str("key", *params.Key).
		Str("bucket", *params.Bucket).Logger()

	uploader := s3manager.NewUploader(sess.Copy(&aws.Config{Logger: getLogger(sdkLog)}))
	fdlLogger.Debug().Msg("Uploading the file")
	return uploader.Upload(params)
}

func (client Client) download(sess *session.Session, params *s3.GetObjectInput) ([]byte, error) {
	fdlLogger := clientLogger.With().
		Str("key", *params.Key).
		Str("bucket", *params.Bucket).Logger()

	sdkLog := sdkLogger.With().
		Str("key", *params.Key).
		Str("bucket", *params.Bucket).Logger()

	downloader := s3manager.NewDownloader(sess.Copy(&aws.Config{Logger: getLogger(sdkLog)}))

	buf := aws.NewWriteAtBuffer([]byte{})

	fdlLogger.Debug().Msg("Downloading file")

	size, err := downloader.Download(buf, params)
	if err != nil {
		fdlLogger.Error().Err(err).Msg("Failed to download file")
		return nil, err
	}
	fdlLogger.Debug().Msgf("Downloaded %v bytes", size)
	return buf.Bytes(), nil
}

func keepSessionRefreshed(client *Client, sessionCh chan<- *session.Session, errorCh <-chan error, closeCh <-chan struct{}) {
	for {
		select {
		case sessionCh <- client.holder.curr:
			continue
		default:
		}
		select {
		case sessionCh <- client.holder.curr:
		case err := <-errorCh:
			clientLogger.Error().Err(err).Msg("Caught error while using S3 session, trying to refresh it")
			if err = client.acquireNewSession(); err != nil {
				clientLogger.Error().Err(err).Msg("Caught error while refreshing S3 session")
				continue
			}
			clientLogger.Info().Msg("Successfully refreshed session")
		case <-closeCh:
			clientLogger.Info().Msg("Closing client")
			return
		}
	}
}

func (client Client) tryRefreshingSession(err error) (*session.Session, error) {
	var sess *session.Session
	select {
	case client.holder.errorCh <- err:
		sess = <-client.holder.requestCh
	case sess = <-client.holder.requestCh:
	}
	if sess == nil {
		return nil, errors.New("failed to refresh session")
	}
	return sess, nil
}

func (client Client) session() (*session.Session, error) {
	sess := <-client.holder.requestCh
	if sess == nil {
		return nil, errors.New("could not get session")
	}
	return sess, nil
}

func (client Client) createEC2Config() *aws.Config {
	return &aws.Config{
		Region:     aws.String(client.region),
		MaxRetries: aws.Int(4),
		LogLevel:   aws.LogLevel(aws.LogDebug),
	}
}
func (client Client) createEnvConfig() *aws.Config {
	creds := credentials.NewStaticCredentials(
		client.env.AccessKeyID,
		client.env.AccessKey,
		"")
	_, err := creds.Get()

	if err != nil {
		clientLogger.Error().Err(err).Msg("Error with credentials from environment")
		panic(err)
	}
	cfg := aws.NewConfig().
		WithRegion(*aws.String(client.region)).
		WithMaxRetries(*aws.Int(4)).
		WithCredentials(creds).
		WithLogLevel(aws.LogDebug)

	inDevEnv := client.env.T2PEnv == "dev"
	if inDevEnv && len(client.env.AwsEndpoint) > 0 {
		cfg = cfg.WithEndpoint(*aws.String(client.env.AwsEndpoint)).
			WithS3ForcePathStyle(true)
	}
	return cfg
}

func (client *Client) acquireNewSession() error {
	sess, err := session.NewSession(
		client.createEC2Config(),
	)
	if err != nil {
		client.holder.curr = nil
		clientLogger.Error().Err(err).Msg("Could not initialize S3 session")
		return err
	}
	_, err = sts.New(sess).GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err == nil {
		client.holder.curr = sess
		clientLogger.Info().Msg("S3 session successfully initialized using EC2")
		return nil
	}
	clientLogger.Info().Msg("Could not initialize S3 session using EC2, trying env credentials")
	sess, err = session.NewSession(
		client.createEnvConfig(),
	)
	if err != nil {
		client.holder.curr = nil
		clientLogger.Error().Err(err).Msg("Could not initialize S3 session")
		return err
	}
	_, err = sts.New(sess).GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		client.holder.curr = nil
		clientLogger.Error().Err(err).Msg("Could not initialize S3 session")
		return errors.New("could not initialize S3 session")
	}
	client.holder.curr = sess
	clientLogger.Info().Msg("S3 session successfully initialized using env credentials")
	return nil
}

type EnvironmentConfig struct {
	BucketName  string `envconfig:"MDL_COMN_STORAGE_CONTAINER_NAME" required:"true"`
	T2PEnv    string `envconfig:"T2P_ENV" required:"true"`
	Region      string `envconfig:"MDL_COMN_AWS_REGION_NAME" required:"true"`
	AwsEndpoint string `envconfig:"MDL_COMN_AWS_ENDPOINT_URL" default:""`
	AccessKeyID string `envconfig:"MDL_COMN_AWS_ACCESS_ID" default:""`
	AccessKey   string `envconfig:"MDL_COMN_AWS_ACCESS_KEY" default:""`
}

func readEnvironment(errLogger *zerolog.Logger) (EnvironmentConfig, error) {
	var config EnvironmentConfig
	err := envconfig.Process("", &config)
	if err != nil {
		errLogger.Err(err).Msg("Got error while processing environment")
		return config, err
	}
	return config, nil
}

type s3Logger struct {
	fdlLogger zerolog.Logger
}

func getLogger(fdlLogger zerolog.Logger) *s3Logger {
	return &s3Logger{
		fdlLogger,
	}
}

func (logger *s3Logger) Log(v ...interface{}) {
	//nolint
	logger.fdlLogger.Debug().Msg(fmt.Sprint(v))
}
