package services

import (
	"encoding/base64"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
)

// Encrypter is an interface for the Encrypt method.
type Encrypter interface {
	// Encrypt takes a string, encrypts it in some manner, and returns the result as a string.
	Encrypt(s string) (string, error)
}

type KMSEncrypter struct {
	KeyID  string
	Client *kms.KMS
}

func (e *KMSEncrypter) Encrypt(s string) (string, error) {
	out, err := e.Client.Encrypt(&kms.EncryptInput{
		KeyId:               aws.String(e.KeyID),
		Plaintext:           []byte(s),
		EncryptionAlgorithm: aws.String(kms.EncryptionAlgorithmSpecSymmetricDefault),
	})
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(out.CiphertextBlob), nil
}

type ROT13Encrypter struct{}

func (e *ROT13Encrypter) Encrypt(s string) (string, error) {
	rot13 := func(r rune) rune {
		switch {
		case 'A' <= r && r <= 'Z':
			return 'A' + (r-'A'+13)%26
		case 'a' <= r && r <= 'z':
			return 'a' + (r-'a'+13)%26
		}
		return r
	}
	return strings.Map(rot13, s), nil
}
