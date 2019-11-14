package dataserver

import (
	"dataserver/internal/pkg/config"
	"dataserver/internal/pkg/dataserver"
	"github.com/minio/minio-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// Start connects and begins parsing and saving data files.
func Start() {
	// Configuration
	configFile, err := config.New("configs/config.yml")
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	// Create our application context
	context, err := dataserver.New(configFile)
	if err != nil {
		log.Fatalf("Got error setting up dataserver: %v", err)
	}
	defer context.Close()

	// Set ourselves up as an FSD server
	go context.SetupServer()

	// Add a fake client to request data with
	go context.AddFSDClient()

	// Begin listening for updates
	go update(context)
	go exposeMetrics()
	go context.RequestATIS()
	go context.RemoveTimedOutClients()
	context.Listen()
}

// exposeMetrics listens for Prometheus scrapes
func exposeMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":2112", nil)
	if err != nil {
		log.Fatal("Failed to expose Prometheus metrics.")
	}
}

// s3Loop pushes the data files to S3
func s3Loop(bucketConfig *config.BucketConfig) {
	contentType := "application/json"
	objectName := "vatsim-data.json"
	userMetaData := map[string]string{"x-amz-acl": "public-read"}

	minioClient, err := minio.New(bucketConfig.Endpoint, bucketConfig.AccessKeyID, bucketConfig.SecretAccessKey, true)
	if err != nil {
		log.WithField("error", err).Error("Failed to create new S3 client.")
	}

	n, err := minioClient.FPutObject(bucketConfig.BucketName, objectName, "directoryvatsim-data.json", minio.PutObjectOptions{ContentType: contentType, UserMetadata: userMetaData})
	if err != nil {
		log.WithField("error", err).Error("Failed to upload object to S3.")
	}
	log.WithFields(log.Fields{
		"object": objectName,
		"size":   n,
	}).Debug("Successfully uploaded object to S3.")
}

// s3Push sets up  the S3 library and begins the goroutine
func s3Push(ctx *dataserver.Context) {
	if ctx.Config.S3Buckets == nil || len(ctx.Config.S3Buckets) == 0 {
		log.Fatal("S3 configuration not defined.")
	}
	for _, bucketConf := range ctx.Config.S3Buckets {
		go s3Loop(bucketConf)
	}
}

// update handles the creation of a 15 second ticker for updating the data file
func update(ctx *dataserver.Context) {
	var nextWriteTimeout <-chan time.Time = nil
	var clientList dataserver.ClientList
	var ok bool // this needs be be defined outside of the select - using := redefines clientList scoped within the for.

	for {
		select {
		case clientList, ok = <-dataserver.Channel:
			if !ok {
				break
			}
			if nextWriteTimeout == nil {
				err := updateFile(ctx, clientList)
				if err != nil {
					log.WithField("error", err).Error("Failed to update data file.")
				} else {
					nextWriteTimeout = time.After(15 * time.Second)
				}
			}
		case <-nextWriteTimeout:
			nextWriteTimeout = nil
			err := updateFile(ctx, clientList)
			if err != nil {
				log.WithField("error", err).Error("Failed to update data file.")
				// retry in 5
				nextWriteTimeout = time.After(5 * time.Second)
			}
		}
	}
}

// updateFile encodes the current clientList and prints to the data file
func updateFile(ctx *dataserver.Context, clientList dataserver.ClientList) error {
	//FIXME: use a stream encoding method rather than marshalling + writing in two separate steps.
	clientJSON, err := dataserver.EncodeJSON(clientList)
	if err != nil {
		return errors.WithStack(err)
	}
	err = ctx.WriteDataFile(clientJSON)
	if err != nil {
		return errors.WithStack(err)
	}
	log.Debug("Data file updated.")
	s3Push(ctx)

	return nil
}
