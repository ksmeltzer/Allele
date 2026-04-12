package alerting

type Event string

const (
	EventBoot      Event = "BOOT"
	EventHeartbeat Event = "HEARTBEAT"
	EventCrash     Event = "CRASH"
	EventShutdown  Event = "SHUTDOWN"
)

type Payload struct {
	Event Event                  `json:"event"`
	Data  map[string]interface{} `json:"data,omitempty"`
}
