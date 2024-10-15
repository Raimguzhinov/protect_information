package common

import "testing"

func TestGCDExtended(t *testing.T) {
	type args struct {
		a int64
		b int64
	}
	tests := []struct {
		name    string
		args    args
		wantGcd int64
		wantX   int64
		wantY   int64
		wantErr bool
	}{
		{
			name:    "28x+19y=gcd(28,19)",
			args:    args{a: 28, b: 19},
			wantGcd: 1,
			wantX:   -2,
			wantY:   3,
			wantErr: false,
		},
		{
			name:    "21x+12y=gcd(21,12)",
			args:    args{a: 21, b: 12},
			wantGcd: 3,
			wantX:   -1,
			wantY:   2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGcd, gotX, gotY, err := GCDExtended(tt.args.a, tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("GCD() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotGcd != tt.wantGcd {
				t.Errorf("GCD() gotGcd = %v, want %v", gotGcd, tt.wantGcd)
			}
			if gotX != tt.wantX {
				t.Errorf("GCD() gotX = %v, want %v", gotX, tt.wantX)
			}
			if gotY != tt.wantY {
				t.Errorf("GCD() gotY = %v, want %v", gotY, tt.wantY)
			}
		})
	}
}
