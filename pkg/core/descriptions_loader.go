package core

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

// GetDescriptionSourceFromJSON получить из json описание ручек
func GetDescriptionSourceFromJSON(reader io.Reader) (*SimpleDescriptionsSource, error) {
	var urlContainer URLContrainer
	// TODO: возможно стоит переделать на работу парсера, чтобы не вычитывать весь файл
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
