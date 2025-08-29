package bark_encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// PKCS7Padding 对数据进行 PKCS7 填充
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// EncryptAESCBC 使用 AES CBC 模式加密数据
func EncryptAESCBC(plainText, key, iv []byte) ([]byte, error) {
	// 校验密钥长度
	switch len(key) {
	case 16, 24, 32:
		break
	default:
		return nil, fmt.Errorf("invalid AES key size: %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		// 这个错误理论上不应该发生，因为我们已经检查了密钥长度
		return nil, fmt.Errorf("failed to create aes cipher: %w", err)
	}

	// 校验 IV 长度
	if len(iv) != block.BlockSize() {
		return nil, fmt.Errorf("invalid IV size: got %d, want %d", len(iv), block.BlockSize())
	}

	paddedPlainText := PKCS7Padding(plainText, block.BlockSize())

	encrypted := make([]byte, len(paddedPlainText))
	blockMode := cipher.NewCBCEncrypter(block, iv)
	blockMode.CryptBlocks(encrypted, paddedPlainText)

	return encrypted, nil
}
