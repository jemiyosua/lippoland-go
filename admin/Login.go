package admin

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
	_ "github.com/go-sql-driver/mysql"
	"github.com/manucorporat/try"
)

type JLoginRequest struct {
	Username string
	Password string
}

func Login(c *gin.Context) {
	db := helper.Connect(c)
	defer db.Close()
	startTime := time.Now()
	startTimeString := startTime.String()

	var (
		bodyBytes []byte
		xRealIp   string
		ip        string
		logFile   string
		paramKey string
		role string
	)

	jLoginRequest := JLoginRequest{}

	errorCode := "1"
	errorMessage := ""

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
	logFile = logFile + "Login_" + dateNow + ".log"
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
		dataLogLogin(jLoginRequest.Username, paramKey, role, errorCode, errorMessage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage = "Error, Body - invalid json data"
		dataLogLogin(jLoginRequest.Username, paramKey, role, errorCode, errorMessage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	try.This(func() {
		if helper.ValidateHeader(bodyString, c) {
			if err := c.ShouldBindJSON(&jLoginRequest); err != nil {
				errorMessage = "Error, Bind Json Data"
				dataLogLogin(jLoginRequest.Username, paramKey, role, errorCode, errorMessage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			} else {
				username := jLoginRequest.Username
				password := jLoginRequest.Password
				errorMessage = ""
	
				// ------ Param Validation ------
				if username == "" {
					errorMessage += "Username can't null value"
				}
	
				if password == "" {
					errorMessage += "Password can't null value"
				}
	
				if errorMessage != "" {
					dataLogLogin(username, paramKey, role, errorCode, errorMessage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}
				// ------ end of Param Validation ------
	
				countLogin := 0
				passwordDB := ""
				statusUser := 0
				role := ""
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt, IFNULL(password, '') password, IFNULL(status, 0) status, IFNULL(role, '') role FROM lippo_login WHERE username = '%s' GROUP BY password, status, role;", username)
				if err := db.QueryRow(query).Scan(&countLogin, &passwordDB, &statusUser, &role); err != nil && err != sql.ErrNoRows {
					errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
					dataLogLogin(username, paramKey, role, errorCode, errorMessage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}
	
				if countLogin == 0 {
					errorMessage = "Username not registered!"
					dataLogLogin(username, paramKey, role, errorCode, errorMessage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				} else {
					if statusUser == 0 {
						errorMessage = "Your account is inactivated, please call administrator!"
						dataLogLogin(username, paramKey, role, errorCode, errorMessage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					} else {
						if password != passwordDB {
							errorMessage = "Password not match!"
							dataLogLogin(username, paramKey, role, errorCode, errorMessage, method, path, ip, logData, allHeader, bodyJson, c)
							return
						} else {
							paramKey = helper.Token()
	
							query := fmt.Sprintf("INSERT INTO lippo_login_session (username, paramkey, tgl_input) VALUES ('%s','%s', ADDTIME(NOW(), '0:20:0'))", username, paramKey)
							_, err := db.Exec(query)
							if err != nil {
								paramKey = ""
								errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
								dataLogLogin(username, paramKey, role, errorCode, errorMessage, method, path, ip, logData, allHeader, bodyJson, c)
								return
							}
	
							dataLogLogin(username, paramKey, role, "0", errorMessage, method, path, ip, logData, allHeader, bodyJson, c)
							return
						}
					}
				}
			}
		}
	}).Finally(func() {
	}).Catch(func(e try.E) {
		// Print crash
		errorMessage := "Error, catch"
		dataLogLogin(jLoginRequest.Username, paramKey, role, "1", errorMessage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	})
}

func dataLogLogin(username string, paramKey string, role string, errorCode string, errorMessage string, method string, path string, ip string, logData string, allHeader string, bodyJson string, c *gin.Context) {
	if errorCode != "0" {
		helper.SendLogError(username, "LOGIN", errorMessage, bodyJson, "", errorCode, allHeader, method, path, ip, c)
	}
	returnLogin(username, paramKey, role, errorCode, errorMessage, logData, c)
	return
}

func returnLogin(username string, paramKey string, role string, errorCode string, errorMessage string, logData string, c *gin.Context) {

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
			"Username":   username,
			"ParamKey":   paramKey,
			"Role":       role,
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

	return
}
