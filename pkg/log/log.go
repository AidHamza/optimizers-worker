package log

import (
	"strings"
	"log/syslog"

	log "gopkg.in/inconshreveable/log15.v2"
)

const (
	logFormat = "json"
	//stdout, stderr, file:///path/to/file, syslog://local/tag or syslog://address:port/tag
	logDest  = "stdout"
	logFiles = true
	logFuncs = true
	//debug, info, warn, error, crit
	logLevel = "debug"
)

var Logger = log.New()

func init() {
	initLogger()
}

func initLogger() {

	var err error
	var handler log.Handler
	var format log.Format

	// Setup logging format
	if logFormat == "text" {
		format = log.LogfmtFormat()
	} else if logFormat == "json" {
		format = log.JsonFormat()
	} else {
		log.Crit("Unrecognized 'core.log.format' option. Possible values: text|json.", "current_value", logFormat)
		panic("Log setup error")
	}

	if logDest == "stdout" {
		handler = log.StdoutHandler
	} else if logDest == "stderr" {
		handler = log.StderrHandler
	} else if strings.HasPrefix(logDest, "file") {
		path := strings.TrimPrefix(logDest, "file://")
		handler, err = log.FileHandler(path, format)
		if err != nil {
			log.Crit("Can't parse logfile path", "err", err.Error(), "value", logDest)
			panic("Log setup error")
		}
	} else if strings.HasPrefix(logDest, "syslog") {
		if strings.HasPrefix(logDest, "syslog://local/") {
			handler, err = log.SyslogHandler(syslog.LOG_DEBUG, strings.TrimPrefix(logDest, "syslog://local/"), format)
			if err != nil {
				log.Crit("Can't parse syslog string", "err", err.Error(), "value", logDest)
				panic("Log setup error")
			}
		} else {
			connstring := strings.Split(strings.TrimPrefix(logDest, "syslog://"), "/")
			addr, tag := connstring[0], connstring[1]
			handler, err = log.SyslogNetHandler("udp", addr, syslog.LOG_DEBUG, tag, format)
			if err != nil {
				log.Crit("Can't parse syslog string", "err", err.Error(), "value", logDest)
				panic("Log setup error")
			}
		}
	} else {
		log.Crit("Unknown destination", "value", logDest)
		panic("Log setup error")
	}

	// Setup funcs/lines logging
	if logFiles {
		handler = log.CallerFileHandler(handler)
	}

	if logFuncs {
		handler = log.CallerFuncHandler(handler)
	}

	// Setup log level
	lvl, err := log.LvlFromString(logLevel)
	if err != nil {
		log.Crit(err.Error())
		panic("Log setup error")
	}

	handler = log.LvlFilterHandler(lvl, handler)
	log.Root().SetHandler(handler)
}
