package entities

import "time"

type User struct {
	ID           int
	Name         string
	Email        string
	PasswordHash string

	AuthKeyHash            string
	KDFAlgorithm           string
	KDFIterations          int
	KDFSalt                []byte
	KDFVersion             int
	WrappedDataKey         []byte
	WrappedDataKeyIV       []byte
	RecoveryWrappedDataKey []byte
	RecoveryDataKeyIV      []byte
	RecoveryVerifier       []byte
	CryptoSetupAt          *time.Time
}

func (u User) CryptoReady() bool {
	return u.CryptoSetupAt != nil
}
