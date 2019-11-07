package dataserver

import (
	"dataserver/internal/pkg/config"
	"dataserver/internal/pkg/dataserver"
	"dataserver/internal/pkg/fsd"
	"fmt"
	"github.com/minio/minio-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

// Start connects and begins parsing and saving data files.
func Start() {
	// Configuration
	config.ReadConfig()
	config.ConfigureSentry()
	producer := config.ConfigureKafka()
	defer producer.Close()

	// Connect to FSD
	conn := fsd.Connect()
	defer func() {
		if err := conn.Close(); err != nil {
			log.Fatal("Failed to close FSD connection.")
		}
	}()

	// Create our application context
	context := dataserver.Context{
		Consumer: conn,
		Producer: producer,
		ClientList: &dataserver.ClientList{
			Mutex: &sync.RWMutex{},
		},
	}

	// Set ourselves up as an FSD server
	go context.SetupServer()

	// Add a fake client to request data with
	go context.AddFSDClient()

	// Begin listening for updates
	go update()
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
func s3Loop(k string) {
	endPoint, _ := config.Cfg.String(fmt.Sprintf("s3.%s.endpoint", k))
	accessKeyID, _ := config.Cfg.String(fmt.Sprintf("s3.%s.accessKeyID", k))
	secretAccessKey, _ := config.Cfg.String(fmt.Sprintf("s3.%s.secretAccessKey", k))
	bucketName, _ := config.Cfg.String(fmt.Sprintf("s3.%s.bucketName", k))
	contentType := "application/json"
	objectName := "vatsim-data.json"
	userMetaData := map[string]string{"x-amz-acl": "public-read"}

	minioClient, err := minio.New(endPoint, accessKeyID, secretAccessKey, true)
	if err != nil {
		log.WithField("error", err).Error("Failed to create new S3 client.")
	}

	n, err := minioClient.FPutObject(bucketName, objectName, "directoryvatsim-data.json", minio.PutObjectOptions{ContentType: contentType, UserMetadata: userMetaData})
	if err != nil {
		log.WithField("error", err).Error("Failed to upload object to S3.")
	}
	log.WithFields(log.Fields{
		"object": objectName,
		"size":   n,
	}).Info("Successfully uploaded object to S3.")
}

// s3Push sets up  the S3 library and begins the goroutine
func s3Push() {
	s3Config, err := config.Cfg.Map("s3")
	if err != nil {
		log.Fatal("S3 configuration not defined.")
	}
	for k := range s3Config {
		go s3Loop(k)
	}
}

// update handles the creation of a 15 second ticker for updating the data file
func update() {
	now := time.Now().UTC()
	for clientList := range dataserver.Channel {
		if time.Since(now) >= (15 * time.Second) {
			err := updateFile(clientList)
			if err != nil {
				log.WithField("error", err).Error("Failed to update data file.")
			}
			now = time.Now().UTC()
		}
	}
}

// updateFile encodes the current clientList and prints to the data file
func updateFile(clientList dataserver.ClientList) error {
	clientJSON, err := dataserver.EncodeJSON(clientList)
	if err != nil {
		return errors.WithStack(err)
	}
	err = dataserver.WriteDataFile(clientJSON)
	if err != nil {
		return errors.WithStack(err)
	}
	log.Info("Data file updated.")
	s3Push()

	return nil
}
