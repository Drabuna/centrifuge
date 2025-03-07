package centrifuge

import (
	"context"

	"github.com/centrifugal/protocol"
)

// ConnectEvent contains fields related to connecting event (when a server
// received Connect protocol command from client).
type ConnectEvent struct {
	// ClientID that was generated by library for client connection.
	ClientID string
	// Token received from client as part of Connect Command.
	Token string
	// Data received from client as part of Connect Command.
	Data []byte
	// Name can contain client name if provided on connect.
	Name string
	// Version can contain client version if provided on connect.
	Version string
	// Transport contains information about transport used by client.
	Transport TransportInfo
	// Channels is a list of channels a client wants to subscribe to
	// (server-side). It's just a way for a client to provide this list.
	// Server should use ConnectReply.Subscriptions to tell Centrifuge
	// the final list of server-side subscriptions for a connection which
	// can differ from the Channels list.
	Channels []string
}

// ConnectReply contains reaction to ConnectEvent.
type ConnectReply struct {
	// Context allows returning a modified context.
	Context context.Context
	// Credentials should be set if app wants to authenticate connection.
	// This field is optional since auth Credentials could be set through
	// HTTP middleware.
	Credentials *Credentials
	// Data allows setting custom data in connect reply.
	Data []byte
	// Subscriptions map contains channels to subscribe connection to on server-side.
	Subscriptions map[string]SubscribeOptions
	// ClientSideRefresh tells library to use client-side refresh logic:
	// i.e. send refresh commands with new connection token. If not set
	// then server-side refresh mechanism will be used.
	ClientSideRefresh bool
}

// ConnectingHandler called when new client authenticates on server.
type ConnectingHandler func(context.Context, ConnectEvent) (ConnectReply, error)

// ConnectHandler called when client connected to server and ready to communicate.
type ConnectHandler func(*Client)

// RefreshEvent contains fields related to refresh event.
type RefreshEvent struct {
	// ClientSideRefresh is true for refresh initiated by client-side refresh workflow.
	ClientSideRefresh bool
	// Token will only be set in case of using client-side refresh mechanism.
	Token string
}

// RefreshReply contains fields determining the reaction on refresh event.
type RefreshReply struct {
	// Expired tells Centrifuge that connection expired. In this case connection will be
	// closed with DisconnectExpired.
	Expired bool
	// ExpireAt defines time in future when connection should expire,
	// zero value means no expiration.
	ExpireAt int64
	// Info allows modifying connection information,
	// zero value means no modification of current connection Info.
	Info []byte
}

// RefreshCallback should be called as soon as handler decides what to do
// with connection refresh event.
type RefreshCallback func(RefreshReply, error)

// RefreshHandler called when it's time to validate client connection and
// update its expiration time if it's still actual.
//
// Centrifuge library supports two ways of refreshing connection: client-side
// and server-side.
//
// The default mechanism is server-side, this means that as soon refresh handler
// set and connection expiration time happens (by timer) – refresh handler will
// be called.
//
// If ClientSideRefresh in ConnectReply inside ConnectingHandler set to true then
// library uses client-side refresh mechanism. In this case library relies on
// Refresh commands sent from client periodically to refresh connection. Refresh
// command contains updated connection token.
type RefreshHandler func(RefreshEvent, RefreshCallback)

// AliveHandler called periodically while connection alive. This is a helper
// to do periodic things which can tolerate some approximation in time. This
// callback will run every ClientPresenceUpdateInterval and can save you a timer.
type AliveHandler func()

// UnsubscribeEvent contains fields related to unsubscribe event.
type UnsubscribeEvent struct {
	// Channel client unsubscribed from.
	Channel string
	// ServerSide set to true for server-side subscription unsubscribe events.
	ServerSide bool
	// Unsubscribe identifies the source of unsubscribe (i.e. why unsubscribed event happened).
	Unsubscribe
	// Disconnect can be additionally set to identify the reason of disconnect when Unsubscribe.Code
	// is UnsubscribeCodeDisconnect - i.e. when unsubscribe caused by a client disconnection process.
	// Otherwise, it's nil.
	Disconnect *Disconnect
}

// UnsubscribeHandler called when client unsubscribed from channel.
type UnsubscribeHandler func(UnsubscribeEvent)

// DisconnectEvent contains fields related to disconnect event.
type DisconnectEvent struct {
	// Disconnect contains a Disconnect object which identifies the code and reason
	// of disconnect process. When disconnect was not initiated by a server this
	// is always DisconnectConnectionClosed.
	Disconnect
}

