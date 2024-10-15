package common

import (
	"testing"
)

func TestGiantBabyStep(t *testing.T) {
	type args struct {
		a int64
		p int64
		y int64
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name:    "2^x mod 23 = 9",
			args:    args{a: 2, p: 23, y: 9},
			want:    5,
			wantErr: false,
		},
		{
			name:    "5^x mod 7 = 4",
			args:    args{a: 5, p: 7, y: 4},
			want:    20,
			wantErr: false,
			//name: "5^20 mod 7",
			//args: args{a: 5, x: 20, p: 7},
			//want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GiantBabyStep(tt.args.a, tt.args.p, tt.args.y)
			if (err != nil) != tt.wantErr {
				t.Errorf("GiantBabyStep() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GiantBabyStep() got = %v, want %v", got, tt.want)
			}
		})
	}
}
