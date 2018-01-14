package session

import (
	"strings"
	"testing"
)

func TestSession_WithDefaultUserAgent(t *testing.T) {
	s := New()
	expected := getDefaultUserAgent()
	actual := s.userAgent
	if actual != expected {
		t.Errorf("UserAgent expected %s, actual %s", expected, actual)
	}
}

func TestAddToUserAgent(t *testing.T) {
	s := New()
	ua := "product/v1 ( myproduct )"
	s.AppendUserAgent(ua)
	if !strings.HasSuffix(s.userAgent, ua) {
		t.Errorf("UserAgent expected to end with %s, but ends with %s", ua, s.userAgent)
	}
	if !strings.HasPrefix(s.userAgent, getDefaultUserAgent()) {
		t.Errorf("UserAgent expected to start with %s, but starts with %s", getDefaultUserAgent(), s.userAgent)
	}
}

func TestResetUserAgent(t *testing.T) {
	s := New()
	ua := "product/v1 ( myproduct )"
	s.AppendUserAgent(ua)
	s.ResetUserAgent()
	if s.userAgent != getDefaultUserAgent() {
		t.Errorf("UserAgent expected to reset to %s, but found to be %s", getDefaultUserAgent(), s.userAgent)
	}
}
