package runner

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type LogCollectorHook struct {
	Logs  map[string][]string
	Mutex sync.RWMutex
}

func NewLogCollectorHook() *LogCollectorHook {
	return &LogCollectorHook{
		Logs:  make(map[string][]string),
		Mutex: sync.RWMutex{},
	}
}

func (h *LogCollectorHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *LogCollectorHook) Fire(e *logrus.Entry) error {
	itemID, ok := e.Data["item"].(string)
	if ok && itemID != "" { // only collect item based logs
		h.Mutex.Lock()
		h.Logs[itemID] = append(h.Logs[itemID], e.Message)
		h.Mutex.Unlock()
	}
	return nil
}

func (h *LogCollectorHook) ClearLogs() {
	h.Mutex.Lock()
	h.Logs = make(map[string][]string)
	h.Mutex.Unlock()
}
func (h *LogCollectorHook) GetLogs() map[string][]string {
	return h.Logs
}
