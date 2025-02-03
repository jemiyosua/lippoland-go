package helper

import (
	"bytes"
	"io/ioutil"

	"github.com/gin-gonic/gin"
)

func ValidateJson(ip string, page string, c *gin.Context) (string, string, string) {
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
	}

	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	bodyString := string(bodyBytes)

	bodyJson := TrimReplace(string(bodyString))

	// ------ Body Json Validation ------
	if string(bodyString) == "" {
		errorMessage := "Error, Body is empty"
		errorCode := "1"
		return errorMessage, errorCode, ""
	}

	is_Json := IsJson(bodyString)
	if is_Json == false {
		errorMessage := "Error, Body - invalid json data"
		errorCode := "1"
		return errorMessage, errorCode, ""
	}
	// ------ end of Body Json Validation ------

	return "", "0", bodyJson
}
