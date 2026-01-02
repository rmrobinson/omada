package omada

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/rmrobinson/omada/api"
)

var (
	ErrObtainTokenFailed  = errors.New("obtaining an API auth token failed")
	ErrRefreshTokenFailed = errors.New("refreshing the API auth token failed")
	ErrTokenInvalid       = errors.New("token invalid")
	ErrTokenExpired       = errors.New("token expired")
)

type Site struct {
	ID   string
	Name string
}

type Client struct {
	*api.ClientWithResponses

	logger     *zap.Logger
	authClient *authClient
	omadaCID   string
}

func NewClient(logger *zap.Logger, controllerURL string, omadaCID string, clientID string, clientSecret string, httpClient *http.Client) *Client {
	authClient := &authClient{
		logger:       logger,
		httpClient:   httpClient,
		url:          controllerURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		omadaCID:     omadaCID,
	}

	apiClient, err := api.NewClientWithResponses(controllerURL, api.WithRequestEditorFn(authClient.addBearerTokenToAPIRequest), api.WithHTTPClient(httpClient))
	if err != nil {
		logger.Fatal("error creating api client", zap.Error(err))
	}

	return &Client{
		ClientWithResponses: apiClient,
		logger:              logger,
		authClient:          authClient,
		omadaCID:            omadaCID,
	}
}

func (c *Client) GetSites(ctx context.Context) ([]Site, error) {
	req := &api.GetSiteListParams{
		Page:     1,
		PageSize: 50,
	}
	resp, err := c.GetSiteListWithResponse(ctx, c.omadaCID, req)
	if err != nil {
		c.logger.Fatal("unable to get site list", zap.Error(err))
	}
	if resp.StatusCode() != http.StatusOK {
		c.logger.Error("http status code not ok", zap.Int("status_code", resp.StatusCode()))
		return nil, ErrTokenInvalid
	} else if resp.JSON200 == nil {
		c.logger.Error("empty json200")
		return nil, ErrTokenInvalid
	} else if *resp.JSON200.ErrorCode == -44112 {
		return nil, ErrTokenExpired
	} else if *resp.JSON200.ErrorCode != 0 {
		c.logger.Error("response error code not ok", zap.Int32("error_code", *resp.JSON200.ErrorCode), zap.String("message", *resp.JSON200.Msg))
		return nil, ErrTokenInvalid
	}

	sites := []Site{}
	for _, site := range *resp.JSON200.Result.Data {
		sites = append(sites, Site{ID: *site.SiteId, Name: *site.Name})
	}

	return sites, nil
}

func IsTokenExpired(errCode int32) bool {
	if errCode == -44112 {
		return true
	}
	return false
}
