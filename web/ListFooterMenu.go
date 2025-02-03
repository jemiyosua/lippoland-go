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

type JListFooterMenuRequest struct {
	Ip string
	Id string
	Flag string
}

type JListFooterMenuLeftResponse struct {
	Id       string
	MenuName    string
	UrlPage    string
}

type JListFooterMenuRightResponse struct {
	Id       string
	MenuName    string
	UrlPage    string
}

func ListFooterMenu(c *gin.Context) {
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

	jListFooterMenuRequest := JListFooterMenuRequest{}

	jListFooterMenuLeftResponse := JListFooterMenuLeftResponse{}
	jListFooterMenuLeftResponses := []JListFooterMenuLeftResponse{}

	jListFooterMenuRightResponse := JListFooterMenuRightResponse{}
	jListFooterMenuRightResponses := []JListFooterMenuRightResponse{}

	errorCode := "1"
	errorMessage := ""

	allHeader := helper.ReadAllHeader(c)
	logFile = os.Getenv("LOGFILE_WEB")
	method := c.Request.Method
	path := c.Request.URL.EscapedPath()

	if xRealIp != "" {
		ip = xRealIp
	} else {
		ip = c.ClientIP()
	}

	// ------ start log file ------
	dateNow := startTime.Format("2006-01-02")
	logFile = logFile + "ListFooterMenu_" + dateNow + ".log"
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
		ReturnListFooterMenu(jListFooterMenuLeftResponses, jListFooterMenuRightResponses, jListFooterMenuRequest.Ip, bodyJson, startTime, logData, errorCode, errorCode, errorMessage, errorMessage, allHeader, method, path, ip, c)
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage = "Error, Body - invalid json data"
		ReturnListFooterMenu(jListFooterMenuLeftResponses, jListFooterMenuRightResponses, jListFooterMenuRequest.Ip, bodyJson, startTime, logData, errorCode, errorCode, errorMessage, errorMessage, allHeader, method, path, ip, c)
		return
	}
	// ------ end of body json validation ------

	errorMessageJson, errorCodeJson, bodyJson := helper.ValidateJson(jListFooterMenuRequest.Ip, "ListFooterMenu", c)
	if errorMessageJson != "" {
		ReturnListFooterMenu(jListFooterMenuLeftResponses, jListFooterMenuRightResponses, jListFooterMenuRequest.Ip, bodyJson, startTime, logData, errorCodeJson, errorCodeJson, errorMessageJson, errorMessageJson, allHeader, method, path, ip, c)
		return
	}

	// ------ Header Validation ------
	if helper.ValidateHeader(bodyJson, c) {
		if err := c.ShouldBindJSON(&jListFooterMenuRequest); err != nil {
			errorMessage := err.Error()
			errorMessageReturn := "Error, bind JSON data"
			ReturnListFooterMenu(jListFooterMenuLeftResponses, jListFooterMenuRightResponses, jListFooterMenuRequest.Ip, bodyJson, startTime, logData, errorCode, errorCode, errorMessage, errorMessageReturn, allHeader, method, path, ip, c)
			return
		} else {
			try.This(func() {

				ipUser := jListFooterMenuRequest.Ip
				id := jListFooterMenuRequest.Id
				flag := jListFooterMenuRequest.Flag

				queryWhere := " status = 1 "
				if id != "" {
					queryWhere += " AND "

					if queryWhere != "" {
						queryWhere += fmt.Sprintf(" id = '%s' ", id)
					}
				}

				if flag != "" {
					queryWhere += " AND "

					if queryWhere != "" {
						queryWhere += fmt.Sprintf(" flag = '%s' ", flag)
					}
				}

				if queryWhere != "" {
					queryWhere = fmt.Sprintf(" WHERE %s ", queryWhere)
				}
				
				query := fmt.Sprintf("SELECT id, menu_name, url FROM lippo_header_menu %s", queryWhere)
				rows, err := db.Query(query)
				if err != nil && err != sql.ErrNoRows {
					errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
					ReturnListFooterMenu(jListFooterMenuLeftResponses, jListFooterMenuRightResponses, jListFooterMenuRequest.Ip, bodyJson, startTime, logData, errorCode, errorCode, errorMessage, errorMessage, allHeader, method, path, ip, c)
					return
				}
				defer rows.Close()
				idMenuPage := ""
				menuName := ""
				urlPage := ""
				for rows.Next() {
					err = rows.Scan(
						&idMenuPage,
						&menuName,
						&urlPage,
					)

					if menuName == "ABOUT US" || menuName == "DEVELOPMENTS" || menuName == "SERVICES" || menuName == "PROMOTION" {
						jListFooterMenuLeftResponse.Id = idMenuPage
						jListFooterMenuLeftResponse.MenuName = menuName
						jListFooterMenuLeftResponse.UrlPage = urlPage

						jListFooterMenuLeftResponses = append(jListFooterMenuLeftResponses, jListFooterMenuLeftResponse)
					}

					if menuName == "SUSTAINABILITY" || menuName == "NEWS" || menuName == "CAREER" {
						jListFooterMenuRightResponse.Id = idMenuPage
						jListFooterMenuRightResponse.MenuName = menuName
						jListFooterMenuRightResponse.UrlPage = urlPage

						jListFooterMenuRightResponses = append(jListFooterMenuRightResponses, jListFooterMenuRightResponse)
					}

					if err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
						ReturnListFooterMenu(jListFooterMenuLeftResponses, jListFooterMenuRightResponses, ipUser, bodyJson, startTime, logData, "1", "1", errorMessage, errorMessage, allHeader, method, path, ip, c)
						return
					}
				}

				ReturnListFooterMenu(jListFooterMenuLeftResponses, jListFooterMenuRightResponses, ipUser, bodyJson, startTime, logData, "0", "0", "", "", allHeader, method, path, ip, c)

			}).Finally(func() {

			}).Catch(func(e try.E) {
				errorMessageReturn := "Error Catch, Data tidak ditemukan"
				ReturnListFooterMenu(jListFooterMenuLeftResponses, jListFooterMenuRightResponses, jListFooterMenuRequest.Ip, bodyJson, startTime, logData, errorCode, errorCode, errorMessageReturn, errorMessageReturn, allHeader, method, path, ip, c)
				return
			})
		}
	}
}

func ReturnListFooterMenu(jListFooterMenuLeftResponses []JListFooterMenuLeftResponse, jListFooterMenuRightResponses []JListFooterMenuRightResponse, ipUser string, bodyJson string, startTime time.Time, logData string, errorCode string, errorCodeReturn string, errorMessage string, errorMessageReturn string, header string, method string, path string, ip string, c *gin.Context) {

	if errorCode != "0" {
		helper.SendLogError(ip, "WEB - LIST FOOTER MENU", errorMessage, bodyJson, "", errorCode, header, method, path, ip, c)
	}

	if strings.Contains(errorMessageReturn, "Error running") {
		errorMessageReturn = "Error Execute data"
	}

	c.PureJSON(http.StatusOK, gin.H{
		"ErrorCode":    errorCodeReturn,
		"ErrorMessage": errorMessageReturn,
		"ResultLeft":       jListFooterMenuLeftResponses,
		"ResultRight":       jListFooterMenuRightResponses,
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
