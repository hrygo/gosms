package server

func SgipHandlers() []TrafficHandler {
	return []TrafficHandler{
		sgipSubmit,
		sgipBind,
		sgipUnbind,
		sgipUnbindRsp,
	}
}
