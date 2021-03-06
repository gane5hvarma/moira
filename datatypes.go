package moira

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"time"
)

var (
	eventStates = [...]string{"OK", "WARN", "ERROR", "NODATA", "EXCEPTION", "TEST"}
)

var scores = map[string]int64{
	"OK":        0,
	"DEL":       0,
	"WARN":      1,
	"ERROR":     100,
	"NODATA":    1000,
	"EXCEPTION": 100000,
}

var eventStateWeight = map[string]int{
	"OK":     0,
	"WARN":   1,
	"ERROR":  100,
	"NODATA": 10000,
}

// NotificationEvent represents trigger state changes event
type NotificationEvent struct {
	IsTriggerEvent bool     `json:"trigger_event,omitempty"`
	Timestamp      int64    `json:"timestamp"`
	Metric         string   `json:"metric"`
	Value          *float64 `json:"value,omitempty"`
	State          string   `json:"state"`
	TriggerID      string   `json:"trigger_id"`
	SubscriptionID *string  `json:"sub_id,omitempty"`
	ContactID      string   `json:"contactId,omitempty"`
	OldState       string   `json:"old_state"`
	Message        *string  `json:"msg,omitempty"`
}

// NotificationEvents represents slice of NotificationEvent
type NotificationEvents []NotificationEvent

// TriggerData represents trigger object
type TriggerData struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Desc       string   `json:"desc"`
	Targets    []string `json:"targets"`
	WarnValue  float64  `json:"warn_value"`
	ErrorValue float64  `json:"error_value"`
	IsRemote   bool     `json:"is_remote"`
	Tags       []string `json:"__notifier_trigger_tags"`
}

// ContactData represents contact object
type ContactData struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	ID    string `json:"id"`
	User  string `json:"user"`
}

// SubscriptionData represents user subscription
type SubscriptionData struct {
	Contacts          []string     `json:"contacts"`
	Tags              []string     `json:"tags"`
	Schedule          ScheduleData `json:"sched"`
	Plotting          PlottingData `json:"plotting"`
	ID                string       `json:"id"`
	Enabled           bool         `json:"enabled"`
	IgnoreWarnings    bool         `json:"ignore_warnings,omitempty"`
	IgnoreRecoverings bool         `json:"ignore_recoverings,omitempty"`
	ThrottlingEnabled bool         `json:"throttling"`
	User              string       `json:"user"`
}

// PlottingData represents plotting settings
type PlottingData struct {
	Enabled bool   `json:"enabled"`
	Theme   string `json:"theme"`
}

// ScheduleData represents subscription schedule
type ScheduleData struct {
	Days           []ScheduleDataDay `json:"days"`
	TimezoneOffset int64             `json:"tzOffset"`
	StartOffset    int64             `json:"startOffset"`
	EndOffset      int64             `json:"endOffset"`
}

// ScheduleDataDay represents week day of schedule
type ScheduleDataDay struct {
	Enabled bool   `json:"enabled"`
	Name    string `json:"name,omitempty"`
}

// ScheduledNotification represent notification object
type ScheduledNotification struct {
	Event     NotificationEvent `json:"event"`
	Trigger   TriggerData       `json:"trigger"`
	Contact   ContactData       `json:"contact"`
	Plotting  PlottingData      `json:"plotting"`
	Throttled bool              `json:"throttled"`
	SendFail  int               `json:"send_fail"`
	Timestamp int64             `json:"timestamp"`
}

// MatchedMetric represents parsed and matched metric data
type MatchedMetric struct {
	Metric             string
	Patterns           []string
	Value              float64
	Timestamp          int64
	RetentionTimestamp int64
	Retention          int
}

// MetricValue represents metric data
type MetricValue struct {
	RetentionTimestamp int64   `json:"step,omitempty"`
	Timestamp          int64   `json:"ts"`
	Value              float64 `json:"value"`
}

const (
	// FallingTrigger represents falling trigger type, in which OK > WARN > ERROR
	FallingTrigger = "falling"
	// RisingTrigger represents rising trigger type, in which OK < WARN < ERROR
	RisingTrigger = "rising"
	// ExpressionTrigger represents trigger type with custom user expression
	ExpressionTrigger = "expression"
)

