package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/goccy/go-json"
)

const (
	latestVersionURL = "https://api.github.com/repos/cloudflare/cloudflared/releases/latest"
	cacheTTL         = 5 * time.Minute
)

var (
	latestVersionMu       = &sync.RWMutex{}
	latestVersionCachedAt time.Time
	latestVersion         string

	versionsMap = &sync.Map{}
)

func GetLatestDaemonVersion(ctx context.Context) (version string, err error) {
	latestVersionMu.RLock()

	if !latestVersionCachedAt.IsZero() && time.Since(latestVersionCachedAt) <= cacheTTL && latestVersion != "" {
		defer latestVersionMu.RUnlock()
		return latestVersion, nil
	}

	latestVersionMu.RUnlock()
	latestVersionMu.Lock()
	defer latestVersionMu.Unlock()

	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, latestVersionURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	var res *http.Response
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		closeErr := res.Body.Close()
		if closeErr != nil {
			if err != nil {
				err = errors.Join(err, closeErr)
			} else {
				err = closeErr
			}
		}
	}()

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var tmp struct {
		TagName string `json:"tag_name"`
	}
	if err = json.UnmarshalNoEscape(rawBody, &tmp); err != nil {
		return "", err
	}

	latestVersion = tmp.TagName
	latestVersionCachedAt = time.Now() // TODO: use clock
	return latestVersion, nil
}

func VerifyDaemonVersion(ctx context.Context, version string) (bool, error) {
	if value, cached := versionsMap.Load(version); cached {
		return value.(bool), nil
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.github.com/repos/cloudflare/cloudflared/releases/tags/"+version,
		nil,
	)
	if err != nil {
		return false, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	if err = res.Body.Close(); err != nil {
		return false, err
	}

	var isValid bool
	switch res.StatusCode {
	case http.StatusOK:
		isValid = true
	case http.StatusNotFound:
		isValid = false
	default:
		return false, fmt.Errorf("unknown status code: %d", res.StatusCode)
	}
	versionsMap.Store(versionsMap, isValid)
	return isValid, nil
}
