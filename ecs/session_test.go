package ecs

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const ssoConfig = `
[profile sso-dev]
sso_session = corp
sso_account_id = 111111111111
sso_role_name = Developer
region = us-east-1

[sso-session corp]
sso_start_url = https://example.awsapps.com/start
sso_region = us-east-1
sso_registration_scopes = sso:account:access
`

// isolateAWSEnv points every AWS file lookup at a temp dir so tests never
// read the real ~/.aws, and resets the cached session.
func isolateAWSEnv(t *testing.T, config string) {
	t.Helper()
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config")
	if err := ioutil.WriteFile(configFile, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", dir)
	t.Setenv("AWS_CONFIG_FILE", configFile)
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", filepath.Join(dir, "credentials"))
	for _, name := range []string{
		"AWS_PROFILE", "AWS_DEFAULT_PROFILE", "AWS_SDK_LOAD_CONFIG",
		"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN",
		"AWS_ACCESS_KEY", "AWS_SECRET_KEY", "AWS_WEB_IDENTITY_TOKEN_FILE", "AWS_ROLE_ARN",
	} {
		t.Setenv(name, "")
	}
	_session = nil
	_sessionOnce = sync.Once{}
	t.Cleanup(func() {
		_session = nil
		_sessionOnce = sync.Once{}
	})
}

func setEnvCredentials(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "AKIDENVTEST")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "envsecret")
}

func assertEnvCredentials(t *testing.T) {
	t.Helper()
	creds, err := getSession().Config.Credentials.Get()
	if err != nil {
		t.Fatalf("expected env credentials, got error: %v", err)
	}
	if creds.AccessKeyID != "AKIDENVTEST" {
		t.Fatalf("expected env credentials, got provider %q key %q", creds.ProviderName, creds.AccessKeyID)
	}
}

func TestSessionUsesEnvCredentials(t *testing.T) {
	isolateAWSEnv(t, "")
	setEnvCredentials(t)
	assertEnvCredentials(t)
}

func TestSessionEnvCredentialsBeatSSOProfile(t *testing.T) {
	isolateAWSEnv(t, ssoConfig)
	setEnvCredentials(t)
	t.Setenv("AWS_PROFILE", "sso-dev")
	assertEnvCredentials(t)
}

func TestSessionEnvCredentialsSurviveBrokenConfig(t *testing.T) {
	isolateAWSEnv(t, "[profile broken\nnot ini at all")
	setEnvCredentials(t)
	assertEnvCredentials(t)
}

func TestSessionKeepsRegionBehavior(t *testing.T) {
	isolateAWSEnv(t, ssoConfig)
	t.Setenv("AWS_PROFILE", "sso-dev")
	t.Setenv("AWS_REGION", "")
	if region := *getSession().Config.Region; region != "us-west-2" {
		t.Fatalf("expected default region us-west-2, got %q", region)
	}
}

func TestSessionResolvesSSOProfile(t *testing.T) {
	isolateAWSEnv(t, ssoConfig)
	t.Setenv("AWS_PROFILE", "sso-dev")
	// No cached token exists in the temp HOME, so resolving must fail inside
	// the SSO provider (rather than falling through to "no credentials").
	_, err := getSession().Config.Credentials.Get()
	if err == nil {
		t.Fatal("expected an SSO token error without a cached token")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "sso") {
		t.Fatalf("expected SSO provider error, got: %v", err)
	}
}
