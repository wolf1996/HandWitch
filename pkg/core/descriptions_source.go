package core

import "errors"

type DescriptionsSource interface {
	GetByName(string) (*UrlRecord, error)
}

type SimpleDescriptionsSource struct {
	descriptions map[string]UrlRecord
}

var (
	ErrNonExistentHand = errors.New("Can't Find key")
)

func (source *SimpleDescriptionsSource) GetByName(name string) (*UrlRecord, error) {
	value, ok := source.descriptions[name]
	if !ok {
		return nil, ErrNonExistentHand
	}
	return &value, nil
}

func NewDescriptionSourceFromDict(urlsDict map[string]UrlRecord) DescriptionsSource {
	return &SimpleDescriptionsSource{
		descriptions: urlsDict,
	}
}
