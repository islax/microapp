package impl

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	mathRand "math/rand"
	"strings"
	"time"

	"github.com/islax/microapp"
	microappError "github.com/islax/microapp/error"
	"github.com/islax/microapp/service"
	"golang.org/x/crypto/scrypt"
)

const (
	extraSaltLength          = 12
	randomPrefixStringLength = 8
	version2Prefix           = "v2.0"
)

type aesDataCryptor struct {
	app *microapp.App
}

// NewAESDataCryptor creates a new AES data cryptor
func NewAESDataCryptor(app *microapp.App) service.DataCryptor {
	return &aesDataCryptor{app}
}

func (cryptor *aesDataCryptor) Encrypt(data string, salt string) (string, error) {
	salt, extraSaltAdded := cryptor.addSomeExtraSalt(salt)
	key, err := cryptor.deriveKey(cryptor.app.Config.GetString("CRYPTO_KEY"), salt)
	// key, err := cryptor.deriveKey("some key###", salt)
	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return "", microappError.NewUnexpectedError(microappError.ErrorCodeCryptoFailure, err)
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return "", microappError.NewUnexpectedError(microappError.ErrorCodeCryptoFailure, err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", microappError.NewUnexpectedError(microappError.ErrorCodeCryptoFailure, err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)
	return version2Prefix + cryptor.getRandomString(randomPrefixStringLength) + base64.StdEncoding.EncodeToString(ciphertext) + extraSaltAdded, nil
}

func (cryptor *aesDataCryptor) Decrypt(data string, salt string) (string, error) {

	extraSaltBeginIndex := len(data) - extraSaltLength
	salt = salt + data[extraSaltBeginIndex:]
	prefixLength := len(version2Prefix) + randomPrefixStringLength
	base64EncodedDataString := data[prefixLength:extraSaltBeginIndex]
	dataBytes, err := base64.StdEncoding.DecodeString(base64EncodedDataString)
	if err != nil {
		return "", microappError.NewUnexpectedError(microappError.ErrorCodeCryptoFailure, err)
	}

	key, err := cryptor.deriveKey(cryptor.app.Config.GetString("CRYPTO_KEY"), salt)
	if err != nil {
		return "", microappError.NewUnexpectedError(microappError.ErrorCodeCryptoFailure, err)
	}
	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return "", microappError.NewUnexpectedError(microappError.ErrorCodeCryptoFailure, err)
	}
	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return "", microappError.NewUnexpectedError(microappError.ErrorCodeCryptoFailure, err)
	}
	nonce, ciphertext := dataBytes[:gcm.NonceSize()], dataBytes[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", microappError.NewUnexpectedError(microappError.ErrorCodeCryptoFailure, err)
	}
	return string(plaintext), nil
}

func (cryptor *aesDataCryptor) deriveKey(password, salt string) ([]byte, error) {

	key, err := scrypt.Key([]byte(password), []byte(salt), 16384, 8, 1, 32)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (cryptor aesDataCryptor) addSomeExtraSalt(salt string) (string, string) {
	extraSalt := cryptor.getRandomString(extraSaltLength)
	return salt + extraSalt, extraSalt
}

func (cryptor aesDataCryptor) getRandomString(length int) string {
	//generate random string
	mathRand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890~!@#$%^&[]{}*()-=_+|;',./:<>?")
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[mathRand.Intn(len(chars))])
	}

	return b.String()
}
