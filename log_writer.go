package main

import "github.com/go-utils/logger"


func InitLogger(configParam *LogConfig){
  logger.Init(configParam.LogLocationDir, configParam.MaxLogFiles, configParam.MaxFileToDelete, configParam.MaxSizeOfLogFiles, configParam.LogTrace)
}

func WriteLog(content string){
  logger.Info(content)
}
