package bark_encrypt

type PushRequestMetadata struct {
	DeviceKey string `json:"device_key"`
	IV        string `json:"iv"` // Base64 编码的 IV
}

// PushResponse 定义了当不转发时，本地返回的 JSON 响应
type PushResponse struct {
	Ciphertext string `json:"ciphertext"` // Base64 编码的密文
	IV         string `json:"iv"`         // Base64 编码的 IV
}
