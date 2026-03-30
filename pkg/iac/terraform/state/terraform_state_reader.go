package state

import (
	"fmt"
	"strings"

	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/terraform"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/states"
	"github.com/hashicorp/terraform/states/statefile"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/pkg/filter"
	"github.com/snyk/driftctl/pkg/iac"
	"github.com/snyk/driftctl/pkg/output"
	"github.com/zclconf/go-cty/cty"
	ctyconvert "github.com/zclconf/go-cty/cty/convert"
	ctyjson "github.com/zclconf/go-cty/cty/json"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/iac/config"
	"github.com/snyk/driftctl/pkg/iac/terraform/state/backend"
	"github.com/snyk/driftctl/pkg/iac/terraform/state/enumerator"
	resdriftctl "github.com/snyk/driftctl/pkg/resource"
)

// TerraformStateReaderSupplier is the supplier key for Terraform state reading.
const TerraformStateReaderSupplier = "tfstate"

type decodedRes struct {
	source resource.Source
	val    cty.Value
}

// TerraformStateReader represents a TerraformStateReader.
type TerraformStateReader struct {
	library        *terraform.ProviderLibrary
	config         config.SupplierConfig
	backend        backend.Backend
	enumerator     enumerator.StateEnumerator
	deserializer   *resource.Deserializer
	backendOptions *backend.Options
	progress       output.Progress
	filter         filter.Filter
	alerter        *alerter.Alerter
	sourceCount    uint
}

func (r *TerraformStateReader) initReader() error {
	enumerator, err := enumerator.GetEnumerator(r.config, r.backendOptions)
	if err != nil {
		return err
	}
	r.enumerator = enumerator
	return nil
}

// NewReader creates a new instance.
func NewReader(config config.SupplierConfig, library *terraform.ProviderLibrary, backendOpts *backend.Options, progress output.Progress, alerter *alerter.Alerter, deserializer *resource.Deserializer, filter filter.Filter) (*TerraformStateReader, error) {
	reader := TerraformStateReader{
		library:        library,
		config:         config,
		deserializer:   deserializer,
		backendOptions: backendOpts,
		progress:       progress,
		alerter:        alerter,
		filter:         filter,
		sourceCount:    0,
	}
	err := reader.initReader()
	if err != nil {
		return nil, err
	}
	return &reader, nil
}

func (r *TerraformStateReader) retrieve() (map[string][]decodedRes, error) {
	b, err := backend.GetBackend(r.config, r.backendOptions)
	if err != nil {
		return nil, err
	}
	r.backend = b

	state, err := read(r.config.Path, r.backend)
	defer func() { _ = r.backend.Close() }()
	if err != nil {
		return nil, err
	}

	resMap := make(map[string][]decodedRes)
	for moduleName, module := range state.Modules {
		logrus.WithFields(logrus.Fields{
			"module":        moduleName,
			"resourceCount": fmt.Sprintf("%d", len(module.Resources)),
		}).Debug("Found module in state")
		for _, stateRes := range module.Resources {
			resName := stateRes.Addr.Resource.Name
			resType := stateRes.Addr.Resource.Type

			if !resdriftctl.IsResourceTypeSupported(resType) {
				logrus.WithFields(logrus.Fields{
					"name": resName,
					"type": resType,
				}).Debug("Ignored unsupported resource from state")
				continue
			}

			if r.filter != nil && r.filter.IsTypeIgnored(resource.Type(resType)) {
				logrus.WithFields(logrus.Fields{
					"name": resName,
					"type": resType,
				}).Debug("Ignored resource from state since it is ignored in filter")
				continue
			}

			if stateRes.Addr.Resource.Mode != addrs.ManagedResourceMode {
				logrus.WithFields(logrus.Fields{
					"mode": stateRes.Addr.Resource.Mode,
					"name": resName,
					"type": resType,
				}).Debug("Skipping state entry as it is not a managed resource")
				continue
			}
			providerType := stateRes.ProviderConfig.Provider.Type
			provider := r.library.Provider(providerType)
			if provider == nil {
				logrus.WithFields(logrus.Fields{
					"providerKey": providerType,
				}).Debug("Unsupported provider found in state")
				continue
			}
			schema := provider.Schema()[stateRes.Addr.Resource.Type]
			for _, instance := range stateRes.Instances {
				decodedVal, err := instance.Current.Decode(schema.Block.ImpliedType())
				if err != nil {
					// Try to do a manual type conversion if we got a path error
					// It will allow driftctl to read state generated with a superior version of provider
					// than the actually supported one
					// by ignoring new fields
					_, isPathError := err.(cty.PathError)
					if isPathError {
						logrus.WithFields(logrus.Fields{
							"name": resName,
							"type": resType,
							"err":  err.Error(),
						}).Debug("Got a cty path error when deserializing state")

						decodedVal, err = r.convertInstance(instance.Current, schema.Block.ImpliedType())
					}

					if err != nil {
						logrus.WithFields(logrus.Fields{
							"name": resName,
							"type": resType,
						}).Error("Unable to decode resource from state")
						return nil, err
					}
				}
				_, exists := resMap[stateRes.Addr.Resource.Type]
				val := decodedRes{
					source: resource.NewTerraformStateSource(r.config.String(), moduleName, resName),
					val:    decodedVal.Value,
				}
				if !exists {
					resMap[stateRes.Addr.Resource.Type] = []decodedRes{val}
				} else {
					resMap[stateRes.Addr.Resource.Type] = append(resMap[stateRes.Addr.Resource.Type], val)
				}
			}
		}
	}

	return resMap, nil
}

