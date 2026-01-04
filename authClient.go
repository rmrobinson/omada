package omada

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

type authToken struct {
	AccessToken   string `json:"accessToken"`
	TokenType     string `json:"tokenType"`
	ExpiresInSecs int    `json:"expiresIn"`
	RefreshToken  string `json:"refreshToken"`
}

type obtainTokenRequest struct {
	OmadaCID     string `json:"omadacId"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}
type authorizeTokenResponse struct {
	ErrorCode int       `json:"errorCode"`
	Message   string    `json:"msg"`
	Result    authToken `json:"result"`
}

type authClient struct {
	logger       *zap.Logger
	httpClient   *http.Client
	url          string
	clientID     string
	clientSecret string
	omadaCID     string

	tokenLock             sync.Mutex // ensure that the updates to token and the expiration fields are done together
	token                 *authToken
	accessTokenExpiresAt  time.Time
	refreshTokenExpiresAt time.Time
}

func (c *authClient) addBearerTokenToAPIRequest(ctx context.Context, req *http.Request) error {
	if c.token == nil {
		if err := c.obtainToken(ctx); err != nil {
			c.logger.Error("unable to obtain API token ahead of adding to request", zap.Error(err), zap.String("request_uri", req.RequestURI))
			return ErrObtainTokenFailed
		}
	} else if time.Now().After(c.accessTokenExpiresAt) {
		if err := c.refreshToken(ctx); err != nil {
			c.logger.Error("unable to refresh API token ahead of adding to request", zap.Error(err), zap.String("request_uri", req.RequestURI))
			return ErrRefreshTokenFailed
		}
	}

	if c.token != nil {
		req.Header.Add("Authorization", fmt.Sprintf("AccessToken=%s", c.token.AccessToken))
	}
	return nil
}

func (c *authClient) obtainToken(ctx context.Context) error {
	tokenReq := &obtainTokenRequest{
		OmadaCID:     c.omadaCID,
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
	}
	jsonTokenReq, err := json.Marshal(tokenReq)
	if err != nil {
		c.logger.Error("unable to serialize token request", zap.Error(err))
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/openapi/authorize/token?grant_type=client_credentials", c.url), bytes.NewBuffer(jsonTokenReq))
	if err != nil {
		c.logger.Error("uanble to create http post request to obtain token", zap.Error(err))
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("unable to do http post request to obtain token", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	tokenResp := &authorizeTokenResponse{}
	if err := json.NewDecoder(resp.Body).Decode(tokenResp); err != nil {
		c.logger.Error("unable to decode json post body to obtain token", zap.Error(err))
		return err
	}

	if tokenResp.ErrorCode != 0 {
		c.logger.Error("unable to obtain token due to omada api error code", zap.Int("error_code", tokenResp.ErrorCode), zap.String("message", tokenResp.Message))
		return err
	}

	c.tokenLock.Lock()
	defer c.tokenLock.Unlock()

	c.token = &tokenResp.Result
	c.refreshTokenExpiresAt = time.Now().Add(time.Hour * 24 * 14)                               // refresh tokens are valid for 14 days
	c.accessTokenExpiresAt = time.Now().Add(time.Second * time.Duration(c.token.ExpiresInSecs)) // access token validity returned

	c.logger.Debug("token retrieved", zap.String("access_token_expires_at", c.accessTokenExpiresAt.String()), zap.String("refresh_token_expires_at", c.refreshTokenExpiresAt.String()))
	return nil
}

func (c *authClient) refreshToken(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/openapi/authorize/token?client_id=%s&client_secret=%s&refresh_token=%s&grant_type=refresh_token", c.clientID, c.clientSecret, c.token.RefreshToken, c.url), nil)
	if err != nil {
		c.logger.Error("uanble to create http post request to refresh token", zap.Error(err))
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("unable to do http post request to refresh token", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	tokenResp := &authorizeTokenResponse{}
	if err := json.NewDecoder(resp.Body).Decode(tokenResp); err != nil {
		c.logger.Error("unable to decode json post body to refresh token", zap.Error(err))
		return err
	}

	// Refresh token is expired; obtain a new one (implicitly gets an access token as well)
	if tokenResp.ErrorCode == -44114 {
		return c.obtainToken(ctx)
	}

	if tokenResp.ErrorCode != 0 {
		c.logger.Error("unable to refresh token due to omada api error code", zap.Int("error_code", tokenResp.ErrorCode), zap.String("message", tokenResp.Message))
		return err
	}

	c.tokenLock.Lock()
	defer c.tokenLock.Unlock()

	c.token = &tokenResp.Result
	c.accessTokenExpiresAt = time.Now().Add(time.Second * time.Duration(c.token.ExpiresInSecs)) // access token validity returned

	c.logger.Debug("token refreshed", zap.String("access_token_expires_at", c.accessTokenExpiresAt.String()), zap.String("refresh_token_expires_at", c.refreshTokenExpiresAt.String()))
	return nil
}
