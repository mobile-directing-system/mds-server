package ws

import (
	"context"
	"fmt"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/wsutil"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// TokenResolver is used for resolving public authentication tokens to internal
// ones for WebSocket communication.
type TokenResolver interface {
	// ResolvePublicToken resolves the given public token to the internal one.
	ResolvePublicToken(ctx context.Context, publicToken string) (string, error)
}

// tokenResolver is the implementation of TokenResolver.
type tokenResolver struct {
	httpClient *http.Client
	resolveURL *url.URL
}

func (res tokenResolver) ResolvePublicToken(ctx context.Context, publicToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, res.resolveURL.String(), strings.NewReader(publicToken))
	if err != nil {
		return "", meh.NewInternalErrFromErr(err, "new request", meh.Details{"resolve_url": res.resolveURL})
	}
	response, err := res.httpClient.Do(req)
	if err != nil {
		return "", meh.NewInternalErrFromErr(err, "new request", meh.Details{"req_url": req.URL.String()})
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", meh.NewInternalErrFromErr(err, "read response body", meh.Details{
			"req_url":              req.URL.String(),
			"response_status":      response.Status,
			"response_status_code": response.StatusCode,
		})
	}
	switch response.StatusCode {
	case http.StatusNotFound:
		return "", meh.NewNotFoundErr("response: not found", meh.Details{
			"req_url":              req.URL.String(),
			"response_status":      response.Status,
			"response_status_code": response.StatusCode,
			"response_body":        body,
		})
	case http.StatusOK:
		return string(body), nil
	default:
		return "", meh.NewInternalErr(fmt.Sprintf("unexpected status code: %d", response.StatusCode), meh.Details{
			"req_url":              req.URL.String(),
			"response_status":      response.Status,
			"response_status_code": response.StatusCode,
			"response_body":        body,
		})
	}
}

// readAuth reads the first message from the given wsutil.Client and returns it's
// content. The use case is that we expect the first message to be the value that
// would normally be set in the Authorization-header.
func readAuth(ctx context.Context, clientClient wsutil.Client, resolver TokenResolver) (string, error) {
	var publicToken string
	select {
	case <-ctx.Done():
		return "", meh.NewBadInputErr("timeout while waiting for authentication message", nil)
	case rawMessage := <-clientClient.RawConnection().ReceiveRaw():
		publicToken = string(rawMessage)
	}
	resolvedInternalToken, err := resolver.ResolvePublicToken(ctx, publicToken)
	if err != nil {
		return "", meh.Wrap(err, "resolve public token", nil)
	}
	return resolvedInternalToken, nil
}