// DisconnectHandler called when client disconnects from server. The important
// thing to remember is that you should not rely entirely on this handler to
// clean up non-expiring resources (in your database for example). Why? Because
// in case of any non-graceful node shutdown (kill -9, process crash, machine lost)
// disconnect handler will never be called (obviously) so you can have stale data.
type DisconnectHandler func(DisconnectEvent)

// SubscribeEvent contains fields related to subscribe event.
type SubscribeEvent struct {
	// Channel client wants to subscribe to.
	Channel string
	// Token will only be set for token channels. This is a task of application
	// to check that subscription to a channel has valid token.
	Token string
	// Data received from client as part of Subscribe Command.
	Data []byte
	// Positioned is true when Client wants to create subscription with positioned property.
	Positioned bool
	// Recoverable is true when Client wants to create subscription with recoverable property.
	Recoverable bool
	// JoinLeave is true when Client wants to receive join/leave messages.
	JoinLeave bool
}

// SubscribeCallback should be called as soon as handler decides what to do
// with connection subscribe event.
type SubscribeCallback func(SubscribeReply, error)

// SubscribeReply contains fields determining the reaction on subscribe event.
type SubscribeReply struct {
	// Options to control subscription.
	Options SubscribeOptions

	// ClientSideRefresh tells library to use client-side refresh logic: i.e. send
	// SubRefresh commands with new Subscription Token. If not set then server-side
	// SubRefresh handler will be used.
	ClientSideRefresh bool
}

// SubscribeHandler called when client wants to subscribe on channel.
type SubscribeHandler func(SubscribeEvent, SubscribeCallback)

// PublishEvent contains fields related to publish event. Note that this event
// called before actual publish to Broker so handler has an option to reject this
// publication returning an error.
type PublishEvent struct {
	// Channel client wants to publish data to.
	Channel string
	// Data client wants to publish.
	Data []byte
	// ClientInfo about client connection.
	ClientInfo *ClientInfo
}

// PublishReply contains fields determining the result on publish.
type PublishReply struct {
	// Options to control publication.
	Options PublishOptions

	// Result if set will tell Centrifuge that message already published to
	// channel by handler code. In this case Centrifuge won't try to publish
	// into channel again after handler returned PublishReply. This can be
	// useful if you need to know new Publication offset in your code, or you
	// want to make sure message successfully published to Broker on server
	// side (otherwise only client will get an error).
	Result *PublishResult
}

// PublishCallback should be called with PublishReply or error.
type PublishCallback func(PublishReply, error)

// PublishHandler called when client publishes into channel.
type PublishHandler func(PublishEvent, PublishCallback)

// SubRefreshEvent contains fields related to subscription refresh event.
type SubRefreshEvent struct {
	// ClientSideRefresh is true for refresh initiated by client-side subscription
	// refresh workflow.
	ClientSideRefresh bool
	// Channel to which SubRefreshEvent belongs to.
	Channel string
	// Token will only be set in case of using client-side subscription refresh mechanism.
	Token string
}

// SubRefreshReply contains fields determining the reaction on
// subscription refresh event.
type SubRefreshReply struct {
	// Expired tells Centrifuge that subscription expired. In this case connection will be
	// closed with DisconnectExpired.
	Expired bool
	// ExpireAt is a new Unix time of expiration. Zero value means no expiration.
	ExpireAt int64
	// Info is a new channel-scope info. Zero value means do not change previous one.
	Info []byte
}

// SubRefreshCallback should be called as soon as handler decides what to do
// with connection SubRefreshEvent.
type SubRefreshCallback func(SubRefreshReply, error)

// SubRefreshHandler called when it's time to validate client subscription to channel and
// update it's state if needed.
//
// If ClientSideRefresh in SubscribeReply inside SubscribeHandler set to true then
// library uses client-side subscription refresh mechanism. In this case library relies on
// SubRefresh commands sent from client periodically to refresh subscription. SubRefresh
// command contains updated subscription token.
type SubRefreshHandler func(SubRefreshEvent, SubRefreshCallback)

// RPCEvent contains fields related to rpc request.
type RPCEvent struct {
	// Method is an optional string that contains RPC method name client wants to call.
	// This is an optional field, by default clients send RPC without any method set.
	Method string
	// Data contains RPC untouched payload.
	Data []byte
}

// RPCReply contains fields determining the reaction on rpc request.
type RPCReply struct {
	// Data to return in RPC reply to client.
	Data []byte
}

