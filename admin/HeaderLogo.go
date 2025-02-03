package admin

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"lippoland/helper"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/manucorporat/try"
)

type JHeaderLogoRequest struct {
	Username string
	ParamKey string
	Method string
	Id string
	FileNameImageDark string
	FileNameImageDarkOld string
	Base64DataImageDark string
	FileNameImageLight string
	FileNameImageLightOld string
	Base64DataImageLight string
	Page        int
	RowPage     int
	OrderBy     string
	Order       string
}

type JHeaderLogoResponse struct {
	Id string
	ImageDark string
	ImageLight string
	FileNameImageDark string
	FileNameImageLight string
	TanggalInput string
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
		totalRecords float64
		totalPage float64
	)

	jHeaderLogoRequest := JHeaderLogoRequest{}
	jHeaderLogoResponse := JHeaderLogoResponse{}
	jHeaderLogoResponses := []JHeaderLogoResponse{}

	errorCode := "1"
	errorMessage := ""
	errorCodeSession := "2"
	errorMessageSession := "Session Expired"

	allHeader := helper.ReadAllHeader(c)
	logFile = os.Getenv("LOGFILE_ADMIN")
	urlImages := os.Getenv("URL_IMAGES")
	method := c.Request.Method
	path := c.Request.URL.EscapedPath()

	// ---------- start get ip ----------
	if Values, _ := c.Request.Header["X-Real-Ip"]; len(Values) > 0 {
		xRealIp = Values[0]
	}

	if xRealIp != "" {
		ip = xRealIp
	} else {
		ip = c.ClientIP()
	}
	// ---------- end of get ip ----------

	// ---------- start log file ----------
	dateNow := startTime.Format("2006-01-02")
	logFile = logFile + "HeaderLogo_" + dateNow + ".log"
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	log.SetOutput(file)
	// ---------- end of log file ----------

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
		dataLogHeaderLogo(jHeaderLogoResponses, jHeaderLogoRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage = "Error, Body - invalid json data"
		dataLogHeaderLogo(jHeaderLogoResponses, jHeaderLogoRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	try.This(func() {
		if helper.ValidateHeader(bodyString, c) {
			if err := c.ShouldBindJSON(&jHeaderLogoRequest); err != nil {
				errorMessage = "Error, Bind Json Data"
				dataLogHeaderLogo(jHeaderLogoResponses, jHeaderLogoRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			} else {
				username := jHeaderLogoRequest.Username
				paramKey := jHeaderLogoRequest.ParamKey
				method := jHeaderLogoRequest.Method

				filenameImageDark := jHeaderLogoRequest.FileNameImageDark
				filenameImageDarkOld := jHeaderLogoRequest.FileNameImageDarkOld
				base64ImageDark := jHeaderLogoRequest.Base64DataImageDark

				filenameImageLight := jHeaderLogoRequest.FileNameImageLight
				filenameImageLightOld := jHeaderLogoRequest.FileNameImageLightOld
				base64ImageLight := jHeaderLogoRequest.Base64DataImageLight

				id := jHeaderLogoRequest.Id
				page := jHeaderLogoRequest.Page
				rowPage := jHeaderLogoRequest.RowPage

				// ------ Param Validation ------
				if username == "" {
					errorMessage += "Username can't null value"
				}

				if paramKey == "" {
					errorMessage += "ParamKey can't null value"
				}

				if method == "" {
					errorMessage += "Method can't null value"
				}

				if method == "SELECT" {
					if page == 0 {
						errorMessage += "Page can't null or 0 value"
					}
	
					if rowPage == 0 {
						errorMessage += "RowPage can't null or 0 value"
					}
				}

				if errorMessage != "" {
					dataLogHeaderLogo(jHeaderLogoResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}
				// ------ end of Param Validation ------

				// ------ start check session paramkey ------
				checkAccessVal := helper.CheckSession(username, paramKey, c)
				if checkAccessVal != "1" {
					dataLogHeaderLogo(jHeaderLogoResponses, username, errorCodeSession, errorMessageSession, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				if method == "INSERT" {

				} else if method == "UPDATE" {

					if (id == "") {
						errorMessage = "Id can't null value"
						dataLogHeaderLogo(jHeaderLogoResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					} 

					queryUpdate := ""
					if (filenameImageDarkOld == "" || filenameImageLightOld == "") {
						errorMessage = "filename image old can't null value"
						dataLogHeaderLogo(jHeaderLogoResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					}

					if (filenameImageDark != "" && base64ImageDark != "") {
						imageFilename, errorCodeImage, errorMessageImage := helper.CreateImageUrl(method, filenameImageDarkOld, filenameImageDark, base64ImageDark, db, c)
						if errorCodeImage != "0" {
							dataLogHeaderLogo(jHeaderLogoResponses, username, errorCodeImage, errorMessageImage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						}

						queryUpdate += fmt.Sprintf(" , logo_dark = '%s' ", imageFilename)
					}

					if (filenameImageLight != "" && base64ImageLight != "") {
						imageFilename, errorCodeImage, errorMessageImage := helper.CreateImageUrl(method, filenameImageLightOld, filenameImageLight, base64ImageLight, db, c)
						if errorCodeImage != "0" {
							dataLogHeaderLogo(jHeaderLogoResponses, username, errorCodeImage, errorMessageImage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						}

						queryUpdate += fmt.Sprintf(" , logo_light = '%s' ", imageFilename)
					}

					query := fmt.Sprintf("UPDATE lippo_header_logo SET tgl_input = NOW() %s WHERE id = '%s'", queryUpdate, id)
					_, err := db.Exec(query)
					if err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
						dataLogHeaderLogo(jHeaderLogoResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					}
					
					dataLogHeaderLogo(jHeaderLogoResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)

				} else if method == "DELETE" {

				} else if method == "SELECT" {
					
					pageNow := (page - 1) * rowPage
					pageNowString := strconv.Itoa(pageNow)
					queryLimit := ""

					queryWhere := ""
					if id != "" {
						if queryWhere != "" {
							queryWhere += " AND "
						}

						queryWhere += fmt.Sprintf(" id = '%s' ", id)
					}

					if queryWhere != "" {
						queryWhere = " WHERE " + queryWhere
					}

					totalRecords = 0
					totalPage = 0
					query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM lippo_header_logo %s", queryWhere)
					if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
						errorMessage = "Error running, " + err.Error()
						dataLogHeaderLogo(jHeaderLogoResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					if rowPage == -1 {
						queryLimit = ""
						totalPage = 1
					} else {
						rowPageString := strconv.Itoa(rowPage)
						queryLimit = "LIMIT " + pageNowString + "," + rowPageString
						totalPage = math.Ceil(float64(totalRecords) / float64(rowPage))
					}

					query1 := fmt.Sprintf("SELECT id, logo_dark, logo_light, tgl_input FROM lippo_header_logo %s %s", queryWhere, queryLimit)
					rows, err := db.Query(query1)
					defer rows.Close()
					if err != nil {
						errorMessage = "Error running, " + err.Error()
						dataLogHeaderLogo(jHeaderLogoResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					for rows.Next() {
						err = rows.Scan(
							&jHeaderLogoResponse.Id,
							&jHeaderLogoResponse.ImageDark,
							&jHeaderLogoResponse.ImageLight,
							&jHeaderLogoResponse.TanggalInput,
						)

						jHeaderLogoResponse.FileNameImageDark = jHeaderLogoResponse.ImageDark
						jHeaderLogoResponse.FileNameImageLight = jHeaderLogoResponse.ImageLight
						jHeaderLogoResponse.ImageDark = urlImages + jHeaderLogoResponse.ImageDark
						jHeaderLogoResponse.ImageLight = urlImages + jHeaderLogoResponse.ImageLight

						if err != nil {
							errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err)
							dataLogHeaderLogo(jHeaderLogoResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
							return
						}

						jHeaderLogoResponses = append(jHeaderLogoResponses, jHeaderLogoResponse)
					}

					dataLogHeaderLogo(jHeaderLogoResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				} else {
					errorMessage = "Method undifined!"
					dataLogHeaderLogo(jHeaderLogoResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}
			}
		}
	}).Finally(func() {
	}).Catch(func(e try.E) {
		// Print crash
		errorMessage := "Error, catch"
		dataLogHeaderLogo(jHeaderLogoResponses, jHeaderLogoRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	})
}

func dataLogHeaderLogo(jHeaderLogoResponses []JHeaderLogoResponse, username string, errorCode string, errorMessage string, totalRecords float64, totalPage float64, method string, path string, ip string, logData string, allHeader string, bodyJson string, c *gin.Context) {
	if errorCode != "0" {
		helper.SendLogError(username, "HEADER LOGO", errorMessage, bodyJson, "", errorCode, allHeader, method, path, ip, c)
	}
	returnHeaderLogo(jHeaderLogoResponses, errorCode, errorMessage, logData, totalRecords, totalPage, c)
}

func returnHeaderLogo(jHeaderLogoResponses []JHeaderLogoResponse, errorCode string, errorMessage string, logData string, totalRecords float64, totalPage float64, c *gin.Context) {

	if strings.Contains(errorMessage, "Error running") {
		errorMessage = "Error Execute data"
	}

	if errorCode == "504" {
		c.String(http.StatusUnauthorized, "")
	} else {
		currentTime := time.Now()
		currentTime1 := currentTime.Format("01/02/2006 15:04:05")

		c.PureJSON(http.StatusOK, gin.H{
			"ErrorCode":    errorCode,
			"ErrorMessage": errorMessage,
			"DateTime":   currentTime1,
			"TotalRecords":   totalRecords,
			"TotalPage":   totalPage,
			"Result": jHeaderLogoResponses, 
		})
	}

	startTime := time.Now()

	rex := regexp.MustCompile(`\r?\n`)
	endTime := time.Now()
	codeError := "200"

	diff := endTime.Sub(startTime)

	logDataNew := rex.ReplaceAllString(logData + codeError + "~" + endTime.String() + "~" + diff.String() + "~" + errorMessage, "")
	log.Println(logDataNew)

	runtime.GC()
}
