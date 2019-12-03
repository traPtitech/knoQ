package logging

import (
	"go.uber.org/zap/zapcore"
)

type HTTPPayload struct {
	RequestMethod string `json:"requestMethod"`
	RequestURL    string `json:"requestUrl"`
	RequestSize   string `json:"requestSize"`
	Status        int    `json:"status"`
	ResponseSize  string `json:"responseSize"`
	UserAgent     string `json:"userAgent"`
	RemoteIP      string `json:"remoteIp"`
	ServerIP      string `json:"serverIp"`
	Referer       string `json:"referer"`
	Latency       string `json:"latency"`
	Protocol      string `json:"protocol"`
	ErrorLocation string `json:"errorLocation"`
}

// MarshalLogObject implements zapcore.ObjectMarshaller interface.
func (p HTTPPayload) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("requestMethod", p.RequestMethod)
	enc.AddString("requestUrl", p.RequestURL)
	enc.AddString("requestSize", p.RequestSize)
	enc.AddInt("status", p.Status)
	enc.AddString("responseSize", p.ResponseSize)
	enc.AddString("userAgent", p.UserAgent)
	enc.AddString("remoteIp", p.RemoteIP)
	enc.AddString("serverIp", p.ServerIP)
	enc.AddString("referer", p.Referer)
	enc.AddString("latency", p.Latency)
	enc.AddString("protocol", p.Protocol)
	enc.AddString("errorLocation", p.ErrorLocation)
	return nil
}
