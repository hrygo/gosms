#!/bin/sh

pkill gosms
sleep 3
pkill -9 gosms

# 非必须参数，系统默认值
export CONF_LOG_PATH="./logs"
export CONF_LOG_TIME_FORMAT="2006-01-02T15:04:05.000"

nohup ./gosms> stout.log 2>&1 &
