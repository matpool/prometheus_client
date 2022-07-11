package promhttp

import (
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	dto "github.com/prometheus/client_model/go"
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
