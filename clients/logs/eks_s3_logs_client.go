package logs

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
	awstrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/aws/aws-sdk-go/aws"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	emrS3LogsBucket    string
	emrS3LogsBasePath  string
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
	//confLogOptions := conf.GetStringMapString("eks_log_driver_options")

	awsRegion := conf.GetString("eks_log_driver_options_awslogs_region")
	if len(awsRegion) == 0 {
		awsRegion = conf.GetString("aws_default_region")
	}

	if len(awsRegion) == 0 {
		return errors.Errorf(
			"EKSS3LogsClient needs one of [eks.log.driver.options.awslogs-region] or [aws_default_region] set in config")
	}

	flotillaMode := conf.GetString("flotilla_mode")
	if flotillaMode != "test" {
		sess := awstrace.WrapSession(session.Must(session.NewSession(&aws.Config{
			Region: aws.String(awsRegion)})))
		sess = awstrace.WrapSession(sess)
		lc.s3Client = s3.New(sess, aws.NewConfig().WithRegion(awsRegion))
	}
	lc.emrS3LogsBucket = conf.GetString("emr_log_bucket")
	lc.emrS3LogsBasePath = conf.GetString("emr_log_base_path")
	s3BucketName := conf.GetString("eks_log_driver_options_s3_bucket_name")

	if len(s3BucketName) == 0 {
		return errors.Errorf(
			"EKSS3LogsClient needs [eks_log_driver_options_s3_bucket_name] set in config")
	}
	lc.s3Bucket = s3BucketName

	s3BucketRootDir := conf.GetString("eks_log_driver_options_s3_bucket_root_dir")

	if len(s3BucketRootDir) == 0 {
		return errors.Errorf(
			"EKSS3LogsClient needs [eks.log.driver.options.s3_bucket_root_dir] set in config")
	}
	lc.s3BucketRootDir = s3BucketRootDir

	lc.logger = log.New(os.Stderr, "[s3logs] ",
		log.Ldate|log.Ltime|log.Lshortfile)
	return nil
}

func (lc *EKSS3LogsClient) emrLogsToMessageString(run state.Run, lastSeen *string, role *string, facility *string) (string, *string, error) {
	s3DirName, err := lc.emrDriverLogsPath(run)
	if err != nil {
		return "", aws.String(""), errors.Errorf("No logs")
	}

	params := &s3.ListObjectsV2Input{
		Bucket:  aws.String(lc.emrS3LogsBucket),
		Prefix:  aws.String(s3DirName),
		MaxKeys: aws.Int64(1000),
	}

	pageNum := 0
	lastModified := &time.Time{}
	var key *string

	err = lc.s3Client.ListObjectsV2Pages(params,
		func(result *s3.ListObjectsV2Output, lastPage bool) bool {
			pageNum++
			if result != nil {
				for _, content := range result.Contents {
					if strings.Contains(*content.Key, *role) && strings.Contains(*content.Key, *facility) && lastModified.Before(*content.LastModified) {
						if content != nil && *content.Size < int64(10000000) {
							key = content.Key
							lastModified = content.LastModified
						}
					}
				}
			}
			if lastPage {
				return false
			}
			return pageNum <= 10
		})

	if key == nil {
		lc.logger.Println(fmt.Sprintf("run=%s emr logging key not found for role=%s facility=%s", run.RunID, *role, *facility))
		return "", aws.String(""), errors.Errorf("No driver logs found")
	}

	startPosition := int64(0)
	if lastSeen != nil {
		parsed, err := strconv.ParseInt(*lastSeen, 10, 64)
		if err == nil {
			startPosition = parsed
		}
	}

	s3Obj, err := lc.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(lc.emrS3LogsBucket),
		Key:    aws.String(*key),
	})

	if s3Obj != nil && err == nil && *s3Obj.ContentLength < int64(10000000) {
		defer s3Obj.Body.Close()
		gr, err := gzip.NewReader(s3Obj.Body)
		if err != nil {
			return "", aws.String(""), errors.Errorf("No driver logs found")
		}
		defer gr.Close()
		reader := bufio.NewReader(gr)
		var b0 bytes.Buffer
		counter := int64(0)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					err = nil
					return b0.String(), aws.String(fmt.Sprintf("%d", counter)), nil
				}

			} else {
				if counter >= startPosition {
					b0.Write(line)
				}
				counter = counter + 1
			}
		}
	}
	return "", aws.String(""), errors.Errorf("No driver logs found")
}

func (lc *EKSS3LogsClient) emrDriverLogsPath(run state.Run) (string, error) {
	if run.SparkExtension.EMRJobId != nil &&
		run.SparkExtension.VirtualClusterId != nil {
		return fmt.Sprintf("%s/%s/jobs/%s/",
			lc.emrS3LogsBasePath,
			*run.SparkExtension.VirtualClusterId,
			*run.SparkExtension.EMRJobId,
		), nil
	}
	return "", errors.New("couldn't construct s3 path.")
}

