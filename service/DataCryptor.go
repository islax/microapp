package service

// DataCryptor encrypts/decrypts data
type DataCryptor interface {
	Encrypt(data string, salt string) (string, error)
	Decrypt(data string, salt string) (string, error)
}
