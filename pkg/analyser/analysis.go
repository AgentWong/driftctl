// Package analyser provides drift analysis and reporting.
package analyser

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/snyk/driftctl/enumeration/alerter"

	"github.com/snyk/driftctl/enumeration/resource"
)

// AttributeChange records a single attribute difference between IaC and remote.
type AttributeChange struct {
	Path   string      `json:"path"`
	Before interface{} `json:"before"`
	After  interface{} `json:"after"`
}

// DriftedResource pairs a resource with its detected attribute changes.
type DriftedResource struct {
	Res              *resource.Resource
	AttributeChanges []AttributeChange
}

// Summary holds aggregate counts from a drift analysis.
type Summary struct {
	TotalResources             int  `json:"total_resources"`
	TotalUnmanaged             int  `json:"total_unmanaged"`
	TotalDeleted               int  `json:"total_missing"`
	TotalManaged               int  `json:"total_managed"`
	TotalDrifted               int  `json:"total_drifted"`
	TotalUnsupported           int  `json:"total_unsupported"`
	TotalCloudFormationManaged int  `json:"total_cloudformation_managed"`
	TotalDefaultResources      int  `json:"total_default_resources"`
	TotalIaCSourceCount        uint `json:"total_iac_source_count"`
}

// Analysis holds the full result of a drift scan.
type Analysis struct {
	unmanaged           []*resource.Resource
	managed             []*resource.Resource
	deleted             []*resource.Resource
	unsupported         []*resource.Resource
	drifted             []*DriftedResource
	summary             Summary
	alerts              alerter.Alerts
	unmanagedCategories map[string]string
	Duration            time.Duration
	Date                time.Time
	ProviderName        string
	ProviderVersion     string
}

type serializableDriftedResource struct {
	Res              resource.SerializableResource `json:"resource"`
	AttributeChanges []AttributeChange             `json:"attribute_changes,omitempty"`
}

type serializableAnalysis struct {
	Summary         Summary                                `json:"summary"`
	Managed         []resource.SerializableResource        `json:"managed"`
	Unmanaged       []resource.SerializableResource        `json:"unmanaged"`
	Deleted         []resource.SerializableResource        `json:"missing"`
	Unsupported     []resource.SerializableResource        `json:"unsupported,omitempty"`
	Drifted         []serializableDriftedResource          `json:"drifted"`
	Coverage        int                                    `json:"coverage"`
	Alerts          map[string][]alerter.SerializableAlert `json:"alerts"`
	ProviderName    string                                 `json:"provider_name"`
	ProviderVersion string                                 `json:"provider_version"`
	ScanDuration    uint                                   `json:"scan_duration,omitempty"`
	Date            time.Time                              `json:"date"`
}

// GenDriftIgnoreOptions configures the generation of driftignore patterns.
type GenDriftIgnoreOptions struct {
	ExcludeUnmanaged bool
	ExcludeDeleted   bool
	ExcludeDrifted   bool
	InputPath        string
	OutputPath       string
}

// NewAnalysis creates an empty Analysis.
func NewAnalysis() *Analysis {
	return &Analysis{}
}

// MarshalJSON serializes the analysis to JSON.
func (a Analysis) MarshalJSON() ([]byte, error) {
	bla := serializableAnalysis{}
	for _, m := range a.managed {
		bla.Managed = append(bla.Managed, *resource.NewSerializableResource(m))
	}
	for _, u := range a.unmanaged {
		sr := *resource.NewSerializableResource(u)
		if a.unmanagedCategories != nil {
			key := u.ResourceType() + "." + u.ResourceID()
			if cat, ok := a.unmanagedCategories[key]; ok {
				sr.Category = cat
			}
		}
		bla.Unmanaged = append(bla.Unmanaged, sr)
	}
	for _, d := range a.deleted {
		bla.Deleted = append(bla.Deleted, *resource.NewSerializableResource(d))
	}
	for _, u := range a.unsupported {
		bla.Unsupported = append(bla.Unsupported, *resource.NewSerializableResource(u))
	}
	for _, dr := range a.drifted {
		bla.Drifted = append(bla.Drifted, serializableDriftedResource{
			Res:              *resource.NewSerializableResource(dr.Res),
			AttributeChanges: dr.AttributeChanges,
		})
	}
	if len(a.alerts) > 0 {
		bla.Alerts = make(map[string][]alerter.SerializableAlert)
		for k, v := range a.alerts {
			for _, al := range v {
				bla.Alerts[k] = append(bla.Alerts[k], alerter.SerializableAlert{Alert: al})
			}
		}
	}
	bla.Summary = a.summary
	bla.Coverage = a.Coverage()
	bla.ProviderName = a.ProviderName
	bla.ProviderVersion = a.ProviderVersion
	bla.ScanDuration = uint(a.Duration.Seconds())
	bla.Date = a.Date

	return json.Marshal(bla)
}

