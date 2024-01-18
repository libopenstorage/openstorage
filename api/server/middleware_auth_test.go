package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/libopenstorage/openstorage/pkg/auth"
	"github.com/libopenstorage/openstorage/pkg/auth/systemtoken"
)

type mockAuthenticator struct {
	claims *auth.Claims
}

func (a *mockAuthenticator) AuthenticateToken(context context.Context, token string) (*auth.Claims, error) {
	return a.claims, nil
}

func (a *mockAuthenticator) Username(claims *auth.Claims) string {
	return a.claims.Name
}

type noTokenGenerator struct{}

func (n *noTokenGenerator) GetToken(opts *auth.Options) (string, error) {
	return "", nil
}

// Issuer returns the token issuer for this generator necessary
// for registering the authenticator in the SDK.
func (n *noTokenGenerator) Issuer() string {
	return ""
}

// GetAuthenticator returns an authenticator for this issuer used by the SDK
func (n *noTokenGenerator) GetAuthenticator() (auth.Authenticator, error) {
	return nil, nil
}

func TestNewSecurityMiddlewareDecorate(t *testing.T) {
	table := []struct {
		description     string
		isAuthEnabled   bool
		authenticators  map[string]auth.Authenticator
		expectDecorated bool
	}{
		{
			expectDecorated: true,
			description:     "auth enabled authenticator not null",
			isAuthEnabled:   true,
			authenticators: map[string]auth.Authenticator{
				"no-authenticators": &mockAuthenticator{},
			},
		},
		{
			expectDecorated: false,
			description:     "auth enabled authenticator is null",
			isAuthEnabled:   true,
			authenticators:  nil,
		},
		{
			expectDecorated: true,
			description:     "auth not enabled authenticator is not null",
			isAuthEnabled:   false,
			authenticators: map[string]auth.Authenticator{
				"no-authenticators": &mockAuthenticator{},
			},
		},
		{
			expectDecorated: false,
			description:     "auth is not enabled authenticator is null",
			authenticators:  nil,
			isAuthEnabled:   false,
		},
		{
			expectDecorated: false,
			description:     "auth enabled authenticator is empty",
			isAuthEnabled:   true,
			authenticators:  map[string]auth.Authenticator{},
		},
	}

	var handlerFunc http.HandlerFunc

	handlerFunc = func(w http.ResponseWriter, r *http.Request) {}

	for _, testCase := range table {

		var (
			stm auth.TokenGenerator
			err error
		)

		if testCase.isAuthEnabled {
			stm, err = systemtoken.NewManager(&systemtoken.Config{
				ClusterId:    "cluster-id",
				NodeId:       "node-id",
				SharedSecret: "shared-secret",
			})

			if err != nil {
				t.Errorf("Failed to create system token manager: %v\n", err)
				continue
			}
		} else {
			stm = &noTokenGenerator{}
		}

		auth.InitSystemTokenManager(stm)

		decorator := newSecurityMiddleware(testCase.authenticators)

		if testCase.expectDecorated {
			if reflect.ValueOf(decorator(handlerFunc)) == reflect.ValueOf(handlerFunc) {
				t.Log(testCase.description)
				t.Errorf("func must be decorated")
			}
		} else {
			if reflect.ValueOf(decorator(handlerFunc)) != reflect.ValueOf(handlerFunc) {
				t.Log(testCase.description)
				t.Errorf("func must not be decorated")
			}
		}
	}
}

func TestNewSecurityMiddlewareCall(t *testing.T) {
	secret := "secret"
	key := []byte(secret)

	table := []struct {
		claims           *auth.Claims
		description      string
		expectedHttpCode int
	}{
		{
			description:      "token invalid",
			claims:           nil,
			expectedHttpCode: http.StatusUnauthorized,
		},
		{
			description: "wrong role",
			claims: &auth.Claims{
				Issuer:  "issuer",
				Email:   "my@email.com",
				Name:    "myname",
				Subject: "mysub",
				Roles:   []string{"user"},
			},
			expectedHttpCode: http.StatusForbidden,
		},
		{
			description: "auth success",
			claims: &auth.Claims{
				Issuer:  "issuer",
				Email:   "my@email.com",
				Name:    "myname",
				Subject: "mysub",
				Roles:   []string{"system.admin"},
			},
			expectedHttpCode: http.StatusOK,
		},
	}

	stm, err := systemtoken.NewManager(&systemtoken.Config{
		ClusterId:    "cluster-id",
		NodeId:       "node-id",
		SharedSecret: secret,
	})

	if err != nil {
		t.Errorf("error creating token manager")
	}

	auth.InitSystemTokenManager(stm)
	// Clean up token manager
	defer func() {
		auth.InitSystemTokenManager(auth.NoAuth())
	}()

	for _, testCase := range table {
		t.Log(testCase.description)

		var (
			handlerFunc http.HandlerFunc
			err         error
		)

		handlerFunc = func(w http.ResponseWriter, r *http.Request) {}

		sig := auth.Signature{
			Type: jwt.SigningMethodHS256,
			Key:  key,
		}

		opts := auth.Options{
			Expiration: time.Now().Add(time.Minute * 10).Unix(),
		}

		rawtoken := "token"

		// Create
		if testCase.claims != nil {
			rawtoken, err = auth.Token(testCase.claims, &sig, &opts)
		}

		if err != nil {
			t.Errorf("Failed to create system token manager: %v\n", err)
			continue
		}

		authenticators := map[string]auth.Authenticator{
			"issuer": &mockAuthenticator{
				claims: testCase.claims,
			},
		}

		decorator := newSecurityMiddleware(authenticators)
		decoratedHandler := decorator(handlerFunc)

		if reflect.ValueOf(decoratedHandler) == reflect.ValueOf(handlerFunc) {
			t.Errorf("func must be decorated")
		}

		rec := httptest.NewRecorder()
		req, err := http.NewRequest("", http.MethodGet, nil)

		if err != nil {
			t.Errorf("error creating request")
			return
		}

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", rawtoken))
		decoratedHandler.ServeHTTP(rec, req)

		if rec.Code != testCase.expectedHttpCode {
			t.Errorf("Wrong code expected %d actual %d", testCase.expectedHttpCode, rec.Code)
		}
	}
}
