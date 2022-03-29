package internal

type MsgRaftCmd struct {
	// TODO: Request *RaftCmdRequest
	Callback *Callback
}

func NewMsgRaftCmd() *MsgRaftCmd {
	return nil
}
