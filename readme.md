### start up

config.yaml

```yaml
port: 9090
bark:
  aes_key: 1234567890123456
  domain: bark.domain.com
```

```shell
cd ci
docker compose up -d 
```

### test

```shell

curl -X POST --location "http://localhost:9090/push-ciphertext" \
    -H "Content-Type: application/json" \
    -d '{
          "title": "test",
          "device_key": "xxx",
          "sound": "newsflash",
          "iv": "1234567890123456"
        }'

```