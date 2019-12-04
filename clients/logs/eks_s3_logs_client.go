package logs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

//
// EKSS3LogsClient corresponds with the aws logs driver
// for ECS and returns logs for runs
//
type EKSS3LogsClient struct {
	logRetentionInDays int64
	logNamespace       string
	s3Client           *s3.S3
	s3Bucket           string
	s3BucketRootDir    string
	logger             *log.Logger
}

type s3Log struct {
	Log    string    `json:"log"`
	Stream string    `json:"stream"`
	Time   time.Time `json:"time"`
}

//
// Name returns the name of the logs client
//
func (lc *EKSS3LogsClient) Name() string {
	return "eks-s3"
}

//
// Initialize sets up the EKSS3LogsClient
//
func (lc *EKSS3LogsClient) Initialize(conf config.Config) error {
	confLogOptions := conf.GetStringMapString("eks.log.driver.options")

	awsRegion := confLogOptions["awslogs-region"]
	if len(awsRegion) == 0 {
		awsRegion = conf.GetString("aws_default_region")
	}

	if len(awsRegion) == 0 {
		return errors.Errorf(
			"EKSS3LogsClient needs one of [eks.log.driver.options.awslogs-region] or [aws_default_region] set in config")
	}

	flotillaMode := conf.GetString("flotilla_mode")
	if flotillaMode != "test" {
		sess := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(awsRegion)}))

		lc.s3Client = s3.New(sess, aws.NewConfig().WithRegion(awsRegion))
	}

	s3BucketName := confLogOptions["s3_bucket_name"]

	if len(s3BucketName) == 0 {
		return errors.Errorf(
			"EKSS3LogsClient needs [eks.log.driver.options.s3_bucket_name] set in config")
	}
	lc.s3Bucket = s3BucketName

	s3BucketRootDir := confLogOptions["s3_bucket_root_dir"]

	if len(s3BucketRootDir) == 0 {
		return errors.Errorf(
			"EKSS3LogsClient needs [eks.log.driver.options.s3_bucket_root_dir] set in config")
	}
	lc.s3BucketRootDir = s3BucketRootDir

	lc.logger = log.New(os.Stderr, "[s3logs] ",
		log.Ldate|log.Ltime|log.Lshortfile)
	return nil
}

func (lc *EKSS3LogsClient) Logs(definition state.Definition, run state.Run, lastSeen *string) (string, *string, error) {
	return "", nil, errors.Errorf("EKSS3LogsClient does not support the Logs method.")
}

//
// Logs returns all logs from the log stream identified by handle since lastSeen
//
func (lc *EKSS3LogsClient) LogsText(definition state.Definition, run state.Run, w http.ResponseWriter) error {
	//Pod isn't there yet - dont return a 404
	if run.PodName == nil {
		return nil
	}
	s3DirName := lc.toS3DirName(run)

	// Get list of S3 objects in the run_id folder.
	result, err := lc.s3Client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(lc.s3Bucket),
		Prefix: aws.String(s3DirName),
	})

	if err != nil {
		return errors.Wrap(err, "problem getting logs")
	}

	// TODO: get latest file.
	if len(result.Contents) == 1 {
		s3Key := result.Contents[0].Key
		result, err := lc.s3Client.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(lc.s3Bucket),
			Key:    aws.String(*s3Key),
		})

		if err != nil {
			return err
		}
		return lc.logsToMessage(result, w)
	}
	return nil
}

func (lc *EKSS3LogsClient) toS3DirName(run state.Run) string {
	return fmt.Sprintf("%s/%s", lc.s3BucketRootDir, run.RunID)
}

func (lc *EKSS3LogsClient) logsToMessage(result *s3.GetObjectOutput, w http.ResponseWriter) error {
	reader := bufio.NewReader(result.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		} else {
			var parsedLine s3Log
			err := json.Unmarshal(line, &parsedLine)
			if err != nil {
				return err
			}
			_, err = io.WriteString(w, parsedLine.Log)
			if err != nil {
				return err
			}
		}
	}
}
