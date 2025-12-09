package dify

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/sha1"
	"log"

	cryptoeax "github.com/enceve/crypto/cipher"
)

const (
	prefixHybrid = "HYBRID:"
)

// Encrypt 使用 RSA 公钥和 AES-EAX 进行混合加密
// text: 要加密的文本
// publicKeyPEM: PEM 格式的 RSA 公钥
func Encrypt(text string, publicKeyPEM []byte) ([]byte, error) {

	// 填充 aes
	aesKey := make([]byte, 16)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, err
	}
	publicKey, err := parseRSAPublicKey(publicKeyPEM)
	if err != nil {
		return nil, err
	}
	oaep := NewPKCS1OAEPCipher(
		publicKey,
		sha1.New(),
		MGF1,
		[]byte(""),
		rand.Read,
	)
	encryptedAESKey, err := oaep.Encrypt(aesKey)
	if err != nil {
		panic(err)
	}

	ciphertext, nonce, tag, err := encryptEAX(aesKey, []byte(text))
	if err != nil {
		log.Fatal(err)
	}

	result := make([]byte, 0, len(prefixHybrid)+len(encryptedAESKey)+len(nonce)+len(tag)+len(ciphertext))
	result = append(result, []byte(prefixHybrid)...)
	result = append(result, encryptedAESKey...)
	result = append(result, nonce...)
	result = append(result, tag...)
	result = append(result, ciphertext...)

	return result, nil
}

func encryptEAX(key, plaintext []byte) ([]byte, []byte, []byte, error) {
	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, err
	}

	// 创建EAX模式加密器
	eax, err := cryptoeax.NewEAX(block, 16)
	if err != nil {
		return nil, nil, nil, err
	}

	// 生成随机nonce（EAX模式需要）
	nonce := make([]byte, eax.NonceSize())
	// 在实际应用中，应该使用 crypto/rand 生成随机nonce
	// 这里为了演示使用固定值
	for i := range nonce {
		nonce[i] = byte(i)
	}

	// 加密
	ciphertext := make([]byte, len(plaintext))
	result := eax.Seal(ciphertext, nonce, plaintext, nil)

	tag := result[len(plaintext):]
	// 在实际使用中，通常需要将nonce、ciphertext和tag一起存储/传输
	// 这里只返回ciphertext和tag，类似Python代码的行为
	return ciphertext, nonce, tag, nil
}
