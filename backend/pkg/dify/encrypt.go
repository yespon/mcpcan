package dify

import (
	"encoding/base64"
	"fmt"
)

type CredentialType string

const (
	SecretInput CredentialType = "secret-input"
)

type ProviderConfig struct {
	Type CredentialType `json:"type"`
	Name string         `json:"name"`
}

type ProviderConfigEncrypter struct {
	encryptPublicKey string
	config           []ProviderConfig
}

func NewProviderConfigEncrypter(encryptPublicKey string, config []ProviderConfig) *ProviderConfigEncrypter {
	return &ProviderConfigEncrypter{
		encryptPublicKey: encryptPublicKey,
		config:           config,
	}
}

// Encrypt 加密敏感数据
// 返回原始数据的深拷贝，其中敏感字段已被加密
func (e *ProviderConfigEncrypter) Encrypt(data map[string]string) (map[string]string, error) {
	// 创建深拷贝
	encryptedData := make(map[string]string)
	for k, v := range data {
		encryptedData[k] = v
	}

	// 构建需要加密的字段映射
	fields := make(map[string]ProviderConfig)
	for _, credential := range e.config {
		fields[credential.Name] = credential
	}

	// 加密每个需要加密的字段
	for fieldName, field := range fields {
		if field.Type == SecretInput {
			if value, exists := encryptedData[fieldName]; exists {
				encrypted, err := EncryptToken(e.encryptPublicKey, value)
				if err != nil {
					return nil, fmt.Errorf("failed to encrypt field %s: %w", fieldName, err)
				}
				encryptedData[fieldName] = encrypted
			}
		}
	}

	return encryptedData, nil
}

func EncryptToken(encryptPublicKey string, token string) (string, error) {
	// 加密 token
	encryptedToken, err := Encrypt(token, []byte(encryptPublicKey))
	if err != nil {
		return "", err
	}

	// 返回 base64 编码的结果
	return base64.StdEncoding.EncodeToString(encryptedToken), nil
}
