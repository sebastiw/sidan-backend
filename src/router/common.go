package router

import (
	"log"
	"net/http"
	"strconv"
)

func MakeDefaultInt(r *http.Request, variableName string, defaultValue string) int {
	rawVal := r.FormValue(variableName)
	if "" == rawVal {
		rawVal = defaultValue
	}
	realVal, err := strconv.Atoi(rawVal)
	if err != nil {
		log.Println(getRequestId(r), "Error parsing '"+variableName+"':", err.Error())
		defaultVal, _ := strconv.Atoi(defaultValue)
		return defaultVal
	}
	return realVal
}
