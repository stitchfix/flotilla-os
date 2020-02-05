package worker

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	"gopkg.in/tomb.v2"
	"strings"
	"time"
)

type cloudtrailWorker struct {
	sm           state.Manager
	qm           queue.Manager
	conf         config.Config
	log          flotillaLog.Logger
	pollInterval time.Duration
	t            tomb.Tomb
	queue        string
	engine       *string
	s3Client     *s3.S3
}

func (ctw *cloudtrailWorker) Initialize(conf config.Config, sm state.Manager, ee engine.Engine, log flotillaLog.Logger, pollInterval time.Duration, engine *string, qm queue.Manager) error {
	ctw.pollInterval = pollInterval
	ctw.conf = conf
	ctw.sm = sm
	ctw.qm = qm
	ctw.log = log
	ctw.engine = engine
	ctw.queue = conf.GetString("cloudtrail_queue")
	_ = ctw.qm.Initialize(ctw.conf, "eks")

	awsRegion := conf.GetString("eks.manifest.storage.options.region")
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion)}))
	ctw.s3Client = s3.New(sess, aws.NewConfig().WithRegion(awsRegion))

	return nil
}

func (ctw *cloudtrailWorker) GetTomb() *tomb.Tomb {
	return &ctw.t
}

//
// Run lists queues, consumes runs from them, and executes them using the execution engine
//
func (ctw *cloudtrailWorker) Run() error {
	for {
		select {
		case <-ctw.t.Dying():
			_ = ctw.log.Log("message", "A CloudTrail worker was terminated")
			return nil
		default:
			ctw.runOnce()
			time.Sleep(ctw.pollInterval)
		}
	}
}

func (ctw *cloudtrailWorker) runOnce() {
	qurl, err := ctw.qm.QurlFor(ctw.queue, false)
	if err != nil {
		_ = ctw.log.Log("message", "Error receiving CloudTrail queue", "error", fmt.Sprintf("%+v", err))
		return
	}
	cloudTrailS3File, err := ctw.qm.ReceiveCloudTrail(qurl)
	if err != nil {
		_ = ctw.log.Log("message", "Error receiving CloudTrail file", "error", fmt.Sprintf("%+v", err))
		return
	}

	ctw.processS3Keys(cloudTrailS3File)
}

func (ctw *cloudtrailWorker) processS3Keys(cloudTrailS3File state.CloudTrailS3File) {
	var ctn state.CloudTrailNotifications
	defaultRegion := ctw.conf.GetString("aws_default_region")
	for _, keyName := range cloudTrailS3File.S3ObjectKey {
		if !strings.Contains(keyName, defaultRegion) {
			continue
		}
		getObjectOutput, err := ctw.s3Client.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(cloudTrailS3File.S3Bucket),
			Key:    aws.String(keyName),
		})

		if err != nil {
			_ = ctw.log.Log("message", "Error receiving CloudTrail file - no object", "error", fmt.Sprintf("%+v", err))
			return
		}

		err = json.NewDecoder(getObjectOutput.Body).Decode(&ctn)

		_ = ctw.log.Log("message", "CloudTrail processing file", "key", fmt.Sprintf("s3://%s/%s", cloudTrailS3File.S3Bucket, keyName), "len", len(ctn.Records))
		ctw.processCloudTrailNotifications(ctn)

		getObjectOutput.Body.Close()
	}
}

func (ctw *cloudtrailWorker) processCloudTrailNotifications(ctn state.CloudTrailNotifications) {
	sa := ctw.conf.GetString("eks.service_account")
	runIdRecordMap := make(map[string][]state.Record)
	for _, record := range ctn.Records {
		if strings.Contains(record.UserIdentity.Arn, sa) && strings.Contains(record.UserIdentity.Arn, "eks-") {
			runId := ctw.getRunId(record)
			runIdRecordMap[runId] = append(runIdRecordMap[runId], record)
		}
	}

	for runId, records := range runIdRecordMap {
		_ = ctw.log.Log("message", "Saving CloudTrail Events", "run_id", runId, len(records))
		run, err := ctw.sm.GetRun(runId)
		if err == nil {
			run.CloudTrailNotifications.Records = append(run.CloudTrailNotifications.Records, records...)
			_, err = ctw.sm.UpdateRun(runId, run)
			if err != nil {
				_ = ctw.log.Log("message", "Error updating run", "error", fmt.Sprintf("%+v", err))
			}
		}
	}
}
func (ctw *cloudtrailWorker) getRunId(record state.Record) string {
	splits := strings.Split(record.UserIdentity.Arn, "/")
	runId := splits[len(splits)-1]
	return runId
}
