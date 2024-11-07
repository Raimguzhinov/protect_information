package common

type Signer interface {
	Sign() error
	Verify() (bool, error)
	SignAndVerify() error
}
