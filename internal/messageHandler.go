package internal

var _ MessageHandler = new(MessageHandleProc)

type MessageHandleProc func(ctx *Context, message *Message) error

func (proc MessageHandleProc) ProcessMessage(ctx *Context, message *Message) error {
	return proc(ctx, message)
}

// //////////////////////////////////////////
var _ MessageHandleProc = StopRecursiveForwardMessageHandler

func StopRecursiveForwardMessageHandler(ctx *Context, msg *Message) error {
	ctx.logger.Fatal("invalid forward; it might be recursive forward message to InvalidMessageHandler")
	return nil
}
