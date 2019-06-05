package parser

import (
	"fmt"
	"strings"
)

const (
	NoLabel = "NoLabel"
)

func LabelsFromString(str string) (map[string]string, error) {
	if len(str) == 0 {
		return nil, nil
	}
	labels := strings.Split(str, ",")
	m := make(map[string]string, len(labels))
	for _, v := range labels {
		if strings.Contains(v, "=") {
			label := strings.SplitN(v, "=", 2)
			if len(label) != 2 {
				return m, fmt.Errorf("Malformed label: %s", v)
			}
			if _, ok := m[label[0]]; ok {
				return m, fmt.Errorf("Duplicate label: %s", v)
			}
			m[label[0]] = label[1]
		} else if len(v) != 0 {
			m[v] = ""
		}
	}
	return m, nil
}

func LabelsToString(labels map[string]string) string {
	l := ""
	for k, v := range labels {
		if len(l) != 0 {
			l += ","
		}
		if len(v) != 0 {
			l += k + "=" + v
		} else if len(k) != 0 {
			l += k
		}
	}
	return l
}

func MergeLabels(old map[string]string, new map[string]string) map[string]string {
	if old == nil {
		return new
	}
	if new == nil {
		return old
	}
	m := make(map[string]string, len(old)+len(new))
	for k, v := range old {
		m[k] = v
	}
	for k, v := range new {
		m[k] = v
	}
	return m
}

func hasLabels(set, subset map[string]string, matchValue bool) bool {
	for k, v1 := range subset {
		if v2, ok := set[k]; !ok || (matchValue && v1 != v2) {
			return false
		}
	}
	return true
}

func HasLabels(set, subset map[string]string) bool {
	return hasLabels(set, subset, true)
}

func HasLabelKeys(set, subset map[string]string) bool {
	return hasLabels(set, subset, false)
}

func hasAnyLabel(set, subset map[string]string, matchValue bool) bool {
	for k, v1 := range subset {
		if v2, ok := set[k]; ok && (!matchValue || v1 == v2) {
			return true
		}
	}
	return false
}

func HasAnyLabel(set, subset map[string]string) bool {
	return hasAnyLabel(set, subset, true)
}

func HasAnyLabelKey(set, subset map[string]string) bool {
	return hasAnyLabel(set, subset, false)
}
