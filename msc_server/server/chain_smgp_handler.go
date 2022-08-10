package server

func SmgpHandlers() []TrafficHandler {
	return []TrafficHandler{
		smgpSubmit,
		smgpActive,
		smgpActiveResp,
		smgpDeliveryResp,
		smgpLogin,
		smgpExit,
		smgpExitResp,
	}
}
