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

type JListStatisticRequest struct {
	Username string
	ParamKey string
	Method string
	Id string
	FileNameImage string
	FileNameImageOld string
	Base64DataImage string
	Title string
	Status string
	ParamUpdate string
	Page        int
	RowPage     int
	OrderBy     string
	Order       string
}

type JListStatisticResponse struct {
	Id string
	Images string
	FileNameImage string
	Title string
	Status string
	TanggalInput string
}

func Statistic(c *gin.Context) {
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

	jListStatisticRequest := JListStatisticRequest{}
	jListStatisticResponse := JListStatisticResponse{}
	jListStatisticResponses := []JListStatisticResponse{}

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
	logFile = logFile + "Statistic_" + dateNow + ".log"
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
		dataLogStatistic(jListStatisticResponses, jListStatisticRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage = "Error, Body - invalid json data"
		dataLogStatistic(jListStatisticResponses, jListStatisticRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	try.This(func() {
		if helper.ValidateHeader(bodyString, c) {
			if err := c.ShouldBindJSON(&jListStatisticRequest); err != nil {
				errorMessage = "Error, Bind Json Data"
				dataLogStatistic(jListStatisticResponses, jListStatisticRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			} else {
				username := jListStatisticRequest.Username
				paramKey := jListStatisticRequest.ParamKey
				method := jListStatisticRequest.Method
				id := jListStatisticRequest.Id
				filenameImage := jListStatisticRequest.FileNameImage
				filenameImageOld := jListStatisticRequest.FileNameImageOld
				base64Image := jListStatisticRequest.Base64DataImage
				title := jListStatisticRequest.Title
				status := jListStatisticRequest.Status
				paramUpdate := jListStatisticRequest.ParamUpdate
				page := jListStatisticRequest.Page
				rowPage := jListStatisticRequest.RowPage
				orderBy := jListStatisticRequest.OrderBy
				order := jListStatisticRequest.Order

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
					dataLogStatistic(jListStatisticResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}
				// ------ end of Param Validation ------

				// ------ start check session paramkey ------
				checkAccessVal := helper.CheckSession(username, paramKey, c)
				if checkAccessVal != "1" {
					dataLogStatistic(jListStatisticResponses, username, errorCodeSession, errorMessageSession, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				if method == "INSERT" {

					if (filenameImage == "" || base64Image == "") {
						errorMessage += "Image Filename or Base64 can't null value"
					}

					if (title == "") {
						errorMessage += "Title can't null value"
					}

					if errorMessage != "" {
						dataLogStatistic(jListStatisticResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					imageFilename, errorCodeImage, errorMessageImage := helper.CreateImageUrl(method, "", filenameImage, base64Image, db, c)
					if errorCodeImage != "0" {
						dataLogStatistic(jListStatisticResponses, username, errorCodeImage, errorMessageImage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					query1 := fmt.Sprintf("INSERT INTO lippo_stats (images, title, status, tgl_input) VALUES ('%s', '%s', 1, NOW());", imageFilename, title)
					_, err1 := db.Exec(query1)
					if err1 != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err1)
						dataLogStatistic(jListStatisticResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					dataLogStatistic(jListStatisticResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return

				} else if method == "UPDATE" {

					if (id == "") {
						errorMessage = "Id can't null value"
						dataLogStatistic(jListStatisticResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					} 

					queryUpdate := ""
					if paramUpdate == "status" {
						if (status != "") {
							queryUpdate += fmt.Sprintf(" , status = '%s' ", status)
						}
					} else {
						if (filenameImageOld == "") {
							errorMessage = "filename image old can't null value"
							dataLogStatistic(jListStatisticResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
							return
						}

						if (filenameImage != "" && base64Image != "") {
							imageFilename, errorCodeImage, errorMessageImage := helper.CreateImageUrl(method, filenameImageOld, filenameImage, base64Image, db, c)
							if errorCodeImage != "0" {
								dataLogStatistic(jListStatisticResponses, username, errorCodeImage, errorMessageImage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
								return
							}
	
							queryUpdate += fmt.Sprintf(" , images = '%s' ", imageFilename)
						}
	
						if (title != "") {
							queryUpdate += fmt.Sprintf(" , title = '%s' ", title)
						}
					}

					query := fmt.Sprintf("UPDATE lippo_stats SET tgl_input = NOW() %s WHERE id = '%s'", queryUpdate, id)
					_, err := db.Exec(query)
					if err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
						dataLogStatistic(jListStatisticResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}
					
					dataLogStatistic(jListStatisticResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return

				} else if method == "DELETE" {

					if (id == "") {
						errorMessage = "Id can't null value"
						dataLogStatistic(jListStatisticResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}
 
					query := fmt.Sprintf("DELETE FROM lippo_stats WHERE id = '%s'", id)
					_, err := db.Exec(query)
					if err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
						dataLogStatistic(jListStatisticResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					dataLogStatistic(jListStatisticResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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
						queryOrder = " ORDER BY tgl_input DESC "
					}

					totalRecords = 0
					totalPage = 0
					query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM lippo_stats %s", queryWhere)
					if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
						errorMessage = "Error running, " + err.Error()
						dataLogStatistic(jListStatisticResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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

					query1 := fmt.Sprintf(`SELECT id, images, title, status, tgl_input FROM lippo_stats %s %s %s`, queryWhere, queryOrder, queryLimit)
					rows, err := db.Query(query1)
					defer rows.Close()
					if err != nil {
						errorMessage = "Error running, " + err.Error()
						dataLogStatistic(jListStatisticResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					for rows.Next() {
						err = rows.Scan(
							&jListStatisticResponse.Id,
							&jListStatisticResponse.Images,
							&jListStatisticResponse.Title,
							&jListStatisticResponse.Status,
							&jListStatisticResponse.TanggalInput,
						)

						jListStatisticResponse.FileNameImage = jListStatisticResponse.Images
						jListStatisticResponse.Images = urlImages + jListStatisticResponse.Images

						if err != nil {
							errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err)
							dataLogStatistic(jListStatisticResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
							return
						}

						jListStatisticResponses = append(jListStatisticResponses, jListStatisticResponse)
					}

					dataLogStatistic(jListStatisticResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return

				} else {
					errorMessage = "Method undifined!"
					dataLogStatistic(jListStatisticResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}
			}
		}
	}).Finally(func() {
	}).Catch(func(e try.E) {
		// Print crash
		errorMessage := "Error, catch"
		dataLogStatistic(jListStatisticResponses, jListStatisticRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	})
}

func dataLogStatistic(jListStatisticResponses []JListStatisticResponse, username string, errorCode string, errorMessage string, totalRecords float64, totalPage float64, method string, path string, ip string, logData string, allHeader string, bodyJson string, c *gin.Context) {
	if errorCode != "0" {
		helper.SendLogError(username, "STATISTIC", errorMessage, bodyJson, "", errorCode, allHeader, method, path, ip, c)
	}
	returnStatistic(jListStatisticResponses, errorCode, errorMessage, logData, totalRecords, totalPage, c)
}

func returnStatistic(jListStatisticResponses []JListStatisticResponse, errorCode string, errorMessage string, logData string, totalRecords float64, totalPage float64, c *gin.Context) {

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
			"Result": jListStatisticResponses, 
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
