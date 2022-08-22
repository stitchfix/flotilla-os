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

		awstrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/aws/aws-sdk-go/aws"
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

func (ctw *cloudtrailWorker) Initialize(conf config.Config, sm state.Manager, eksEngine engine.Engine, emrEngine engine.Engine, log flotillaLog.Logger, pollInterval time.Duration, qm queue.Manager) error {
	ctw.pollInterval = pollInterval
	ctw.conf = conf
	ctw.sm = sm
	ctw.qm = qm
	ctw.log = log
	ctw.queue = conf.GetString("cloudtrail_queue")
	_ = ctw.qm.Initialize(ctw.conf, "eks")

	awsRegion := conf.GetString("eks_manifest_storage_options_region")
	sess := awstrace.WrapSession(session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion)})))
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
		ctw.processCloudTrailNotifications(ctn)
		getObjectOutput.Body.Close()
	}
}

func (ctw *cloudtrailWorker) processCloudTrailNotifications(ctn state.CloudTrailNotifications) {
	sa := ctw.conf.GetString("eks_service_account")
	runIdRecordMap := make(map[string][]state.Record)
	for _, record := range ctn.Records {
		if strings.Contains(record.UserIdentity.Arn, sa) && strings.Contains(record.UserIdentity.Arn, "eks-") {
			runId := ctw.getRunId(record)
			runIdRecordMap[runId] = append(runIdRecordMap[runId], record)
		}
	}

	for runId, records := range runIdRecordMap {
		run, err := ctw.sm.GetRun(runId)
		if err == nil {
			var rawRecords []state.Record
			if run.CloudTrailNotifications == nil || len((*run.CloudTrailNotifications).Records) == 0 {
				rawRecords = records
			} else {
				rawRecords = append((*run.CloudTrailNotifications).Records, records...)
			}
			run.CloudTrailNotifications = &state.CloudTrailNotifications{Records: ctw.makeSet(rawRecords)}
			_, err = ctw.sm.UpdateRun(runId, run)
			if err != nil {
				_ = ctw.log.Log("message", "Error updating run", "error", fmt.Sprintf("%+v", err))
			}
		}
	}
}

func (ctw *cloudtrailWorker) makeSet(records []state.Record) []state.Record {
	keys := make(map[string]bool)
	var set []state.Record
	for _, record := range records {
		if _, value := keys[record.String()]; !value {
			keys[record.String()] = true
			set = append(set, record)
		}
	}
	return set
}

func (ctw *cloudtrailWorker) getRunId(record state.Record) string {
	splits := strings.Split(record.UserIdentity.Arn, "/")
	runId := splits[len(splits)-1]
	return runId
}
