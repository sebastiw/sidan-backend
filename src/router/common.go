package router

import (
	"log"
	"net/http"
	"strconv"
	ru "github.com/sebastiw/sidan-backend/src/router_util"
)

func MakeDefaultInt(r *http.Request, variableName string, defaultValue string) int {
	rawVal := r.FormValue(variableName)
	if "" == rawVal {
		rawVal = defaultValue
	}
	realVal, err := strconv.Atoi(rawVal)
	if err != nil {
		log.Println(ru.GetRequestId(r), "Error parsing '"+variableName+"':", err.Error())
		defaultVal, _ := strconv.Atoi(defaultValue)
		return defaultVal
	}
	return realVal
}

func MakeDefaultBool(r *http.Request, variableName string, defaultValue string) bool {
	rawVal := r.FormValue(variableName)
	if "" == rawVal {
		rawVal = defaultValue
	}
	realVal, err := strconv.ParseBool(rawVal)
	if err != nil {
		log.Println(ru.GetRequestId(r), "Error parsing '"+variableName+"':", err.Error())
		defaultVal, _ := strconv.ParseBool(defaultValue)
		return defaultVal
	}
	return realVal
}

func CheckError(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		log.Println(ru.GetRequestId(r), err)
		panic(err.Error())
	}
}