// UnmarshalJSON deserializes an analysis from JSON.
func (a *Analysis) UnmarshalJSON(bytes []byte) error {
	bla := serializableAnalysis{}
	if err := json.Unmarshal(bytes, &bla); err != nil {
		return err
	}
	for _, u := range bla.Unmanaged {
		a.AddUnmanaged(&resource.Resource{
			ID:   u.ID,
			Type: u.Type,
		})
	}
	for _, d := range bla.Deleted {
		a.AddDeleted(&resource.Resource{
			ID:   d.ID,
			Type: d.Type,
		})
	}
	for _, u := range bla.Unsupported {
		a.unsupported = append(a.unsupported, &resource.Resource{
			ID:   u.ID,
			Type: u.Type,
		})
		a.summary.TotalUnsupported++
		a.summary.TotalResources++
	}
	for _, m := range bla.Managed {
		res := &resource.Resource{
			ID:   m.ID,
			Type: m.Type,
		}
		if m.Source != nil {
			// We loose the source type in the serialization process, for now everything is serialized back to a
			// TerraformStateSource.
			// TODO: Add a discriminator field to be able to serialize back to the right type
			// when we'll introduce a new source type
			res.Source = &resource.TerraformStateSource{
				State:  m.Source.S,
				Module: m.Source.Ns,
				Name:   m.Source.Name,
			}
		}
		a.AddManaged(res)
	}
	if len(bla.Alerts) > 0 {
		a.alerts = make(alerter.Alerts)
		for k, v := range bla.Alerts {
			for _, al := range v {
				a.alerts[k] = append(a.alerts[k], &alerter.SerializedAlert{
					Msg: al.Message(),
				})
			}
		}
	}
	a.ProviderName = bla.ProviderName
	a.ProviderVersion = bla.ProviderVersion
	a.SetIaCSourceCount(bla.Summary.TotalIaCSourceCount)
	a.Duration = time.Duration(bla.ScanDuration) * time.Second //nolint:gosec // G115: ScanDuration is always a small positive integer
	a.Date = bla.Date
	return nil
}

// IsSync reports whether infrastructure is in sync (no drift).
func (a *Analysis) IsSync() bool {
	return a.summary.TotalUnmanaged == 0 && a.summary.TotalDeleted == 0 && a.summary.TotalDrifted == 0
}

// AddDeleted records resources found in IaC but missing from the cloud.
func (a *Analysis) AddDeleted(resources ...*resource.Resource) {
	a.deleted = append(a.deleted, resources...)
	a.summary.TotalResources += len(resources)
	a.summary.TotalDeleted += len(resources)
}

// AddUnmanaged records resources found in the cloud but not in IaC.
func (a *Analysis) AddUnmanaged(resources ...*resource.Resource) {
	a.unmanaged = append(a.unmanaged, resources...)
	a.summary.TotalResources += len(resources)
	a.summary.TotalUnmanaged += len(resources)
}

// AddManaged records resources present in both IaC and the cloud.
func (a *Analysis) AddManaged(resources ...*resource.Resource) {
	a.managed = append(a.managed, resources...)
	a.summary.TotalResources += len(resources)
	a.summary.TotalManaged += len(resources)
}

// SetAlerts sets the alerts for this analysis.
func (a *Analysis) SetAlerts(alerts alerter.Alerts) {
	a.alerts = alerts
}

// SetIaCSourceCount records the number of IaC source files.
func (a *Analysis) SetIaCSourceCount(i uint) {
	a.summary.TotalIaCSourceCount = i
}

// Coverage returns the percentage of managed resources.
func (a *Analysis) Coverage() int {
	if a.summary.TotalResources > 0 {
		return int((float32(a.summary.TotalManaged) / float32(a.summary.TotalResources)) * 100.0)
	}
	return 0
}

// Managed returns the managed resources.
func (a *Analysis) Managed() []*resource.Resource {
	return a.managed
}

// Unmanaged returns the unmanaged resources.
func (a *Analysis) Unmanaged() []*resource.Resource {
	return a.unmanaged
}

// Deleted returns the deleted resources.
func (a *Analysis) Deleted() []*resource.Resource {
	return a.deleted
}

// Unsupported returns the unsupported resources.
func (a *Analysis) Unsupported() []*resource.Resource {
	return a.unsupported
}

// Drifted returns the drifted resources.
func (a *Analysis) Drifted() []*DriftedResource {
	return a.drifted
}

// AddDrifted records a resource with detected drift.
func (a *Analysis) AddDrifted(d *DriftedResource) {
	a.drifted = append(a.drifted, d)
	a.summary.TotalResources++
	a.summary.TotalDrifted++
}

// Summary returns the aggregate analysis summary.
func (a *Analysis) Summary() Summary {
	return a.summary
}

