package functions

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func init() {
	HttpAuthBearerTokenKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

var HttpAuthBearerTokenKey *ecdsa.PrivateKey

var ValidServiceName = []string{"node"}

func VerifyAuthorization(authorization string) (string, string, string) {
	if authorization == "" {
		return "0", "guest", "_"
	} else {
		token, err := jwt.ParseWithClaims(authorization, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return &HttpAuthBearerTokenKey.PublicKey, nil
		}, jwt.WithIssuedAt(), jwt.WithIssuer("tmpro"), jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{"ES256"})) // TODO, jwt.WithAudience(aud)
		if err != nil || !token.Valid {
			return "0", "guest", "_"
		} else {
			claims := token.Claims.(*jwt.RegisteredClaims)
			now := time.Now()
			//exp & nbf
			if now.After(claims.ExpiresAt.Time) || now.Before(claims.NotBefore.Time) {
				return "0", "guest", "_"
			}
			id, _ := claims.GetSubject()
			_aud, _ := claims.GetAudience()
			if len(_aud) != 1 {
				return "0", "guest", "_"
			}

			aud := _aud[0]

			switch aud {
			case "node":
				// if !slices.Contains(ValidServiceName, aud) {
				// 	return "0", "guest", "_"
				// }

				numID, _ := strconv.ParseInt(id, 10, 64)
				if numID <= 0 {
					return "0", "guest", "_"
				}

				if true {
					return id, aud, strconv.Itoa(1) // <- force UID:1
				} else {
					return "0", "guest", "_"
				}
			default:
				return "0", "guest", "_"
			}
		}
	}
}
