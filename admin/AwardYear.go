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

type JListAwardsYearRequest struct {
	Username string
	ParamKey string
	Method string
	Year string
	Page        int
	RowPage     int
	OrderBy     string
	Order       string
}

type JListAwardsYearResponse struct {
	Year string
}

func AwardsYear(c *gin.Context) {
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

	jListAwardsYearRequest := JListAwardsYearRequest{}
	jListAwardsYearResponse := JListAwardsYearResponse{}
	jListAwardsYearResponses := []JListAwardsYearResponse{}

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
	logFile = logFile + "AwardsYear_" + dateNow + ".log"
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
		dataLogAwardsYear(jListAwardsYearResponses, jListAwardsYearRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage = "Error, Body - invalid json data"
		dataLogAwardsYear(jListAwardsYearResponses, jListAwardsYearRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	try.This(func() {
		if helper.ValidateHeader(bodyString, c) {
			if err := c.ShouldBindJSON(&jListAwardsYearRequest); err != nil {
				errorMessage = "Error, Bind Json Data"
				dataLogAwardsYear(jListAwardsYearResponses, jListAwardsYearRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			} else {
				username := jListAwardsYearRequest.Username
				paramKey := jListAwardsYearRequest.ParamKey
				method := jListAwardsYearRequest.Method
				year := jListAwardsYearRequest.Year
				page := jListAwardsYearRequest.Page
				rowPage := jListAwardsYearRequest.RowPage
				orderBy := jListAwardsYearRequest.OrderBy
				order := jListAwardsYearRequest.Order

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
					dataLogAwardsYear(jListAwardsYearResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}
				// ------ end of Param Validation ------

				// ------ start check session paramkey ------
				checkAccessVal := helper.CheckSession(username, paramKey, c)
				if checkAccessVal != "1" {
					dataLogAwardsYear(jListAwardsYearResponses, username, errorCodeSession, errorMessageSession, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				if method == "INSERT" {

				} else if method == "UPDATE" {

				} else if method == "DELETE" {

				} else if method == "SELECT" {

					pageNow := (page - 1) * rowPage
					pageNowString := strconv.Itoa(pageNow)
					queryLimit := ""

					queryWhere := ""
					if year != "" {
						if queryWhere != "" {
							queryWhere += " AND "
						}
						
						queryWhere += fmt.Sprintf(" year = '%s' ", year)
					}

					if queryWhere != "" {
						queryWhere = " WHERE " + queryWhere
					}

					queryGroupBy := " GROUP BY year "

					queryOrder := ""
					if orderBy != "" {
						queryOrder = fmt.Sprintf(" ORDER BY %s %s", orderBy, order)
					} else {
						queryOrder = ""
					}

					totalRecords = 0
					totalPage = 0
					query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM lippo_award %s %s", queryWhere, queryGroupBy)
					if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
						errorMessage = "Error running, " + err.Error()
						dataLogAwardsYear(jListAwardsYearResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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

					query1 := fmt.Sprintf(`SELECT year FROM lippo_award %s %s %s %s`, queryWhere, queryGroupBy, queryOrder, queryLimit)
					rows, err := db.Query(query1)
					defer rows.Close()
					if err != nil {
						errorMessage = "Error running, " + err.Error()
						dataLogAwardsYear(jListAwardsYearResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					for rows.Next() {
						err = rows.Scan(
							&jListAwardsYearResponse.Year,
						)

						if err != nil {
							errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err)
							dataLogAwardsYear(jListAwardsYearResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
							return
						}

						jListAwardsYearResponses = append(jListAwardsYearResponses, jListAwardsYearResponse)
					}

					dataLogAwardsYear(jListAwardsYearResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return

				} else {
					errorMessage = "Method undifined!"
					dataLogAwardsYear(jListAwardsYearResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}
			}
		}
	}).Finally(func() {
	}).Catch(func(e try.E) {
		// Print crash
		errorMessage := "Error, catch"
		dataLogAwardsYear(jListAwardsYearResponses, jListAwardsYearRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	})
}

func dataLogAwardsYear(jListAwardsYearResponses []JListAwardsYearResponse, username string, errorCode string, errorMessage string, totalRecords float64, totalPage float64, method string, path string, ip string, logData string, allHeader string, bodyJson string, c *gin.Context) {
	if errorCode != "0" {
		helper.SendLogError(username, "AwardsYear", errorMessage, bodyJson, "", errorCode, allHeader, method, path, ip, c)
	}
	returnAwardsYear(jListAwardsYearResponses, errorCode, errorMessage, logData, totalRecords, totalPage, c)
}

func returnAwardsYear(jListAwardsYearResponses []JListAwardsYearResponse, errorCode string, errorMessage string, logData string, totalRecords float64, totalPage float64, c *gin.Context) {

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
			"Result": jListAwardsYearResponses, 
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
