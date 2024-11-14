package common

import (
	"fmt"
	"math/big"
	"testing"
)

func TestModularExponentiation(t *testing.T) {
	type args struct {
		a int64
		x int64
		p int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "5^20 mod 7",
			args: args{a: 5, x: 20, p: 7},
			want: 4,
		},
		{
			name: "2^10 mod 5",
			args: args{a: 2, x: 10, p: 5},
			want: 4,
		},
		{
			name: "3^100 mod 7",
			args: args{a: 3, x: 100, p: 7},
			want: 4,
		},
		{
			name: "3^21 mod 11",
			args: args{a: 3, x: 21, p: 11},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ModularExponentiation(tt.args.a, tt.args.x, tt.args.p); got != tt.want {
				t.Errorf("ModularExponentiation() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Тест функции ModularExponentiationBig
func TestModularExponentiationBig(t *testing.T) {
	base := big.NewInt(2)
	exp := big.NewInt(5)
	mod := big.NewInt(13)
	result := ModularExponentiationBig(base, exp, mod)

	// Ожидается, что (2^5) % 13 == 6
	fmt.Printf("Result of 2^5 mod 13: %s (ожидается 6)\n", result.String())
}
