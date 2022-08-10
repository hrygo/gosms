package msc

import (
	"github.com/hrygo/gosms/auth"
	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/utils"
)

func FindAuthConf(isp, clientId string) (ac *codec.AuthConf) {
	c := auth.Cache.FindByCid(isp, clientId)
	ac = &codec.AuthConf{}
	utils.StructCopy(c, ac)
	return
}
