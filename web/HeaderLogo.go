package web

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"lippoland/helper"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/manucorporat/try"
)

type JHeaderLogoRequest struct {
	Ip string
	Id string
}

type JHeaderLogoResponse struct {
	Id       string
	ImagesDark    string
	ImagesLight    string
}

func HeaderLogo(c *gin.Context) {
	db := helper.Connect(c)
	defer db.Close()
	startTime := time.Now()
	startTimeString := startTime.String()

	var (
		bodyBytes []byte
		xRealIp   string
		ip        string
		logFile   string
	)

	jHeaderLogoRequest := JHeaderLogoRequest{}
	jHeaderLogoResponse := JHeaderLogoResponse{}
	jHeaderLogoResponses := []JHeaderLogoResponse{}

	errorCode := "1"
	errorMessage := ""

	allHeader := helper.ReadAllHeader(c)
	logFile = os.Getenv("LOGFILE_WEB")
	urlImages := os.Getenv("URL_IMAGES")
	method := c.Request.Method
	path := c.Request.URL.EscapedPath()

	if xRealIp != "" {
		ip = xRealIp
	} else {
		ip = c.ClientIP()
	}

	// ------ start log file ------
	dateNow := startTime.Format("2006-01-02")
	logFile = logFile + "HeaderLogo_" + dateNow + ".log"
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	log.SetOutput(file)
	// ------ end log file ------

	// ------ start body json validation ------
	if c.Request.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	bodyString := string(bodyBytes)

	bodyJson := helper.TrimReplace(string(bodyString))
	logData := startTimeString + "~" + ip + "~" + method + "~" + path + "~" + allHeader + "~"
	rex := regexp.MustCompile(`\r?\n`)
	logData = logData + rex.ReplaceAllString(bodyJson, "") + "~"

	if string(bodyString) == "" {
		errorMessage = "Error, Body is empty"
		ReturnHeaderLogo(jHeaderLogoResponses, jHeaderLogoRequest.Ip, bodyJson, startTime, logData, errorCode, errorCode, errorMessage, errorMessage, allHeader, method, path, ip, c)
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage = "Error, Body - invalid json data"
		ReturnHeaderLogo(jHeaderLogoResponses, jHeaderLogoRequest.Ip, bodyJson, startTime, logData, errorCode, errorCode, errorMessage, errorMessage, allHeader, method, path, ip, c)
		return
	}
	// ------ end of body json validation ------

	errorMessageJson, errorCodeJson, bodyJson := helper.ValidateJson(jHeaderLogoRequest.Ip, "HeaderLogo", c)
	if errorMessageJson != "" {
		ReturnHeaderLogo(jHeaderLogoResponses, jHeaderLogoRequest.Ip, bodyJson, startTime, logData, errorCodeJson, errorCodeJson, errorMessageJson, errorMessageJson, allHeader, method, path, ip, c)
		return
	}

	// ------ Header Validation ------
	if helper.ValidateHeader(bodyJson, c) {
		if err := c.ShouldBindJSON(&jHeaderLogoRequest); err != nil {
			errorMessage := err.Error()
			errorMessageReturn := "Error, bind JSON data"
			ReturnHeaderLogo(jHeaderLogoResponses, jHeaderLogoRequest.Ip, bodyJson, startTime, logData, errorCode, errorCode, errorMessage, errorMessageReturn, allHeader, method, path, ip, c)
			return
		} else {
			try.This(func() {

				ipUser := jHeaderLogoRequest.Ip
				id := jHeaderLogoRequest.Id

				queryWhere := ""
				if id != "" {
					queryWhere += " AND "

					if queryWhere != "" {
						queryWhere += fmt.Sprintf(" id = '%s' ", id)
					}
				}

				if queryWhere != "" {
					queryWhere = fmt.Sprintf(" WHERE %s ", queryWhere)
				}

				query := fmt.Sprintf("SELECT id, logo_dark, logo_light FROM lippo_header_logo %s", queryWhere)
				rows, err := db.Query(query)
				if err != nil && err != sql.ErrNoRows {
					errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
					ReturnHeaderLogo(jHeaderLogoResponses, jHeaderLogoRequest.Ip, bodyJson, startTime, logData, errorCode, errorCode, errorMessage, errorMessage, allHeader, method, path, ip, c)
					return
				}
				defer rows.Close()
				for rows.Next() {
					err = rows.Scan(
						&jHeaderLogoResponse.Id,
						&jHeaderLogoResponse.ImagesDark,
						&jHeaderLogoResponse.ImagesLight,
					)

					jHeaderLogoResponse.ImagesDark = urlImages + jHeaderLogoResponse.ImagesDark
					jHeaderLogoResponse.ImagesLight = urlImages + jHeaderLogoResponse.ImagesLight

					if err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
						ReturnHeaderLogo(jHeaderLogoResponses, ipUser, bodyJson, startTime, logData, "1", "1", errorMessage, errorMessage, allHeader, method, path, ip, c)
						return
					}

					jHeaderLogoResponses = append(jHeaderLogoResponses, jHeaderLogoResponse)

				}

				ReturnHeaderLogo(jHeaderLogoResponses, ipUser, bodyJson, startTime, logData, "0", "0", "", "", allHeader, method, path, ip, c)

			}).Finally(func() {

			}).Catch(func(e try.E) {
				errorMessageReturn := "Error Catch, Data tidak ditemukan"
				ReturnHeaderLogo(jHeaderLogoResponses, jHeaderLogoRequest.Ip, bodyJson, startTime, logData, errorCode, errorCode, errorMessageReturn, errorMessageReturn, allHeader, method, path, ip, c)
				return
			})
		}
	}
}

func ReturnHeaderLogo(jHeaderLogoResponses []JHeaderLogoResponse, ipUser string, bodyJson string, startTime time.Time, logData string, errorCode string, errorCodeReturn string, errorMessage string, errorMessageReturn string, header string, method string, path string, ip string, c *gin.Context) {

	if errorCode != "0" {
		helper.SendLogError(ip, "WEB - HEADER LOGO", errorMessage, bodyJson, "", errorCode, header, method, path, ip, c)
	}

	if strings.Contains(errorMessageReturn, "Error running") {
		errorMessageReturn = "Error Execute data"
	}

	c.PureJSON(http.StatusOK, gin.H{
		"ErrorCode":    errorCodeReturn,
		"ErrorMessage": errorMessageReturn,
		"Result":       jHeaderLogoResponses,
	})

	Rex := regexp.MustCompile(`\r?\n`)
	EndTime := time.Now()
	CodeError := "200"

	if errorMessage != "" {
		CodeError = "500"
	}

	Diff := EndTime.Sub(startTime)

	logDataNew := Rex.ReplaceAllString(logData+CodeError+""+EndTime.String()+""+Diff.String()+"~"+errorMessage, "")
	//
	log.Println(logDataNew)
	runtime.GC()
	return
}
