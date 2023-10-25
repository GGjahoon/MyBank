package token

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

const minSecretKeySize = 32

type JWTMaker struct {
	SecretKey string
}

// NewJWTMaker creates a new JWTMaker
func NewJWTMaker(secretKey string) (Maker, error) {
	// length of the secret key should not too short
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size : must be at least %d characters", minSecretKeySize)
	}
	return &JWTMaker{SecretKey: secretKey}, nil
}
func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
	//create a new payload
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", payload, err
	}
	//the input : claims(payload) should have valid method
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	accessToken, err := jwtToken.SignedString([]byte(maker.SecretKey))
	return accessToken, payload, err
}

// VerifyToken checks if the token is valid or not
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		//get its signing algorithm
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		//means the algorithm of the token is not match with our signing algorithm
		if !ok {
			return nil, ErrInvalidToken
		}
		//the algorithm matches,return the secretKey that using to sign the token
		return []byte(maker.SecretKey), nil
	}

	//parse the token
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	//此处判断err类型，在jwt.ParseWithClaims定义中可以看到其调用了token.Claims.Valid()。即此前定义的payload中的valid方法
	//返回错误时以wt.ValidationError结构体对象返回，隐去了实际上的错误信息
	//将返回的错误断言为jwt.ValidationError，从结构体中取出inner以而判断令牌过期/无效
	if err != nil {
		//two different scenarios : the token is valid or expired
		vErr, ok := err.(*jwt.ValidationError)
		//if assert success and token is expired
		if ok && errors.Is(vErr.Inner, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	//the token is successfully parsed and verified,get the payload of token
	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}
	return payload, nil

}
