package utils

import "github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/client"

func WithAuth(username, password, token string) client.ConfigFn {
	if token != "" {
		return client.WithBearerAuth(token)
	}
	if username != "" && password != "" {
		return client.WithBasicAuth(username, password)
	}
	return nil
}
