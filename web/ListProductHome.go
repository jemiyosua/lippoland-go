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

type JListProductHomeRequest struct {
	Ip string
	Id string
}

type JListProductHomeResponse struct {
	Id       string
	Images    string
	Title     string
}

func ListProductHome(c *gin.Context) {
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

	jListProductHomeRequest := JListProductHomeRequest{}
	jListProductHomeResponse := JListProductHomeResponse{}
	jListProductHomeResponses := []JListProductHomeResponse{}

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
	logFile = logFile + "ListProductHome_" + dateNow + ".log"
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
		ReturnListProductHome(jListProductHomeResponses, jListProductHomeRequest.Ip, bodyJson, startTime, logData, errorCode, errorCode, errorMessage, errorMessage, allHeader, method, path, ip, c)
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage = "Error, Body - invalid json data"
		ReturnListProductHome(jListProductHomeResponses, jListProductHomeRequest.Ip, bodyJson, startTime, logData, errorCode, errorCode, errorMessage, errorMessage, allHeader, method, path, ip, c)
		return
	}
	// ------ end of body json validation ------

	errorMessageJson, errorCodeJson, bodyJson := helper.ValidateJson(jListProductHomeRequest.Ip, "ListProductHome", c)
	if errorMessageJson != "" {
		ReturnListProductHome(jListProductHomeResponses, jListProductHomeRequest.Ip, bodyJson, startTime, logData, errorCodeJson, errorCodeJson, errorMessageJson, errorMessageJson, allHeader, method, path, ip, c)
		return
	}

	// ------ Header Validation ------
	if helper.ValidateHeader(bodyJson, c) {
		if err := c.ShouldBindJSON(&jListProductHomeRequest); err != nil {
			errorMessage := err.Error()
			errorMessageReturn := "Error, bind JSON data"
			ReturnListProductHome(jListProductHomeResponses, jListProductHomeRequest.Ip, bodyJson, startTime, logData, errorCode, errorCode, errorMessage, errorMessageReturn, allHeader, method, path, ip, c)
			return
		} else {
			try.This(func() {

				ipUser := jListProductHomeRequest.Ip
				id := jListProductHomeRequest.Id

				queryWhere := " status = 1 "
				if id != "" {
					queryWhere += " AND "

					if queryWhere != "" {
						queryWhere += fmt.Sprintf(" id = '%s' ", id)
					}
				}

				if queryWhere != "" {
					queryWhere = fmt.Sprintf(" WHERE %s ", queryWhere)
				}
				
				query := fmt.Sprintf("SELECT id, images, title FROM lippo_products_home %s", queryWhere)
				rows, err := db.Query(query)
				if err != nil && err != sql.ErrNoRows {
					errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
					ReturnListProductHome(jListProductHomeResponses, jListProductHomeRequest.Ip, bodyJson, startTime, logData, errorCode, errorCode, errorMessage, errorMessage, allHeader, method, path, ip, c)
					return
				}
				defer rows.Close()
				for rows.Next() {
					err = rows.Scan(
						&jListProductHomeResponse.Id,
						&jListProductHomeResponse.Images,
						&jListProductHomeResponse.Title,
					)

					jListProductHomeResponse.Images = urlImages + jListProductHomeResponse.Images

					if err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
						ReturnListProductHome(jListProductHomeResponses, ipUser, bodyJson, startTime, logData, "1", "1", errorMessage, errorMessage, allHeader, method, path, ip, c)
						return
					}

					jListProductHomeResponses = append(jListProductHomeResponses, jListProductHomeResponse)
				}

				ReturnListProductHome(jListProductHomeResponses, ipUser, bodyJson, startTime, logData, "0", "0", "", "", allHeader, method, path, ip, c)

			}).Finally(func() {

			}).Catch(func(e try.E) {
				errorMessageReturn := "Error Catch, Data tidak ditemukan"
				ReturnListProductHome(jListProductHomeResponses, jListProductHomeRequest.Ip, bodyJson, startTime, logData, errorCode, errorCode, errorMessageReturn, errorMessageReturn, allHeader, method, path, ip, c)
				return
			})
		}
	}
}

func ReturnListProductHome(jListProductHomeResponses []JListProductHomeResponse, ipUser string, bodyJson string, startTime time.Time, logData string, errorCode string, errorCodeReturn string, errorMessage string, errorMessageReturn string, header string, method string, path string, ip string, c *gin.Context) {

	if errorCode != "0" {
		helper.SendLogError(ip, "WEB - LIST PRODUCT HOME", errorMessage, bodyJson, "", errorCode, header, method, path, ip, c)
	}

	if strings.Contains(errorMessageReturn, "Error running") {
		errorMessageReturn = "Error Execute data"
	}

	c.PureJSON(http.StatusOK, gin.H{
		"ErrorCode":    errorCodeReturn,
		"ErrorMessage": errorMessageReturn,
		"Result":       jListProductHomeResponses,
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
