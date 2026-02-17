package structure

import (
	"go.uber.org/zap/zapcore"
)

type ID struct {
	TraceId string
	UserId  int
}

func (id ID) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("trace_id", id.TraceId)
	enc.AddInt("user_id", id.UserId)
	return nil
}

type Website struct {
	Endpoint string
	Method   string
}

func (w Website) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("endpoint", w.Endpoint)
	enc.AddString("method", w.Method)
	return nil
}

type LogConfig struct {
	Status  bool
	Service string
	Error string
}

func (lc LogConfig) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddBool("status", lc.Status)
	enc.AddString("service", lc.Service)
	enc.AddString("error", lc.Service)
	return nil
}

type LogHSR struct {
	Status  bool
	ID      ID
	Service string
	Website Website
	Error   string
}

func (lhsr LogHSR) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddBool("status", lhsr.Status)
	enc.AddObject("id_id", lhsr.ID)
	enc.AddString("service", lhsr.Service)
	enc.AddObject("website_website", lhsr.Website)
	enc.AddString("error", lhsr.Error)
	return nil
}
