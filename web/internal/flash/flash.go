package flash

type MessageType string

const (
	Success = MessageType("success")
	Info    = MessageType("info")
	Error   = MessageType("error")
)

type Message struct {
	Type    MessageType
	Content string
}