// Alerts returns the alerts for this analysis.
func (a *Analysis) Alerts() alerter.Alerts {
	return a.alerts
}

// HasCategories reports whether unmanaged resource categories are set.
func (a *Analysis) HasCategories() bool {
	return a.unmanagedCategories != nil
}

// UnmanagedCategory returns the category label for an unmanaged resource.
func (a *Analysis) UnmanagedCategory(r *resource.Resource) string {
	if a.unmanagedCategories == nil {
		return ""
	}
	return a.unmanagedCategories[r.ResourceType()+"."+r.ResourceID()]
}

// SetUnmanagedCategories sets the category labels for unmanaged resources.
func (a *Analysis) SetUnmanagedCategories(cats map[string]string) {
	a.unmanagedCategories = cats
}

// ReclassifyMissingAsUnsupported moves deleted resources whose Terraform type
// is not discoverable by AWS Config into a separate "unsupported" bucket.
func (a *Analysis) ReclassifyMissingAsUnsupported(supportedTypes map[string]bool) {
	var trueMissing []*resource.Resource
	for _, r := range a.deleted {
		if supportedTypes[r.ResourceType()] {
			trueMissing = append(trueMissing, r)
		} else {
			a.unsupported = append(a.unsupported, r)
		}
	}
	moved := len(a.deleted) - len(trueMissing)
	a.deleted = trueMissing
	a.summary.TotalDeleted -= moved
	a.summary.TotalUnsupported += moved
}

// AdjustSummaryForCloudFormation shifts cfnCount resources from the unmanaged
// total into the managed total so CloudFormation-managed resources count as IaC.
func (a *Analysis) AdjustSummaryForCloudFormation(cfnCount int) {
	a.summary.TotalCloudFormationManaged = cfnCount
	a.summary.TotalManaged += cfnCount
	a.summary.TotalUnmanaged -= cfnCount
}

// AdjustSummaryForDefaultResources removes default resources from the unmanaged
// total because they are auto-created by AWS, not user-managed drift.
func (a *Analysis) AdjustSummaryForDefaultResources(count int) {
	a.summary.TotalDefaultResources = count
	a.summary.TotalUnmanaged -= count
	a.summary.TotalResources -= count
}

// FilterUnmanagedByCategory removes unmanaged resources whose category is in the exclude set.
func (a *Analysis) FilterUnmanagedByCategory(excludeCategories map[string]bool) {
	if a.unmanagedCategories == nil {
		return
	}
	var filtered []*resource.Resource
	removed := 0
	for _, r := range a.unmanaged {
		key := r.ResourceType() + "." + r.ResourceID()
		cat := a.unmanagedCategories[key]
		if !excludeCategories[cat] {
			filtered = append(filtered, r)
		} else {
			removed++
		}
	}
	a.unmanaged = filtered
	a.summary.TotalUnmanaged -= removed
	a.summary.TotalResources -= removed
}

// SortResources sorts unmanaged and deleted resources.
func (a *Analysis) SortResources() {
	a.unmanaged = resource.Sort(a.unmanaged)
	a.deleted = resource.Sort(a.deleted)
}

// Merge combines another Analysis into this one, accumulating all resource lists
// and recalculating the summary. Duration and Date are taken from the receiver.
func (a *Analysis) Merge(other *Analysis) {
	if other == nil {
		return
	}
	a.AddManaged(other.managed...)
	a.AddUnmanaged(other.unmanaged...)
	a.AddDeleted(other.deleted...)
	for _, d := range other.drifted {
		a.AddDrifted(d)
	}
	// merge unsupported (no public adder, so append directly)
	a.unsupported = append(a.unsupported, other.unsupported...)
	a.summary.TotalUnsupported += other.summary.TotalUnsupported
}

// DriftIgnoreList builds a .driftignore file from the analysis results.
func (a *Analysis) DriftIgnoreList(opts GenDriftIgnoreOptions) (int, string) {
	var list []string

	resourceCount := 0

	addResources := func(res ...*resource.Resource) {
		for _, r := range res {
			list = append(list, fmt.Sprintf("%s.%s", r.ResourceType(), escapeKey(r.ResourceID())))
		}
		resourceCount += len(res)
	}

	if !opts.ExcludeUnmanaged && a.Summary().TotalUnmanaged > 0 {
		list = append(list, "# Resources not covered by IaC")
		addResources(a.Unmanaged()...)
	}
	if !opts.ExcludeDeleted && a.Summary().TotalDeleted > 0 {
		list = append(list, "# Missing resources")
		addResources(a.Deleted()...)
	}

	return resourceCount, strings.Join(list, "\n")
}

func escapeKey(line string) string {
	line = strings.ReplaceAll(line, `\`, `\\`)
	line = strings.ReplaceAll(line, `.`, `\.`)

	return line
}
