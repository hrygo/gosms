#!/bin/sh

pkill gosmsn
sleep 3
pkill -9 gosmsn

# 非必须参数，系统默认值
export CONF_LOG_PATH="./logs"
export CONF_LOG_TIME_FORMAT="2006-01-02T15:04:05.000"

nohup ./gosmsn> stout.log 2>&1 &
