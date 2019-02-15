package sm

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// DefaultRetrieve retrieves data from AWS Secrets Manager.
func DefaultRetrieve(name string) (secret string, err error) {
	cfg := aws.NewConfig()
	if region, ok := getRegionFromARN(name); ok {
		cfg = cfg.WithRegion(region)
	}
	svc := secretsmanager.New(session.New(cfg))
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(name),
		VersionStage: aws.String("AWSCURRENT"),
	}
	var result *secretsmanager.GetSecretValueOutput
	result, err = svc.GetSecretValue(input)
	if err != nil {
		return
	}
	secret = *result.SecretString
	return
}

func getRegionFromARN(arn string) (region string, ok bool) {
	// arn:partition:service:region:account-id:resource
	split := strings.Split(arn, ":")
	if len(split) < 4 {
		return
	}
	region = split[3]
	ok = true
	return
}
