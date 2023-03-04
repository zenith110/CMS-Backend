package routes

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	"github.com/zenith110/CMS-Backend/graph/model"
)

type ArticleJson struct {
	Name        string
	Author      string
	Date        string
	ImageUrl    string
	Content     string
	Description string
}

/*
Creates a reusable session for AWS S3 usage
*/
func CreateAWSSession() *session.Session {
	session, err := session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	})
	if err != nil {
		panic(fmt.Errorf("session connection error has occured!\n%v", err))
	}

	return session
}

/*
Takes in a map, and processes the image to be able to upload to AWS
*/
func ProcessImages(storage map[string]any) bytes.Buffer {
	srcImage := storage["srcImage"].(image.Image)
	buffer := storage["buffer"].(bytes.Buffer)
	var err error
	imageContentType := storage["imageContentType"].(string)

	switch imageContentType {
	case "image/png":
		err = png.Encode(&buffer, srcImage)
		if err != nil {
			panic(fmt.Errorf("error has occured! could not convert image to png\n%v", err))
		}
	case "image/jpeg":
		options := jpeg.Options{
			Quality: 100,
		}
		err = jpeg.Encode(&buffer, srcImage, &options)
		if err != nil {
			panic(fmt.Errorf("error has occured! could not convert image to png\n%v", err))
		}
	case "image/jpg":
		options := jpeg.Options{
			Quality: 100,
		}
		err = jpeg.Encode(&buffer, srcImage, &options)
		if err != nil {
			panic(fmt.Errorf("error has occured! could not convert image to png\n%v", err))
		}
	}
	return buffer
}
func UploadImage(information map[string]any) string {
	var err error
	s3ConnectionUploader := information["s3ConnectionUploader"].(*s3manager.Uploader)
	bucketName := information["bucketName"].(string)
	URL := information["URL"].(string)
	titleCardName := information["titleCardName"].(string)
	contentType := information["contentType"].(string)
	username := information["username"].(string)
	password := information["password"].(string)
	projectUuid := information["projectuuid"].(string)
	finalImage := information["finalImage"].(*bytes.Reader)
	imageUUID := information["imageUUID"].(string)
	_, err = s3ConnectionUploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(fmt.Sprintf("%s/%s", URL, titleCardName)),
		Body:        finalImage,
		ACL:         aws.String("public-read"),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		panic(fmt.Errorf("error has occured in uploading! %s", err))
	}
	url := fmt.Sprintf("https://%s-%s-images.s3.%s.amazonaws.com/%s/%s", username, projectUuid, os.Getenv("AWS_REGION"), URL, titleCardName)
	if err != nil {
		panic(fmt.Errorf("error has occured! %s", err))
	}
	image := model.Image{URL: URL, Type: contentType, Name: titleCardName, UUID: uuid.NewString()}
	UploadImageDB(image, url, username, projectUuid)

	zincData := fmt.Sprintf(`{
		"Url": "%s",
		"Type": "%s",
		"Name": "%s"
	}`, url, contentType, titleCardName)
	CreateDocument(bucketName, zincData, imageUUID, username, password)
	return url
}
func UploadFileToS3(input *model.CreateArticleInfo) string {
	session := CreateAWSSession()
	s3ConnectionUploader := s3manager.NewUploader(session)
	srcImage, _, err := image.Decode(input.TitleCard.FileData.File)

	if err != nil {
		panic(fmt.Errorf("error has occured!\n%v", err))
	}
	var buffer bytes.Buffer
	storage := map[string]any{
		"buffer":           buffer,
		"imageContentType": *input.TitleCard.ContentType,
		"srcImage":         srcImage,
	}

	finalImageBuffer := ProcessImages(storage)

	// Create S3 service client
	s3sc := s3.New(session)
	bucketName := fmt.Sprintf("%s-%s-images", input.Username, input.ProjectUUID)
	bucketExist := CheckIfBucketExist(s3sc, bucketName)
	finalImage := bytes.NewReader(finalImageBuffer.Bytes())

	if bucketExist == true {
		uploadImageMap := map[string]any{
			"s3ConnectionUploader": s3ConnectionUploader,
			"bucketName":           bucketName,
			"URL":                  *input.URL,
			"titleCardName":        *input.TitleCard.Name,
			"contentType":          *input.TitleCard.ContentType,
			"username":             input.Username,
			"password":             input.Password,
			"projectuuid":          input.ProjectUUID,
			"finalImage":           finalImage,
			"imageUUID":            *input.UUID,
		}

		return UploadImage(uploadImageMap)
	} else {
		CreateProjectBucket(s3sc, bucketName)
		uploadImageMap := map[string]any{
			"s3ConnectionUploader": s3ConnectionUploader,
			"bucketName":           bucketName,
			"URL":                  *input.URL,
			"titleCardName":        *input.TitleCard.Name,
			"contentType":          *input.TitleCard.ContentType,
			"username":             input.Username,
			"password":             input.Password,
			"projectuuid":          input.ProjectUUID,
			"finalImage":           finalImage,
			"imageUUID":            *input.UUID,
		}
		return UploadImage(uploadImageMap)
	}
}

