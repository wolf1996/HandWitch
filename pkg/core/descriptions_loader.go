package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

// ValidationError error reporting about
type ValidationError struct {
	Field        string
	WrappedError []error
}

func (val *ValidationError) Error() string {
	if len(val.WrappedError) == 0 {
		panic(fmt.Sprintf("Empty error requesting! %v", val))
	}
	var builder strings.Builder
	for _, err := range val.WrappedError {
		builder.WriteString(err.Error() + "\n")
	}
	return fmt.Sprintf("Error(s) on processing entity %s: %s", val.Field, builder.String())
}

func (val *ValidationError) Unwrap() error {
	if len(val.WrappedError) == 0 {
		panic(fmt.Sprintf("Empty error unwrap requesting! %v", val))
	}
	return val.WrappedError[0]
}

func newValidationError(field string, wrappedError []error) error {
	if len(wrappedError) == 0 {
		return nil
	}
	return &ValidationError{
		Field:        field,
		WrappedError: wrappedError,
	}
}

func validateParam(paramInfo *ParamInfo) []error {
	errs := make([]error, 0)
	if (paramInfo.Destination == URLPlaced) && paramInfo.Optional && paramInfo.DefaultValue == nil {
		errs = append(errs, errors.New("UrlPlaced param can't be marked as optional"))
	}
	if paramInfo.DefaultValue != nil {
		str, ok := paramInfo.DefaultValue.(string)
		if ok {
			val, err := parseValue(paramInfo.Type, str)
			if err != nil {
				errs = append(errs, fmt.Errorf("Error on default value %w", err))
			} else {
				paramInfo.DefaultValue = val
			}
		} else {
			errs = append(errs, fmt.Errorf("Error on default value: failed to get default value as a string"))
		}
	}
	return errs
}

func validateHand(urlRecord *URLRecord) []error {
	errs := make([]error, 0)
	for paramName, param := range urlRecord.Parameters {
		handErrs := validateParam(&param)
		if len(handErrs) != 0 {
			err := newValidationError(paramName, handErrs)
			errs = append(errs, err)
		}
	}
	return errs
}

func validateContainer(container *URLContrainer) error {
	errs := make([]error, 0)
	for handName, hand := range *container {
		handErrs := validateHand(&hand)

		if hand.URLName != handName {
			handErrs = append(handErrs, fmt.Errorf("difference between hand name in field %s and in map %s", hand.URLName, handName))
		}

		if len(handErrs) != 0 {
			err := newValidationError(handName, handErrs)
			errs = append(errs, err)
		}
	}
	err := newValidationError("", errs)
	return err
}

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
	err = validateContainer(&urlContainer)
	if err != nil {
		return nil, err
	}
	return NewDescriptionSourceFromDict(urlContainer), nil
}

// GetDescriptionSourceFromYAML получить из json описание ручек
func GetDescriptionSourceFromYAML(reader io.Reader) (*SimpleDescriptionsSource, error) {
	var urlContainer URLContrainer
	// TODO: возможно стоит переделать на работу парсера, чтобы не вычитывать весь файл
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(bytes, &urlContainer)
	if err != nil {
		return nil, err
	}
	err = validateContainer(&urlContainer)
	if err != nil {
		return nil, err
	}
	return NewDescriptionSourceFromDict(urlContainer), nil
}
