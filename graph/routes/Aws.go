package routes

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	"github.com/zenith110/portfilo/graph/model"
	"golang.org/x/image/draw"
)

func UploadFileToS3(input *model.CreateArticleInfo) string {
	session, err := session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	})
	if err != nil {
		panic(fmt.Errorf("error has occured!\n%v", err))
	}
	s3ConnectionUploader := s3manager.NewUploader(session)
	srcImage, _, err := image.Decode(input.TitleCard.FileData.File)

	if err != nil {
		panic(fmt.Errorf("error has occured!\n%v", err))
	}
	var buffer bytes.Buffer
	fmt.Println(*input.TitleCard.ContentType)
	switch *input.TitleCard.ContentType {
	case "image/png":
		newImage := ImageScale(srcImage)
		err = png.Encode(&buffer, newImage)
		if err != nil {
			panic(fmt.Errorf("error has occured! could not convert image to png\n%v", err))
		}
	case "image/jpeg":
		newImage := ImageScale(srcImage)
		options := jpeg.Options{
			Quality: 100,
		}
		err = jpeg.Encode(&buffer, newImage, &options)
		if err != nil {
			panic(fmt.Errorf("error has occured! could not convert image to png\n%v", err))
		}
	case "image/jpg":
		newImage := ImageScale(srcImage)
		options := jpeg.Options{
			Quality: 100,
		}
		err = jpeg.Encode(&buffer, newImage, &options)
		if err != nil {
			panic(fmt.Errorf("error has occured! could not convert image to png\n%v", err))
		}
	}

	finalImage := bytes.NewReader(buffer.Bytes())
	_, err = s3ConnectionUploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(os.Getenv("BLOG_BUCKET")),
		Key:         aws.String(*input.URL + "/" + *input.TitleCard.Name),
		Body:        finalImage,
		ACL:         aws.String("public-read"),
		ContentType: aws.String(*input.TitleCard.ContentType),
	})

	if err != nil {
		panic(fmt.Errorf("error has occured! %s", err))
	}
	url := "https://" + os.Getenv("BLOG_BUCKET") + ".s3." + os.Getenv("AWS_REGION") + ".amazonaws.com/" + *input.URL + "/" + *input.TitleCard.Name
	image := model.Image{URL: *input.URL, Type: *input.TitleCard.ContentType, Name: *input.TitleCard.Name, UUID: uuid.NewString()}
	UploadImageDB(image, url)

	zincData := `{
		"Url":  *input.URL,
		"Type": *input.TitleCard.ContentType,
		"Name": *input.TitleCard.Name,
	}`
	CreateDocument("images", zincData, *input.UUID)
	return url
}

func UploadUpdatedFileToS3(input *model.UpdatedArticleInfo) string {
	session, err := session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	})
	if err != nil {
		panic(fmt.Errorf("error has occured!\n%v", err))
	}
	s3ConnectionUploader := s3manager.NewUploader(session)
	srcImage, _, err := image.Decode(input.TitleCard.FileData.File)

	if err != nil {
		panic(fmt.Errorf("error has occured!\n%v", err))
	}
	var buffer bytes.Buffer
	switch *input.TitleCard.ContentType {
	case "png":
		newImage := ImageScale(srcImage)
		err = png.Encode(&buffer, newImage)
		if err != nil {
			panic(fmt.Errorf("error has occured! could not convert image to png\n%v", err))
		}
	case "jpeg":
		newImage := ImageScale(srcImage)
		options := jpeg.Options{
			Quality: 100,
		}
		err = jpeg.Encode(&buffer, newImage, &options)
		if err != nil {
			panic(fmt.Errorf("error has occured! could not convert image to png\n%v", err))
		}
	case "jpg":
		newImage := ImageScale(srcImage)
		options := jpeg.Options{
			Quality: 100,
		}
		err = jpeg.Encode(&buffer, newImage, &options)
		if err != nil {
			panic(fmt.Errorf("error has occured! could not convert image to png\n%v", err))
		}
	}
	finalImage := bytes.NewReader(buffer.Bytes())
	_, err = s3ConnectionUploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(os.Getenv("BLOG_BUCKET")),
		Key:         aws.String(*input.URL + "/" + *input.TitleCard.Name),
		Body:        finalImage,
		ACL:         aws.String("public-read"),
		ContentType: aws.String(*input.TitleCard.ContentType),
	})

	if err != nil {
		panic(fmt.Errorf("error has occured! %s", err))
	}
	url := "https://" + os.Getenv("BLOG_BUCKET") + ".s3." + os.Getenv("AWS_REGION") + ".amazonaws.com/" + *input.URL + "/" + *input.TitleCard.Name
	return url
}

