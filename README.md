# gcs_sample

## CMEK

```
gsutil kms authorize -p sinmetal-playground-20211227 -k projects/sinmetal-playground-20211227/locations/asia-northeast1/keyRings/gcs/cryptoKeys/sample
Authorized project sinmetal-playground-20211227 to encrypt and decrypt with key:
projects/sinmetal-playground-20211227/locations/asia-northeast1/keyRings/gcs/cryptoKeys/sample
```

```
gsutil kms encryption -k projects/sinmetal-playground-20211227/locations/asia-northeast1/keyRings/gcs/cryptoKeys/sample gs://sinmetal-playground-20211227
```

```
gsutil stat gs://sinmetal-playground-20211227/logo_only.jpg
gs://sinmetal-playground-20211227/logo_only.jpg:
    Creation time:          Mon, 27 Dec 2021 09:52:16 GMT
    Update time:            Mon, 27 Dec 2021 09:52:16 GMT
    Storage class:          STANDARD
    KMS key:                projects/sinmetal-playground-20211227/locations/asia-northeast1/keyRings/gcs/cryptoKeys/sample/cryptoKeyVersions/1
    Content-Length:         97148
    Content-Type:           image/jpeg
    Hash (crc32c):          /gdB4Q==
    Hash (md5):             YPiH8r5W05/qEnOZfpm85Q==
    ETag:                   CPfHr8fag/UCEAE=
    Generation:             1640598736724983
    Metageneration:         1
```