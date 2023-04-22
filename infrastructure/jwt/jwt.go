package jwt

import (
	"NetDesk/models"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
	UserID  string `json:"user_ID"`
	UserPwd string `json:"user_pwd"`
	Email   string `json:"user_email"`
	jwt.StandardClaims
}

// 生成jwt
func Encode(t models.Token, keys []byte, expire int64) (string, error) {
	c := &Claims{}
	// 拼接claims
	if t.Expire == 0 {
		c.ExpiresAt = time.Now().Unix() + expire
	}
	c.UserID = t.ID
	c.Email = t.Email
	c.UserPwd = t.Password

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	token, err := tokenClaims.SignedString(keys)

	return token, err
}

// 解析jwt
func Decode(token string, keys []byte) (*models.Token, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return keys, nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return &models.Token{
				ID:       claims.UserID,
				Password: claims.UserPwd,
				Email:    claims.Email,
				Expire:   claims.ExpiresAt,
			}, nil
		}
	}

	return nil, err
}
