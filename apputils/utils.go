package apputils

import (
	"os/user"

	"github.com/kpango/glg"
)

func InitLoggin() {
	tempUser, err := user.Current()
	if CheckError(err) != true {
		logFileWriter := glg.FileWriter(tempUser.HomeDir+"/localeConverter/log.txt", 0666)
		glg.Get().SetMode(glg.BOTH).AddWriter(logFileWriter)
	}
}

func WriteLog(msg string, level glg.LEVEL) {
	switch level {
	case glg.INFO:
		glg.Info(msg)
		break

	case glg.ERR:
		glg.Error(msg)
		break

	case glg.WARN:
		glg.Warn(msg)
		break
	}
}

func StopRunning() {
	glg.Fatalln("Some problems occurs. Stop running")
}

func CheckError(tempError error) bool {
	if tempError != nil {
		glg.Error(tempError)
		return true
	}

	return false
}
