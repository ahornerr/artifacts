package client

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"math"
	"net/http"
)

func New(token string) (*client.ClientWithResponses, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = math.MaxInt32 // Effectively infinite retries
	retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		shouldRetry, checkErr := retryablehttp.DefaultRetryPolicy(ctx, resp, err)
		if shouldRetry || checkErr != nil {
			return shouldRetry, err
		}

		switch resp.StatusCode {
		case 486:
			// character is locked. Action is already in progress
			return true, nil
		case 499:
			// character in cooldown.
			// This shouldn't happen because we should already be checking cooldowns, but is it retryable.
			return true, nil
		}

		return false, nil
	}

	opts := []client.ClientOption{
		client.WithHTTPClient(retryClient.StandardClient()),
	}

	if token != "" {
		opts = append(opts, client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			return nil
		}))
	}

	return client.NewClientWithResponses("https://api.artifactsmmo.com", opts...)
}
