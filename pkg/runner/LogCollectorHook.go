package runner

import "github.com/sirupsen/logrus"

type LogCollectorHook struct {
	Logs map[string][]string
}

func NewLogCollectorHook() *LogCollectorHook {
	return &LogCollectorHook{Logs: make(map[string][]string)}
}

func (h *LogCollectorHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *LogCollectorHook) Fire(e *logrus.Entry) error {
	itemID, ok := e.Data["item"].(string)
	if ok && itemID != "" { // only collect item based logs
		h.Logs[itemID] = append(h.Logs[itemID], e.Message)
	}
	return nil
}

func (h *LogCollectorHook) ClearLogs() {
	h.Logs = make(map[string][]string)
}
func (h *LogCollectorHook) GetLogs() map[string][]string {
	return h.Logs
}
