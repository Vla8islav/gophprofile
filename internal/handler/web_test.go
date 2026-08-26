package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Vla8islav/gophprofile/internal/config"
	"github.com/Vla8islav/gophprofile/internal/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func newWebTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg, err := config.ReadFlagsServer(nil, zap.NewNop())
	require.NoError(t, err)

	h := NewHandler(mocks.NewMockGophprofileService(ctrl), zap.NewNop())
	server := httptest.NewServer(NewRouter(h, cfg))
	t.Cleanup(server.Close)
	return server
}

func getBody(t *testing.T, url string) (*http.Response, string) {
	t.Helper()
	res, err := http.Get(url)
	require.NoError(t, err)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return res, string(body)
}

func TestWebUploadPage(t *testing.T) {
	t.Parallel()
	server := newWebTestServer(t)

	res, body := getBody(t, server.URL+"/web/upload")
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, res.Header.Get("Content-Type"), "text/html")
	// key wiring the page depends on: auth panel, API field name, poller
	require.Contains(t, body, `id="uploadForm"`)
	require.Contains(t, body, `formData.append('file'`)
	require.Contains(t, body, "/api/user/register")
	require.Contains(t, body, "/api/v1/users/") // the integrated gallery's list endpoint
}

func TestWebGalleryRedirectsToUpload(t *testing.T) {
	t.Parallel()
	server := newWebTestServer(t)

	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	res, err := client.Get(server.URL + "/web/gallery/42")
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusMovedPermanently, res.StatusCode)
	require.Equal(t, "/web/upload", res.Header.Get("Location"))
}

func TestWebRootRedirectsToUpload(t *testing.T) {
	t.Parallel()
	server := newWebTestServer(t)

	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse // don't follow: assert the redirect itself
	}}
	res, err := client.Get(server.URL + "/web")
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusMovedPermanently, res.StatusCode)
	require.Equal(t, "/web/upload", res.Header.Get("Location"))
}

func TestWebAuthJS(t *testing.T) {
	t.Parallel()
	server := newWebTestServer(t)

	res, body := getBody(t, server.URL+"/web/static/auth.js")
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, res.Header.Get("Content-Type"), "javascript")
	require.True(t, strings.Contains(body, "currentUserID"))
}
