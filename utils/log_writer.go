package utils

import "github.com/deep-patel/log-collection-server/logger"
import "fmt"

type LogConfig struct{
	LogLocationDir string
	MaxLogFiles int
	MaxFileToDelete int
	MaxSizeOfLogFiles uint32
	LogTrace bool
}

func InitLogger(configParam *LogConfig){
  fmt.Printf("-----Inititalizing logger----\nLog Dir: %s\nMaxLogFiles: %d\nMaxFileToDelete: %d\nMaxSizeOfLogFiles: %d\n", configParam.LogLocationDir, configParam.MaxLogFiles, configParam.MaxFileToDelete, configParam.MaxSizeOfLogFiles)
  logger.Init(configParam.LogLocationDir, configParam.MaxLogFiles, configParam.MaxFileToDelete, configParam.MaxSizeOfLogFiles, configParam.LogTrace, false)
}

func WriteLog(content string){
  logger.Info(content)
}
