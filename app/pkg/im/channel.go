package im

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
)

type WebsocketInterface interface {
	Open(client *Client)
	Message(message *RecvMessage)
	Close(client *Client, code int, text string)
}

// ChannelManager WebSocket 渠道管理（多渠道划分，实现不同业务之间隔离）
type ChannelManager struct {
	Name     string             // 渠道名称
	Count    int                // 客户端连接数
	Clients  map[string]*Client // 客户端列表
	RecvChan chan *RecvMessage  // 消息接收通道
	SendChan chan *SendMessage  // 消息发送通道
	Lock     *sync.Mutex        // 互斥锁
	Handle   WebsocketInterface
}

// RegisterClient 注册客户端
func (c *ChannelManager) RegisterClient(client *Client) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	c.Clients[client.Uuid] = client

	c.Count++
}

// RemoveClient 删除客户端
func (c *ChannelManager) RemoveClient(client *Client) bool {
	_, ok := c.Clients[client.Uuid]
	if !ok {
		return false
	}

	c.Lock.Lock()
	defer c.Lock.Unlock()

	delete(c.Clients, client.Uuid)

	c.Count--
	return true
}

// GetClient 获取客户端
func (c *ChannelManager) GetClient(uuid string) (*Client, bool) {
	client, ok := c.Clients[uuid]

	return client, ok
}

// SendMessage 推送消息到消费通道
func (c *ChannelManager) SendMessage(message *SendMessage) {
	c.SendChan <- message
}

// SetCallbackHandler 设置 WebSocket 处理事件
func (c *ChannelManager) SetCallbackHandler(handle WebsocketInterface) *ChannelManager {
	c.Handle = handle

	return c
}

// Process 渠道消费协程
func (c *ChannelManager) Process() {
	fmt.Printf("[%s] 消费协程已启动\n", c.Name)

	for {
		select {
		// 处理接收消息
		case value, ok := <-c.RecvChan:
			if ok {
				c.Handle.Message(value)
			}

		// 处理发送消息
		case value, ok := <-c.SendChan:
			if !ok {
				fmt.Printf("消费通道[%s]，读取数据失败...", c.Name)
				continue
			}

			content, _ := jsoniter.Marshal(value)

			// 判断是否推送所有客户端
			if value.IsAll {
				for _, client := range c.Clients {
					_ = client.Conn.WriteMessage(websocket.TextMessage, content)
				}
			} else {
				for _, uuid := range value.Clients {
					client, ok := c.Clients[uuid]
					if ok {
						_ = client.Conn.WriteMessage(websocket.TextMessage, content)
					}
				}
			}
		case <-time.After(5 * time.Second):
		}
	}
}