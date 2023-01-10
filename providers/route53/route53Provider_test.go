package route53

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	r53Types "github.com/aws/aws-sdk-go-v2/service/route53/types"
)

func TestUnescape(t *testing.T) {
	var tests = []struct {
		experiment, expected string
	}{
		{"foo", "foo"},
		{"foo.", "foo"},
		{"foo..", "foo."},
		{"foo...", "foo.."},
		{`\052`, "*"},
		{`\052.foo..`, "*.foo."},
		// {`\053.foo`, "+.foo"},  // Not implemented yet.
	}

	for i, test := range tests {
		actual := unescape(&test.experiment)
		if test.expected != actual {
			t.Errorf("%d: Expected %s, got %s", i, test.expected, actual)
		}
	}
}

type batch struct {
	start int
	end   int
}

func (b batch) String() string {
	return fmt.Sprintf("%d:%d", b.start, b.end)
}

func Test_changeBatcher(t *testing.T) {
	genChanges := func(action r53Types.ChangeAction, typ r53Types.RRType, namePattern string, n int, targets ...string) []r53Types.Change {
		changes := make([]r53Types.Change, n)
		for i := 0; i < n; i++ {
			changes[i].Action = action
			changes[i].ResourceRecordSet = &r53Types.ResourceRecordSet{
				Name: aws.String(fmt.Sprintf(namePattern, i)),
				Type: typ,
			}
			for j := 0; j < len(targets); j++ {
				changes[i].ResourceRecordSet.ResourceRecords = append(changes[i].ResourceRecordSet.ResourceRecords, r53Types.ResourceRecord{
					Value: aws.String(targets[j]),
				})
			}
		}
		return changes
	}

	type fields struct {
		changes  []r53Types.Change
		maxSize  int
		maxChars int
	}
	tests := []struct {
		name    string
		fields  fields
		want    []batch
		wantErr bool
	}{
		{
			name: "one_batch",
			fields: fields{
				changes:  genChanges(r53Types.ChangeActionUpsert, r53Types.RRTypeA, "rec%04d", 99, "1.2.3.4"),
				maxSize:  1000,
				maxChars: 32000,
			},
			want: []batch{
				{start: 0, end: 99},
			},
			wantErr: false,
		},
		{
			name: "multi_batch_size",
			fields: fields{
				changes:  genChanges(r53Types.ChangeActionUpsert, r53Types.RRTypeA, "rec%04d", 2000, "1.2.3.4"),
				maxSize:  1000,
				maxChars: 32000,
			},
			want: []batch{
				{start: 0, end: 500},
				{start: 500, end: 1000},
				{start: 1000, end: 1500},
				{start: 1500, end: 2000},
			},
			wantErr: false,
		},
		{
			name: "multi_batch_chars",
			fields: fields{
				changes:  genChanges(r53Types.ChangeActionCreate, r53Types.RRTypeTxt, "rec%04d", 1000, "1.2.3.4", "1.2.3.5", "1.2.3.6", "1.2.3.7", "1.2.3.8", "1.2.3.9"),
				maxSize:  1000,
				maxChars: 32000,
			},
			want: []batch{
				{start: 0, end: 166},
				{start: 166, end: 332},
				{start: 332, end: 498},
				{start: 498, end: 664},
				{start: 664, end: 830},
				{start: 830, end: 996},
				{start: 996, end: 1000},
			},
			wantErr: false,
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &changeBatcher{
				changes:  tt.fields.changes,
				maxSize:  tt.fields.maxSize,
				maxChars: tt.fields.maxChars,
			}
			got := make([]batch, 0)
			for b.Next() {
				start, end := b.Batch()
				got = append(got, batch{
					start: start,
					end:   end,
				})
			}
			err := b.Err()
			if tt.wantErr && err == nil {
				t.Errorf("%d: Expected an error, got nil", i)
			} else if !tt.wantErr && err != nil {
				t.Errorf("%d: Expected no error, got '%s'", i, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%d: Expected %s, got %s", i, tt.want, got)
			}
		})
	}
}
