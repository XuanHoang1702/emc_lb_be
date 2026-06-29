package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	ctx := context.Background()
	region := "us-east-1"
	endpoint := "http://localhost:4566"
	accessKeyID := "test"
	secretAccessKey := "test"
	bucket := "ecommerce-images"

	cfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.UsePathStyle = true
		options.BaseEndpoint = aws.String(endpoint)
	})

	// Create bucket
	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			log.Fatalf("failed to create bucket: %v", err)
		}
	}

	// Set bucket policy to public read
	policy := fmt.Sprintf(`{
		"Version":"2012-10-17",
		"Statement":[
		  {
			"Sid":"PublicReadGetObject",
			"Effect":"Allow",
			"Principal": "*",
			"Action":["s3:GetObject"],
			"Resource":["arn:aws:s3:::%s/*"]
		  }
		]
	  }`, bucket)

	_, err = client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucket),
		Policy: aws.String(policy),
	})
	if err != nil {
		log.Printf("warning: failed to set public bucket policy: %v", err)
	}

	images := map[string]string{
		"hero.png":    `C:\Users\tranx\.gemini\antigravity-ide\brain\1ab68181-e8d8-48ce-9164-7ad889461191\hero_banner_women_1782743408931.png`,
		"dress.png":   `C:\Users\tranx\.gemini\antigravity-ide\brain\1ab68181-e8d8-48ce-9164-7ad889461191\product_dress_1782743420653.png`,
		"jewelry.png": `C:\Users\tranx\.gemini\antigravity-ide\brain\1ab68181-e8d8-48ce-9164-7ad889461191\product_jewelry_1782743433067.png`,
	}

	for key, path := range images {
		file, err := os.Open(path)
		if err != nil {
			log.Printf("failed to open %s: %v", path, err)
			continue
		}
		
		_, err = client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(bucket),
			Key:         aws.String(key),
			Body:        file,
			ContentType: aws.String("image/png"),
		})
		file.Close()
		
		if err != nil {
			log.Printf("failed to upload %s: %v", key, err)
		} else {
			fmt.Printf("Successfully uploaded %s to %s/%s/%s\n", path, endpoint, bucket, key)
		}
	}
}
