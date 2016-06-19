package sdp

import (
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestDecodeInterval(t *testing.T) {
	var dt = []struct {
		in  string
		out time.Duration
	}{
		{"0", 0},
		{"7d", time.Hour * 24 * 7},
		{"25h", time.Hour * 25},
		{"63050", time.Second * 63050},
		{"5", time.Second * 5},
	}
	for i, dtt := range dt {
		var (
			d time.Duration
			v = []byte(dtt.in)
		)
		if err := decodeInterval(v, &d); err != nil {
			t.Errorf("dtt[%d]: dec(%s) err: %s", i, dtt.in, err)
			continue
		}
		if d != dtt.out {
			t.Errorf("dtt[%d]: dec(%s) = %s != %s", i, dtt.in, d, dtt.out)
		}
	}
}

func TestDecoder_Decode(t *testing.T) {
	m := new(Message)
	tData := loadData(t, "sdp_session_ex_full")
	session, err := DecodeSession(tData, nil)
	if err != nil {
		t.Fatal(err)
	}
	decoder := Decoder{
		s: session,
	}
	if err := decoder.Decode(m); err != nil {
		errors.Fprint(os.Stderr, err)
		t.Fatal(err)
	}
	if m.Version != 0 {
		t.Error("wat", m.Version)
	}
	if !m.Flag("recvonly") {
		t.Error("flag recvonly not found")
	}
	if len(m.Medias) != 2 {
		t.Error("len(medias)", len(m.Medias))
	}
	if m.Medias[1].Attributes.Value("rtpmap") != "99 h263-1998/90000" {
		log.Println(m.Medias[1].Attributes)
		log.Println(m.Medias[0].Attributes)
		t.Error("rtpmap", m.Medias[1].Attributes.Value("rtpmap"))
	}
	if m.Bandwidths[BandwidthConferenceTotal] != 154798 {
		t.Error("bandwidth bad value", m.Bandwidths[BandwidthConferenceTotal])
	}
	expectedEncryption := Encryption{"clear", "ab8c4df8b8f4as8v8iuy8re"}
	if m.Encryption != expectedEncryption {
		t.Error("bad encryption", m.Encryption, "!=", expectedEncryption)
	}
	expectedEncryption = Encryption{Method: "prompt"}
	if m.Medias[1].Encryption != expectedEncryption {
		t.Error("bad encryption",
			m.Medias[1].Encryption, "!=", expectedEncryption)
	}
	if m.Start() != NTPToTime(2873397496) {
		t.Error(m.Start(), "!=", NTPToTime(2873397496))
	}
	if m.End() != NTPToTime(2873404696) {
		t.Error(m.End(), "!=", NTPToTime(2873404696))
	}
	cExpected := ConnectionData{
		IP:          net.ParseIP("224.2.17.12"),
		TTL:         127,
		NetworkType: "IN",
		AddressType: "IP4",
	}
	if !cExpected.Equal(m.Connection) {
		t.Error(cExpected, "!=", m.Connection)
	}
	// jdoe 2890844526 2890842807 IN IP4 10.47.16.5
	oExpected := Origin{
		IP:             net.ParseIP("10.47.16.5"),
		SessionID:      2890844526,
		SessionVersion: 2890842807,
		NetworkType:    "IN",
		AddressType:    "IP4",
		Username:       "jdoe",
	}
	if !oExpected.Equal(m.Origin) {
		t.Error(oExpected, "!=", m.Origin)
	}
	if len(m.TZAdjustments) < 2 {
		t.Error("tz adjustments count unexpected")
	} else {
		tzExp := TimeZone{
			Start:  NTPToTime(2882844526),
			Offset: -1 * time.Hour,
		}
		if m.TZAdjustments[0] != tzExp {
			t.Error("tz", m.TZAdjustments[0], "!=", tzExp)
		}
		tzExp = TimeZone{
			Start:  NTPToTime(2898848070),
			Offset: 0,
		}
		if m.TZAdjustments[1] != tzExp {
			t.Error("tz", m.TZAdjustments[0], "!=", tzExp)
		}
	}

	if len(m.Medias) != 2 {
		t.Error("media count unexpected")
	} else {
		// audio 49170 RTP/AVP 0
		mExp := MediaDescription{
			Type:     "audio",
			Port:     49170,
			Protocol: "RTP/AVP",
			Format:   "0",
		}
		if m.Medias[0].Description != mExp {
			t.Error("m", m.Medias[0].Description, "!=", mExp)
		}
		// video 51372 RTP/AVP 99
		mExp = MediaDescription{
			Type:     "video",
			Port:     51372,
			Protocol: "RTP/AVP",
			Format:   "99",
		}
		if m.Medias[1].Description != mExp {
			t.Error("m", m.Medias[1].Description, "!=", mExp)
		}
	}
}

func TestDecoder_WebRTC1(t *testing.T) {
	m := new(Message)
	tData := loadData(t, "spd_session_ex_webrtc1")
	session, err := DecodeSession(tData, nil)
	if err != nil {
		t.Fatal(err)
	}
	//for _, l := range session {
	//	fmt.Println(l)
	//}
	decoder := Decoder{
		s: session,
	}
	if err := decoder.Decode(m); err != nil {
		errors.Fprint(os.Stderr, err)
		t.Error(err)
	}
	if m.Version != 0 {
		t.Error("wat", m.Version)
	}
}

func BenchmarkDecoder_Decode(b *testing.B) {
	m := new(Message)
	b.ReportAllocs()
	tData := loadData(b, "sdp_session_ex_full")
	session, err := DecodeSession(tData, nil)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoder := Decoder{
			s: session,
		}
		if err := decoder.Decode(m); err != nil {
			errors.Fprint(os.Stderr, err)
			b.Error(err)
		}
		m.Medias = m.Medias[:0]
	}
}
