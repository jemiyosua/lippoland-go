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

type JListCompanyOverviewRequest struct {
	Username string
	ParamKey string
	Method string
	Id string
	FileNameImage string
	FileNameImageOld string
	Base64DataImage string
	Title string
	Description string
	Status string
	Page        int
	RowPage     int
	OrderBy     string
	Order       string
}

type JListCompanyOverviewResponse struct {
	Id string
	Images string
	FileNameImage string
	Title string
	Description string
	Status string
	TanggalInput string
}

func CompanyOverview(c *gin.Context) {
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

	jListCompanyOverviewRequest := JListCompanyOverviewRequest{}
	jListCompanyOverviewResponse := JListCompanyOverviewResponse{}
	jListCompanyOverviewResponses := []JListCompanyOverviewResponse{}

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
	logFile = logFile + "CompanyOverview_" + dateNow + ".log"
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
		dataLogCompanyOverview(jListCompanyOverviewResponses, jListCompanyOverviewRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage = "Error, Body - invalid json data"
		dataLogCompanyOverview(jListCompanyOverviewResponses, jListCompanyOverviewRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	try.This(func() {
		if helper.ValidateHeader(bodyString, c) {
			if err := c.ShouldBindJSON(&jListCompanyOverviewRequest); err != nil {
				errorMessage = "Error, Bind Json Data"
				dataLogCompanyOverview(jListCompanyOverviewResponses, jListCompanyOverviewRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			} else {
				username := jListCompanyOverviewRequest.Username
				paramKey := jListCompanyOverviewRequest.ParamKey
				method := jListCompanyOverviewRequest.Method
				id := jListCompanyOverviewRequest.Id
				filenameImage := jListCompanyOverviewRequest.FileNameImage
				filenameImageOld := jListCompanyOverviewRequest.FileNameImageOld
				base64Image := jListCompanyOverviewRequest.Base64DataImage
				title := jListCompanyOverviewRequest.Title
				description := jListCompanyOverviewRequest.Description
				status := jListCompanyOverviewRequest.Status
				page := jListCompanyOverviewRequest.Page
				rowPage := jListCompanyOverviewRequest.RowPage
				orderBy := jListCompanyOverviewRequest.OrderBy
				order := jListCompanyOverviewRequest.Order

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
					dataLogCompanyOverview(jListCompanyOverviewResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}
				// ------ end of Param Validation ------

				// ------ start check session paramkey ------
				checkAccessVal := helper.CheckSession(username, paramKey, c)
				if checkAccessVal != "1" {
					dataLogCompanyOverview(jListCompanyOverviewResponses, username, errorCodeSession, errorMessageSession, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				if method == "INSERT" {

					if (filenameImage == "" || base64Image == "") {
						errorMessage += "Image Filename or Base64 can't null value"
					}

					if (title == "") {
						errorMessage += "Title can't null value"
					}

					if (description == "") {
						errorMessage += "Description can't null value"
					}

					if errorMessage != "" {
						dataLogCompanyOverview(jListCompanyOverviewResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					imageFilename, errorCodeImage, errorMessageImage := helper.CreateImageUrl(method, "", filenameImage, base64Image, db, c)
					if errorCodeImage != "0" {
						dataLogCompanyOverview(jListCompanyOverviewResponses, username, errorCodeImage, errorMessageImage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					query1 := fmt.Sprintf("INSERT INTO lippo_desc_about_us (images, title, description, status, tgl_input) VALUES ('%s', '%s', '%s', %d, 1, NOW());", imageFilename, title, description)
					_, err1 := db.Exec(query1)
					if err1 != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err1)
						dataLogCompanyOverview(jListCompanyOverviewResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					dataLogCompanyOverview(jListCompanyOverviewResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return

				} else if method == "UPDATE" {

					if (id == "" || filenameImageOld == "") {
						errorMessage = "Hero Id or filename image old can't null value"
						dataLogCompanyOverview(jListCompanyOverviewResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					} 

					queryUpdate := ""
					if (filenameImage != "" && base64Image != "") {

						imageFilename, errorCodeImage, errorMessageImage := helper.CreateImageUrl(method, filenameImageOld, filenameImage, base64Image, db, c)
						if errorCodeImage != "0" {
							dataLogCompanyOverview(jListCompanyOverviewResponses, username, errorCodeImage, errorMessageImage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
							return
						}

						queryUpdate += fmt.Sprintf(" , images = '%s' ", imageFilename)
					}

					if (title != "") {
						queryUpdate += fmt.Sprintf(" , title = '%s' ", title)
					}

					if (description != "") {
						queryUpdate += fmt.Sprintf(" , description = '%s' ", description)
					}

					if (status != "") {
						queryUpdate += fmt.Sprintf(" , status = '%s' ", status)
					}

					

					query := fmt.Sprintf("UPDATE lippo_desc_about_us SET tgl_input = NOW() %s WHERE id = '%s'", queryUpdate, id)
					_, err := db.Exec(query)
					if err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
						dataLogCompanyOverview(jListCompanyOverviewResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}
					
					dataLogCompanyOverview(jListCompanyOverviewResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return

				} else if method == "DELETE" {

					if (id == "") {
						errorMessage = "Hero Id can't null value"
						dataLogCompanyOverview(jListCompanyOverviewResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}
 
					query := fmt.Sprintf("DELETE FROM lippo_desc_about_us WHERE id = '%s'", id)
					_, err := db.Exec(query)
					if err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
						dataLogCompanyOverview(jListCompanyOverviewResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					dataLogCompanyOverview(jListCompanyOverviewResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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

					if description != "" {
						if queryWhere != "" {
							queryWhere += " AND "
						}
						
						queryWhere += fmt.Sprintf(" description LIKE '%%%s%%' ", description)
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
					query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM lippo_desc_about_us %s", queryWhere)
					if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
						errorMessage = "Error running, " + err.Error()
						dataLogCompanyOverview(jListCompanyOverviewResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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

					query1 := fmt.Sprintf(`SELECT id, images, title, description, status, tgl_input FROM lippo_desc_about_us %s %s %s`, queryWhere, queryOrder, queryLimit)
					rows, err := db.Query(query1)
					defer rows.Close()
					if err != nil {
						errorMessage = "Error running, " + err.Error()
						dataLogCompanyOverview(jListCompanyOverviewResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					for rows.Next() {
						err = rows.Scan(
							&jListCompanyOverviewResponse.Id,
							&jListCompanyOverviewResponse.Images,
							&jListCompanyOverviewResponse.Title,
							&jListCompanyOverviewResponse.Description,
							&jListCompanyOverviewResponse.Status,
							&jListCompanyOverviewResponse.TanggalInput,
						)

						jListCompanyOverviewResponse.FileNameImage = jListCompanyOverviewResponse.Images
						jListCompanyOverviewResponse.Images = urlImages + jListCompanyOverviewResponse.Images

						if err != nil {
							errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err)
							dataLogCompanyOverview(jListCompanyOverviewResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
							return
						}

						jListCompanyOverviewResponses = append(jListCompanyOverviewResponses, jListCompanyOverviewResponse)
					}

					dataLogCompanyOverview(jListCompanyOverviewResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return

				} else {
					errorMessage = "Method undifined!"
					dataLogCompanyOverview(jListCompanyOverviewResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}
			}
		}
	}).Finally(func() {
	}).Catch(func(e try.E) {
		// Print crash
		errorMessage := "Error, catch"
		dataLogCompanyOverview(jListCompanyOverviewResponses, jListCompanyOverviewRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	})
}

func dataLogCompanyOverview(jListCompanyOverviewResponses []JListCompanyOverviewResponse, username string, errorCode string, errorMessage string, totalRecords float64, totalPage float64, method string, path string, ip string, logData string, allHeader string, bodyJson string, c *gin.Context) {
	if errorCode != "0" {
		helper.SendLogError(username, "COMPANY OVERVIEW", errorMessage, bodyJson, "", errorCode, allHeader, method, path, ip, c)
	}
	returnCompanyOverview(jListCompanyOverviewResponses, errorCode, errorMessage, logData, totalRecords, totalPage, c)
}

func returnCompanyOverview(jListCompanyOverviewResponses []JListCompanyOverviewResponse, errorCode string, errorMessage string, logData string, totalRecords float64, totalPage float64, c *gin.Context) {

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
			"Result": jListCompanyOverviewResponses, 
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
