package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

type SlackApp struct {
	token  string
	Client http.Client
}

type errorResponse struct {
	Error  string          `json:"error"`
	Errors json.RawMessage `json:"errors"`
}

func New(token string) SlackApp {
	return SlackApp{
		token:  token,
		Client: http.Client{},
	}
}

func (a *SlackApp) Request(ctx context.Context, method string, body []byte, resultJson interface{}) error {

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://slack.com/api/"+method, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-type", "application/json")
	request.Header.Add("Authorization", "Bearer "+a.token)
	result, err := a.Client.Do(request)
	if err != nil {
		return err
	}
	defer result.Body.Close()
	resultBody, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return err
	}
	if result.StatusCode != http.StatusOK {
		return errors.New(string(resultBody))
	}
	var errorJson errorResponse
	err = json.Unmarshal(resultBody, &errorJson)
	if err != nil {
		return err
	}
	if errorJson.Error != "" {
		return errors.New(errorJson.Error + string(errorJson.Errors))
	}
	if resultJson != nil {
		err = json.Unmarshal(resultBody, resultJson)
		if err != nil {
			return err
		}
	}
	return nil
}
