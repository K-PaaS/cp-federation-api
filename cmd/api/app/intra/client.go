package intra

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/karmada-io/dashboard/cmd/api/app/domain"
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type ClusterResponse struct {
	ResultCode    string           `json:"resultCode"`
	ResultMessage string           `json:"resultMessage"`
	Items         []domain.Cluster `json:"items"`
}

type AdminCheck struct {
	IsSuperAdmin bool `json:"isSuperAdmin"`
}

const (
	CommonApi           = "CommonApi"
	ResultStatusSuccess = "SUCCESS"
	ResultStatusFail    = "FAIL"
	ClaimUserIdKey      = "userAuthId"
)

func ApiCall(target string, method, path string, body interface{}, model interface{}) error {
	var reqBody io.Reader
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	if body != nil {
		sendData, err := json.Marshal(body)
		if err != nil {
			slog.Error("failed to marshal request body", "err", err.Error())
			return err
		}
		reqBody = bytes.NewBuffer(sendData)
	}

	url, auth := setApiUrlAuthorization(target)
	slog.Info("ApiCall SendUrl", "url", url+path)
	request, err := http.NewRequest(method, url+path, reqBody)
	if err != nil {
		slog.Error("failed to new create request", "err", err.Error())
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", "Basic "+auth)
	response, err := client.Do(request)
	if err != nil {
		slog.Error("request failed", "err", err.Error())
		return err
	}
	defer response.Body.Close()
	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		slog.Error("failed to read body", "err", err.Error())
		return err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		slog.Error("unexpected status", response.StatusCode, string(respBody))
		return apperrors.FailedRequest
	}
	if err = json.Unmarshal(respBody, model); err != nil {
		slog.Error("failed to unmarshal data", "err", err.Error())
		return err
	}

	return nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func setApiUrlAuthorization(target string) (url string, auth string) {
	if target == CommonApi {
		return Env.CommonApiUrl, basicAuth(Env.CommonApiUserName, Env.CommonApiUserPassword)
	}
	return "", ""
}

func ApiCallManagedClusters() (ClusterResponse, error) {
	var resp ClusterResponse
	err := ApiCall(CommonApi, "GET", Env.GetFederatedClusterListUrl, nil, &resp)
	if err != nil {
		return ClusterResponse{}, err
	}
	if resp.ResultCode != ResultStatusSuccess {
		return ClusterResponse{}, apperrors.FailedRequest
	}
	return resp, nil
}
