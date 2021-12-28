package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/kelseyhightower/envconfig"
	"github.com/sinmetal/gcs_sample/encryption"
	metadatabox "github.com/sinmetalcraft/gcpbox/metadata"
	"go.opencensus.io/trace"
	"google.golang.org/api/cloudkms/v1"
)

type Config struct {
	// BaseBucket is 適当に扱うファイルを置いておくBucket
	BaseBucket string

	// CloudKMSKeyName is CSEKで扱うCloud KMS Key Name
	// format: projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s
	CloudKMSKeyName string
}

// CSEKEncryptBucket1 is 暗号化したファイルを置くBucket
func (c *Config) CSEKEncryptBucket1() string {
	return fmt.Sprintf("%s-encrypt1", c.BaseBucket)
}

// CSEKEncryptBucket2 is CSEKEncryptBucket1からCopyしたファイルを置くBucket
func (c *Config) CSEKEncryptBucket2() string {
	return fmt.Sprintf("%s-encrypt2", c.BaseBucket)
}

// CMEKEncryptBucket is Default Keyを指定したBucket
func (c *Config) CMEKEncryptBucket() string {
	return fmt.Sprintf("%s-cmek-encrypt", c.BaseBucket)
}

func main() {
	ctx := context.Background()

	log.Print("starting server...")
	http.HandleFunc("/", helloHandler)

	projectID, err := metadatabox.ProjectID()
	if err != nil {
		log.Fatal(err.Error())
	}

	if metadatabox.OnGCP() {
		// Create and register a OpenCensus Stackdriver Trace exporter.
		exporter, err := stackdriver.NewExporter(stackdriver.Options{
			ProjectID: projectID,
		})
		if err != nil {
			log.Fatal(err)
		}
		trace.AlwaysSample()
		trace.RegisterExporter(exporter)
	}

	var cfg Config
	err = envconfig.Process("SINMETAL", &cfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Printf("BaseBucketName:%s\n", cfg.BaseBucket)
	fmt.Printf("CloudKMSKeyName:%s\n", cfg.CloudKMSKeyName)

	gcs, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	kms, err := cloudkms.NewService(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	csekService, err := encryption.NewCSEKService(ctx, gcs, kms)
	if err != nil {
		log.Fatal(err.Error())
	}
	cmekService, err := encryption.NewCMEKService(ctx, gcs)
	if err != nil {
		log.Fatal(err.Error())
	}

	handlers := Handlers{
		Config:      &cfg,
		GCS:         gcs,
		CSEKService: csekService,
		CMEKService: cmekService,
	}
	http.HandleFunc("/encryption/csek/upload", handlers.UploadCSEKHandler)
	http.HandleFunc("/encryption/csek/copy", handlers.CopyCSEKHandler)

	http.HandleFunc("/encryption/cmek/upload", handlers.UploadCMEKHandler)
	http.HandleFunc("/encryption/cmek/re-encrypt", handlers.ReEncryptCMEKHandler)

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!\n")
}
