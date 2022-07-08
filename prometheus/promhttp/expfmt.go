package promhttp

import (
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	dto "github.com/prometheus/client_model/go"
	"io"
)

type Closer interface {
	Close() error
}

// Encoder types encode metric families into an underlying wire protocol.
type Encoder interface {
	Encode(*dto.MetricFamily) error
}

type encoderCloser struct {
	encode func(*dto.MetricFamily) error
	close  func() error
}

func (ec encoderCloser) Encode(v *dto.MetricFamily) error {
	return ec.encode(v)
}

func (ec encoderCloser) Close() error {
	return ec.close()
}

// Format specifies the HTTP content type of the different wire protocols.
type Format string

const (
	TextVersion        = "0.0.4"
	ProtoType          = `application/vnd.google.protobuf`
	ProtoProtocol      = `io.prometheus.client.MetricFamily`
	ProtoFmt           = ProtoType + "; proto=" + ProtoProtocol + ";"
	OpenMetricsType    = `application/openmetrics-text`
	OpenMetricsVersion = "0.0.1"

	// The Content-Type values for the different wire protocols.
	FmtUnknown      Format = `<unknown>`
	FmtText         Format = `text/plain; version=` + TextVersion + `; charset=utf-8`
	FmtProtoDelim   Format = ProtoFmt + ` encoding=delimited`
	FmtProtoText    Format = ProtoFmt + ` encoding=text`
	FmtProtoCompact Format = ProtoFmt + ` encoding=compact-text`
	FmtOpenMetrics  Format = OpenMetricsType + `; version=` + OpenMetricsVersion + `; charset=utf-8`
)

func NewEncoder(w io.Writer, format Format) Encoder {
	switch format {
	case FmtProtoDelim:
		return encoderCloser{
			encode: func(v *dto.MetricFamily) error {
				_, err := pbutil.WriteDelimited(w, v)
				return err
			},
			close: func() error { return nil },
		}
	case FmtProtoCompact:
		return encoderCloser{
			encode: func(v *dto.MetricFamily) error {
				_, err := fmt.Fprintln(w, v.String())
				return err
			},
			close: func() error { return nil },
		}
	case FmtProtoText:
		return encoderCloser{
			encode: func(v *dto.MetricFamily) error {
				_, err := fmt.Fprintln(w, proto.MarshalTextString(v))
				return err
			},
			close: func() error { return nil },
		}
	case FmtText:
		return encoderCloser{
			encode: func(v *dto.MetricFamily) error {
				_, err := MetricFamilyToText(w, v)
				return err
			},
			close: func() error { return nil },
		}
	case FmtOpenMetrics:
		return encoderCloser{
			encode: func(v *dto.MetricFamily) error {
				_, err := MetricFamilyToOpenMetrics(w, v)
				return err
			},
			close: func() error {
				_, err := FinalizeOpenMetrics(w)
				return err
			},
		}
	}
	panic(fmt.Errorf("expfmt.NewEncoder: unknown format %q", format))
}

// WriteDelimited encodes and dumps a message to the provided writer prefixed
// with a 32-bit varint indicating the length of the encoded message, producing
// a length-delimited record stream, which can be used to chain together
// encoded messages of the same type together in a file.  It returns the total
// number of bytes written and any applicable error.  This is roughly
// equivalent to the companion Java API's MessageLite#writeDelimitedTo.
func WriteDelimited(w io.Writer, m proto.Message) (n int, err error) {
	buffer, err := proto.Marshal(m)
	if err != nil {
		return 0, err
	}

	var buf [binary.MaxVarintLen32]byte
	encodedLength := binary.PutUvarint(buf[:], uint64(len(buffer)))

	sync, err := w.Write(buf[:encodedLength])
	if err != nil {
		return sync, err
	}

	n, err = w.Write(buffer)
	return n + sync, err
}
