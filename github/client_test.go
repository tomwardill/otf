package github

import (
	"bytes"
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestGetUser(t *testing.T) {
	ctx := context.Background()
	org := otf.NewTestOrganization(t)
	team := otf.NewTeam("fake-team", org)
	want := otf.NewUser("fake-user", otf.WithOrganizationMemberships(org), otf.WithTeamMemberships(team))
	client := newTestServerClient(t, WithUser(want))

	got, err := client.GetUser(ctx)
	require.NoError(t, err)

	assert.Equal(t, want.Username(), got.Username())
	if assert.Equal(t, 1, len(got.Organizations())) {
		assert.Equal(t, org.Name(), got.Organizations()[0].Name())
	}
	if assert.Equal(t, 1, len(got.Teams())) {
		assert.Equal(t, team.Name(), got.Teams()[0].Name())
	}
}

func TestGetRepoTarball(t *testing.T) {
	ctx := context.Background()
	want, err := os.ReadFile("../testdata/github.tar.gz")
	require.NoError(t, err)
	client := newTestServerClient(t,
		WithRepo(&otf.Repo{Identifier: "acme/terraform", Branch: "master"}),
		WithArchive(want),
	)

	got, err := client.GetRepoTarball(ctx, otf.GetRepoTarballOptions{
		Identifier: "acme/terraform",
		Ref:        "master",
	})
	require.NoError(t, err)

	dst := t.TempDir()
	err = otf.Unpack(bytes.NewReader(got), dst)
	require.NoError(t, err)
}

func TestCreateWebhook(t *testing.T) {
	ctx := context.Background()

	client := newTestServerClient(t,
		WithRepo(&otf.Repo{Identifier: "acme/terraform", Branch: "master"}),
	)

	_, err := client.CreateWebhook(ctx, otf.CreateWebhookOptions{
		Identifier: "acme/terraform",
		Secret:     "me-secret",
	})
	require.NoError(t, err)
}

// newTestServerClient creates a github server for testing purposes and
// returns a client configured to access the server.
func newTestServerClient(t *testing.T, opts ...TestServerOption) *Client {
	srv := NewTestServer(t, opts...)

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client, err := NewClient(context.Background(), otf.CloudClientOptions{
		Hostname:            u.Host,
		SkipTLSVerification: true,
		CloudCredentials: otf.CloudCredentials{
			OAuthToken: &oauth2.Token{AccessToken: "fake-token"},
		},
	})
	require.NoError(t, err)

	return client
}
