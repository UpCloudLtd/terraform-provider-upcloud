package config

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Version contains the current software version from git.
var Version = "dev"

type Config struct {
	Username string
	Password string
	Token    string
}

func (c Config) WithAuth() (client.ConfigFn, error) {
	if c.Token != "" {
		return client.WithBearerAuth(c.Token), nil
	}
	if c.Username != "" && c.Password != "" {
		return client.WithBasicAuth(c.Username, c.Password), nil
	}
	return nil, errors.New("Either token or username and password must be configured. Define the credentials either in the provider configuration or as environment variables.") //nolint // This error message is printed to console as is.
}

func (c Config) NewUpCloudServiceConnection(httpClient *http.Client, requestTimeout time.Duration, userAgents ...string) (*service.Service, error) {
	authFn, err := c.WithAuth()
	if err != nil {
		return nil, err
	}
	providerClient := client.New(
		"",
		"",
		client.WithHTTPClient(httpClient),
		client.WithTimeout(requestTimeout),
		client.WithLogger(logDebug),
		authFn,
	)

	if len(userAgents) == 0 {
		userAgents = []string{DefaultUserAgent()}
	}
	providerClient.UserAgent = strings.Join(userAgents, " ")

	svc := service.New(providerClient)
	return svc, checkLogin(svc)
}

func DefaultUserAgent() string {
	return fmt.Sprintf("terraform-provider-upcloud/%s", Version)
}

// logDebug converts slog style key-value varargs to a map compatible with tflog methods and calls tflog.Debug.
func logDebug(ctx context.Context, format string, args ...any) {
	meta := make(map[string]interface{})
	for i := 0; i < len(args); i += 2 {
		key := fmt.Sprintf("%+v", args[i])

		var value interface{}
		if i+1 < len(args) {
			value = args[i+1]
		}

		meta[key] = value
	}
	tflog.Debug(ctx, format, meta)
}

func checkLogin(svc *service.Service) error {
	const numRetries = 10
	var (
		err     error
		problem *upcloud.Problem
	)

	for trys := 0; trys < numRetries; trys++ {
		_, err = svc.GetAccount(context.Background())
		if err == nil {
			break
		}
		if errors.As(err, &problem) && problem.Status == http.StatusUnauthorized {
			return fmt.Errorf("Failed to get account, error was %s: %s", problem.ErrorCode(), problem.Title) //nolint // This error message is printed to console as is.
		}
		time.Sleep(time.Millisecond * 500)
	}

	return err
}
