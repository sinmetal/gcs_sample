package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func (handlers *Handlers) UploadCMEKHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	object := r.FormValue("object")

	file, err := handlers.GCS.Bucket(handlers.Config.BaseBucket).Object(object).NewReader(ctx)
	if err != nil {
		fmt.Printf("failed object.NewReader: %s: %s\n", object, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Printf("failed object read: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := file.Close(); err != nil {
		fmt.Printf("warn objectReader.Close: %s\n", err.Error())
	}

	size, err := handlers.CMEKService.Upload(ctx, handlers.Config.CMEKEncryptBucket(), object, data)
	if err != nil {
		fmt.Printf("failed upload to gcs: kmsKey=%s, object=%s: %s\n", handlers.Config.CloudKMSKeyName, object, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(fmt.Sprintf("finish.\nsize=%d", size)))
	if err != nil {
		fmt.Printf("warn write response. %s", err)
	}
}

func (handlers *Handlers) ReEncryptCMEKHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	object := r.FormValue("object")

	if err := handlers.CMEKService.ReEncrypt(ctx, handlers.Config.CMEKEncryptBucket(), object); err != nil {
		fmt.Printf("failed copy object: kmsKey=%s, object=%s: %s\n", handlers.Config.CloudKMSKeyName, object, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("finish."))
	if err != nil {
		fmt.Printf("warn write response. %s", err)
	}
}
