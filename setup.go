package vegeta

import (
	"os"
	"path/filepath"
	"time"

	"github.com/Code-Hex/exit"
	"github.com/Code-Hex/vegeta/internal/utils"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	xslate "github.com/lestrrat/go-xslate"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	LogDir  = "log"
	LogName = "vegeta_log"
)

func (e *Engine) setup() error {
	e.Server.Handler = e.router
	e.HTTPErrorHandler = e.DefaultHTTPErrorHandler
	e.pool.New = func() interface{} {
		return e.NewContext(nil, nil)
	}
	if err := e.setupXSlate(); err != nil {
		return errors.Wrap(err, "Failed to setup xslate")
	}

	err := e.setupLogger(
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)
	if err != nil {
		return errors.Wrap(err, "Failed to setup logger")
	}

	return nil
}

func (e *Engine) setupLogger(opts ...zap.Option) error {
	config := genLoggerConfig()
	enc := zapcore.NewJSONEncoder(config.EncoderConfig)

	dir := LogDir
	ok, err := utils.Exists(dir)
	if err != nil {
		return exit.MakeUnAvailable(err)
	}
	if !ok {
		os.Mkdir(dir, os.ModeDir|os.ModePerm)
	}
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return exit.MakeUnAvailable(err)
	}
	logf, err := rotatelogs.New(
		filepath.Join(absPath, LogName+".%Y%m%d%H%M"),
		rotatelogs.WithLinkName(filepath.Join(absPath, LogName)),
		rotatelogs.WithMaxAge(24*time.Hour),
		rotatelogs.WithRotationTime(time.Hour),
	)
	if err != nil {
		return exit.MakeUnAvailable(err)
	}
	core := zapcore.NewCore(enc, zapcore.AddSync(logf), config.Level)
	e.Logger = zap.New(core, opts...)

	return nil
}

func (e *Engine) setupXSlate() (err error) {
	e.Xslate, err = xslate.New(xslate.Args{
		"Loader": xslate.Args{
			"LoadPaths": []string{"./templates"},
		},
		"Parser": xslate.Args{"Syntax": "TTerse"},
	})
	if err != nil {
		return errors.Wrap(err, "Failed to construct xslate")
	}
	return // nil
}

func genLoggerConfig() zap.Config {
	if isProduction() {
		return zap.NewProductionConfig()
	}
	return zap.NewDevelopmentConfig()
}

func isProduction() bool {
	return os.Getenv("STAGE") == "production"
}
