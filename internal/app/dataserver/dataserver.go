package dataserver

import (
	"dataserver/internal/pkg/config"
	"dataserver/internal/pkg/dataserver"
	"dataserver/internal/pkg/fsd"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/minio/minio-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"net/http"
	"net/textproto"
	"time"
)

var (
	packetsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dataserver_packets_processed",
		Help: "The total number of processed packets.",
	})

	timeToProcessPacket = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "dataserver_time_to_process_packet",
		Help: "The time to process a packet sent by FSD.",
	})
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
			sentry.CaptureException(err)
		}
	}()
	fsd.Sync(conn)

	// Instantiate empty client list and begin listening for updates
	clientList := &dataserver.ClientList{}
	go update()
	go exposeMetrics()
	listen(clientList, conn, producer)
}

func exposeMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":2112", nil)
	if err != nil {
		panic(err)
	}
}

// update handles the creation of a 15 second ticker for updating the data file
func update() {
	now := time.Now().UTC()
	for clientList := range dataserver.Channel {
		if time.Since(now) >= (15 * time.Second) {
			checkError(updateFile(clientList))
			now = time.Now().UTC()
		}
	}
}

// listen continually reads, parses and handles FSD packets.
func listen(clientList *dataserver.ClientList, conn *textproto.Conn, producer *kafka.Producer) {
	for {
		bytes, err := fsd.ReadMessage(conn)
		timer := prometheus.NewTimer(timeToProcessPacket)
		if err != nil {
			sentry.CaptureException(err)
			continue
		}
		split := fsd.ParseMessage(bytes)
		processMessage(split, clientList, conn, producer)
		timer.ObserveDuration()
	}
}

// processMessage classifies the FSD packet and performs the appropriate action
func processMessage(split []string, clientList *dataserver.ClientList, conn *textproto.Conn, producer *kafka.Producer) {
	if split[0] == "ADDCLIENT" {
		checkError(dataserver.AddClient(split, clientList, producer))
	} else if split[0] == "RMCLIENT" {
		checkError(dataserver.RemoveClient(split, clientList, producer))
	} else if split[0] == "PD" {
		checkError(dataserver.UpdatePosition(split, clientList, producer))
	} else if split[0] == "AD" {
		checkError(dataserver.UpdateControllerData(split, clientList, producer))
	} else if split[0] == "PLAN" {
		checkError(dataserver.UpdateFlightPlan(split, clientList, producer))
	} else if split[0] == "PING" && len(split) >= 6 {
		fsd.Pong(conn, split)
	}
	packetsProcessed.Inc()
}

// checkError checks if an error occurred and reports it to Sentry
func checkError(err error) {
	if err != nil {
		sentry.CaptureException(err)
	}
}

// updateFile encodes the current clientList and prints to the data file
func updateFile(clientList dataserver.ClientList) error {
	clientJSON, err := dataserver.EncodeJSON(clientList)
	if err != nil {
		return err
	}
	err = dataserver.WriteDataFile(clientJSON)
	if err != nil {
		return err
	}
	fmt.Printf("%+v Data file updated\n", time.Now().UTC().Format(time.RFC3339))
	s3Push()

	return nil
}

// s3Push sets up  the S3 library and begins the goroutine
func s3Push() {
	s3Config, err := config.Cfg.Map("s3")
	if err != nil {
		panic(err)
	}
	for k := range s3Config {
		go s3Loop(k)
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
		sentry.CaptureException(err)
	}

	n, err := minioClient.FPutObject(bucketName, objectName, "directoryvatsim-data.json", minio.PutObjectOptions{ContentType: contentType, UserMetadata: userMetaData})
	if err != nil {
		sentry.CaptureException(err)
	}
	fmt.Printf("%+v Successfully uploaded %s of size %d\n", time.Now().UTC().Format(time.RFC3339), objectName, n)
}