// RPCCallback should be called as soon as handler decides what to do
// with connection RPCEvent.
type RPCCallback func(RPCReply, error)

// RPCHandler must handle incoming command from client.
type RPCHandler func(RPCEvent, RPCCallback)

// MessageEvent contains fields related to message request.
type MessageEvent struct {
	// Data contains message untouched payload.
	Data []byte
}

// MessageHandler must handle incoming async message from client.
type MessageHandler func(MessageEvent)

// PresenceEvent has channel operation called for.
type PresenceEvent struct {
	Channel string
}

// PresenceReply contains fields determining the reaction on presence request.
type PresenceReply struct {
	Result *PresenceResult
}

// PresenceCallback should be called with PresenceReply or error.
type PresenceCallback func(PresenceReply, error)

// PresenceHandler called when presence request received from client.
type PresenceHandler func(PresenceEvent, PresenceCallback)

// PresenceStatsEvent has channel operation called for.
type PresenceStatsEvent struct {
	Channel string
}

// PresenceStatsReply contains fields determining the reaction on presence request.
type PresenceStatsReply struct {
	Result *PresenceStatsResult
}

// PresenceStatsCallback should be called with PresenceStatsReply or error.
type PresenceStatsCallback func(PresenceStatsReply, error)

// PresenceStatsHandler must handle incoming command from client.
type PresenceStatsHandler func(PresenceStatsEvent, PresenceStatsCallback)

// HistoryEvent has channel operation called for.
type HistoryEvent struct {
	Channel string
	Filter  HistoryFilter
}

// HistoryReply contains fields determining the reaction on history request.
type HistoryReply struct {
	Result *HistoryResult
}

// HistoryCallback should be called with HistoryReply or error.
type HistoryCallback func(HistoryReply, error)

// HistoryHandler must handle incoming command from client.
type HistoryHandler func(HistoryEvent, HistoryCallback)

// StateSnapshotHandler must return a copy of current client's
// internal state. Returning a copy is important to avoid data races.
type StateSnapshotHandler func() (interface{}, error)

// SurveyEvent with Op and Data of survey.
type SurveyEvent struct {
	Op   string
	Data []byte
}

// SurveyReply contains survey reply fields.
type SurveyReply struct {
	Code uint32
	Data []byte
}

// SurveyCallback should be called with SurveyReply as soon as survey completed.
type SurveyCallback func(SurveyReply)

// SurveyHandler allows setting survey handler function.
type SurveyHandler func(SurveyEvent, SurveyCallback)

// NotificationEvent with Op and Data.
type NotificationEvent struct {
	FromNodeID string
	Op         string
	Data       []byte
}

// NotificationHandler allows handling notifications.
type NotificationHandler func(NotificationEvent)

// NodeInfoSendReply can modify sending Node control frame in some ways.
type NodeInfoSendReply struct {
	// Data allows setting an arbitrary data to the control node frame which is
	// published by each Node periodically, so it will be available in the
	// result of Node.Info call for the current Node description. Keep this
	// data reasonably small.
	Data []byte
}

// NodeInfoSendHandler called every time the control node frame is published
// and allows modifying Node control frame sending. Currently, attaching an
// arbitrary data to it. See NodeInfoSendReply.
type NodeInfoSendHandler func() NodeInfoSendReply

// TransportWriteEvent called just before sending data into the client connection. The
// event is triggered from inside each client's message queue consumer – so it should
// not directly affect Hub broadcast latencies.
type TransportWriteEvent struct {
	// Data represents single Centrifuge protocol message which is going to be sent
	// into the connection. For unidirectional transports this is an encoded protocol.Push
	// type, for bidirectional transports this is an encoded protocol.Reply type.
	Data []byte
}

// TransportWriteHandler called just before writing data to the Transport.
// At this moment application can skip sending data to a client returning
// false from a handler. The main purpose of this handler is not a message
// filtering based on data content but rather tracing stuff.
type TransportWriteHandler func(*Client, TransportWriteEvent) bool

// CommandReadEvent contains protocol.Command processed by Client.
type CommandReadEvent struct {
	Command *protocol.Command
}

// CommandReadHandler allows setting a callback which will be called after
// Client processed a protocol.Command. This exists mostly for real-time connection
// tracing purposes. Theoretically CommandReadHandler may be called after the
// corresponding Reply written to connection and TransportWriteHandler called. But
// for tracing purposes this seems tolerable as commands and replies may be matched
// by id.
type CommandReadHandler func(*Client, CommandReadEvent)
