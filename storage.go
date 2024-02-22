package main

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func getConfig(regionName string) aws.Config {
	cfg, err := config.LoadDefaultConfig(context.Background(), 
	config.WithRegion(regionName))
	if err != nil {
		log.Fatalf("Unable to load SDK config %v", err)
	}
	return cfg
}


func getPresignURL(cfg aws.Config, bucket, key string) string {
	s3client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(s3client)
	presignedurl, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key: aws.String(key),
	})
	if err != nil {
		log.Fatal(err)
	}
	_ = s3.WithPresignExpires(15 * time.Minute)
	return presignedurl.URL
}


func putPresignURL(cfg aws.Config, bucket, key string) string {

	s3client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(s3client)

	presignedurl, err := presignClient.PresignPutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key: aws.String(key),
		ContentType: aws.String("multipart/form-data"),
	})

	if err != nil {
		log.Fatal(err)
	}
	_ = s3.WithPresignExpires(15 * time.Minute)
	return presignedurl.URL
}



