package service

import (
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jcmturner/gokrb5/v8/iana/errorcode"
	"github.com/jcmturner/gokrb5/v8/iana/nametype"
	"github.com/jcmturner/gokrb5/v8/keytab"
	"github.com/jcmturner/gokrb5/v8/messages"
	"github.com/jcmturner/gokrb5/v8/rcache"
	"github.com/jcmturner/gokrb5/v8/test/testdata"
	"github.com/jcmturner/gokrb5/v8/types"
)

func TestVerifyAPREQ_ReplayCacheDuplicateError(t *testing.T) {
	resetReplayCacheForTest(t)
	t.Cleanup(func() {
		resetReplayCacheForTest(t)
	})

	apReq, s := newValidAPReqAndSettings(t)
	rc := &stubReplayCache{addErr: rcache.ErrAlreadyExists}
	if err := SetReplayCache(rc); err != nil {
		t.Fatalf("error setting replay cache: %v", err)
	}

	ok, _, err := VerifyAPREQ(apReq, s)
	if ok || err == nil {
		t.Fatal("expected VerifyAPREQ to fail with replay error")
	}

	krbErr, ok := err.(messages.KRBError)
	if !ok {
		t.Fatalf("expected KRBError, got %T: %v", err, err)
	}
	if krbErr.ErrorCode != errorcode.KRB_AP_ERR_REPEAT {
		t.Fatalf("expected KRB_AP_ERR_REPEAT, got %d", krbErr.ErrorCode)
	}

	if rc.addCalls != 1 {
		t.Fatalf("expected Add to be called once, got %d", rc.addCalls)
	}
	if rc.lastExpiry != s.MaxClockSkew() {
		t.Fatalf("expected Add expiry %v, got %v", s.MaxClockSkew(), rc.lastExpiry)
	}
	if !strings.Contains(rc.lastKey, "/") {
		t.Fatalf("expected replay cache key to include '/' separators, got %q", rc.lastKey)
	}
}

func TestVerifyAPREQ_ReplayCacheOperationalError(t *testing.T) {
	resetReplayCacheForTest(t)
	t.Cleanup(func() {
		resetReplayCacheForTest(t)
	})

	apReq, s := newValidAPReqAndSettings(t)
	backendErr := errors.New("backend unavailable")
	rc := &stubReplayCache{addErr: backendErr}
	if err := SetReplayCache(rc); err != nil {
		t.Fatalf("error setting replay cache: %v", err)
	}

	ok, _, err := VerifyAPREQ(apReq, s)
	if ok || err == nil {
		t.Fatal("expected VerifyAPREQ to fail")
	}
	if !errors.Is(err, backendErr) {
		t.Fatalf("expected wrapped backend error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "replay cache error") {
		t.Fatalf("expected replay cache wrapper error, got: %v", err)
	}
}

func newValidAPReqAndSettings(t *testing.T) (*messages.APReq, *Settings) {
	t.Helper()

	cl := getClient()
	sname := types.PrincipalName{
		NameType:   nametype.KRB_NT_PRINCIPAL,
		NameString: []string{"HTTP", "host.test.gokrb5"},
	}

	b, _ := hex.DecodeString(testdata.HTTP_KEYTAB)
	kt := keytab.New()
	_ = kt.Unmarshal(b)

	st := time.Now().UTC()
	tkt, sessionKey, err := messages.NewTicket(cl.Credentials.CName(), cl.Credentials.Domain(),
		sname, "TEST.GOKRB5",
		types.NewKrbFlags(),
		kt,
		18,
		1,
		st,
		st,
		st.Add(24*time.Hour),
		st.Add(48*time.Hour),
	)
	if err != nil {
		t.Fatalf("error getting test ticket: %v", err)
	}

	apReq, err := messages.NewAPReq(
		tkt,
		sessionKey,
		newTestAuthenticator(*cl.Credentials),
	)
	if err != nil {
		t.Fatalf("error getting test AP_REQ: %v", err)
	}

	h, _ := types.GetHostAddress("127.0.0.1:1234")
	s := NewSettings(kt, ClientAddress(h))

	return &apReq, s
}
