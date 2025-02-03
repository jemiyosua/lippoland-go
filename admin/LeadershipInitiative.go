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

type JListLeadershipInitiativeRequest struct {
	Username string
	ParamKey string
	Method string
	Id string
	Title string
	Description string
	Status string
	ParamUpdate string
	Page        int
	RowPage     int
	OrderBy     string
	Order       string
}

type JListLeadershipInitiativeResponse struct {
	Id string
	Title string
	Description string
	Status string
	TanggalInput string
}

func LeadershipInitiative(c *gin.Context) {
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

	jListLeadershipInitiativeRequest := JListLeadershipInitiativeRequest{}
	jListLeadershipInitiativeResponse := JListLeadershipInitiativeResponse{}
	jListLeadershipInitiativeResponses := []JListLeadershipInitiativeResponse{}

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
	logFile = logFile + "LeadershipInitiative_" + dateNow + ".log"
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
		dataLogLeadershipInitiative(jListLeadershipInitiativeResponses, jListLeadershipInitiativeRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage = "Error, Body - invalid json data"
		dataLogLeadershipInitiative(jListLeadershipInitiativeResponses, jListLeadershipInitiativeRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	try.This(func() {
		if helper.ValidateHeader(bodyString, c) {
			if err := c.ShouldBindJSON(&jListLeadershipInitiativeRequest); err != nil {
				errorMessage = "Error, Bind Json Data"
				dataLogLeadershipInitiative(jListLeadershipInitiativeResponses, jListLeadershipInitiativeRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			} else {
				username := jListLeadershipInitiativeRequest.Username
				paramKey := jListLeadershipInitiativeRequest.ParamKey
				method := jListLeadershipInitiativeRequest.Method
				id := jListLeadershipInitiativeRequest.Id
				title := jListLeadershipInitiativeRequest.Title
				description := jListLeadershipInitiativeRequest.Description
				status := jListLeadershipInitiativeRequest.Status
				paramUpdate := jListLeadershipInitiativeRequest.ParamUpdate
				page := jListLeadershipInitiativeRequest.Page
				rowPage := jListLeadershipInitiativeRequest.RowPage
				orderBy := jListLeadershipInitiativeRequest.OrderBy
				order := jListLeadershipInitiativeRequest.Order

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
					dataLogLeadershipInitiative(jListLeadershipInitiativeResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}
				// ------ end of Param Validation ------

				// ------ start check session paramkey ------
				checkAccessVal := helper.CheckSession(username, paramKey, c)
				if checkAccessVal != "1" {
					dataLogLeadershipInitiative(jListLeadershipInitiativeResponses, username, errorCodeSession, errorMessageSession, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				if method == "INSERT" {

				} else if method == "UPDATE" {

					if (id == "") {
						errorMessage = "Id can't null value"
						dataLogLeadershipInitiative(jListLeadershipInitiativeResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					queryUpdate := ""
					if paramUpdate == "status" {
						if (status != "") {
							queryUpdate += fmt.Sprintf(" , status = '%s' ", status)
						}
					} else {
						if (title != "") {
							queryUpdate += fmt.Sprintf(" , title = '%s' ", title)
						}

						if (description != "") {
							queryUpdate += fmt.Sprintf(" , description = '%s' ", description)
						}
					}

					query := fmt.Sprintf("UPDATE lippo_leader_initiative SET tgl_input = NOW() %s WHERE id = '%s'", queryUpdate, id)
					_, err := db.Exec(query)
					if err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
						dataLogLeadershipInitiative(jListLeadershipInitiativeResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}
					
					dataLogLeadershipInitiative(jListLeadershipInitiativeResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return

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

					if title != "" {
						if queryWhere != "" {
							queryWhere += " AND "
						}
						
						queryWhere += fmt.Sprintf(" title LIKE '%%%s%%' ", title)
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

					queryOrder := ""
					if orderBy != "" {
						queryOrder = fmt.Sprintf(" ORDER BY %s %s", orderBy, order)
					} else {
						queryOrder = ""
					}

					totalRecords = 0
					totalPage = 0
					query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM lippo_leader_initiative %s", queryWhere)
					if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
						errorMessage = "Error running, " + err.Error()
						dataLogLeadershipInitiative(jListLeadershipInitiativeResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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

					query1 := fmt.Sprintf("SELECT id, title, description, status, tgl_input FROM lippo_leader_initiative %s %s %s", queryWhere, queryOrder, queryLimit)
					rows, err := db.Query(query1)
					defer rows.Close()
					if err != nil {
						errorMessage = "Error running, " + err.Error()
						dataLogLeadershipInitiative(jListLeadershipInitiativeResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					for rows.Next() {
						err = rows.Scan(
							&jListLeadershipInitiativeResponse.Id,
							&jListLeadershipInitiativeResponse.Title,
							&jListLeadershipInitiativeResponse.Description,
							&jListLeadershipInitiativeResponse.Status,
							&jListLeadershipInitiativeResponse.TanggalInput,
						)

						if err != nil {
							errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err)
							dataLogLeadershipInitiative(jListLeadershipInitiativeResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
							return
						}

						jListLeadershipInitiativeResponses = append(jListLeadershipInitiativeResponses, jListLeadershipInitiativeResponse)
					}

					dataLogLeadershipInitiative(jListLeadershipInitiativeResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return

				} else {
					errorMessage = "Method undifined!"
					dataLogLeadershipInitiative(jListLeadershipInitiativeResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}
			}
		}
	}).Finally(func() {
	}).Catch(func(e try.E) {
		// Print crash
		errorMessage := "Error, catch"
		dataLogLeadershipInitiative(jListLeadershipInitiativeResponses, jListLeadershipInitiativeRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	})
}

func dataLogLeadershipInitiative(jListLeadershipInitiativeResponses []JListLeadershipInitiativeResponse, username string, errorCode string, errorMessage string, totalRecords float64, totalPage float64, method string, path string, ip string, logData string, allHeader string, bodyJson string, c *gin.Context) {
	if errorCode != "0" {
		helper.SendLogError(username, "LeadershipInitiative", errorMessage, bodyJson, "", errorCode, allHeader, method, path, ip, c)
	}
	returnLeadershipInitiative(jListLeadershipInitiativeResponses, errorCode, errorMessage, logData, totalRecords, totalPage, c)
}

func returnLeadershipInitiative(jListLeadershipInitiativeResponses []JListLeadershipInitiativeResponse, errorCode string, errorMessage string, logData string, totalRecords float64, totalPage float64, c *gin.Context) {

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
			"Result": jListLeadershipInitiativeResponses, 
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