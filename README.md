# gosms golang开发的短信网关

## Quick Start

从源码构建并启动

```bash
# 构建服务端
make linux
# make darwin

# 构建客户端 
make client

cd publish
mv gosms-* gosms
# 启动服务端 
./start.sh

# 启动客户端
cd cli
./smscli -p 13800001111 -m 'hello world, 你好世界！' -i 10000
# -p 手机号
# -m 短信内容
# -i 迭代次数
```

### 采用mongodb存储客户端认证配置

/config/config.yaml

```yaml
AuthClient:
  StoreType: "mongo"
  Mongo:
    URI: "mongodb+srv://<user>:<passwd>@cluster0.ppiyq4w.mongodb.net/test"
    ConnectTimeout: 15s
    ReadTimeout: 15s
    WriteTimeout: 15s
    HeartbeatInterval: 60s
    MinPoolSize: 2
    MaxPoolSize: 10
```

然后通过环境变量设置用户名密码

```bash
export MONGO_USER=xxx
export MONGO_PASSWD=xxx
```

如果不启用MongoDB，需设置

```yaml
AuthClient:
  StoreType: "yml"
````

### 采用mongodb存储客户端消息发送记录

/config/sms.yml

```yaml
Mongo:
  URI: "mongodb+srv://<user>:<passwd>@cluster0.ppiyq4w.mongodb.net/test"
  ConnectTimeout: 15s
  ReadTimeout: 15s
  WriteTimeout: 15s
  HeartbeatInterval: 60s
  MinPoolSize: 2
  MaxPoolSize: 10
```

同样，然后通过环境变量设置用户名密码

```bash
export MONGO_USER=xxx
export MONGO_PASSWD=xxx
```

如果不启用MongoDB，不设置 `Mongo.URI` 即可。

## 功能及原理说明

TODO 其他说明文档待补充