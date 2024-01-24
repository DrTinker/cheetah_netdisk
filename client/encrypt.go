package client

import (
	"NetDisk/infrastructure/jwt"
	"NetDisk/models"
)

type Encryption interface {
	JWTInit(expire int64, key []byte)
	JwtEncode(t models.Token) (string, error)
	JwtDecode(s string) (*models.Token, error)
}

type encryption struct {
	JWTKey    []byte
	JWTExpire int64
}

var EncryptionClient encryption

func (e *encryption) JwtEncode(t models.Token) (string, error) {
	token, err := jwt.Encode(t, e.JWTKey, e.JWTExpire)
	return token, err
}

func (e *encryption) JwtDecode(s string) (*models.Token, error) {
	token, err := jwt.Decode(s, e.JWTKey)
	return token, err
}

func (e *encryption) JWTInit(expire int64, key []byte) {
	EncryptionClient = encryption{}
	e.JWTKey = key
	e.JWTExpire = expire
}
