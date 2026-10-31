package aws

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// quietChecksums silences the SDK warning logged for objects uploaded
// without checksums, which is most of the existing content.
func quietChecksums(o *s3.Options) { o.DisableLogOutputChecksumValidationSkipped = true }

// ObjectInfo describes an S3 object as returned by a bucket listing.
type ObjectInfo struct {
	Key          string
	ETag         string
	Size         int64
	LastModified time.Time
}

// ListAllObjects returns every object in the bucket, following pagination.
func ListAllObjects(bucket string) ([]ObjectInfo, error) {
	var objects []ObjectInfo
	paginator := s3.NewListObjectsV2Paginator(getS3Client(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("listing s3://%s: %w", bucket, err)
		}
		for _, obj := range page.Contents {
			objects = append(objects, ObjectInfo{
				Key:          aws.ToString(obj.Key),
				ETag:         aws.ToString(obj.ETag),
				Size:         aws.ToInt64(obj.Size),
				LastModified: aws.ToTime(obj.LastModified),
			})
		}
	}
	return objects, nil
}

// GetObjectBytes downloads an object fully into memory, returning its body and content type.
func GetObjectBytes(bucket, key string) ([]byte, string, error) {
	resp, err := getS3Client().GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, quietChecksums)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return b, aws.ToString(resp.ContentType), nil
}

// PutObjectFromFile uploads a local file to the given bucket and key.
func PutObjectFromFile(bucket, key, path, contentType string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = getS3Client().PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        f,
		ContentType: aws.String(contentType),
	})
	return err
}

// GetObjectToWriter streams an object into w. It reports found=false
// (with no error) when the key does not exist.
func GetObjectToWriter(bucket, key string, w io.Writer) (found bool, err error) {
	resp, err := getS3Client().GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return false, nil
		}
		return false, err
	}
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	return true, err
}
