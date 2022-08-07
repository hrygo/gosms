# gosmsn golang开发的短信网关

## 快速开始

从源码构建并启动

```bash
make linux
# make darwin 
make client

cd publish
 
./start.sh

./cli/smscli -p 13800001111 -m 'hello world, 你好世界！' -i 10000
```

## 功能及原理说明



