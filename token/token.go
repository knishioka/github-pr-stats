package token

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"sync"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/knishioka/github-pr-stats/conf"
)

// JWTInterface represents an agent to obtain, use, store, and schedule renewal of bearer access tokens
type JWTInterface interface {
	ScheduleRenewal(context.Context)
	Bearer() string
	Renew() error
}

// JWTAgent implements TokenInterface
type JWTAgent struct {
	GithubKey string
	AppID     string
	token     string
	mutex     *sync.RWMutex
}

// NewJWTAgent returns a new TokenAgent with the token and renewal set.
// The passed context is used to cancel the scheduled renewal.
func NewJWTAgent(ctx context.Context) JWTInterface {
	agent := &JWTAgent{
		GithubKey: conf.Configs.GithubKey,
		AppID:     conf.Configs.AppID,
		mutex:     &sync.RWMutex{},
	}
	err := agent.Renew()
	if err != nil {
		log.Fatal(err)
	}

	go agent.ScheduleRenewal(ctx)

	return agent
}

// Bearer returns the string to set in Authorization headers for requests to Github API
func (a *JWTAgent) Bearer() string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.token
}

// Renew sets a new JWT token
func (a *JWTAgent) Renew() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	token, err := a.generateJWT()
	if err != nil {
		return err
	}

	a.token = token

	return nil
}

// ScheduleRenewal calls Renew() after a timeout unless the passed context is canceled
func (a *JWTAgent) ScheduleRenewal(ctx context.Context) {
	select {
	// TODO make renewal time configurable with env variable
	case <-time.After(time.Minute * 9):
		err := a.Renew()
		if err != nil {
			log.Fatal(err)
		}

		a.ScheduleRenewal(ctx)
	case <-ctx.Done():
		fmt.Print(ctx.Err())
		return
	}
}

//generateJWT generate a JWT with 10 minutes expiry to enable authentication as github app
func (a *JWTAgent) generateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(10 * time.Minute).Unix()
	claims["iss"] = a.AppID

	raw, err := ioutil.ReadFile(a.GithubKey)
	if err != nil {
		return "", err
	}

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(raw)
	if err != nil {
		return "", err
	}

	tokenString, err := token.SignedString(signKey)

	if err != nil {
		return "", fmt.Errorf("Unable to generate token: %s", err.Error())
	}

	return tokenString, nil
}