func (lc *EKSS3LogsClient) Logs(executable state.Executable, run state.Run, lastSeen *string, role *string, facility *string) (string, *string, error) {
	if *run.Engine == state.EKSSparkEngine {
		return lc.emrLogsToMessageString(run, lastSeen, role, facility)
	}

	result, err := lc.getS3Object(run)
	startPosition := int64(0)
	if lastSeen != nil {
		parsed, err := strconv.ParseInt(*lastSeen, 10, 64)
		if err == nil {
			startPosition = parsed
		}
	}

	if result != nil && err == nil {
		acc, position, err := lc.logsToMessageString(result, startPosition)
		newLastSeen := fmt.Sprintf("%d", position)
		return acc, &newLastSeen, err
	}

	return "", aws.String(""), errors.Errorf("No logs.")
}

//
// Logs returns all logs from the log stream identified by handle since lastSeen
//
func (lc *EKSS3LogsClient) LogsText(executable state.Executable, run state.Run, w http.ResponseWriter) error {
	if run.Engine == nil || *run.Engine == state.EKSEngine {
		result, err := lc.getS3Object(run)

		if result != nil && err == nil {
			return lc.logsToMessage(result, w)
		}
	}
	if *run.Engine == state.EKSSparkEngine {
		return lc.logsEMR(w)
	}
	return nil
}

//
// Fetch S3Object associated with the pod's log.
//
func (lc *EKSS3LogsClient) getS3Object(run state.Run) (*s3.GetObjectOutput, error) {
	//Pod isn't there yet - dont return a 404
	//if run.PodName == nil {
	//	return nil, errors.New("no pod associated with the run.")
	//}
	s3DirName := lc.toS3DirName(run)

	// Get list of S3 objects in the run_id folder.
	result, err := lc.s3Client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(lc.s3Bucket),
		Prefix: aws.String(s3DirName),
	})

	if err != nil {
		return nil, errors.Wrap(err, "problem getting logs")
	}

	if result == nil || result.Contents == nil || len(result.Contents) == 0 {
		return nil, errors.New("no s3 files associated with the run.")
	}
	var key *string
	lastModified := &time.Time{}

	//Find latest log file (could have multiple log files per pod - due to pod retries)
	for _, content := range result.Contents {
		if strings.Contains(*content.Key, run.RunID) && lastModified.Before(*content.LastModified) {
			if content != nil && *content.Size < int64(10000000) {
				key = content.Key
				lastModified = content.LastModified
			}
		}
	}
	if key != nil {
		return lc.getS3Key(key)
	} else {
		return nil, errors.New("no s3 files associated with the run.")
	}
}

func (lc *EKSS3LogsClient) getS3Key(s3Key *string) (*s3.GetObjectOutput, error) {
	result, err := lc.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(lc.s3Bucket),
		Key:    aws.String(*s3Key),
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

//
// Formulate dir name on S3.
//
func (lc *EKSS3LogsClient) toS3DirName(run state.Run) string {
	return fmt.Sprintf("%s/%s", lc.s3BucketRootDir, run.RunID)
}

//
// Converts log messages from S3 to strings - returns the contents of the entire file.
//
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

func (lc *EKSS3LogsClient) logsEMR(w http.ResponseWriter) error {
	_, _ = io.WriteString(w, "todo!!!")
	return nil
}

//
// Converts log messages from S3 to strings, takes a starting offset.
//
func (lc *EKSS3LogsClient) logsToMessageString(result *s3.GetObjectOutput, startingPosition int64) (string, int64, error) {
	acc := ""
	currentPosition := int64(0)
	// if less than/equal to 0, read entire log.
	if startingPosition <= 0 {
		startingPosition = currentPosition
	}

	// No S3 file or object, return "", 0, err
	if result == nil {
		return acc, startingPosition, errors.New("s3 object not present.")
	}

	reader := bufio.NewReader(result.Body)

	// Reading until startingPosition and discard unneeded lines.
	for currentPosition < startingPosition {
		currentPosition = currentPosition + 1
		_, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return acc, startingPosition, err
		}
	}

	// Read upto MaxLogLines
	for currentPosition <= startingPosition+state.MaxLogLines {
		currentPosition = currentPosition + 1
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return acc, currentPosition, err
		} else {
			var parsedLine s3Log
			err := json.Unmarshal(line, &parsedLine)
			if err == nil {
				acc = fmt.Sprintf("%s%s", acc, parsedLine.Log)
			}
		}
	}

	_ = result.Body.Close()
	return acc, currentPosition, nil
}
