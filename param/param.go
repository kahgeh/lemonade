package param

type OpenParam struct {
	URI           string
	TransLoopback bool
}

type TmuxSendKeysParam struct {
	Target string
	Keys   string
}
