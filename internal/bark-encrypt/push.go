package bark_encrypt

import (
	"bark-encrypt/internal/config"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

type PushHandler struct {
	Cfg config.BarkConfig
}

// NewPushHandler 创建一个新的 PushHandler 实例
func NewPushHandler(cfg config.BarkConfig) *PushHandler {
	return &PushHandler{
		Cfg: cfg,
	}
}

// EncryptAndPush 是处理 /push-ciphertext 的 Echo handler
func (h *PushHandler) EncryptAndPush(c echo.Context) error {
	// 1. 绑定并解析请求体
	// 先读取原始 body 用于加密
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read request body"})
	}

	var req PushRequestMetadata
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON format"})
	}

	if req.DeviceKey == "" || req.IV == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "device_key and iv are required"})
	}

	// 2. 执行加密
	encrypted, err := EncryptAESCBC(bodyBytes, []byte(h.Cfg.AesKey), []byte(req.IV))
	if err != nil {
		log.Printf("Encryption failed: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Encryption failed"})
	}
	ciphertextBase64 := base64.StdEncoding.EncodeToString(encrypted)

	// 3. 根据配置决定是直接返回还是转发
	// 如果没有配置 BarkDomain，则直接返回加密后的数据
	if h.Cfg.Domain == "" {
		return c.JSON(http.StatusOK, PushResponse{
			Ciphertext: ciphertextBase64,
			IV:         req.IV,
		})
	}

	// 否则，将请求转发到 Bark 服务器
	resp, err := ForwardToBark(h.Cfg.Domain, req.DeviceKey, ciphertextBase64, req.IV)
	if err != nil {
		log.Printf("Failed to forward request to Bark API: %v", err)
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Failed to forward request to Bark API"})
	}
	defer resp.Body.Close()

	log.Printf("Forwarded request for deviceKey ending in ...%s, status: %d", req.DeviceKey[len(req.DeviceKey)-4:], resp.StatusCode)

	// 将 Bark 服务器的响应头和响应体透传给客户端
	for k, v := range resp.Header {
		c.Response().Header().Set(k, v[0])
	}
	c.Response().WriteHeader(resp.StatusCode)
	_, _ = io.Copy(c.Response().Writer, resp.Body)
	return nil
}

// ForwardToBark 将加密后的数据转发到指定的 Bark 服务器
func ForwardToBark(barkDomain, deviceKey, ciphertextBase64, iv string) (*http.Response, error) {
	formData := url.Values{}
	formData.Set("ciphertext", ciphertextBase64)
	formData.Set("iv", iv)

	// 构建目标 URL，确保 URL 路径拼接正确
	targetURL, err := url.JoinPath(barkDomain, deviceKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create target URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, targetURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request for bark: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	return client.Do(req)
}
