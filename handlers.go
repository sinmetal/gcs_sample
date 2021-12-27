package main

import (
	"cloud.google.com/go/storage"
	"github.com/sinmetal/gcs_sample/encryption"
)

type Handlers struct {
	Config      *Config
	GCS         *storage.Client
	CSEKService *encryption.CSEKService
}