func UploadUpdatedFileToS3(input *model.UpdatedArticleInfo) string {
	session := CreateAWSSession()
	s3ConnectionUploader := s3manager.NewUploader(session)
	srcImage, _, err := image.Decode(input.TitleCard.FileData.File)
	if err != nil {
		panic(fmt.Errorf("error has occured!\n%v", err))
	}
	var buffer bytes.Buffer
	storage := map[string]any{
		"buffer":           buffer,
		"imageContentType": *input.TitleCard.ContentType,
		"srcImage":         srcImage,
	}
	finalIMageBuffer := ProcessImages(storage)
	finalImage := bytes.NewReader(finalIMageBuffer.Bytes())
	s3sc := s3.New(session)
	bucketName := fmt.Sprintf("%s-%s-images", input.Username, input.ProjectUUID)
	bucketExist := CheckIfBucketExist(s3sc, bucketName)
	if bucketExist == true {
		_, err = s3ConnectionUploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String(bucketName),
			Key:         aws.String(*input.URL + "/" + *input.TitleCard.Name),
			Body:        finalImage,
			ACL:         aws.String("public-read"),
			ContentType: aws.String(*input.TitleCard.ContentType),
		})

		if err != nil {
			panic(fmt.Errorf("error has occured! %s", err))
		}
		url := fmt.Sprintf("https://%s-%s-%s-images.s3.%s.amazonaws.com/%s/%s", input.Username, input.ProjectUUID, *input.UUID, os.Getenv("AWS_REGION"), *input.URL, *input.TitleCard.Name)
		return url
	} else {
		CreateProjectBucket(s3sc, bucketName)
		_, err = s3ConnectionUploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String(bucketName),
			Key:         aws.String(*input.URL + "/" + *input.TitleCard.Name),
			Body:        finalImage,
			ACL:         aws.String("public-read"),
			ContentType: aws.String(*input.TitleCard.ContentType),
		})

		if err != nil {
			panic(fmt.Errorf("error has occured! %s", err))
		}
		url := fmt.Sprintf("https://%s-%s-%s-images.s3.%s.amazonaws.com/%s/%s", input.Username, input.ProjectUUID, *input.UUID, os.Getenv("AWS_REGION"), *input.URL, *input.TitleCard.Name)
		return url
	}
}
func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
func CheckIfBucketExist(s3sc *s3.S3, bucketName string) bool {
	result, err := s3sc.ListBuckets(nil)
	if err != nil {
		exitErrorf("Unable to list buckets, %v", err)
	}
	exist := false

	for _, b := range result.Buckets {
		if aws.StringValue(b.Name) == bucketName {
			exist = true
		} else {
			continue
		}
	}
	return exist
}
func CreateProjectBucket(s3sc *s3.S3, bucketName string) {
	var err error
	_, err = s3sc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		exitErrorf("Unable to create bucket %q, %v", bucketName, err)
	}

	// Wait until bucket is created before finishing
	fmt.Printf("Waiting for bucket %q to be created...\n", bucketName)

	err = s3sc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
}

