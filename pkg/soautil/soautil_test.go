package soautil

import (
	"reflect"
	"testing"
)

func Test_RFC5322MailToBind(t *testing.T) {
	tests := []struct {
		name        string
		rfc5322Mail string
		bindMail    string
	}{
		{"0", "hostmaster@example.com", "hostmaster.example.com"},
		{"1", "admin.dns@example.com", "admin\\.dns.example.com"},
		{"2", "hostmaster@sub.example.com", "hostmaster.sub.example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RFC5322MailToBind(tt.rfc5322Mail); !reflect.DeepEqual(got, tt.bindMail) {
				t.Errorf("RFC5322MailToBind(%v) = %v, want %v", tt.rfc5322Mail, got, tt.bindMail)
			}
		})
	}
}
