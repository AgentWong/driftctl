package analyser

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/snyk/driftctl/enumeration/alerter"

	"github.com/snyk/driftctl/enumeration/resource"
)

type AttributeChange struct {
	Path   string      `json:"path"`
	Before interface{} `json:"before"`
	After  interface{} `json:"after"`
}

type DriftedResource struct {
	Res              *resource.Resource
	AttributeChanges []AttributeChange
}

type Summary struct {
	TotalResources             int  `json:"total_resources"`
	TotalUnmanaged             int  `json:"total_unmanaged"`
	TotalDeleted               int  `json:"total_missing"`
	TotalManaged               int  `json:"total_managed"`
	TotalDrifted               int  `json:"total_drifted"`
	TotalUnsupported           int  `json:"total_unsupported"`
	TotalCloudFormationManaged int  `json:"total_cloudformation_managed"`
	TotalIaCSourceCount        uint `json:"total_iac_source_count"`
}

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

type GenDriftIgnoreOptions struct {
	ExcludeUnmanaged bool
	ExcludeDeleted   bool
	ExcludeDrifted   bool
	InputPath        string
	OutputPath       string
}

func NewAnalysis() *Analysis {
	return &Analysis{}
}

func (a Analysis) MarshalJSON() ([]byte, error) {
	bla := serializableAnalysis{}
	for _, m := range a.managed {
		bla.Managed = append(bla.Managed, *resource.NewSerializableResource(m))
	}
	for _, u := range a.unmanaged {
		sr := *resource.NewSerializableResource(u)
		if a.unmanagedCategories != nil {
			key := u.ResourceType() + "." + u.ResourceId()
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

func (a *Analysis) UnmarshalJSON(bytes []byte) error {
	bla := serializableAnalysis{}
	if err := json.Unmarshal(bytes, &bla); err != nil {
		return err
	}
	for _, u := range bla.Unmanaged {
		a.AddUnmanaged(&resource.Resource{
			Id:   u.Id,
			Type: u.Type,
		})
	}
	for _, d := range bla.Deleted {
		a.AddDeleted(&resource.Resource{
			Id:   d.Id,
			Type: d.Type,
		})
	}
	for _, u := range bla.Unsupported {
		a.unsupported = append(a.unsupported, &resource.Resource{
			Id:   u.Id,
			Type: u.Type,
		})
		a.summary.TotalUnsupported++
		a.summary.TotalResources++
	}
	for _, m := range bla.Managed {
		res := &resource.Resource{
			Id:   m.Id,
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
	a.Duration = time.Duration(bla.ScanDuration) * time.Second
	a.Date = bla.Date
	return nil
}

func (a *Analysis) IsSync() bool {
	return a.summary.TotalUnmanaged == 0 && a.summary.TotalDeleted == 0 && a.summary.TotalDrifted == 0
}

func (a *Analysis) AddDeleted(resources ...*resource.Resource) {
	a.deleted = append(a.deleted, resources...)
	a.summary.TotalResources += len(resources)
	a.summary.TotalDeleted += len(resources)
}

func (a *Analysis) AddUnmanaged(resources ...*resource.Resource) {
	a.unmanaged = append(a.unmanaged, resources...)
	a.summary.TotalResources += len(resources)
	a.summary.TotalUnmanaged += len(resources)
}

func (a *Analysis) AddManaged(resources ...*resource.Resource) {
	a.managed = append(a.managed, resources...)
	a.summary.TotalResources += len(resources)
	a.summary.TotalManaged += len(resources)
}

func (a *Analysis) SetAlerts(alerts alerter.Alerts) {
	a.alerts = alerts
}

func (a *Analysis) SetIaCSourceCount(i uint) {
	a.summary.TotalIaCSourceCount = i
}

func (a *Analysis) Coverage() int {
	if a.summary.TotalResources > 0 {
		return int((float32(a.summary.TotalManaged) / float32(a.summary.TotalResources)) * 100.0)
	}
	return 0
}

func (a *Analysis) Managed() []*resource.Resource {
	return a.managed
}

func (a *Analysis) Unmanaged() []*resource.Resource {
	return a.unmanaged
}

func (a *Analysis) Deleted() []*resource.Resource {
	return a.deleted
}

func (a *Analysis) Unsupported() []*resource.Resource {
	return a.unsupported
}

func (a *Analysis) Drifted() []*DriftedResource {
	return a.drifted
}

func (a *Analysis) AddDrifted(d *DriftedResource) {
	a.drifted = append(a.drifted, d)
	a.summary.TotalResources++
	a.summary.TotalDrifted++
}

func (a *Analysis) Summary() Summary {
	return a.summary
}

func (a *Analysis) Alerts() alerter.Alerts {
	return a.alerts
}

func (a *Analysis) HasCategories() bool {
	return a.unmanagedCategories != nil
}

func (a *Analysis) UnmanagedCategory(r *resource.Resource) string {
	if a.unmanagedCategories == nil {
		return ""
	}
	return a.unmanagedCategories[r.ResourceType()+"."+r.ResourceId()]
}

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

func (a *Analysis) FilterUnmanagedByCategory(excludeCategories map[string]bool) {
	if a.unmanagedCategories == nil {
		return
	}
	var filtered []*resource.Resource
	removed := 0
	for _, r := range a.unmanaged {
		key := r.ResourceType() + "." + r.ResourceId()
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

func (a *Analysis) SortResources() {
	a.unmanaged = resource.Sort(a.unmanaged)
	a.deleted = resource.Sort(a.deleted)
}

func (a *Analysis) DriftIgnoreList(opts GenDriftIgnoreOptions) (int, string) {
	var list []string

	resourceCount := 0

	addResources := func(res ...*resource.Resource) {
		for _, r := range res {
			list = append(list, fmt.Sprintf("%s.%s", r.ResourceType(), escapeKey(r.ResourceId())))
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
