package core

import (
	"errors"
	"sort"
)

//DescriptionsSource DescriptionsSource get URL record by name
type DescriptionsSource interface {
	GetByName(string) (*URLRecord, error)
	GetAllRecords() ([]URLRecord, error)
}

//SimpleDescriptionsSource implementation of DescriptionsSource with
// mp as an data source
type SimpleDescriptionsSource struct {
	descriptions map[string]URLRecord
}

var (
	//ErrNonExistentHand hand with specified parameters
	// doesn't exists
	ErrNonExistentHand = errors.New("Can't Find key")
)

//GetByName get url data by name
func (source *SimpleDescriptionsSource) GetByName(name string) (*URLRecord, error) {
	value, ok := source.descriptions[name]
	if !ok {
		return nil, ErrNonExistentHand
	}
	return &value, nil
}

//GetAllRecords get url data by name
func (source *SimpleDescriptionsSource) GetAllRecords() ([]URLRecord, error) {
	result := make([]URLRecord, 0, len(source.descriptions))
	keys := make([]string, 0, len(source.descriptions))

	for key := range source.descriptions {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		result = append(result, source.descriptions[key])
	}

	return result, nil
}

//NewDescriptionSourceFromDict get simple descriprion source with data from map
func NewDescriptionSourceFromDict(URLsDict map[string]URLRecord) *SimpleDescriptionsSource {
	return &SimpleDescriptionsSource{
		descriptions: URLsDict,
	}
}
