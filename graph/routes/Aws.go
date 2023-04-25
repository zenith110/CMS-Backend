package routes

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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
	case "image/gif":
		options := gif.Options{
			NumColors: 256,
		}
		err = gif.Encode(&buffer, srcImage, &options)
		if err != nil {
			panic(fmt.Errorf("error has occured! could not convert image to gif\n%v", err))
		}
	}
	return buffer
}
func UploadImageArticle(information map[string]any, jwt string) string {
	var err error
	s3ConnectionUploader := information["s3ConnectionUploader"].(*s3manager.Uploader)
	bucketName := information["bucketName"].(string)
	URL := information["URL"].(string)
	titleCardName := information["titleCardName"].(string)
	contentType := information["contentType"].(string)
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

	url := fmt.Sprintf("https://%s-images.s3.%s.amazonaws.com/%s/%s", projectUuid, os.Getenv("AWS_REGION"), URL, titleCardName)
	if err != nil {
		panic(fmt.Errorf("error has occured! %s", err))
	}
	image := model.Image{URL: URL, Type: contentType, Name: titleCardName, ArticleUUID: imageUUID, ProjectUUID: projectUuid}
	UploadImageDB(image, url, jwt, projectUuid)
	return url
}
func UploadFileToS3(input *model.CreateArticleInfo, zincpassword string) string {
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
	redisClient := RedisClientInstation()
	redisData := RedisUserInfo(input.Jwt, redisClient)
	username := redisData["username"]
	zincusername, _ := ZincLogin(input.ProjectUUID)
	bucketName := fmt.Sprintf("%s-images", zincusername)
	bucketExist := CheckIfBucketExist(s3sc, bucketName)
	finalImage := bytes.NewReader(finalImageBuffer.Bytes())

	if bucketExist == true {
		uploadImageMap := map[string]any{
			"s3ConnectionUploader": s3ConnectionUploader,
			"bucketName":           bucketName,
			"URL":                  *input.URL,
			"titleCardName":        *input.TitleCard.Name,
			"contentType":          *input.TitleCard.ContentType,
			"username":             username,
			"password":             zincpassword,
			"projectuuid":          input.ProjectUUID,
			"finalImage":           finalImage,
			"imageUUID":            *input.UUID,
		}

		return UploadImageArticle(uploadImageMap, input.Jwt)
	} else {
		CreateProjectBucket(s3sc, bucketName)
		uploadImageMap := map[string]any{
			"s3ConnectionUploader": s3ConnectionUploader,
			"bucketName":           bucketName,
			"URL":                  *input.URL,
			"titleCardName":        *input.TitleCard.Name,
			"contentType":          *input.TitleCard.ContentType,
			"username":             input.ProjectUUID,
			"password":             zincpassword,
			"projectuuid":          input.ProjectUUID,
			"finalImage":           finalImage,
			"imageUUID":            *input.UUID,
		}
		return UploadImageArticle(uploadImageMap, input.Jwt)
	}
}

func UploadUpdatedFileToS3(input *model.UpdatedArticleInfo, zincpassword string) string {
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
	finalImage := bytes.NewReader(finalImageBuffer.Bytes())
	s3sc := s3.New(session)
	zincusername, _ := ZincLogin(input.ProjectUUID)
	bucketName := fmt.Sprintf("%s-images", zincusername)
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
		url := fmt.Sprintf("https://%s-images.s3.%s.amazonaws.com/%s/%s", zincusername, os.Getenv("AWS_REGION"), *input.URL, *input.TitleCard.Name)
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
		url := fmt.Sprintf("https://%s-images.s3.%s.amazonaws.com/%s/%s", zincusername, os.Getenv("AWS_REGION"), *input.URL, *input.TitleCard.Name)
		return url
	}
}
func UploadAvatarImageCreation(input *model.UserCreation) string {
	session := CreateAWSSession()
	s3ConnectionUploader := s3manager.NewUploader(session)
	srcImage, _, err := image.Decode(input.ProfilePic.FileData.File)
	if err != nil {
		panic(fmt.Errorf("error has occured!\n%v", err))
	}
	var buffer bytes.Buffer
	storage := map[string]any{
		"buffer":           buffer,
		"imageContentType": *input.ProfilePic.ContentType,
		"srcImage":         srcImage,
	}
	finalImageBuffer := ProcessImages(storage)
	finalImage := bytes.NewReader(finalImageBuffer.Bytes())
	s3sc := s3.New(session)
	bucketName := "graphql-cms-profilepics"
	bucketExist := CheckIfBucketExist(s3sc, bucketName)
	if bucketExist == true {
		_, err = s3ConnectionUploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String(bucketName),
			Key:         aws.String(fmt.Sprintf("%s/%s", input.Username, *input.ProfilePic.Name)),
			Body:        finalImage,
			ACL:         aws.String("public-read"),
			ContentType: aws.String(*input.ProfilePic.ContentType),
		})

		if err != nil {
			panic(fmt.Errorf("error has occured! %s", err))
		}
		url := fmt.Sprintf("https://graphql-cms-profilepics.s3.%s.amazonaws.com/%s/%s", os.Getenv("AWS_REGION"), input.Username, *input.ProfilePic.Name)
		return url
	} else {
		CreateProjectBucket(s3sc, bucketName)
		_, err = s3ConnectionUploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String(bucketName),
			Key:         aws.String(fmt.Sprintf("%s/%s", input.Username, *input.ProfilePic.Name)),
			Body:        finalImage,
			ACL:         aws.String("public-read"),
			ContentType: aws.String(*input.ProfilePic.ContentType),
		})

		if err != nil {
			panic(fmt.Errorf("error has occured! %s", err))
		}
		url := fmt.Sprintf("https://graphql-cms-profilepics.s3.%s.amazonaws.com/%s/%s", os.Getenv("AWS_REGION"), input.Username, *input.ProfilePic.Name)
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
func DeleteArticleFolder(s3sc *s3.S3, iterator s3manager.BatchDeleteIterator, bucketName string) error {
	// Handle an edge case if attempting to deleting a bucket that does not exist
	if err := s3manager.NewBatchDeleteWithClient(s3sc).Delete(context.Background(), iterator); err != nil {
		fmt.Print("Skipping, bucket does not exist!\n")
		return err
	}
	fmt.Printf("Deleted object(s) from bucket: %s\n", bucketName)
	var err error
	return err
}

/*
Deletes a bucket specified by bucketname
*/
func DeleteBucket(s3sc *s3.S3, bucketName string) {
	iter := s3manager.NewDeleteListIterator(s3sc, &s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
	})
	// Empty the bucket before deleting
	err := DeleteArticleFolder(s3sc, iter, bucketName)
	if err != nil {
		return
	}
	// Makes an s3 service client
	s3sc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketName)})
	fmt.Printf("Successfully deleted %s\n", bucketName)
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
		if strings.Contains(*b.Name, "images") {
			DeleteBucket(s3sc, *b.Name)
		}
		continue
	}
}
