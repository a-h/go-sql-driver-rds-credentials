package sm

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// DefaultRetrieve retrieves data from AWS Secrets Manager.
func DefaultRetrieve(name string) (secret string, err error) {
	svc := secretsmanager.New(session.New())
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
