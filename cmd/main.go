package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"log"
	"os"
	"strconv"

	"image/jpeg"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nfnt/resize"
)

var (
	s3Session     *s3.S3
	bucket        string
	resizedBucket string
)

func initAWS() {
	sess := session.Must(
		session.NewSession(
			&aws.Config{
				Region: aws.String(os.Getenv("AWS_REGION")),
			},
		),
	)
	s3Session = s3.New(sess)
	bucket = os.Getenv("AWS_BUCKET_NAME")
	resizedBucket = os.Getenv("AWS_RESIZED_BUCKET_NAME")
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, message := range sqsEvent.Records {
		var msgBody S3Event
		err := json.Unmarshal([]byte(message.Body), &msgBody)
		if err != nil || len(msgBody.Records) == 0 {
			log.Println("Error unmarshalling message body:", err)
			continue
		}

		fmt.Println("Body: ", msgBody.Records)

		resizeOperation(msgBody.Records[0].S3.Object.Key)
	}
	return nil
}

func downloadImage(key string) ([]byte, error) {
	resp, err := s3Session.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func resizeImage(imageData []byte, key string) {
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		log.Println("Error decoding image:", err)
		return
	}

	sizes := []uint{100, 500, 1000}
	for _, size := range sizes {
		resizedImg := resize.Resize(size, 0, img, resize.Lanczos3)

		var buf bytes.Buffer
		err = jpeg.Encode(&buf, resizedImg, nil)
		if err != nil {
			log.Println("Error encoding resized image:", err)
			continue
		}

		// Upload the resized image to S3
		err = uploadImage(buf.Bytes(), size, key)
		if err != nil {
			log.Println("Error uploading resized image to S3:", err)
		}
	}
}

func uploadImage(imageData []byte, size uint, key string) error {
	_, err := s3Session.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(resizedBucket),
		Key:    aws.String(key + "/" + strconv.Itoa(int(size)) + ".jpg"),
		Body:   bytes.NewReader(imageData),
	})
	return err
}

func resizeOperation(key string) {
	image, err := downloadImage(key)
	if err != nil {
		log.Println("Error downloading image from S3:", err)
		return
	}

	resizeImage(image, key)
}

type S3Event struct {
	Records []Record `json:"Records"`
}

type Record struct {
	EventVersion      string            `json:"eventVersion"`
	EventSource       string            `json:"eventSource"`
	AwsRegion         string            `json:"awsRegion"`
	EventTime         string            `json:"eventTime"`
	EventName         string            `json:"eventName"`
	UserIdentity      UserIdentity      `json:"userIdentity"`
	RequestParameters RequestParameters `json:"requestParameters"`
	ResponseElements  ResponseElements  `json:"responseElements"`
	S3                S3                `json:"s3"`
}

type UserIdentity struct {
	PrincipalId string `json:"principalId"`
}

type RequestParameters struct {
	SourceIPAddress string `json:"sourceIPAddress"`
}

type ResponseElements struct {
	XAmzRequestId string `json:"x-amz-request-id"`
	XAmzId2       string `json:"x-amz-id-2"`
}

type S3 struct {
	S3SchemaVersion string   `json:"s3SchemaVersion"`
	ConfigurationId string   `json:"configurationId"`
	Bucket          Bucket   `json:"bucket"`
	Object          S3Object `json:"object"`
}

type Bucket struct {
	Name          string        `json:"name"`
	OwnerIdentity OwnerIdentity `json:"ownerIdentity"`
	Arn           string        `json:"arn"`
}

type OwnerIdentity struct {
	PrincipalId string `json:"principalId"`
}

type S3Object struct {
	Key       string `json:"key"`
	Size      int    `json:"size"`
	ETag      string `json:"eTag"`
	Sequencer string `json:"sequencer"`
}

func main() {
	initAWS()
	lambda.Start(handler)
}
