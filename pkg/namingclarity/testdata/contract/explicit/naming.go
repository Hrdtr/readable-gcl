package explicit

type fileTransferHandler struct{}
type handlerAlias = *fileTransferHandler

var transfer handlerAlias
var handler handlerAlias // want: namingclarity: "handler" should be named "transfer"
