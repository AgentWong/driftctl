package alerter

import (
	"fmt"

	"github.com/snyk/driftctl/enumeration/resource"
)

// Interface defines the interface for sending alerts during enumeration.
type Interface interface {
	SendAlert(key string, alert Alert)
}

// Alerter collects and stores alerts produced during a scan via an internal channel.
type Alerter struct {
	alerts   Alerts
	alertsCh chan Alerts
	doneCh   chan bool
}

// NewAlerter creates a new Alerter and starts its internal goroutine for collecting alerts.
func NewAlerter() *Alerter {
	var alerter = &Alerter{
		alerts:   make(Alerts),
		alertsCh: make(chan Alerts),
		doneCh:   make(chan bool),
	}

	go alerter.run()

	return alerter
}

func (a *Alerter) run() {
	defer func() { a.doneCh <- true }()
	for alert := range a.alertsCh {
		for k, v := range alert {
			if val, ok := a.alerts[k]; ok {
				a.alerts[k] = append(val, v...)
			} else {
				a.alerts[k] = v
			}
		}
	}
}

// SetAlerts replaces the current alert map with the provided one.
func (a *Alerter) SetAlerts(alerts Alerts) {
	a.alerts = alerts
}

// Retrieve closes the alert channel, waits for the internal goroutine to finish, and returns collected alerts.
func (a *Alerter) Retrieve() Alerts {
	close(a.alertsCh)
	<-a.doneCh
	return a.alerts
}

// SendAlert sends an alert for the given resource key to the internal channel.
func (a *Alerter) SendAlert(key string, alert Alert) {
	a.alertsCh <- Alerts{
		key: []Alert{alert},
	}
}

// IsResourceIgnored reports whether any alert for the given resource has the ignore flag set.
func (a *Alerter) IsResourceIgnored(res *resource.Resource) bool {
	alert, alertExists := a.alerts[fmt.Sprintf("%s.%s", res.ResourceType(), res.ResourceID())]
	wildcardAlert, wildcardAlertExists := a.alerts[res.ResourceType()]
	shouldIgnoreAlert := a.shouldBeIgnored(alert)
	shouldIgnoreWildcardAlert := a.shouldBeIgnored(wildcardAlert)
	return (alertExists && shouldIgnoreAlert) || (wildcardAlertExists && shouldIgnoreWildcardAlert)
}

func (a *Alerter) shouldBeIgnored(alert []Alert) bool {
	for _, a := range alert {
		if a.ShouldIgnoreResource() {
			return true
		}
	}
	return false
}