// Trigger represents trigger data object
type Trigger struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	Desc             *string       `json:"desc,omitempty"`
	Targets          []string      `json:"targets"`
	WarnValue        *float64      `json:"warn_value"`
	ErrorValue       *float64      `json:"error_value"`
	TriggerType      string        `json:"trigger_type"`
	Tags             []string      `json:"tags"`
	TTLState         *string       `json:"ttl_state,omitempty"`
	TTL              int64         `json:"ttl,omitempty"`
	Schedule         *ScheduleData `json:"sched,omitempty"`
	Expression       *string       `json:"expression,omitempty"`
	PythonExpression *string       `json:"python_expression,omitempty"`
	Patterns         []string      `json:"patterns"`
	IsRemote         bool          `json:"is_remote"`
	MuteNewMetrics   bool          `json:"mute_new_metrics"`
}

// TriggerCheck represents trigger data with last check data and check timestamp
type TriggerCheck struct {
	Trigger
	Throttling int64     `json:"throttling"`
	LastCheck  CheckData `json:"last_check"`
}

// CheckData represents last trigger check data
type CheckData struct {
	Metrics                      map[string]MetricState `json:"metrics"`
	Score                        int64                  `json:"score"`
	State                        string                 `json:"state"`
	Maintenance                  int64                  `json:"maintenance,omitempty"`
	Timestamp                    int64                  `json:"timestamp,omitempty"`
	EventTimestamp               int64                  `json:"event_timestamp,omitempty"`
	LastSuccessfulCheckTimestamp int64                  `json:"last_successful_check_timestamp"`
	Suppressed                   bool                   `json:"suppressed,omitempty"`
	SuppressedState              string                 `json:"suppressed_state,omitempty"`
	Message                      string                 `json:"msg,omitempty"`
}

// MetricState represents metric state data for given timestamp
type MetricState struct {
	EventTimestamp  int64    `json:"event_timestamp"`
	State           string   `json:"state"`
	Suppressed      bool     `json:"suppressed"`
	SuppressedState string   `json:"suppressed_state,omitempty"`
	Timestamp       int64    `json:"timestamp"`
	Value           *float64 `json:"value,omitempty"`
	Maintenance     int64    `json:"maintenance,omitempty"`
}

// MetricEvent represents filter metric event
type MetricEvent struct {
	Metric  string `json:"metric"`
	Pattern string `json:"pattern"`
}

// GetSubjectState returns the most critical state of events
func (events NotificationEvents) GetSubjectState() string {
	result := ""
	states := make(map[string]bool)
	for _, event := range events {
		states[event.State] = true
	}
	for _, state := range eventStates {
		if states[state] {
			result = state
		}
	}
	return result
}

// GetTags returns "[tag1][tag2]...[tagN]" string
func (trigger *TriggerData) GetTags() string {
	var buffer bytes.Buffer
	for _, tag := range trigger.Tags {
		buffer.WriteString(fmt.Sprintf("[%s]", tag))
	}
	return buffer.String()
}

// GetKey return notification key to prevent duplication to the same contact
func (notification *ScheduledNotification) GetKey() string {
	return fmt.Sprintf("%s:%s:%s:%s:%s:%d:%f:%d:%t:%d",
		notification.Contact.Type,
		notification.Contact.Value,
		notification.Event.TriggerID,
		notification.Event.Metric,
		notification.Event.State,
		notification.Event.Timestamp,
		UseFloat64(notification.Event.Value),
		notification.SendFail,
		notification.Throttled,
		notification.Timestamp,
	)
}

