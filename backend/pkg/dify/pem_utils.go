package dify

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// parseRSAPublicKey 解析 PEM 格式的 RSA 公钥
func parseRSAPublicKey(publicKeyPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	// 尝试解析为 PKIX 格式的公钥
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPubKey, nil
}
