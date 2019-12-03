package core

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

// ReadJSON получить из json описание ручек
func ReadJSON(reader io.Reader) (*SimpleDescriptionsSource, error) {
	var urlContainer URLContrainer
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &urlContainer)
	if err != nil {
		return nil, err
	}
	return NewDescriptionSourceFromDict(urlContainer), nil
}
