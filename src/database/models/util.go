package database

import(
	"fmt"
)

func add_sp(s []string, t string, m *string) []string {
	if(nil != m) {
		s = append(s, fmt.Sprintf("%s: %s", t, *m))
	}
	return s
}
func add_s(s []string, t string, m string) []string {
	if("" != m) {
		s = append(s, fmt.Sprintf("%s: %s", t, m))
	}
	return s
}
func add_i(s []string, t string, m int64) []string {
	if(0 != m) {
		s = append(s, fmt.Sprintf("%s: %d", t, m))
	}
	return s
}
func add_bp(s []string, t string, m *bool) []string {
	if(nil != m) {
		s = append(s, fmt.Sprintf("%s: %t", t, *m))
	}
	return s
}
