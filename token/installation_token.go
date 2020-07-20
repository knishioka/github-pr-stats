package token

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// InsTokenInterface represents an agent to obtain, use, storage of installation access tokens
type InsTokenInterface interface {
	GenerateNew() error
	AccountName() string
	Bearer() string
}

//InsTokenAgent handles installation access tokens
type InsTokenAgent struct {
	installationID int64
	accountName    string
	token          string
	c              *http.Client
	ta             JWTInterface
}

//NewInsTokenAgent returns a GitHelper
func NewInsTokenAgent(ctx context.Context, installationID int64, accName string) InsTokenInterface {
	return &InsTokenAgent{
		c: &http.Client{
			Timeout: time.Second * 23,
		},
		ta:             NewJWTAgent(ctx),
		installationID: installationID,
		accountName:    accName,
	}
}

//AccountName returns the installation account name
func (h *InsTokenAgent) AccountName() string {
	return h.accountName
}

//Bearer returns installation access token
func (h *InsTokenAgent) Bearer() string {
	return h.token
}

func (h *InsTokenAgent) getInstallationTokenURL() string {
	return fmt.Sprintf("https://api.github.com/app/installations/%v/access_tokens", h.installationID)
}

// GenerateNew generate new installation token
func (h *InsTokenAgent) GenerateNew() error {
	data := make(map[string]interface{})
	req, err := http.NewRequest("POST", h.getInstallationTokenURL(), &bytes.Buffer{})
	if err != nil {
		return fmt.Errorf("create new HTTP request: %v", err.Error())
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", h.ta.Bearer()))
	req.Header.Add("Accept", "application/vnd.github.machine-man-preview+json")

	resp, err := h.c.Do(req)
	if err != nil {
		return fmt.Errorf("make request error: %v", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode > 204 {
		return fmt.Errorf("make request error: unexpected response status %v", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body error: %v", err.Error())
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if data["token"] == nil {
		return fmt.Errorf("token not available in response: %v", data)
	}

	h.token = fmt.Sprintf("%s", data["token"])

	return nil
}
