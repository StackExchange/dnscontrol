package gandi

import (
	"testing"

	"github.com/StackExchange/dnscontrol/v2/models"
	"github.com/prasmussen/gandi-api/live_dns/record"
	"github.com/stretchr/testify/assert"
)

func makeRC(label, domain, target string, rc models.RecordConfig) *models.RecordConfig {
	rc.SetLabel(label, domain)
	rc.SetTarget(target)
	return &rc
}
func TestRecordConfigFromInfo(t *testing.T) {

	for _, data := range []struct {
		info   *record.Info
		config []*models.RecordConfig
	}{
		{
			&record.Info{
				Name:   "www",
				Type:   "A",
				TTL:    500,
				Values: []string{"127.0.0.1", "127.1.0.1"},
			},
			[]*models.RecordConfig{
				makeRC("www", "example.com", "127.0.0.1", models.RecordConfig{
					Type: "A",
					TTL:  500,
				}),
				makeRC("www", "example.com", "127.1.0.1", models.RecordConfig{
					Type: "A",
					TTL:  500,
				}),
			},
		},
		{
			&record.Info{
				Name:   "www",
				Type:   "TXT",
				TTL:    500,
				Values: []string{"\"test 2\"", "\"test message test message test message\""},
			},
			[]*models.RecordConfig{
				makeRC("www", "example.com", "test 2", models.RecordConfig{
					Type:       "TXT",
					TxtStrings: []string{"test 2", "test message test message test message"},
					TTL:        500,
				}),
			},
		},
		{
			&record.Info{
				Name: "www",
				Type: "CAA",
				TTL:  500,
				// examples from https://sslmate.com/caa/
				Values: []string{"0 issue \"www.certinomis.com\"", "0 issuewild \"buypass.com\""},
			},
			[]*models.RecordConfig{
				makeRC("www", "example.com", "www.certinomis.com", models.RecordConfig{
					Type:    "CAA",
					CaaFlag: 0,
					CaaTag:  "issue",
					TTL:     500,
				}),
				makeRC("www", "example.com", "buypass.com", models.RecordConfig{
					Type:    "CAA",
					CaaFlag: 0,
					CaaTag:  "issuewild",
					TTL:     500,
				}),
			},
		},
		{
			&record.Info{
				Name:   "test",
				Type:   "SRV",
				TTL:    500,
				Values: []string{"20 0 5060 backupbox.example.com."},
			},
			[]*models.RecordConfig{
				makeRC("test", "example.com", "backupbox.example.com.", models.RecordConfig{
					Type:        "SRV",
					SrvPriority: 20,
					SrvWeight:   0,
					SrvPort:     5060,
					TTL:         500,
				}),
			},
		},
		{
			&record.Info{
				Name:   "mail",
				Type:   "MX",
				TTL:    500,
				Values: []string{"50 fb.mail.gandi.net.", "10 spool.mail.gandi.net."},
			},
			[]*models.RecordConfig{
				makeRC("mail", "example.com", "fb.mail.gandi.net.", models.RecordConfig{
					Type:         "MX",
					MxPreference: 50,
					TTL:          500,
				}),
				makeRC("mail", "example.com", "spool.mail.gandi.net.", models.RecordConfig{
					Type:         "MX",
					MxPreference: 10,
					TTL:          500,
				}),
			},
		},
	} {
		t.Run("with record type "+data.info.Type, func(t *testing.T) {
			c := liveClient{}
			for _, c := range data.config {
				c.Original = data.info
			}
			t.Run("Convert gandi info to record config", func(t *testing.T) {
				recordConfig := c.recordConfigFromInfo([]*record.Info{data.info}, "example.com")
				assert.Equal(t, data.config, recordConfig)
			})
			t.Run("Convert record config to gandi info", func(t *testing.T) {
				_, recordInfos, err := c.recordsToInfo(data.config)
				assert.NoError(t, err)
				assert.Equal(t, []*record.Info{data.info}, recordInfos)
			})
		})
	}
}
