package client

import (
	"testing"

	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/keytab"
	"github.com/jcmturner/gokrb5/v8/krberror"
)

func TestAssumePreauthentication(t *testing.T) {
	t.Parallel()

	cl := NewWithKeytab("username", "REALM", &keytab.Keytab{}, &config.Config{}, AssumePreAuthentication(true))
	if !cl.settings.assumePreAuthentication {
		t.Fatal("assumePreAuthentication should be true")
	}
	if !cl.settings.AssumePreAuthentication() {
		t.Fatal("AssumePreAuthentication() should be true")
	}
}

func TestAutoPAFXFASTDefaultsEnabled(t *testing.T) {
	t.Parallel()

	cl := NewWithKeytab("username", "REALM", &keytab.Keytab{}, &config.Config{})
	if !cl.settings.AutoPAFXFAST() {
		t.Fatal("AutoPAFXFAST() should default to true")
	}
}

func TestASExchangeInitialPAFXFASTDecision(t *testing.T) {
	t.Parallel()

	cl := NewWithKeytab(
		"username",
		"REALM",
		&keytab.Keytab{},
		&config.Config{},
		AutoPAFXFAST(false),
		DisablePAFXFAST(true),
	)
	if cl.usePAFXFASTOnInitialASExchangeAttempt() {
		t.Fatal("PA_FX_FAST should not be used when auto mode is disabled and DisablePAFXFAST is true")
	}

	cl = NewWithKeytab(
		"username",
		"REALM",
		&keytab.Keytab{},
		&config.Config{},
		AutoPAFXFAST(true),
		DisablePAFXFAST(true),
	)
	if !cl.usePAFXFASTOnInitialASExchangeAttempt() {
		t.Fatal("PA_FX_FAST should be used on initial attempt when auto mode is enabled")
	}
}

func TestIsPAFXFASTUnsupportedError(t *testing.T) {
	t.Parallel()

	unsupported := krberror.NewErrorf(krberror.PAFXFASTUnsupportedError, "KDC did not respond appropriately to FAST negotiation")
	if !isPAFXFASTUnsupportedError(unsupported) {
		t.Fatal("error should be classified as PA_FX_FAST unsupported")
	}

	nonFAST := krberror.NewErrorf(krberror.KRBMsgError, "some other AS_REP verification failure")
	if isPAFXFASTUnsupportedError(nonFAST) {
		t.Fatal("error should not be classified as PA_FX_FAST unsupported")
	}
}
