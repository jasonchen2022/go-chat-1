package im

// Message 客户端交互的消息体
type Message struct {
	Event   string      `json:"event"`   // 事件名称
	Content interface{} `json:"content"` // 消息内容
}

// SenderContent 推送的消息
type SenderContent struct {
	broadcast bool     // 是否广播消息
	exclude   []int64  // 排除的用户(预留)
	receives  []int64  // 推送的用户
	message   *Message // 消息体
}

func NewSenderContent() *SenderContent {
	return &SenderContent{
		broadcast: false,
		exclude:   make([]int64, 0),
		receives:  make([]int64, 0),
	}
}

// SetBroadcast 设置广播推送
func (s *SenderContent) SetBroadcast(value bool) *SenderContent {
	s.broadcast = value
	return s
}

// SetMessage 设置推送数据
func (s *SenderContent) SetMessage(msg *Message) *SenderContent {
	s.message = msg
	return s
}

// SetReceive 设置推送客户端
func (s *SenderContent) SetReceive(cid ...int64) *SenderContent {
	s.receives = append(s.receives, cid...)
	return s
}

// SetExclude 设置广播推送中需要过滤的客户端
func (s *SenderContent) SetExclude(cid ...int64) *SenderContent {
	s.exclude = append(s.exclude, cid...)
	return s
}

// IsBroadcast 判断是否是广播推送
func (s *SenderContent) IsBroadcast() bool {
	return s.broadcast
}

// GetMessage 获取消息内容
func (s *SenderContent) GetMessage() interface{} {
	return s.message
}
