package view

import (
	"errors"
	"fmt"
)

const (
	additionalProtoDirsKey = "additional-proto-dirs"
)

type Settings struct {
	AdditionalProtoDirs []string
}

var (
	ErrRepackingSettings = errors.New("failed repacking settings")
)

func SettingsFromInterface(in interface{}) (*Settings, error) {
	settingsMap, ok := in.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: settings should have a map[string]interface{} type", ErrRepackingSettings)
	}

	// vscode-languageclient wraps settings under a section named after the
	// language server. Unwrap it if the flat key isn't present at the top level.
	if _, ok := settingsMap[additionalProtoDirsKey]; !ok {
		if unwrapped := findNestedSettingsMap(settingsMap); unwrapped != nil {
			settingsMap = unwrapped
		}
	}

	var settings Settings

	if value, ok := settingsMap[additionalProtoDirsKey]; ok {
		protoDirs, err := StringsSliceFromInterface(value)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: key = %s", ErrRepackingSettings, err.Error(), additionalProtoDirsKey)
		}
		settings.AdditionalProtoDirs = protoDirs
	}

	return &settings, nil
}

// findNestedSettingsMap searches one level deep for a nested map that contains
// the additional-proto-dirs key. Returns nil if none is found.
func findNestedSettingsMap(settingsMap map[string]interface{}) map[string]interface{} {
	for _, value := range settingsMap {
		nested, ok := value.(map[string]interface{})
		if !ok {
			continue
		}
		if _, ok := nested[additionalProtoDirsKey]; ok {
			return nested
		}
	}
	return nil
}

func StringsSliceFromInterface(in interface{}) ([]string, error) {
	interfaceSlice, ok := in.([]interface{})
	if !ok {
		return nil, errors.New("field should have a []interface{} type")
	}

	var result []string

	for i, item := range interfaceSlice {
		str, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("item [%d] should have a string type", i)
		}
		result = append(result, str)
	}

	return result, nil
}
