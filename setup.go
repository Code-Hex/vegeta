package vegeta

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Code-Hex/exit"
	"github.com/Code-Hex/vegeta/internal/utils"
	"github.com/jinzhu/gorm"
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

func (v *Vegeta) setup() error {
	if err := v.setupDatabase(); err != nil {
		return err
	}
	if err := v.setupXslate(); err != nil {
		return errors.Wrap(err, "Failed to setup xslate")
	}
	err := v.setupLogger(
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)
	if err != nil {
		return errors.Wrap(err, "Failed to setup logger")
	}
	return v.setupHandlers()
}

func (v *Vegeta) setupDatabase() error {
	user := os.Getenv("MYSQL_USERNAME")
	passwd := os.Getenv("MYSQL_PASSWORD")
	database := os.Getenv("MYSQL_DATABASE")
	info := fmt.Sprintf(
		"%s:%s@/%s?charset=utf8&parseTime=True&loc=Local",
		user,
		passwd,
		database,
	)
	db, err := gorm.Open("mysql", info)
	if err != nil {
		return err
	}
	v.DB = db
	return nil
}

func (v *Vegeta) setupLogger(opts ...zap.Option) error {
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
	v.Logger = zap.New(core, opts...)

	return nil
}

func (v *Vegeta) setupXslate() (err error) {
	v.Xslate, err = xslate.New(xslate.Args{
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
