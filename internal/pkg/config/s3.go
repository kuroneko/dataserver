package config

import (
	"fmt"
	"github.com/olebedev/config"
)

type BucketConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
}

func newBucketConfigFromConfig(c *config.Config, key string) (bc *BucketConfig, err error) {
	bc = &BucketConfig{}
	bc.BucketName, err = c.String(fmt.Sprintf("s3.%s.bucketName", key))
	if err != nil {
		return nil, err
	}
	bc.Endpoint = c.UString(fmt.Sprintf("s3.%s.endpoint", key), "")
	bc.AccessKeyID = c.UString(fmt.Sprintf("s3.%s.accessKeyID", key), "")
	bc.SecretAccessKey = c.UString(fmt.Sprintf("s3.%s.secretAccessKey", key), "")
	return
}

func allBucketConfigs(c *config.Config) (allConfigs []*BucketConfig, err error) {
	allConfigs = make([]*BucketConfig, 0)
	s3map, err := c.Map("s3")
	if err != nil {
		for k := range s3map {
			var bucketConfig *BucketConfig
			bucketConfig, err = newBucketConfigFromConfig(c, k)
			allConfigs = append(allConfigs, bucketConfig)
		}
	}
	return allConfigs, nil
}