func DeleteArticleBucket(bucketName string) {
	session, err := session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	})
	if err != nil {
		panic(fmt.Errorf("session connection error has occured!\n%v", err))
	}
	// Makes an s3 service client
	s3sc := s3.New(session)

	iterator := s3manager.NewDeleteListIterator(s3sc, &s3.ListObjectsInput{
		Bucket: aws.String(os.Getenv("BLOG_BUCKET")),
		Prefix: aws.String(bucketName + "/"),
	})
	if err := s3manager.NewBatchDeleteWithClient(s3sc).Delete(context.Background(), iterator); err != nil {
		panic(fmt.Errorf("unable to delete objects from bucket %q, %v", bucketName, err))
	}

}
func ImageScale(srcImage image.Image) *image.RGBA {
	newImage := image.NewRGBA(image.Rect(0, 0, 345, 140))
	draw.ApproxBiLinear.Scale(newImage, newImage.Rect, srcImage, srcImage.Bounds(), draw.Over, nil)
	return newImage
}
func UploadToGallery(file *model.File) string {

	session, err := session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	})
	if err != nil {
		panic(fmt.Errorf("error has occured!\n%v", err))
	}
	s3ConnectionUploader := s3manager.NewUploader(session)
	srcImage, _, err := image.Decode(file.FileData.File)

	if err != nil {
		panic(fmt.Errorf("error has occured!\n%v", err))
	}
	var buffer bytes.Buffer
	switch *file.ContentType {
	case "image/png":
		newImage := ImageScale(srcImage)
		err = png.Encode(&buffer, newImage)
		if err != nil {
			panic(fmt.Errorf("error has occured! could not convert image to png\n%v", err))
		}
	case "image/jpeg":
		newImage := ImageScale(srcImage)
		options := jpeg.Options{
			Quality: 100,
		}
		err = jpeg.Encode(&buffer, newImage, &options)
		if err != nil {
			panic(fmt.Errorf("error has occured! could not convert image to png\n%v", err))
		}
	case "image/jpg":
		newImage := ImageScale(srcImage)
		options := jpeg.Options{
			Quality: 100,
		}
		err = jpeg.Encode(&buffer, newImage, &options)
		if err != nil {
			panic(fmt.Errorf("error has occured! could not convert image to png\n%v", err))
		}

	}

	finalImage := bytes.NewReader(buffer.Bytes())
	_, err = s3ConnectionUploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(os.Getenv("BLOG_BUCKET")),
		Key:         aws.String(*file.URL + "/" + *file.Name),
		Body:        finalImage,
		ACL:         aws.String("public-read"),
		ContentType: aws.String(*file.ContentType),
	})

	if err != nil {
		panic(fmt.Errorf("error has occured! %s", err))
	}
	url := "https://" + os.Getenv("BLOG_BUCKET") + ".s3." + os.Getenv("AWS_REGION") + ".amazonaws.com/" + *file.URL + "/" + *file.Name
	image := model.Image{URL: url, Type: *file.ContentType, Name: *file.Name, UUID: uuid.NewString()}
	UploadImageDB(image, url)
	uuid := uuid.New()
	zincData := `{
		"Url":  url,
		"Type": *file.ContentType,
		"Name": *file.Name,
	}`

	CreateDocument("images", zincData, uuid.String())
	return url
}
