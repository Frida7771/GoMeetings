package helper

import (
	"GoMeetings/internal/define"
	"encoding/base64"
	"encoding/json"
	"errors"

	"crypto/md5"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/satori/go.uuid"
)

type UserClaims struct {
	Id      uint   `json:"id"`
	Name    string `json:"name"`
	IsAdmin int    `json:"is_admin"`
	jwt.RegisteredClaims
}

func GetMd5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
func GenerateUUID() string {
	return uuid.NewV4().String()
}

// GenerateToken creates a signed JWT for the user.
func GenerateToken(id uint, name string) (string, error) {
	userClaim := &UserClaims{
		Id:   id,
		Name: name,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaim)

	tokenString, err := token.SignedString([]byte(define.MyKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// AnalyzeToken parses and validates a JWT string.
func AnalyzeToken(tokenString string) (*UserClaims, error) {
	userClaim := &UserClaims{}
	claims, err := jwt.ParseWithClaims(tokenString, userClaim, func(token *jwt.Token) (interface{}, error) {
		return []byte(define.MyKey), nil
	})
	if err != nil {
		return nil, err
	}
	if !claims.Valid {
		return nil, errors.New("token is invalid")
	}
	return userClaim, nil
}

func Encode(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

func Decode(in string, obj interface{}) {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, obj)
	if err != nil {
		panic(err)
	}
}
