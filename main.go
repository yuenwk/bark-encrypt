package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func encrypt(plainText, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plainText = PKCS7Padding(plainText, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, iv)
	encrypted := make([]byte, len(plainText))
	blockMode.CryptBlocks(encrypted, plainText)
	return encrypted, nil
}

type RequestData struct {
	DeviceKey string `json:"device_key"`
	Iv        string `json:"iv"`
}

func handler(w http.ResponseWriter, r *http.Request, aesKey []byte, barkDomain string) {
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	var data RequestData
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	deviceKey := data.DeviceKey
	iv := data.Iv

	encrypted, err := encrypt(body, aesKey, []byte(iv))
	if err != nil {
		http.Error(w, "Encryption failed", http.StatusInternalServerError)
		return
	}
	ciphertextBase64 := base64.StdEncoding.EncodeToString(encrypted)

	if len(barkDomain) <= 0 {
		data := map[string]interface{}{
			"ciphertext": ciphertextBase64,
			"iv":         iv,
		}

		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
			return
		}
		return
	}

	formData := url.Values{}
	formData.Set("ciphertext", ciphertextBase64)
	formData.Set("iv", iv)

	targetURL := barkDomain + "/" + deviceKey
	resp, err := http.Post(targetURL, "application/x-www-form-urlencoded", strings.NewReader(formData.Encode()))
	if err != nil {
		http.Error(w, "Failed to forward request to Bark API", http.StatusBadGateway)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	log.Printf("Forwarded request for deviceKey ending in ...%s, status: %d", deviceKey[len(deviceKey)-4:], resp.StatusCode)

	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func main() {
	aesKeyStr := os.Getenv("BARK_AES_KEY")
	if len(aesKeyStr) != 16 {
		log.Fatal("Fatal: BARK_AES_KEY environment variable not set or not 16 characters long.")
	}

	barkDomain := os.Getenv("BARK_DOMAIN")

	aesKeyBytes := []byte(aesKeyStr)

	http.HandleFunc("/push-ciphertext", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, aesKeyBytes, barkDomain)
	})

	log.Println("Starting encryption service on :9090")
	if err := http.ListenAndServe(":9090", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
