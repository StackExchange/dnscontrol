package cloudflare

import (
	"testing"
)

func Test_normalizeURL(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		want    string
		want1   string
		want2   string
		wantErr bool
	}{
		{
			s:     "foo.com",
			want:  "https://foo.com",
			want1: "foo.com",
			want2: "",
		},
		{
			s:     "http://foo.com",
			want:  "https://foo.com",
			want1: "foo.com",
			want2: "",
		},
		{
			s:     "https://foo.com",
			want:  "https://foo.com",
			want1: "foo.com",
			want2: "",
		},

		{
			s:     "foo.com/bar",
			want:  "https://foo.com/bar",
			want1: "foo.com",
			want2: "/bar",
		},
		{
			s:     "http://foo.com/bar",
			want:  "https://foo.com/bar",
			want1: "foo.com",
			want2: "/bar",
		},
		{
			s:     "https://foo.com/bar",
			want:  "https://foo.com/bar",
			want1: "foo.com",
			want2: "/bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := normalizeURL(tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("normalizeURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("normalizeURL() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("normalizeURL() got1 = %v, want1 %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("normalizeURL() got2 = %v, want2 %v", got2, tt.want2)
			}
		})
	}
}

func Test_makeSingleDirectRule(t *testing.T) {
	tests := []struct {
		name string
		//
		pattern string
		replace string
		//
		wantMatch string
		wantExpr  string
		wantErr   bool
	}{
		{
			name:      "000",
			pattern:   "example.com/",
			replace:   "foo.com",
			wantMatch: `http.host eq "example.com" and http.request.uri.path eq "/"`,
			wantExpr:  `"https://foo.com"`,
			wantErr:   false,
		},

		/*
			All the test-cases I could find in dnsconfig.js

			Generated with this:

			dnscontrol print-ir --pretty |grep '"target' |grep , | sed -e 's@"target":@@g' > /tmp/list
			vim /tmp/list    # removed the obvious duplicates
			awk < /tmp/list -v q='"' -F, '{ print "{" ; print "name: " q NR  q ","  ; print "pattern: " $1 q "," ; print "replace: " q $2 "," ; print "wantMatch: `FIXME`," ; print "wantExpr:  `FIXME`," ; print "wantErr:   false," ; print "},"  }' | pbcopy

		*/

		// {
		// 	name:      "1",
		// 	pattern:   "https://i-dev.sstatic.net/",
		// 	replace:   "https://stackexchange.com/",
		// 	wantMatch: `http.host eq "i-dev.sstatic.net" and http.request.uri.path eq "/"`,
		// 	wantExpr:  `concat("https://i-dev.sstatic.net", http.request.uri.path)`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "2",
		// 	pattern:   "https://i.stack.imgur.com/*",
		// 	replace:   "https://i.sstatic.net/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "3",
		// 	pattern:   "https://img.stack.imgur.com/*",
		// 	replace:   "https://i.sstatic.net/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "4",
		// 	pattern:   "https://insights.stackoverflow.com/",
		// 	replace:   "https://survey.stackoverflow.co",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "5",
		// 	pattern:   "https://insights.stackoverflow.com/trends",
		// 	replace:   "https://trends.stackoverflow.co",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "6",
		// 	pattern:   "https://insights.stackoverflow.com/trends/",
		// 	replace:   "https://trends.stackoverflow.co",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "7",
		// 	pattern:   "https://insights.stackoverflow.com/survey/2021",
		// 	replace:   "https://survey.stackoverflow.co/2021",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "8",
		// 	pattern:   "https://looker.ds.stackexchange.com/*",
		// 	replace:   "https://stackoverflow.cloud.looker.com/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "9",
		// 	pattern:   "https://looker-api.ds.stackexchange.com/*",
		// 	replace:   "https://stackoverflow.cloud.looker.com/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "10",
		// 	pattern:   "https://moderators.meta.stackexchange.com/*",
		// 	replace:   "https://communitybuilding.meta.stackexchange.com/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "11",
		// 	pattern:   "https://meta.moderators.stackexchange.com/*",
		// 	replace:   "https://communitybuilding.meta.stackexchange.com/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "12",
		// 	pattern:   "https://meta.writers.stackexchange.com/*",
		// 	replace:   "https://writing.meta.stackexchange.com/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "13",
		// 	pattern:   "https://fantasy.meta.stackexchange.com/*",
		// 	replace:   "https://scifi.meta.stackexchange.com/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "14",
		// 	pattern:   "https://meta.fantasy.stackexchange.com/*",
		// 	replace:   "https://scifi.meta.stackexchange.com/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "15",
		// 	pattern:   "https://meta.beer.stackexchange.com/*",
		// 	replace:   "https://alcohol.meta.stackexchange.com/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "16",
		// 	pattern:   "https://meta.photography.stackexchange.com/*",
		// 	replace:   "https://photo.meta.stackexchange.com/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "17",
		// 	pattern:   "https://garage.meta.stackexchange.com/*",
		// 	replace:   "https://mechanics.meta.stackexchange.com/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "18",
		// 	pattern:   "https://meta.garage.stackexchange.com/*",
		// 	replace:   "https://mechanics.meta.stackexchange.com/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "19",
		// 	pattern:   "https://meta.health.stackexchange.com/*",
		// 	replace:   "https://medicalsciences.meta.stackexchange.com/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "20",
		// 	pattern:   "https://ui.meta.stackexchange.com/*",
		// 	replace:   "https://ux.meta.stackexchange.com/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "21",
		// 	pattern:   "https://meta.ui.stackexchange.com/*",
		// 	replace:   "https://ux.meta.stackexchange.com/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "22",
		// 	pattern:   "https://meta.mathoverflow.stackexchange.com/*",
		// 	replace:   "https://meta.mathoverflow.net/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "23",
		// 	pattern:   "https://meta.programmers.stackexchange.com/*",
		// 	replace:   "https://softwareengineering.meta.stackexchange.com/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "24",
		// 	pattern:   "*stackpromos.com/*",
		// 	replace:   "https://contests.stackoverflow.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "25",
		// 	pattern:   "*isstackoverflowdownforeveryoneorjustme.com/*",
		// 	replace:   "https://www.stackstatus.net/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "26",
		// 	pattern:   "https://stackoverflow.help/*",
		// 	replace:   "https://stackoverflowteams.help/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "27",
		// 	pattern:   "*www.stackoverflow.help/*",
		// 	replace:   "https://stackoverflow.help/$1",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "28",
		// 	pattern:   "*stackoverflow.help/support/solutions/articles/36000241656-write-an-article",
		// 	replace:   "https://stackoverflow.help/en/articles/4397209-write-an-article",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "29",
		// 	pattern:   "*stackoverflow.careers/*",
		// 	replace:   "https://careers.stackoverflow.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "30",
		// 	pattern:   "*stackoverflowenterprise.com/*",
		// 	replace:   "https://www.stackoverflowbusiness.com/enterprise/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "31",
		// 	pattern:   "stackenterprise.com/*",
		// 	replace:   "https://stackoverflow.co/teams/",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "32",
		// 	pattern:   "www.stackenterprise.com/*",
		// 	replace:   "https://stackoverflow.co/teams/",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "33",
		// 	pattern:   "meta.*yodeya.com/*",
		// 	replace:   "https://judaism.meta.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "34",
		// 	pattern:   "chat.*yodeya.com/*",
		// 	replace:   "https://chat.stackexchange.com/?tab=site\u0026host=judaism.stackexchange.com",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "35",
		// 	pattern:   "*yodeya.com/*",
		// 	replace:   "https://judaism.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "36",
		// 	pattern:   "meta.*seasonedadvice.com/*",
		// 	replace:   "https://cooking.meta.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "37",
		// 	pattern:   "chat.*seasonedadvice.com/*",
		// 	replace:   "https://chat.stackexchange.com/?tab=site\u0026host=cooking.stackexchange.com",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "38",
		// 	pattern:   "*seasonedadvice.com/*",
		// 	replace:   "https://cooking.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "39",
		// 	pattern:   "meta.*askpatents.com/*",
		// 	replace:   "https://patents.meta.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "40",
		// 	pattern:   "chat.*askpatents.com/*",
		// 	replace:   "https://chat.stackexchange.com/?tab=site\u0026host=patents.stackexchange.com",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "41",
		// 	pattern:   "*askpatents.com/*",
		// 	replace:   "https://patents.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "42",
		// 	pattern:   "meta.*arqade.com/*",
		// 	replace:   "https://gaming.meta.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "43",
		// 	pattern:   "chat.*arqade.com/*",
		// 	replace:   "https://chat.stackexchange.com/?tab=site\u0026host=gaming.stackexchange.com",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "44",
		// 	pattern:   "*arqade.com/*",
		// 	replace:   "https://gaming.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "45",
		// 	pattern:   "meta.*askdifferent.com/*",
		// 	replace:   "https://apple.meta.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "46",
		// 	pattern:   "chat.*askdifferent.com/*",
		// 	replace:   "https://chat.stackexchange.com/?tab=site\u0026host=apple.stackexchange.com",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "47",
		// 	pattern:   "*askdifferent.com/*",
		// 	replace:   "https://apple.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "48",
		// 	pattern:   "meta.*basicallymoney.com/*",
		// 	replace:   "https://money.meta.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "49",
		// 	pattern:   "chat.*basicallymoney.com/*",
		// 	replace:   "https://chat.stackexchange.com/?tab=site\u0026host=money.stackexchange.com",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "50",
		// 	pattern:   "*basicallymoney.com/*",
		// 	replace:   "https://money.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "51",
		// 	pattern:   "meta.*chiphacker.com/*",
		// 	replace:   "https://electronics.meta.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "52",
		// 	pattern:   "chat.*chiphacker.com/*",
		// 	replace:   "https://chat.stackexchange.com/?tab=site\u0026host=electronics.stackexchange.com",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "53",
		// 	pattern:   "*chiphacker.com/*",
		// 	replace:   "https://electronics.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "54",
		// 	pattern:   "meta.*crossvalidated.com/*",
		// 	replace:   "https://stats.meta.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "55",
		// 	pattern:   "chat.*crossvalidated.com/*",
		// 	replace:   "https://chat.stackexchange.com/?tab=site\u0026host=stats.stackexchange.com",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "56",
		// 	pattern:   "*crossvalidated.com/*",
		// 	replace:   "https://stats.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "57",
		// 	pattern:   "meta.*miyodeya.com/*",
		// 	replace:   "https://judaism.meta.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "58",
		// 	pattern:   "chat.*miyodeya.com/*",
		// 	replace:   "https://chat.stackexchange.com/?tab=site\u0026host=judaism.stackexchange.com",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "59",
		// 	pattern:   "*miyodeya.com/*",
		// 	replace:   "https://judaism.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "60",
		// 	pattern:   "meta.*skepticexchange.com/*",
		// 	replace:   "https://skeptics.meta.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "61",
		// 	pattern:   "chat.*skepticexchange.com/*",
		// 	replace:   "https://chat.stackexchange.com/?tab=site\u0026host=skeptics.stackexchange.com",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "62",
		// 	pattern:   "*skepticexchange.com/*",
		// 	replace:   "https://skeptics.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "63",
		// 	pattern:   "meta.*thearqade.com/*",
		// 	replace:   "https://gaming.meta.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "64",
		// 	pattern:   "chat.*thearqade.com/*",
		// 	replace:   "https://chat.stackexchange.com/?tab=site\u0026host=gaming.stackexchange.com",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "65",
		// 	pattern:   "*thearqade.com/*",
		// 	replace:   "https://gaming.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "66",
		// 	pattern:   "meta.*nothingtoinstall.com/*",
		// 	replace:   "https://webapps.meta.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "67",
		// 	pattern:   "chat.*nothingtoinstall.com/*",
		// 	replace:   "https://chat.stackexchange.com/?tab=site\u0026host=webapps.stackexchange.com",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "68",
		// 	pattern:   "*nothingtoinstall.com/*",
		// 	replace:   "https://webapps.stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "69",
		// 	pattern:   "*slackoverflow.com/*",
		// 	replace:   "https://stackoverflow.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "70",
		// 	pattern:   "collectivesonstackoverflow.co/*",
		// 	replace:   "https://stackoverflow.com/collectives-on-stack-overflow",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "71",
		// 	pattern:   "*collectivesonstackoverflow.co/*",
		// 	replace:   "https://stackoverflow.com/collectives-on-stack-overflow",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "72",
		// 	pattern:   "collectivesonstackoverflow.io/*",
		// 	replace:   "https://stackoverflow.com/collectives-on-stack-overflow",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "73",
		// 	pattern:   "*collectivesonstackoverflow.io/*",
		// 	replace:   "https://stackoverflow.com/collectives-on-stack-overflow",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "74",
		// 	pattern:   "collectivesonstackoverflow.com/*",
		// 	replace:   "https://stackoverflow.com/collectives-on-stack-overflow",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "75",
		// 	pattern:   "*collectivesonstackoverflow.com/*",
		// 	replace:   "https://stackoverflow.com/collectives-on-stack-overflow",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "76",
		// 	pattern:   "*stackexchange.ca/*",
		// 	replace:   "https://stackexchange.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "77",
		// 	pattern:   "*stackoverflow.ca/*",
		// 	replace:   "https://stackoverflow.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
		// {
		// 	name:      "78",
		// 	pattern:   "*stackoverflow.tv/*",
		// 	replace:   "https://stackoverflow.com/$2",
		// 	wantMatch: `FIXME`,
		// 	wantExpr:  `FIXME`,
		// 	wantErr:   false,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatch, gotExpr, err := makeRuleFromPattern(tt.pattern, tt.replace, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeSingleDirectRule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotMatch != tt.wantMatch {
				t.Errorf("makeSingleDirectRule() MATCH = %v, want %v", gotMatch, tt.wantMatch)
			}
			if gotExpr != tt.wantExpr {
				t.Errorf("makeSingleDirectRule()  EXPR = %v, want %v", gotExpr, tt.wantExpr)
			}
		})
	}
}
