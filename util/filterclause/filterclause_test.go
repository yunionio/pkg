package filterclause

import (
	"reflect"
	"testing"
)

func TestParseFilterClause(t *testing.T) {
	for _, c := range []string{
		"abc.in(1,2,3)",
		"test.equals(1)",
	} {
		fc := ParseFilterClause(c)
		t.Logf("%s => %s", c, fc.String())
	}
}

func TestParseJointFilterClause(t *testing.T) {
	type args struct {
		jointFilter string
	}
	tests := []struct {
		name string
		args args
		want *SJointFilterClause
	}{
		{
			name: "test parse guestnetworks",
			args: args{
				jointFilter: "guestnetworks(guest_id).ip_addr.equals(10.168.222.232)",
			},
			want: &SJointFilterClause{
				SFilterClause: SFilterClause{
					field:    "ip_addr",
					funcName: "equals",
					params:   []string{"10.168.222.232"},
				},
				JointModel:  "guestnetworks",
				ReleatedKey: "guest_id",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseJointFilterClause(tt.args.jointFilter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseJointFilterClause() = %v, want %v", got, tt.want)
			}
		})
	}
}
