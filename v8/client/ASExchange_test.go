package client

import (
	"testing"

	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/iana/patype"
	"github.com/jcmturner/gokrb5/v8/keytab"
	"github.com/jcmturner/gokrb5/v8/krberror"
	"github.com/jcmturner/gokrb5/v8/messages"
	"github.com/jcmturner/gokrb5/v8/types"
)

func TestSetPADataPAFXFASTToggle(t *testing.T) {
	t.Parallel()

	cl := NewWithKeytab("username", "REALM", &keytab.Keytab{}, &config.Config{})
	asReq := messages.ASReq{
		KDCReqFields: messages.KDCReqFields{
			PAData: types.PADataSequence{{PADataType: patype.PA_REQ_ENC_PA_REP}},
		},
	}

	err := setPAData(cl, nil, &asReq, true)
	if err != nil {
		t.Fatalf("setPAData returned error: %v", err)
	}
	if got := countPADataType(asReq.PAData, patype.PA_REQ_ENC_PA_REP); got != 1 {
		t.Fatalf("expected exactly one PA_REQ_ENC_PA_REP entry, got %d", got)
	}

	err = setPAData(cl, nil, &asReq, true)
	if err != nil {
		t.Fatalf("setPAData returned error: %v", err)
	}
	if got := countPADataType(asReq.PAData, patype.PA_REQ_ENC_PA_REP); got != 1 {
		t.Fatalf("expected exactly one PA_REQ_ENC_PA_REP entry after repeat call, got %d", got)
	}

	err = setPAData(cl, nil, &asReq, false)
	if err != nil {
		t.Fatalf("setPAData returned error: %v", err)
	}
	if got := countPADataType(asReq.PAData, patype.PA_REQ_ENC_PA_REP); got != 0 {
		t.Fatalf("expected no PA_REQ_ENC_PA_REP entries when disabled, got %d", got)
	}
}

func countPADataType(pas types.PADataSequence, paType int32) int {
	count := 0
	for _, pa := range pas {
		if pa.PADataType == paType {
			count++
		}
	}
	return count
}

func TestShouldRetryASExchangeWithoutPAFXFAST(t *testing.T) {
	t.Parallel()

	unsupported := krberror.NewErrorf(krberror.PAFXFASTUnsupportedError, "KDC did not respond appropriately to FAST negotiation")
	nonUnsupported := krberror.NewErrorf(krberror.KRBMsgError, "some other AS_REP verification failure")

	tests := []struct {
		name         string
		autoPAFXFAST bool
		usePAFXFAST  bool
		err          error
		expect       bool
	}{
		{
			name:         "auto enabled and unsupported fast retries",
			autoPAFXFAST: true,
			usePAFXFAST:  true,
			err:          unsupported,
			expect:       true,
		},
		{
			name:         "auto disabled does not retry",
			autoPAFXFAST: false,
			usePAFXFAST:  true,
			err:          unsupported,
			expect:       false,
		},
		{
			name:         "fast not used does not retry",
			autoPAFXFAST: true,
			usePAFXFAST:  false,
			err:          unsupported,
			expect:       false,
		},
		{
			name:         "other errors do not retry",
			autoPAFXFAST: true,
			usePAFXFAST:  true,
			err:          nonUnsupported,
			expect:       false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cl := NewWithKeytab("username", "REALM", &keytab.Keytab{}, &config.Config{}, AutoPAFXFAST(tc.autoPAFXFAST))
			got := cl.shouldRetryASExchangeWithoutPAFXFAST(tc.usePAFXFAST, tc.err)
			if got != tc.expect {
				t.Fatalf("expected retry=%t, got %t", tc.expect, got)
			}
		})
	}
}

func TestRemovePADataType(t *testing.T) {
	t.Parallel()

	const keepType int32 = patype.PA_ENC_TIMESTAMP
	const removeType int32 = patype.PA_REQ_ENC_PA_REP

	tests := []struct {
		name       string
		in         types.PADataSequence
		removeType int32
		wantCount  int
		wantNoType bool
	}{
		{
			name:       "empty sequence",
			in:         types.PADataSequence{},
			removeType: removeType,
			wantCount:  0,
			wantNoType: true,
		},
		{
			name: "type absent unchanged",
			in: types.PADataSequence{
				{PADataType: keepType},
				{PADataType: patype.PA_ETYPE_INFO},
			},
			removeType: removeType,
			wantCount:  2,
			wantNoType: true,
		},
		{
			name: "single matching type removed",
			in: types.PADataSequence{
				{PADataType: keepType},
				{PADataType: removeType},
			},
			removeType: removeType,
			wantCount:  1,
			wantNoType: true,
		},
		{
			name: "multiple matching types removed",
			in: types.PADataSequence{
				{PADataType: removeType},
				{PADataType: keepType},
				{PADataType: removeType},
				{PADataType: patype.PA_ETYPE_INFO2},
				{PADataType: removeType},
			},
			removeType: removeType,
			wantCount:  2,
			wantNoType: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out := removePADataType(tc.in, tc.removeType)
			if len(out) != tc.wantCount {
				t.Fatalf("expected count=%d, got %d", tc.wantCount, len(out))
			}
			hasType := false
			for _, pa := range out {
				if pa.PADataType == tc.removeType {
					hasType = true
					break
				}
			}
			if tc.wantNoType && hasType {
				t.Fatalf("expected removed type %d to be absent", tc.removeType)
			}
		})
	}
}
