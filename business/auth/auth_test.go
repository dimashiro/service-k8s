package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/dimashiro/service/business/auth"
	"github.com/golang-jwt/jwt/v4"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

func TestAuth(t *testing.T) {

	t.Logf("\tTest:\tWhen handling a single user.")
	{
		const keyID = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatalf("\t%s\tTest:\tShould be able to create a private key: %v", failed, err)
		}
		t.Logf("\t%s\tTest:\tShould be able to create a private key.", success)

		a, err := auth.New(keyID, &keyStore{pk: privateKey})
		if err != nil {
			t.Fatalf("\t%s\tTest:\tShould be able to create an authenticator: %v", failed, err)
		}
		t.Logf("\t%s\tTest:\tShould be able to create an authenticator.", success)

		claims := auth.Claims{
			StandardClaims: jwt.StandardClaims{
				Issuer:    "service",
				Subject:   "5cf37266-3473-4006-984f-9325122678b7",
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
				IssuedAt:  time.Now().UTC().Unix(),
			},
			Roles: []string{auth.RoleAdmin},
		}

		token, err := a.GenerateToken(claims)
		if err != nil {
			t.Fatalf("\t%s\tTest:\tShould be able to generate a JWT: %v", failed, err)
		}
		t.Logf("\t%s\tTest:\tShould be able to generate a JWT.", success)

		parsedClaims, err := a.ValidateToken(token)
		if err != nil {
			t.Fatalf("\t%s\tTest:\tShould be able to parse the claims: %v", failed, err)
		}
		t.Logf("\t%s\tTest:\tShould be able to parse the claims.", success)

		if exp, got := len(claims.Roles), len(parsedClaims.Roles); exp != got {
			t.Logf("\t\tTest:\texp: %d", exp)
			t.Logf("\t\tTest:\tgot: %d", got)
			t.Fatalf("\t%s\tTest:\tShould have the expected number of roles: %v", failed, err)
		}
		t.Logf("\t%s\tTest:\tShould have the expected number of roles.", success)

		if exp, got := claims.Roles[0], parsedClaims.Roles[0]; exp != got {
			t.Logf("\t\tTest:\texp: %v", exp)
			t.Logf("\t\tTest:\tgot: %v", got)
			t.Fatalf("\t%s\tTest:\tShould have the expected roles: %v", failed, err)
		}
		t.Logf("\t%s\tTest:\tShould have the expected roles.", success)
	}
}

type keyStore struct {
	pk *rsa.PrivateKey
}

func (ks *keyStore) PrivateKey(kid string) (*rsa.PrivateKey, error) {
	return ks.pk, nil
}

func (ks *keyStore) PublicKey(kid string) (*rsa.PublicKey, error) {
	return &ks.pk.PublicKey, nil
}
