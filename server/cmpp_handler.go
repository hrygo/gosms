package server

func CmppHandlers() []TrafficHandler {
	return []TrafficHandler{
		cmppSubmit,
		cmppActive,
		cmppActiveResp,
		cmppDeliveryResp,
		cmppConnect,
		cmppTerminate,
		cmppTerminateResp,
	}
}
