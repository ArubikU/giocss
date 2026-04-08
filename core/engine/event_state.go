package engine

import (
	"image"
	"strconv"
	"strings"

	coreinput "github.com/ArubikU/giocss/core/input"
)

func ToggleCheckboxState(path string, boolValues map[string]bool, timestamp int64) map[string]any {
	checked := !boolValues[path]
	boolValues[path] = checked
	return BuildCheckboxEventPayload(path, checked, timestamp)
}

func SelectRadioState(path string, groupKey string, radioValue string, inputValues map[string]string, timestamp int64) map[string]any {
	inputValues[groupKey] = radioValue
	return BuildRadioEventPayload(path, radioValue, timestamp)
}

func ResolveSliderPointerState(path string, boundsForPath map[string]image.Rectangle, lastPointer map[string]image.Point, sliderValues map[string]float64, minV float64, maxV float64, component string, timestamp int64) (map[string]any, bool) {
	bounds, ok := boundsForPath[path]
	if !ok {
		return nil, false
	}
	bw := bounds.Max.X - bounds.Min.X
	px := lastPointer[path].X
	newVal, ok := coreinput.ResolveSliderValueFromPointer(px, bounds.Min.X, bw, minV, maxV)
	if !ok {
		return nil, false
	}
	sliderValues[path] = newVal
	return BuildSliderEventPayload(path, component, newVal, timestamp), true
}

func CycleSelectState(path string, key string, inputValues map[string]string, labels []string, values []string, enabledIndexes []int, currentIndex int, timestamp int64) (map[string]any, bool) {
	if len(labels) == 0 {
		return nil, false
	}
	if len(enabledIndexes) == 0 {
		return nil, false
	}
	if len(values) != len(labels) {
		values = labels
	}
	next := nextEnabledSelectIndex(currentIndex, enabledIndexes, len(labels))
	if idxStr := inputValues[key]; idxStr != "" {
		if n, err := strconv.Atoi(idxStr); err == nil {
			next = nextEnabledSelectIndex(n, enabledIndexes, len(labels))
		}
	}
	if next < 0 {
		next = 0
	}
	if next >= len(labels) {
		next = len(labels) - 1
	}
	inputValues[key] = strconv.Itoa(next)
	selectedValue := strings.TrimSpace(values[next])
	if selectedValue == "" {
		selectedValue = labels[next]
	}
	return BuildSelectEventPayload(path, selectedValue, next, timestamp), true
}

func nextEnabledSelectIndex(current int, enabledIndexes []int, total int) int {
	if total <= 0 {
		return 0
	}
	if len(enabledIndexes) == 0 {
		return coreinput.NextSelectIndex(current, total)
	}
	for i, idx := range enabledIndexes {
		if idx == current {
			return enabledIndexes[(i+1)%len(enabledIndexes)]
		}
	}
	for _, idx := range enabledIndexes {
		if idx > current {
			return idx
		}
	}
	return enabledIndexes[0]
}
