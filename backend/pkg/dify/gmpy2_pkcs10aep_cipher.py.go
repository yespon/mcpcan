package dify

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"errors"
	"hash"
	"math/big"
)

// PKCS1OAEPCipher 实现 PKCS#1 OAEP 加密
type PKCS1OAEPCipher struct {
	key      *rsa.PublicKey
	hash     hash.Hash
	mgf      func(seed []byte, length int) ([]byte, error)
	label    []byte
	randFunc func([]byte) (int, error)
}

// NewPKCS1OAEPCipher 创建新的 OAEP 密码器
func NewPKCS1OAEPCipher(
	key *rsa.PublicKey,
	hash hash.Hash,
	mgf func(seed []byte, length int) ([]byte, error),
	label []byte,
	randFunc func([]byte) (int, error),
) *PKCS1OAEPCipher {
	if hash == nil {
		hash = sha1.New()
	}
	if mgf == nil {
		mgf = MGF1
	}
	if randFunc == nil {
		randFunc = rand.Read
	}
	return &PKCS1OAEPCipher{
		key:      key,
		hash:     hash,
		mgf:      mgf,
		label:    label,
		randFunc: randFunc,
	}
}

// Encrypt 加密消息
func (c *PKCS1OAEPCipher) Encrypt(message []byte) ([]byte, error) {
	// 步骤 1b：计算参数
	modBits := c.key.N.BitLen()
	k := (modBits + 7) / 8 // 从位转换为字节
	hLen := c.hash.Size()
	mLen := len(message)

	// 计算 PS 长度
	psLen := k - mLen - 2*hLen - 2
	if psLen < 0 {
		return nil, errors.New("plaintext is too long")
	}

	// 步骤 2a：计算 lHash
	lHash := sha1.Sum(c.label)

	// 步骤 2b：创建 PS (全零填充)
	ps := make([]byte, psLen)

	// 步骤 2c：构造 DB
	db := make([]byte, 0, hLen+psLen+1+mLen)
	db = append(db, lHash[:]...)
	db = append(db, ps...)
	db = append(db, 0x01)
	db = append(db, message...)

	// 步骤 2d：生成随机种子
	seed := make([]byte, hLen)
	if _, err := c.randFunc(seed); err != nil {
		return nil, err
	}

	// 步骤 2e：生成 dbMask
	dbMask, err := c.mgf(seed, k-hLen-1)
	if err != nil {
		return nil, err
	}

	// 步骤 2f：异或得到 maskedDB
	maskedDB := xorBytes(db, dbMask)

	// 步骤 2g：生成 seedMask
	seedMask, err := c.mgf(maskedDB, hLen)
	if err != nil {
		return nil, err
	}

	// 步骤 2h：异或得到 maskedSeed
	maskedSeed := xorBytes(seed, seedMask)

	// 步骤 2i：构造 EM
	em := make([]byte, 0, k)
	em = append(em, 0x00)
	em = append(em, maskedSeed...)
	em = append(em, maskedDB...)

	// 步骤 3a：将 EM 转换为整数
	emInt := new(big.Int).SetBytes(em)

	// 步骤 3b：RSA 加密
	// c = m^e mod n
	cInt := new(big.Int).Exp(emInt, big.NewInt(int64(c.key.E)), c.key.N)

	// 步骤 3c：将整数转换回字节
	ciphertext := cInt.Bytes()

	// 确保密文长度为 k 字节
	if len(ciphertext) < k {
		// 左侧填充零
		padding := make([]byte, k-len(ciphertext))
		ciphertext = append(padding, ciphertext...)
	}

	return ciphertext, nil
}

// MGF1 掩码生成函数
func MGF1(seed []byte, length int) ([]byte, error) {
	if length < 0 {
		return nil, errors.New("mask length must be non-negative")
	}

	//hlen := sha1.Size
	T := make([]byte, 0, length)

	for counter := 0; len(T) < length; counter++ {
		// C = I2OSP(counter, 4)
		C := []byte{
			byte(counter >> 24),
			byte(counter >> 16),
			byte(counter >> 8),
			byte(counter),
		}

		// hash = Hash(seed || C)
		hash := sha1.New()
		hash.Write(seed)
		hash.Write(C)
		hashBytes := hash.Sum(nil)

		T = append(T, hashBytes...)
	}

	return T[:length], nil
}

// xorBytes 字节数组异或
func xorBytes(a, b []byte) []byte {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}

	result := make([]byte, n)
	for i := 0; i < n; i++ {
		result[i] = a[i] ^ b[i]
	}
	return result
}