func (r *TerraformStateReader) convertInstance(instance *states.ResourceInstanceObjectSrc, ty cty.Type) (*states.ResourceInstanceObject, error) {
	inputType, err := ctyjson.ImpliedType(instance.AttrsJSON)
	if err != nil {
		return nil, err
	}
	input, err := ctyjson.Unmarshal(instance.AttrsJSON, inputType)
	if err != nil {
		return nil, err
	}

	// pad missing attributes with typed nulls so conversion succeeds when
	// state was written by a newer provider that dropped attributes
	input = padMissingAttributes(input, ty)

	convertedVal, err := ctyconvert.Convert(input, ty)
	if err != nil {
		// per-attribute fallback: convert each attribute individually,
		// using a typed null for any that fail (e.g. type mismatches
		// between provider versions)
		convertedVal, err = convertAttributeByAttribute(input, ty)
		if err != nil {
			return nil, err
		}
	}

	instanceObj := &states.ResourceInstanceObject{
		Value:               convertedVal,
		Status:              instance.Status,
		Dependencies:        instance.Dependencies,
		Private:             instance.Private,
		CreateBeforeDestroy: instance.CreateBeforeDestroy,
	}

	logrus.Debug("Successfully converted resource")

	return instanceObj, nil
}

// padMissingAttributes adds typed null values for any object attributes that
// exist in the target type but are absent from the input value. This handles
// state files written by a newer provider version that removed attributes
// still present in the older schema driftctl uses.
func padMissingAttributes(input cty.Value, targetType cty.Type) cty.Value {
	if !targetType.IsObjectType() || !input.Type().IsObjectType() {
		return input
	}

	inputAttrs := input.Type().AttributeTypes()
	targetAttrs := targetType.AttributeTypes()

	needsPadding := false
	for name := range targetAttrs {
		if _, ok := inputAttrs[name]; !ok {
			needsPadding = true
			break
		}
	}
	if !needsPadding {
		return input
	}

	vals := make(map[string]cty.Value, len(targetAttrs))
	for name, attrType := range targetAttrs {
		if _, ok := inputAttrs[name]; ok {
			vals[name] = input.GetAttr(name)
		} else {
			vals[name] = cty.NullVal(attrType)
		}
	}
	// preserve input attributes not in the target (handled by convert later)
	for name := range inputAttrs {
		if _, ok := targetAttrs[name]; !ok {
			vals[name] = input.GetAttr(name)
		}
	}

	return cty.ObjectVal(vals)
}

