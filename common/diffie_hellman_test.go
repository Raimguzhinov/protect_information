package common

import (
	"testing"
)

func TestIsPrime(t *testing.T) {
	type args struct {
		p int64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "5 is prime",
			args: args{p: 5},
			want: true,
		},
		{
			name: "7 is prime",
			args: args{p: 7},
			want: true,
		},
		{
			name: "11 is prime",
			args: args{p: 11},
			want: true,
		},
		{
			name: "13 is prime",
			args: args{p: 13},
			want: true,
		},
		{
			name: "17 is prime",
			args: args{p: 17},
			want: true,
		},
		{
			name: "293 is prime",
			args: args{p: 293},
			want: true,
		},
		{
			name: "283 is prime",
			args: args{p: 283},
			want: true,
		},
		{
			name: "10 is not prime",
			args: args{p: 10},
			want: false,
		},
		{
			name: "9 is not prime",
			args: args{p: 9},
			want: false,
		},
		{
			name: "15 is not prime",
			args: args{p: 15},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPrime(tt.args.p); got != tt.want {
				t.Errorf("IsPrime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiffieHellman(t *testing.T) {
	t.Run("общий ключ", func(t *testing.T) {
		sharedKey, err := DiffieHellman()
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("sharedKey = %d", sharedKey)
	})
}
