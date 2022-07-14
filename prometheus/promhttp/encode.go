package promhttp

import "net/http"

const (
	hdrAccept = "Accept"
)

// Negotiate returns the Content-Type based on the given Accept header. If no
// appropriate accepted type is found, FmtText is returned (which is the
// Prometheus text format). This function will never negotiate FmtOpenMetrics,
// as the support is still experimental. To include the option to negotiate
// FmtOpenMetrics, use NegotiateOpenMetrics.
func Negotiate(h http.Header) Format {
	for _, ac := range ParseAccept(h.Get(hdrAccept)) {
		ver := ac.Params["version"]
		if ac.Type+"/"+ac.SubType == ProtoType && ac.Params["proto"] == ProtoProtocol {
			switch ac.Params["encoding"] {
			case "delimited":
				return FmtProtoDelim
			case "text":
				return FmtProtoText
			case "compact-text":
				return FmtProtoCompact
			}
		}
		if ac.Type == "text" && ac.SubType == "plain" && (ver == TextVersion || ver == "") {
			return FmtText
		}
	}
	return FmtText
}

// NegotiateIncludingOpenMetrics works like Negotiate but includes
// FmtOpenMetrics as an option for the result. Note that this function is
// temporary and will disappear once FmtOpenMetrics is fully supported and as
// such may be negotiated by the normal Negotiate function.
func NegotiateIncludingOpenMetrics(h http.Header) Format {
	for _, ac := range ParseAccept(h.Get(hdrAccept)) {
		ver := ac.Params["version"]
		if ac.Type+"/"+ac.SubType == ProtoType && ac.Params["proto"] == ProtoProtocol {
			switch ac.Params["encoding"] {
			case "delimited":
				return FmtProtoDelim
			case "text":
				return FmtProtoText
			case "compact-text":
				return FmtProtoCompact
			}
		}
		if ac.Type == "text" && ac.SubType == "plain" && (ver == TextVersion || ver == "") {
			return FmtText
		}
		if ac.Type+"/"+ac.SubType == OpenMetricsType && (ver == OpenMetricsVersion || ver == "") {
			return FmtOpenMetrics
		}
	}
	return FmtText
}