// IsScheduleAllows check if the time is in the allowed schedule interval
func (schedule *ScheduleData) IsScheduleAllows(ts int64) bool {
	if schedule == nil {
		return true
	}
	endOffset, startOffset := schedule.EndOffset, schedule.StartOffset
	if schedule.EndOffset < schedule.StartOffset {
		endOffset = schedule.EndOffset + 24*60
	}
	timestamp := ts - ts%60 - schedule.TimezoneOffset*60
	date := time.Unix(timestamp, 0).UTC()
	if !schedule.Days[int(date.Weekday()+6)%7].Enabled {
		return false
	}
	dayStart := time.Unix(timestamp-timestamp%(24*3600), 0).UTC()
	startDayTime := dayStart.Add(time.Duration(startOffset) * time.Minute)
	endDayTime := dayStart.Add(time.Duration(endOffset) * time.Minute)
	if endOffset < 24*60 {
		if (date.After(startDayTime) || date.Equal(startDayTime)) && (date.Before(endDayTime) || date.Equal(endDayTime)) {
			return true
		}
	} else {
		endDayTime = endDayTime.Add(-time.Hour * 24)
		if date.Before(endDayTime) || date.After(startDayTime) {
			return true
		}
	}
	return false
}

func (eventData NotificationEvent) String() string {
	return fmt.Sprintf("TriggerId: %s, Metric: %s, Value: %v, OldState: %s, State: %s, Message: '%s', Timestamp: %v", eventData.TriggerID, eventData.Metric, UseFloat64(eventData.Value), eventData.OldState, eventData.State, UseString(eventData.Message), eventData.Timestamp)
}

// GetOrCreateMetricState gets metric state from check data or create new if CheckData has no state for given metric
func (checkData *CheckData) GetOrCreateMetricState(metric string, emptyTimestampValue int64, muteNewMetric bool) MetricState {
	_, ok := checkData.Metrics[metric]
	if !ok {
		checkData.Metrics[metric] = createEmptyMetricState(emptyTimestampValue, !muteNewMetric)
	}
	return checkData.Metrics[metric]
}

func createEmptyMetricState(defaultTimestampValue int64, firstStateIsNodata bool) MetricState {
	if firstStateIsNodata {
		return MetricState{
			State:     "NODATA",
			Timestamp: defaultTimestampValue,
		}
	}

	unixNow := time.Now().Unix()

	return MetricState{
		State:          "OK",
		Timestamp:      unixNow,
		EventTimestamp: unixNow,
	}
}

// GetCheckPoint gets check point for given MetricState
// CheckPoint is the timestamp from which to start checking the current state of the metric
func (metricState *MetricState) GetCheckPoint(checkPointGap int64) int64 {
	return int64(math.Max(float64(metricState.Timestamp-checkPointGap), float64(metricState.EventTimestamp)))
}

// GetEventTimestamp gets event timestamp for given metric
func (metricState MetricState) GetEventTimestamp() int64 {
	if metricState.EventTimestamp == 0 {
		return metricState.Timestamp
	}
	return metricState.EventTimestamp
}

// GetEventTimestamp gets event timestamp for given check
func (checkData CheckData) GetEventTimestamp() int64 {
	if checkData.EventTimestamp == 0 {
		return checkData.Timestamp
	}
	return checkData.EventTimestamp
}

// IsSimple checks triggers patterns
// If patterns more than one or it contains standard graphite wildcard symbols,
// when this target can contain more then one metrics, and is it not simple trigger
func (trigger *Trigger) IsSimple() bool {
	if len(trigger.Targets) > 1 || len(trigger.Patterns) > 1 {
		return false
	}
	for _, pattern := range trigger.Patterns {
		if strings.ContainsAny(pattern, "*{?[") {
			return false
		}
	}
	return true
}

// UpdateScore update and return checkData score, based on metric states and checkData state
func (checkData *CheckData) UpdateScore() int64 {
	checkData.Score = scores[checkData.State]
	for _, metricData := range checkData.Metrics {
		checkData.Score += scores[metricData.State]
	}
	return checkData.Score
}

// MustIgnore returns true if given state transition must be ignored
func (subscription *SubscriptionData) MustIgnore(eventData *NotificationEvent) bool {
	if oldStateWeight, ok := eventStateWeight[eventData.OldState]; ok {
		if newStateWeight, ok := eventStateWeight[eventData.State]; ok {
			delta := newStateWeight - oldStateWeight
			if delta < 0 {
				if delta == -1 && (subscription.IgnoreRecoverings || subscription.IgnoreWarnings) {
					return true
				}
				return subscription.IgnoreRecoverings
			}
			if delta == 1 {
				return subscription.IgnoreWarnings
			}
		}
	}
	return false
}
