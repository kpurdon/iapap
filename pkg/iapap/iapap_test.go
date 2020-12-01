package iapap_test

import (
	"testing"

	"github.com/kpurdon/iapap/pkg/iapap"
	"github.com/stretchr/testify/require"
)

func TestNewVerifier(t *testing.T) {
	v := iapap.NewVerifier("https://test.com")
	require.NotNil(t, v)
}

func TestVerifierGetPublicKey(t *testing.T) {
	t.Skip("TODO")
}

func TestVerifierVerify(t *testing.T) {
	t.Skip("TODO")
}

func TestVerifierApply(t *testing.T) {
	t.Skip("TODO")
}
