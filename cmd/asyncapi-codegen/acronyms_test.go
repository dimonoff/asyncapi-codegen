package main

import (
	"testing"

	tpl "github.com/dimonoff/asyncapi-codegen/pkg/utils/template"
	"github.com/stretchr/testify/require"
)

func TestDefaultAcronymRegistryConfigure(t *testing.T) {
	tpl.SetKnownAcronyms(nil)
	require.NoError(t, tpl.SetConvertKeyFn("camel"))
	t.Cleanup(func() {
		tpl.SetKnownAcronyms(nil)
		require.NoError(t, tpl.SetConvertKeyFn("none"))
	})

	DefaultAcronymRegistry().Configure()

	require.Equal(t, "UserID", tpl.ConvertKey("user_id"))
	require.Equal(t, "APIURL", tpl.ConvertKey("api_url"))
	require.Equal(t, "OIDCConfigURL", tpl.ConvertKey("oidc_config_url"))
	require.Equal(t, "OAuthToken", tpl.ConvertKey("oauth_token"))
}

