package messages

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/jcmturner/gofork/encoding/asn1"
	"github.com/jcmturner/gokrb5/v8/asn1tools"
	"github.com/jcmturner/gokrb5/v8/crypto"
	"github.com/jcmturner/gokrb5/v8/iana/asnAppTag"
	"github.com/jcmturner/gokrb5/v8/iana/keyusage"
	"github.com/jcmturner/gokrb5/v8/iana/msgtype"
	"github.com/jcmturner/gokrb5/v8/krberror"
	"github.com/jcmturner/gokrb5/v8/types"
)

// APRep implements RFC 4120 KRB_AP_REP: https://tools.ietf.org/html/rfc4120#section-5.5.2.
type APRep struct {
	PVNO    int                 `asn1:"explicit,tag:0"`
	MsgType int                 `asn1:"explicit,tag:1"`
	EncPart types.EncryptedData `asn1:"explicit,tag:2"`
}

// EncAPRepPart is the encrypted part of KRB_AP_REP.
type EncAPRepPart struct {
	CTime          time.Time           `asn1:"generalized,explicit,tag:0"`
	Cusec          int                 `asn1:"explicit,tag:1"`
	Subkey         types.EncryptionKey `asn1:"optional,explicit,tag:2"`
	SequenceNumber int64               `asn1:"optional,explicit,tag:3"`
}

// NewAPRep prepares and returns a new APRep message in response to a validated APReq.
func NewAPRep(req APReq) (APRep, error) {
	var a APRep
	seq, err := rand.Int(rand.Reader, big.NewInt(math.MaxUint32))
	if err != nil {
		return a, err
	}
	usage := uint32(keyusage.AP_REP_ENCPART)
	key := req.Ticket.DecryptedEncPart.Key
	encAPRepPart := EncAPRepPart{
		CTime:          req.Authenticator.CTime.UTC(),
		Cusec:          req.Authenticator.Cusec,
		SequenceNumber: seq.Int64() & 0x3fffffff,
		Subkey:         req.Authenticator.SubKey,
	}
	fmt.Printf("Creating EncAPRepPart: %#v\n", encAPRepPart)
	b, err := encAPRepPart.Marshal()
	if err != nil {
		return a, err
	}
	fmt.Printf("Marshaled EncAPRepPart: %v\n", b)
	fmt.Printf("Encrypting EncAPRepPart with key %#v\n", key)
	ed, err := crypto.GetEncryptedData(b, key, usage, 0)
	if err != nil {
		return a, err
	}
	fmt.Printf("encrypted EncAPRepPart: %#v\n", ed)

	a = APRep{
		PVNO:    5,
		MsgType: msgtype.KRB_AP_REP,
		EncPart: ed,
	}
	return a, nil
}

// Unmarshal bytes b into the APRep struct.
func (a *APRep) Unmarshal(b []byte) error {
	_, err := asn1.UnmarshalWithParams(b, a, fmt.Sprintf("application,explicit,tag:%v", asnAppTag.APREP))
	if err != nil {
		return processUnmarshalReplyError(b, err)
	}
	expectedMsgType := msgtype.KRB_AP_REP
	if a.MsgType != expectedMsgType {
		return krberror.NewErrorf(krberror.KRBMsgError, "message ID does not indicate a KRB_AP_REP. Expected: %v; Actual: %v", expectedMsgType, a.MsgType)
	}
	return nil
}

// Unmarshal bytes b into the APRep encrypted part struct.
func (a *EncAPRepPart) Unmarshal(b []byte) error {
	_, err := asn1.UnmarshalWithParams(b, a, fmt.Sprintf("application,explicit,tag:%v", asnAppTag.EncAPRepPart))
	if err != nil {
		return krberror.Errorf(err, krberror.EncodingError, "AP_REP unmarshal error")
	}
	return nil
}

// Marshal the EncAPRepPart.
func (a *EncAPRepPart) Marshal() ([]byte, error) {
	b, err := asn1.Marshal(*a)
	if err != nil {
		return nil, krberror.Errorf(err, krberror.EncodingError, "AP_REP marshal error")
	}
	b = asn1tools.AddASNAppTag(b, asnAppTag.EncAPRepPart)
	return b, nil
}

// Marshal APRep struct.
func (a *APRep) Marshal() ([]byte, error) {
	mk, err := asn1.Marshal(*a)
	if err != nil {
		return mk, krberror.Errorf(err, krberror.EncodingError, "marshaling error of AP_REP")
	}
	mk = asn1tools.AddASNAppTag(mk, asnAppTag.APREP)
	return mk, nil
}
