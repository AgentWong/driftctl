// Package plan provides types and utilities for parsing and running Terraform plan files.
package plan

import (
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
)

// Action represents the type of drift detected from a terraform plan.
type Action string

// Action constants represent the possible drift actions for a Terraform resource.
const (
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionNoOp   Action = "no-op"
	ActionRead   Action = "read"
)

// AttributeChange represents a single attribute that has drifted.
type AttributeChange struct {
	Path   string      `json:"path"`
	Before interface{} `json:"before"`
	After  interface{} `json:"after"`
}

// DriftResult represents the drift status of a single resource.
type DriftResult struct {
	Type             string
	ID               string
	Address          string
	Action           Action
	AttributeChanges []AttributeChange
}

// ParsePlan extracts DriftResults from a terraform plan, including
// resources in both root and child modules.
func ParsePlan(p *tfjson.Plan) ([]DriftResult, error) {
	var results []DriftResult

	if p.ResourceChanges == nil {
		return results, nil
	}

	for _, rc := range p.ResourceChanges {
		if rc.Change == nil {
			continue
		}

		result := DriftResult{
			Type:    rc.Type,
			Address: rc.Address,
		}

		if rc.Change.Before != nil {
			if beforeMap, ok := rc.Change.Before.(map[string]interface{}); ok {
				if id, ok := beforeMap["id"].(string); ok {
					result.ID = id
				}
			}
		}

		result.Action = mapActions(rc.Change.Actions)

		if result.Action == ActionUpdate && rc.Change != nil {
			result.AttributeChanges = extractAttributeChanges(rc.Change)
		}

		results = append(results, result)
	}

	return results, nil
}

// mapActions converts tfjson.Actions into our Action type.
// terraform-json Actions is a slice; e.g. ["update"], ["delete","create"] for replace.
func mapActions(actions tfjson.Actions) Action {
	if actions.NoOp() {
		return ActionNoOp
	}
	if actions.Read() {
		return ActionRead
	}
	if actions.Update() {
		return ActionUpdate
	}
	if actions.Create() && actions.Delete() {
		// replace counts as an update (resource is recreated due to drift)
		return ActionUpdate
	}
	if actions.Delete() {
		return ActionDelete
	}
	if actions.Create() {
		return ActionCreate
	}
	return ActionNoOp
}

// extractAttributeChanges diffs Before/After maps to find changed attributes.
func extractAttributeChanges(change *tfjson.Change) []AttributeChange {
	beforeMap, beforeOk := change.Before.(map[string]interface{})
	afterMap, afterOk := change.After.(map[string]interface{})
	if !beforeOk || !afterOk {
		return nil
	}

	var changes []AttributeChange

	// collect all keys from both maps
	keys := make(map[string]struct{})
	for k := range beforeMap {
		keys[k] = struct{}{}
	}
	for k := range afterMap {
		keys[k] = struct{}{}
	}

	for k := range keys {
		bv := beforeMap[k]
		av := afterMap[k]
		if fmt.Sprintf("%v", bv) != fmt.Sprintf("%v", av) {
			changes = append(changes, AttributeChange{
				Path:   k,
				Before: bv,
				After:  av,
			})
		}
	}

	return changes
}
