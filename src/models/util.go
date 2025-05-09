package models

import (
	"fmt"
)

func addSp(s []string, t string, m *string) []string {
	if nil != m {
		s = append(s, fmt.Sprintf("%s: %s", t, *m))
	}
	return s
}
func addS(s []string, t string, m string) []string {
	if "" != m {
		s = append(s, fmt.Sprintf("%s: %s", t, m))
	}
	return s
}
func addI(s []string, t string, m int64) []string {
	if 0 != m {
		s = append(s, fmt.Sprintf("%s: %d", t, m))
	}
	return s
}
func addFp(s []string, t string, m *float64) []string {
	if nil != m {
		s = append(s, fmt.Sprintf("%s: %f", t, *m))
	}
	return s
}
func addB(s []string, t string, m bool) []string {
	s = append(s, fmt.Sprintf("%s: %t", t, m))
	return s
}
func addBp(s []string, t string, m *bool) []string {
	if nil != m {
		s = append(s, fmt.Sprintf("%s: %t", t, *m))
	}
	return s
}
