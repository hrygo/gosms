package msc

const (
	ErrorsFuncEventAlreadyExists = "注册函数类事件失败，键名已经被注册"
	ErrorsFuncEventNotRegister   = "没有找到键名对应的函数"
	_                            = ""
	ErrorsSessionThreshReached   = "Reached total sessions threshold"
	ErrorsNoEffectiveAction      = "No effective action more than %v"
	ErrorsNoneActiveTestResponse = "None response active test more than 3 times"
	ErrorsDecodePacketBody       = "Decode packet body error"
	ErrorsIllegalPacketLength    = "Illegal packet length %d"
	ErrorsIllegalCommand         = "Illegal command %0x"
	ErrorsSubmitFlowControl      = "Submit message flow control"
)