/*
Deletes individual article folders within a user's project bucket
*/
func DeleteArticleFolder(s3sc *s3.S3, iterator s3manager.BatchDeleteIterator, bucketName string) {
	if err := s3manager.NewBatchDeleteWithClient(s3sc).Delete(context.Background(), iterator); err != nil {
		panic(fmt.Errorf("unable to delete objects from bucket %q, %v", bucketName, err))
	}
	fmt.Printf("Deleted object(s) from bucket: %s", bucketName)
}

/*
Deletes a bucket specified by bucketname
*/
func DeleteBucket(s3sc *s3.S3, bucketName string) {
	iter := s3manager.NewDeleteListIterator(s3sc, &s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
	})
	// Empty the bucket before deleting
	DeleteArticleFolder(s3sc, iter, bucketName)
	// Makes an s3 service client
	s3sc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketName)})
	fmt.Printf("Successfully deleted %s", bucketName)
}

/*
Deletes all projects buckets tied to a specific user
*/
func DeleteAllProjectsBuckets(username string) {
	session := CreateAWSSession()
	// Makes an s3 service client
	s3sc := s3.New(session)
	result, err := s3sc.ListBuckets(nil)
	if err != nil {
		exitErrorf("Unable to list buckets, %v", err)
	}

	for _, b := range result.Buckets {
		if strings.Contains(*b.Name, username) {
			DeleteBucket(s3sc, *b.Name)
		}
		continue
	}
}

// func UploadToGallery(file *model.File) string {

// 	session, err := session.NewSession(&aws.Config{
// 		Region:      aws.String(os.Getenv("AWS_REGION")),
// 		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
// 	})
// 	if err != nil {
// 		panic(fmt.Errorf("error has occured!\n%v", err))
// 	}
// 	s3ConnectionUploader := s3manager.NewUploader(session)
// 	srcImage, _, err := image.Decode(file.FileData.File)

// 	if err != nil {
// 		panic(fmt.Errorf("error has occured!\n%v", err))
// 	}
// 	var buffer bytes.Buffer
// 	switch *file.ContentType {
// 	case "image/png":
// 		newImage := ImageScale(srcImage)
// 		err = png.Encode(&buffer, newImage)
// 		if err != nil {
// 			panic(fmt.Errorf("error has occured! could not convert image to png\n%v", err))
// 		}
// 	case "image/jpeg":
// 		newImage := ImageScale(srcImage)
// 		options := jpeg.Options{
// 			Quality: 100,
// 		}
// 		err = jpeg.Encode(&buffer, newImage, &options)
// 		if err != nil {
// 			panic(fmt.Errorf("error has occured! could not convert image to png\n%v", err))
// 		}
// 	case "image/jpg":
// 		newImage := ImageScale(srcImage)
// 		options := jpeg.Options{
// 			Quality: 100,
// 		}
// 		err = jpeg.Encode(&buffer, newImage, &options)
// 		if err != nil {
// 			panic(fmt.Errorf("error has occured! could not convert image to png\n%v", err))
// 		}

// 	}

// 	finalImage := bytes.NewReader(buffer.Bytes())
// 	_, err = s3ConnectionUploader.Upload(&s3manager.UploadInput{
// 		Bucket:      aws.String(os.Getenv("BLOG_BUCKET")),
// 		Key:         aws.String(*file.URL + "/" + *file.Name),
// 		Body:        finalImage,
// 		ACL:         aws.String("public-read"),
// 		ContentType: aws.String(*file.ContentType),
// 	})

// 	if err != nil {
// 		panic(fmt.Errorf("error has occured! %s", err))
// 	}
// 	url := "https://" + os.Getenv("BLOG_BUCKET") + ".s3." + os.Getenv("AWS_REGION") + ".amazonaws.com/" + *file.URL + "/" + *file.Name
// 	image := model.Image{URL: url, Type: *file.ContentType, Name: *file.Name, UUID: uuid.NewString()}
// 	UploadImageDB(image, url)
// 	uuid := uuid.New()
// 	zincData := fmt.Sprintf(`{
// 		"Url": "%s",
// 		"Type": "%s",
// 		"Name": "%s"
// 	}`, url, *file.ContentType, *file.Name)
// 	CreateDocument("images", zincData, uuid.String())
// 	return url
// }
