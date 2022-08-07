#!/bin/sh

nohup ./smscli> stout.log 2>&1 &

tail -10f ./logs/gosms.log
