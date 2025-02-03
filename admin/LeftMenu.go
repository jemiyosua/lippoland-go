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

type JLeftMenuRequest struct {
	Username 	string
	ParamKey 	string
	Method 		string
	Menu 		string
	Status 	string
	Page        int
	RowPage     int
	OrderBy     string
	Order       string
}

type JLeftMenuResponse struct {
	Id string
	MenuNameId string
	MenuNameEn string
	Status 		 string
	TanggalInput string
	Item []JLeftMenuSubResponse
}

type JLeftMenuSubResponse struct {
	Id string
	MenuNameId string
	MenuNameEn string
	UrlPage string
	Status string
	TanggalInput string
}

func LeftMenu(c *gin.Context) {
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

	jLeftMenuRequest := JLeftMenuRequest{}
	jLeftMenuResponse := JLeftMenuResponse{}
	jLeftMenuResponses := []JLeftMenuResponse{}

	errorCode := "1"
	errorMessage := ""
	errorCodeSession := "2"
	errorMessageSession := "Session Expired"

	allHeader := helper.ReadAllHeader(c)
	logFile = os.Getenv("LOGFILE_ADMIN")
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
	logFile = logFile + "Menu_" + dateNow + ".log"
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
		dataLogMenu(jLeftMenuResponses, jLeftMenuRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage = "Error, Body - invalid json data"
		dataLogMenu(jLeftMenuResponses, jLeftMenuRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	try.This(func() {
	if helper.ValidateHeader(bodyString, c) {
		if err := c.ShouldBindJSON(&jLeftMenuRequest); err != nil {
			errorMessage = "Error, Bind Json Data"
			dataLogMenu(jLeftMenuResponses, jLeftMenuRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
			return
		} else {
			username := jLeftMenuRequest.Username
			paramKey := jLeftMenuRequest.ParamKey
			method := jLeftMenuRequest.Method
			menu := jLeftMenuRequest.Menu
			status := jLeftMenuRequest.Status
			page := jLeftMenuRequest.Page
			rowPage := jLeftMenuRequest.RowPage

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

			if page == 0 {
				errorMessage += "Page can't null or 0 value"
			}

			if rowPage == 0 {
				errorMessage += "Page can't null or 0 value"
			}

			if errorMessage != "" {
				dataLogMenu(jLeftMenuResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			}
			// ------ end of Param Validation ------

			// ------ start check session paramkey ------
			checkAccessVal := helper.CheckSession(username, paramKey, c)
			if checkAccessVal != "1" {
				dataLogMenu(jLeftMenuResponses, username, errorCodeSession, errorMessageSession, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			}

			if method == "INSERT" {

			} else if method == "UPDATE" {

			} else if method == "DELETE" {

			} else if method == "SELECT" {
				pageNow := (page - 1) * rowPage
				pageNowString := strconv.Itoa(pageNow)
				queryLimit := ""

				queryWhere := " status = 1 "
				if menu != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += fmt.Sprintf(" menu LIKE '%%%s%%' ", menu)
				}

				if status != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}
					
					queryWhere += fmt.Sprintf(" status = '%s' ", status)
				}

				if queryWhere != "" {
					queryWhere = " WHERE " + queryWhere
				}

				totalRecords = 0
				totalPage = 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM lippo_login_menu_parent %s", queryWhere)
				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage = "Error running, " + err.Error()
					dataLogMenu(jLeftMenuResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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

				// ---------- start query get menu ----------
				query1 := fmt.Sprintf("SELECT id, menu_name_id, menu_name_en, status, tgl_input FROM lippo_login_menu_parent %s %s", queryWhere, queryLimit)
				rows, err := db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage = "Error running, " + err.Error()
					dataLogMenu(jLeftMenuResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				for rows.Next() {
					err = rows.Scan(
						&jLeftMenuResponse.Id,
						&jLeftMenuResponse.MenuNameId,
						&jLeftMenuResponse.MenuNameEn,
						&jLeftMenuResponse.Status,
						&jLeftMenuResponse.TanggalInput,
					)

					if err != nil {
						errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err)
						dataLogMenu(jLeftMenuResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					jLeftMenuSubResponse := JLeftMenuSubResponse{}
					jLeftMenuSubResponses := []JLeftMenuSubResponse{}

					query1 := fmt.Sprintf("SELECT id, menu_name_id, menu_name_en, url, status, tgl_input FROM lippo_login_menu_parent_sub WHERE status = 1 AND parent_id = '%s'", jLeftMenuResponse.Id)
					rows, err := db.Query(query1)
					defer rows.Close()
					if err != nil {
						errorMessage = "Error running, " + err.Error()
						dataLogMenu(jLeftMenuResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					for rows.Next() {
						err = rows.Scan(
							&jLeftMenuSubResponse.Id,
							&jLeftMenuSubResponse.MenuNameId,
							&jLeftMenuSubResponse.MenuNameEn,
							&jLeftMenuSubResponse.UrlPage,
							&jLeftMenuSubResponse.Status,
							&jLeftMenuSubResponse.TanggalInput,
						)

						if err != nil {
							errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err)
							dataLogMenu(jLeftMenuResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
							return
						}

						jLeftMenuSubResponses = append(jLeftMenuSubResponses, jLeftMenuSubResponse)	
					}

					jLeftMenuResponse.Item = jLeftMenuSubResponses
					jLeftMenuResponses = append(jLeftMenuResponses, jLeftMenuResponse)

					
				}
				// ---------- end of query get menu ----------

				dataLogMenu(jLeftMenuResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			} else {
				errorMessage = "Method undifined!"
				dataLogMenu(jLeftMenuResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			}
		}
	}
	}).Finally(func() {
	}).Catch(func(e try.E) {
		// Print crash
		errorMessage := "Error, catch"
		dataLogMenu(jLeftMenuResponses, jLeftMenuRequest.Username, "1", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	})
}

func dataLogMenu(jLeftMenuResponses []JLeftMenuResponse, username string, errorCode string, errorMessage string, totalRecords float64, totalPage float64, method string, path string, ip string, logData string, allHeader string, bodyJson string, c *gin.Context) {
	if errorCode != "0" {
		helper.SendLogError(username, "LEFT MENU", errorMessage, bodyJson, "", errorCode, allHeader, method, path, ip, c)
	}
	returnMenu(jLeftMenuResponses, errorCode, errorMessage, logData, totalRecords, totalPage, c)
}

func returnMenu(jLeftMenuResponses []JLeftMenuResponse, errorCode string, errorMessage string, logData string, totalRecords float64, totalPage float64, c *gin.Context) {

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
			"Result": jLeftMenuResponses, 
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
