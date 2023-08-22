package uservaults

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"strconv"

	pbCommon "davensi.com/core/gen/common"
	pbUserVaults "davensi.com/core/gen/uservaults"
)

const (
	encryptKey = "the-key-has-to-be-32-bytes-long!"
)

func encrypt(plaintext, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decrypt(ciphertext, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func bytesToValue(data []byte, valueType pbUserVaults.ValueType) (*pbUserVaults.Value, error) {
	valStr := string(data)
	if valueType == pbUserVaults.ValueType_VALUE_TYPE_BOOL {
		if valStr == "true" {
			return &pbUserVaults.Value{
				Value: &pbUserVaults.Value_Bool{
					Bool: true,
				},
			}, nil
		} else {
			return &pbUserVaults.Value{
				Value: &pbUserVaults.Value_Bool{
					Bool: false,
				},
			}, nil
		}
	}
	if valueType == pbUserVaults.ValueType_VALUE_TYPE_INT64 {
		n, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			return nil, err
		}
		return &pbUserVaults.Value{
			Value: &pbUserVaults.Value_Integer{
				Integer: n,
			},
		}, nil
	}
	if valueType == pbUserVaults.ValueType_VALUE_TYPE_DECIMAL {
		return &pbUserVaults.Value{
			Value: &pbUserVaults.Value_Decimal{
				Decimal: &pbCommon.Decimal{
					Value: valStr,
				},
			},
		}, nil
	}
	if valueType == pbUserVaults.ValueType_VALUE_TYPE_STRING {
		return &pbUserVaults.Value{
			Value: &pbUserVaults.Value_String_{
				String_: valStr,
			},
		}, nil
	}
	if valueType == pbUserVaults.ValueType_VALUE_TYPE_BYTE {
		return &pbUserVaults.Value{
			Value: &pbUserVaults.Value_Bytes{
				Bytes: data,
			},
		}, nil
	}

	return nil, nil
}

func bytesToKeyValue(data []byte, valueType pbUserVaults.ValueType) (*pbUserVaults.KeyValue, error) {
	valStr := string(data)

	if valueType == pbUserVaults.ValueType_VALUE_TYPE_INT64 {
		n, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			return nil, err
		}
		return &pbUserVaults.KeyValue{
			Value: &pbUserVaults.KeyValue_Integer{
				Integer: n,
			},
		}, nil
	}
	if valueType == pbUserVaults.ValueType_VALUE_TYPE_BOOL {
		if valStr == "true" {
			return &pbUserVaults.KeyValue{
				Value: &pbUserVaults.KeyValue_Bool{
					Bool: true,
				},
			}, nil
		} else {
			return &pbUserVaults.KeyValue{
				Value: &pbUserVaults.KeyValue_Bool{
					Bool: false,
				},
			}, nil
		}
	}
	if valueType == pbUserVaults.ValueType_VALUE_TYPE_DECIMAL {
		return &pbUserVaults.KeyValue{
			Value: &pbUserVaults.KeyValue_Decimal{
				Decimal: &pbCommon.Decimal{
					Value: valStr,
				},
			},
		}, nil
	}
	if valueType == pbUserVaults.ValueType_VALUE_TYPE_STRING {
		return &pbUserVaults.KeyValue{
			Value: &pbUserVaults.KeyValue_String_{
				String_: valStr,
			},
		}, nil
	}
	if valueType == pbUserVaults.ValueType_VALUE_TYPE_BYTE {
		return &pbUserVaults.KeyValue{
			Value: &pbUserVaults.KeyValue_Bytes{
				Bytes: data,
			},
		}, nil
	}

	return nil, nil
}
