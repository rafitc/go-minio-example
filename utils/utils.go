package utils

import (
	"fmt"
	"log"
	"log/slog"
	"mime/multipart"
	"minio-example/model"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Connect to a minIo server
func ConnectToMinIo() *minio.Client {
	endpoint := "play.min.io"
	accessKeyID := "Q3AM3UQ867SPQQA43P2F"
	secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	useSSL := true

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	slog.Debug("MinIo client created")
	return minioClient
}

func PutImageInBucket(ctx *gin.Context, bucket_name string, file *multipart.FileHeader, client *minio.Client, FileUploadStatus []model.Uploadstatus, wg *sync.WaitGroup, ch chan model.Uploadstatus) {
	var uploadStatusOfGoRoutine model.Uploadstatus
	// Create a unique file name
	object_name := fmt.Sprintf("%s-%s", uuid.NewString(), file.Filename) // uuid + file name (to makesure file name is unique)

	// Update the static vars
	uploadStatusOfGoRoutine.BucketName = bucket_name
	uploadStatusOfGoRoutine.ObjectName = object_name

	// open file
	reader, err := file.Open()
	if err != nil {
		// update status
		uploadStatusOfGoRoutine.Status = false
		slog.Error("Error processing file", "filename", file.Filename, "error", err.Error())
		// pass value into channel and exit
		ch <- uploadStatusOfGoRoutine
		defer wg.Done()
	}
	defer reader.Close()

	info, err := client.PutObject(ctx, bucket_name, object_name, reader, file.Size, minio.PutObjectOptions{ContentType: "application/image"})
	if err != nil {
		// update status
		uploadStatusOfGoRoutine.Status = false
		slog.Error("Error while uploading file", "filename", file.Filename, "error", err.Error())
		// pass value into channel and exit
		ch <- uploadStatusOfGoRoutine
		defer wg.Done()
	}
	slog.Info("Successfully uploaded file %v Size : %d", file.Filename, info.Size)
	uploadStatusOfGoRoutine.Status = true
	// pass value into channel and exit
	ch <- uploadStatusOfGoRoutine
}

func GeneratePresignedURL(ctx *gin.Context, client *minio.Client, bucket_name string) []model.Files {
	var filePaths []model.Files
	// Create a done channel.
	doneCh := make(chan struct{})
	defer close(doneCh)

	// Read the obejct information from given bucket
	// After each reading use the object name to generate the PreSigned URLs
	for message := range client.ListObjects(ctx, bucket_name, minio.ListObjectsOptions{Prefix: "", Recursive: true}) {
		objectName := message.Key

		// With object name and bucket create a presigned URL with 60 Sec validity
		reqParams := make(url.Values)
		reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", objectName))

		url, err := client.PresignedGetObject(ctx, bucket_name, objectName, time.Duration(60*int(time.Second)), reqParams)
		if err != nil {
			slog.Error("Error while retrieving preSigned URL", "error", err.Error())
			continue
		}
		filePaths = append(filePaths, model.Files{FilePath: url.String()})
	}
	slog.Info("Successfully Generated presigned URLs for objects in ", "bucket name", bucket_name)
	return filePaths
}