// convertAttributeByAttribute builds a target-typed object by converting each
// attribute individually, falling back to a typed null when an attribute's type
// is incompatible (e.g. a string in state where the schema expects a number).
func convertAttributeByAttribute(input cty.Value, ty cty.Type) (cty.Value, error) {
	if !ty.IsObjectType() {
		return cty.NilVal, fmt.Errorf("target type is not an object")
	}

	targetAttrs := ty.AttributeTypes()
	vals := make(map[string]cty.Value, len(targetAttrs))

	for name, attrType := range targetAttrs {
		if !input.Type().HasAttribute(name) {
			vals[name] = cty.NullVal(attrType)
			continue
		}
		srcVal := input.GetAttr(name)
		converted, err := ctyconvert.Convert(srcVal, attrType)
		if err != nil {
			vals[name] = cty.NullVal(attrType)
			continue
		}
		vals[name] = converted
	}

	return cty.ObjectVal(vals), nil
}

func (r *TerraformStateReader) decode(valFromState map[string][]decodedRes) ([]*resource.Resource, error) {
	results := make([]*resource.Resource, 0)

	for ty, val := range valFromState {
		for _, stateVal := range val {
			res, err := r.deserializer.DeserializeOne(ty, stateVal.val)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"type":  ty,
					"name":  stateVal.source.InternalName(),
					"state": stateVal.source.Source(),
				}).Warnf("Could not read from state: %+v", err)
				continue
			}
			res.Source = stateVal.source
			results = append(results, res)
		}
	}

	return results, nil
}

// Resources implements the TerraformStateReader interface.
func (r *TerraformStateReader) Resources() ([]*resource.Resource, error) {
	if r.enumerator == nil {
		return r.retrieveForState(r.config.Path)
	}

	return r.retrieveMultiplesStates()
}

// SourceCount implements the TerraformStateReader interface.
func (r *TerraformStateReader) SourceCount() uint {
	return r.sourceCount
}

func (r *TerraformStateReader) retrieveForState(path string) ([]*resource.Resource, error) {
	r.config.Path = path
	r.sourceCount++
	logrus.WithFields(logrus.Fields{
		"path":    r.config.Path,
		"backend": r.config.Backend,
	}).Debug("Reading resources from state")
	r.progress.Inc()
	values, err := r.retrieve()
	if err != nil {
		return nil, errors.Wrap(err, r.config.String())
	}
	decode, err := r.decode(values)
	return decode, errors.Wrap(err, r.config.String())
}

func (r *TerraformStateReader) retrieveMultiplesStates() ([]*resource.Resource, error) {
	keys, err := r.enumerator.Enumerate()
	if err != nil {
		r.alerter.SendAlert("", NewReadingAlert(r.enumerator.Origin(), err))
		return nil, errors.Wrap(err, r.config.String())
	}

	logrus.WithFields(logrus.Fields{
		"keys": keys,
	}).Debug("Enumerated keys")

	results := make([]*resource.Resource, 0)
	isSuccess := false
	readingError := iac.NewStateReadingError()

	for _, key := range keys {
		resources, err := r.retrieveForState(key)
		if err != nil {
			readingError.Add(err)
			r.alerter.SendAlert("", NewReadingAlert(key, err))
			continue
		}
		isSuccess = true
		results = append(results, resources...)
	}

	if !isSuccess {
		// all key failed, throw an error
		return results, readingError
	}

	return results, nil
}

func read(path string, reader backend.Backend) (*states.State, error) {
	state, err := readState(path, reader)
	if err != nil {
		if _, ok := reader.(*backend.HTTPBackend); ok && strings.Contains(err.Error(), "The state file could not be parsed as JSON") {
			return nil, errors.Errorf("given url is not a valid state file")
		}
		return nil, err
	}
	return state, nil
}

func readState(path string, reader backend.Backend) (*states.State, error) {
	state, err := statefile.Read(reader)
	if err != nil {
		return nil, err
	}

	supported, err := IsVersionSupported(state.TerraformVersion.String())
	if err != nil {
		return nil, err
	}

	if !supported {
		return nil, &UnsupportedVersionError{
			StateFile: path,
			Version:   state.TerraformVersion,
		}
	}

	return state.State, nil
}
