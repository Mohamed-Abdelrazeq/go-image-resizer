# Go Image Resizer

This project is a Go-based AWS Lambda function that resizes images stored in an S3 bucket. The resized images are then uploaded to another S3 bucket.



## Configuration

The project uses environment variables to configure AWS settings. These variables should be set in the `.env` file:

- `AWS_REGION`: The AWS region where your S3 buckets are located.
- `AWS_BUCKET_NAME`: The name of the S3 bucket where the original images are stored.
- `AWS_RESIZED_BUCKET_NAME`: The name of the S3 bucket where the resized images will be uploaded.

## Building and Running

To build the project for deployment, use the following command:

```sh
GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap cmd/main.go
```

To run the project locally, you can use the air tool for live reloading. Ensure you have air installed and then run:

```sh
air
```

# AWS Lambda Deployment
The project is designed to be deployed as an AWS Lambda function. Ensure you have the AWS CLI installed and configured. Then, package and deploy the Lambda function using the AWS CLI or any other deployment tool of your choice.

# Usage
The Lambda function listens for SQS events that contain information about new images uploaded to the S3 bucket. When an image is uploaded, the function:

Downloads the image from the S3 bucket.
Resizes the image to predefined sizes (100px, 500px, 1000px).
Uploads the resized images to the resized S3 bucket.
# Dependencies
The project relies on several Go packages, including:
- github.com/aws/aws-lambda-go
- github.com/aws/aws-sdk-go
- github.com/nfnt/resize
- github.com/gin-gonic/gin
For a full list of dependencies, see the go.mod file.

License
This project is licensed under the MIT License.