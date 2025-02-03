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

type JDevelopmentSectionRequest struct {
	Username string
	ParamKey string
	Method string
	Id string
	CategoryId string
	Title string
	Description string
	Status string
	FileNameImage string
	FileNameImageOld string
	Base64DataImage string
	Page        int
	RowPage     int
	OrderBy     string
	Order       string
}

type JDevelopmentSectionResponse struct {
	Id string
	CategoryId string
	Images string
	FileNameImage string
	Title string
	Description string
	OrderDev string
	Status string
	TanggalInput string
}

func DevelopmentSection(c *gin.Context) {
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

	jDevelopmentSectionRequest := JDevelopmentSectionRequest{}
	jDevelopmentSectionResponse := JDevelopmentSectionResponse{}
	jDevelopmentSectionResponses := []JDevelopmentSectionResponse{}

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
	logFile = logFile + "DevelopmentSection_" + dateNow + ".log"
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
		dataLogDevelopmentSection(jDevelopmentSectionResponses, jDevelopmentSectionRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage = "Error, Body - invalid json data"
		dataLogDevelopmentSection(jDevelopmentSectionResponses, jDevelopmentSectionRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	try.This(func() {
		if helper.ValidateHeader(bodyString, c) {
			if err := c.ShouldBindJSON(&jDevelopmentSectionRequest); err != nil {
				errorMessage = "Error, Bind Json Data"
				dataLogDevelopmentSection(jDevelopmentSectionResponses, jDevelopmentSectionRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			} else {
				username := jDevelopmentSectionRequest.Username
				paramKey := jDevelopmentSectionRequest.ParamKey
				method := jDevelopmentSectionRequest.Method
				id := jDevelopmentSectionRequest.Id
				categoryId := jDevelopmentSectionRequest.CategoryId
				title := jDevelopmentSectionRequest.Title
				description := jDevelopmentSectionRequest.Description
				status := jDevelopmentSectionRequest.Status
				filenameImage := jDevelopmentSectionRequest.FileNameImage
				filenameImageOld := jDevelopmentSectionRequest.FileNameImageOld
				base64DataImage := jDevelopmentSectionRequest.Base64DataImage
				page := jDevelopmentSectionRequest.Page
				rowPage := jDevelopmentSectionRequest.RowPage
				orderBy := jDevelopmentSectionRequest.OrderBy
				order := jDevelopmentSectionRequest.Order

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
					dataLogDevelopmentSection(jDevelopmentSectionResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}
				// ------ end of Param Validation ------

				// ------ start check session paramkey ------
				checkAccessVal := helper.CheckSession(username, paramKey, c)
				if checkAccessVal != "1" {
					dataLogDevelopmentSection(jDevelopmentSectionResponses, username, errorCodeSession, errorMessageSession, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				if method == "INSERT" {

					if categoryId == "" {
						errorMessage += "- Category Id can't null value"
					}
					
					if (filenameImage == "" || base64DataImage == "") {
						errorMessage += "- Image Filename or Base64 can't null value"
					}

					if (title == "") {
						errorMessage += "- Title can't null value"
					}

					if (description == "") {
						errorMessage += "- Description can't null value"
					}

					if errorMessage != "" {
						dataLogDevelopmentSection(jDevelopmentSectionResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					imageFilename, errorCodeImage, errorMessageImage := helper.CreateImageUrl(method, "", filenameImage, base64DataImage, db, c)
					if errorCodeImage != "0" {
						dataLogDevelopmentSection(jDevelopmentSectionResponses, username, errorCodeImage, errorMessageImage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					maxDevOrder := 0
					query := fmt.Sprintf("SELECT MAX(order_dev) AS order_dev FROM lippo_dev_section WHERE category_id = '%s'", categoryId)
					if err := db.QueryRow(query).Scan(&maxDevOrder); err != nil {
						errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
						dataLogDevelopmentSection(jDevelopmentSectionResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}
					maxDevOrder += 1

					query1 := fmt.Sprintf("INSERT INTO lippo_dev_section (category_id, images, title, description, order_dev, status, tgl_input) VALUES ('%s', '%s', '%s', '%s', %d, 1, NOW());", categoryId, imageFilename, title, description, maxDevOrder)
					_, err1 := db.Exec(query1)
					if err1 != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err1)
						dataLogDevelopmentSection(jDevelopmentSectionResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					dataLogDevelopmentSection(jDevelopmentSectionResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return

				} else if method == "UPDATE" {

					if (id == "" || filenameImageOld == "") {
						errorMessage = "Id or filename image old can't null value"
						dataLogDevelopmentSection(jDevelopmentSectionResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					} 

					queryUpdate := ""
					if (filenameImage != "" && base64DataImage != "") {

						imageFilename, errorCodeImage, errorMessageImage := helper.CreateImageUrl(method, filenameImageOld, filenameImage, base64DataImage, db, c)
						if errorCodeImage != "0" {
							dataLogDevelopmentSection(jDevelopmentSectionResponses, username, errorCodeImage, errorMessageImage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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

					if (description != "") {
						queryUpdate += fmt.Sprintf(" , description = '%s' ", description)
					}

					if (status != "") {
						queryUpdate += fmt.Sprintf(" , status = '%s' ", status)
					}

					query := fmt.Sprintf("UPDATE lippo_dev_section SET tgl_input = NOW() %s WHERE id = '%s'", queryUpdate, id)
					_, err := db.Exec(query)
					if err != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
						dataLogDevelopmentSection(jDevelopmentSectionResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}
					
					dataLogDevelopmentSection(jDevelopmentSectionResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return

				} else if method == "DELETE" {

					if id == "" {
						errorMessage += "- Id can't null value"
					}

					if categoryId == "" {
						errorMessage += "- Category Id can't null value"
					} 

					if errorMessage != "" {
						dataLogDevelopmentSection(jDevelopmentSectionResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					query1 := fmt.Sprintf("DELETE FROM lippo_dev_section WHERE id = '%s' AND category_id = '%s'", id, categoryId)
					_, err1 := db.Exec(query1)
					if err1 != nil {
						errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err1)
						dataLogDevelopmentSection(jDevelopmentSectionResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					dataLogDevelopmentSection(jDevelopmentSectionResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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

					if categoryId != "" {
						if queryWhere != "" {
							queryWhere += " AND "
						}

						queryWhere += fmt.Sprintf(" category_id = '%s' ", categoryId)
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
					if queryOrder != "" {
						queryOrder = fmt.Sprintf(" ORDER BY %s %s", orderBy, order)
					} else {
						queryOrder = " ORDER BY order_dev ASC "
					}

					totalRecords = 0
					totalPage = 0
					query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM lippo_dev_section %s", queryWhere)
					if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
						errorMessage = "Error running, " + err.Error()
						dataLogDevelopmentSection(jDevelopmentSectionResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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

					query1 := fmt.Sprintf(`SELECT id, category_id, images, title, description, order_dev, status, tgl_input FROM lippo_dev_section %s %s`, queryWhere, queryLimit)
					rows, err := db.Query(query1)
					defer rows.Close()
					if err != nil {
						errorMessage = "Error running, " + err.Error()
						dataLogDevelopmentSection(jDevelopmentSectionResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					for rows.Next() {
						err = rows.Scan(
							&jDevelopmentSectionResponse.Id,
							&jDevelopmentSectionResponse.CategoryId,
							&jDevelopmentSectionResponse.Images,
							&jDevelopmentSectionResponse.Title,
							&jDevelopmentSectionResponse.Description,
							&jDevelopmentSectionResponse.OrderDev,
							&jDevelopmentSectionResponse.Status,
							&jDevelopmentSectionResponse.TanggalInput,
						)

						jDevelopmentSectionResponse.FileNameImage = jDevelopmentSectionResponse.Images
						jDevelopmentSectionResponse.Images = urlImages + jDevelopmentSectionResponse.Images

						if err != nil {
							errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err)
							dataLogDevelopmentSection(jDevelopmentSectionResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
							return
						}

						jDevelopmentSectionResponses = append(jDevelopmentSectionResponses, jDevelopmentSectionResponse)
					}

					dataLogDevelopmentSection(jDevelopmentSectionResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return

				} else {

					errorMessage = "Method undifined!"
					dataLogDevelopmentSection(jDevelopmentSectionResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return

				}
			}
		}
	}).Finally(func() {
	}).Catch(func(e try.E) {
		// Print crash
		errorMessage := "Error, catch"
		dataLogDevelopmentSection(jDevelopmentSectionResponses, jDevelopmentSectionRequest.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	})
}

func dataLogDevelopmentSection(jDevelopmentSectionResponses []JDevelopmentSectionResponse, username string, errorCode string, errorMessage string, totalRecords float64, totalPage float64, method string, path string, ip string, logData string, allHeader string, bodyJson string, c *gin.Context) {
	if errorCode != "0" {
		helper.SendLogError(username, "DEVELOPMENT SECTION", errorMessage, bodyJson, "", errorCode, allHeader, method, path, ip, c)
	}
	returnDevelopmentSection(jDevelopmentSectionResponses, errorCode, errorMessage, logData, totalRecords, totalPage, c)
}

func returnDevelopmentSection(jDevelopmentSectionResponses []JDevelopmentSectionResponse, errorCode string, errorMessage string, logData string, totalRecords float64, totalPage float64, c *gin.Context) {

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
			"Result": jDevelopmentSectionResponses, 
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
