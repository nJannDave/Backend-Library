package log

import (
	"context"
	"stmnplibrary/constanta"
	"stmnplibrary/log/structure"

	"go.uber.org/zap"
)

var ZapLog *zap.Logger

func LogInit(l *zap.Logger) {
	ZapLog = l
}

func LogConfig(m string, s string, errr error) {
	err := &structure.LogConfig{
		Status:  false,
		Service: s,
		Error: errr.Error(),
	}
	ZapLog.Error(m, zap.Object("log_config: ", err))
}

func LogHSR(ctx context.Context, m string, s string, e string, me string, errStr string) {
	uI, exists := ctx.Value(constanta.UI).(int)
	if !exists {
		uI = 0
	}
	tI, exists := ctx.Value(constanta.TI).(string)
	if !exists {
		tI = "0"
	}
	err := &structure.LogHSR{
		Status: false,
		ID: structure.ID{
			TraceId: tI,
			UserId:  uI,
		},
		Service: s,
		Website: structure.Website{
			Endpoint: e,
			Method:   me,
		},
		Error: errStr,
	}
	ZapLog.Error(m, zap.Object("log_hsr: ", err))
}
