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

type JCategoryDevelopmentsRequest struct {
	Username string
	ParamKey string
	Method string
	ParamUpdate string
	Id string
	Name string
	Status string
	Page        int
	RowPage     int
	OrderBy     string
	Order       string
}

type JCategoryDevelopmentsResponse struct {
	Id string
	Name string
	Status string
	TanggalInput string
}

func CategoryDevelopments(c *gin.Context) {
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

	jCategoryDevelopmentsRequest := JCategoryDevelopmentsRequest{}
	jCategoryDevelopmentsResponse := JCategoryDevelopmentsResponse{}
	jCategoryDevelopmentsResponses := []JCategoryDevelopmentsResponse{}

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
	logFile = logFile + "CategoryDevelopments_" + dateNow + ".log"
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
		dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, jCategoryDevelopmentsRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage = "Error, Body - invalid json data"
		dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, jCategoryDevelopmentsRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	try.This(func() {
		if helper.ValidateHeader(bodyString, c) {
			if err := c.ShouldBindJSON(&jCategoryDevelopmentsRequest); err != nil {
				errorMessage = "Error, Bind Json Data"
				dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, jCategoryDevelopmentsRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			} else {
				username := jCategoryDevelopmentsRequest.Username
				paramKey := jCategoryDevelopmentsRequest.ParamKey
				method := jCategoryDevelopmentsRequest.Method
				paramUpdate := jCategoryDevelopmentsRequest.ParamUpdate
				id := jCategoryDevelopmentsRequest.Id
				name := jCategoryDevelopmentsRequest.Name
				status := jCategoryDevelopmentsRequest.Status
				page := jCategoryDevelopmentsRequest.Page
				rowPage := jCategoryDevelopmentsRequest.RowPage

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
					dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}
				// ------ end of Param Validation ------

				// ------ start check session paramkey ------
				checkAccessVal := helper.CheckSession(username, paramKey, c)
				if checkAccessVal != "1" {
					dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, username, errorCodeSession, errorMessageSession, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				if method == "INSERT" {

					if name == "" {
						errorMessage += "Name can't null value"
					} else {
						countCategoryName := 0
						query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM lippo_category_dev WHERE UPPER(name) = UPPER('%s')", name)
						if err := db.QueryRow(query).Scan(&countCategoryName); err != nil {
							errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
							dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
							return
						}

						if countCategoryName > 0 {
							errorMessage += "Name can't duplicate"
						}
					}

					if errorMessage != "" {
						dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					query1 := fmt.Sprintf("INSERT INTO lippo_category_dev (name, status, tgl_input) VALUES ('%s', 1, NOW());", name)
					_, err1 := db.Exec(query1)
					if err1 != nil {
						errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err1)
						dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return

				} else if method == "UPDATE" {

					if id == "" {
						errorMessage += "Id can't null value"
						dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					queryUpdate := ""
					if paramUpdate == "status" {
						queryUpdate += fmt.Sprintf(" , status = '%s' ", status)
					} else {
						countCategoryName := 0
						query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM lippo_category_dev WHERE UPPER(name) = UPPER('%s')", name)
						if err := db.QueryRow(query).Scan(&countCategoryName); err != nil {
							errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
							dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
							return
						}

						if countCategoryName > 0 {
							errorMessage += "Name can't duplicate"
						} else {
							if name != "" {
								queryUpdate = fmt.Sprintf(" , name = '%s' ", name)
							}
						}
					}

					if errorMessage != "" {
						dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					query1 := fmt.Sprintf("UPDATE lippo_category_dev SET tgl_input = NOW() %s WHERE id = '%s'", queryUpdate, id)
					_, err1 := db.Exec(query1)
					if err1 != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err1)
						dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return

				} else if method == "DELETE" {

					if id == "" {
						errorMessage += "Id can't null value"
						dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					query1 := fmt.Sprintf("DELETE FROM lippo_category_dev WHERE id = '%s'", id)
					_, err1 := db.Exec(query1)
					if err1 != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err1)
						dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return

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

					if name != "" {
						if queryWhere != "" {
							queryWhere += " AND "
						}
						
						queryWhere += fmt.Sprintf(" name LIKE '%%%s%%' ", name)
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
					query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM lippo_category_dev %s", queryWhere)
					if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
						errorMessage = "Error running, " + err.Error()
						dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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
					query1 := fmt.Sprintf(`SELECT id, name, status, tgl_input FROM lippo_category_dev %s %s`, queryWhere, queryLimit)
					rows, err := db.Query(query1)
					defer rows.Close()
					if err != nil {
						errorMessage = "Error running, " + err.Error()
						dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					for rows.Next() {
						err = rows.Scan(
							&jCategoryDevelopmentsResponse.Id,
							&jCategoryDevelopmentsResponse.Name,
							&jCategoryDevelopmentsResponse.Status,
							&jCategoryDevelopmentsResponse.TanggalInput,
						)

						jCategoryDevelopmentsResponses = append(jCategoryDevelopmentsResponses, jCategoryDevelopmentsResponse)

						if err != nil {
							errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err)
							dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
							return
						}
					}
					// ---------- end of query get menu ----------

					dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				} else {
					errorMessage = "Method undifined!"
					dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}
			}
		}
	}).Finally(func() {
	}).Catch(func(e try.E) {
		// Print crash
		errorMessage := "Error, catch"
		dataLogCategoryDevelopments(jCategoryDevelopmentsResponses, jCategoryDevelopmentsRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	})
}

func dataLogCategoryDevelopments(jCategoryDevelopmentsResponses []JCategoryDevelopmentsResponse, username string, errorCode string, errorMessage string, totalRecords float64, totalPage float64, method string, path string, ip string, logData string, allHeader string, bodyJson string, c *gin.Context) {
	if errorCode != "0" {
		helper.SendLogError(username, "CATEGORY DEVELOPMENTS", errorMessage, bodyJson, "", errorCode, allHeader, method, path, ip, c)
	}
	returnCategoryDevelopments(jCategoryDevelopmentsResponses, errorCode, errorMessage, logData, totalRecords, totalPage, c)
}

func returnCategoryDevelopments(jCategoryDevelopmentsResponses []JCategoryDevelopmentsResponse, errorCode string, errorMessage string, logData string, totalRecords float64, totalPage float64, c *gin.Context) {

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
			"Result": jCategoryDevelopmentsResponses, 
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
