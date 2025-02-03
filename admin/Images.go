package admin

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"lippoland/helper"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Images(c *gin.Context) {

	db := helper.Connect(c)
	defer db.Close()

	var (
		startTime    time.Time
		logfile      string
		errorMessage string
	)

	logfile = os.Getenv("LOGFILE_ADMIN")
	urlImages := os.Getenv("URLIMAGES")

	log.SetFlags(0)

	// ------ start log file ------
	startTime = time.Now()
	// dateNow := startTime.Format("2006-01-02")

	logFILE := logfile + "StarPoinWebLog.log"

	file, err := os.OpenFile(logFILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	log.SetOutput(file)
	// ------ end log file ------

	startTimeStr := startTime.String()

	// ------ start Read Header ------
	allHeader := helper.ReadAllHeader(c)
	// ------ end Read Header ------

	method := c.Request.Method
	path := c.Request.URL.EscapedPath()

	var XRealIp string
	if values, _ := c.Request.Header["X-Real-Ip"]; len(values) > 0 {
		XRealIp = values[0]
	}

	var ip string
	if XRealIp != "" {
		ip = XRealIp
	} else {
		ip = c.ClientIP()
	}

	logData := startTimeStr + "~" + ip + "~" + method + "~" + path + "~" + allHeader + "~"

	NamaFile := c.Param("NamaFile")

	body := `<!DOCTYPE html>
	<html style="height:100%">
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no" >
		<title> 404 Not Found</title>
	</head>
	<body style="color: #444; margin:0;font: normal 14px/20px Arial, Helvetica, sans-serif; height:100%; background-color: #fff;">
		<div style="height:auto; min-height:100%; ">
			<div style="text-align: center; width:800px; margin-left: -400px; position:absolute; top: 30%; left:50%;">
				<h1 style="margin:0; font-size:150px; line-height:150px; font-weight:bold;">404</h1>
				<h2 style="margin-top:20px;font-size: 30px;">Not Found</h2>
				<p>The resource requested could not be found on this server!</p>
			</div>
		</div>
	</body>
	</html> `

	if NamaFile == "" {
		c.Header("Content-Type", "text/html")
		c.String(200, body)
		return
	}

	errorMessage = ""
	query := ""
	dataImage := ""
	cnt := ""
	query = fmt.Sprintf("SELECT COUNT(1) AS cnt , CONVERT(Base64Data using utf8) AS dataImage FROM lippo_images WHERE filename = '%s' GROUP BY dataImage;", NamaFile)
	if err := db.QueryRow(query).Scan(&cnt, &dataImage); err != nil {
		imgFile, err := os.Open(urlImages + NamaFile)
		if err != nil {
			errorMessage = "Error, Select Data starpoin_image" + " | " + query
			dataLogImages("", logData, ip, allHeader, NamaFile, "1", "1", errorMessage, errorMessage, c)
			c.Writer.Header().Add("Content-Type", "text/html")
			c.String(200, body)
			return
		}

		defer imgFile.Close()

		// create a new buffer base on file size
		fInfo, _ := imgFile.Stat()
		var size int64 = fInfo.Size()
		buf := make([]byte, size)

		// read file content into buffer
		fReader := bufio.NewReader(imgFile)
		fReader.Read(buf)

		dataImage = base64.StdEncoding.EncodeToString(buf)
	}

	if cnt == "0" {
		imgFile, err := os.Open(urlImages + NamaFile)
		if err != nil {
			errorMessage = "Error, Select Data starpoin_image" + " | " + query
			dataLogImages("", logData, ip, allHeader, NamaFile, "1", "1", errorMessage, errorMessage, c)
			c.Writer.Header().Add("Content-Type", "text/html")
			c.String(200, body)
			return
		}

		defer imgFile.Close()

		// create a new buffer base on file size
		fInfo, _ := imgFile.Stat()
		var size int64 = fInfo.Size()
		buf := make([]byte, size)

		// read file content into buffer
		fReader := bufio.NewReader(imgFile)
		fReader.Read(buf)

		dataImage = base64.StdEncoding.EncodeToString(buf)
	}

	body = dataImage

	dec, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		errorMessage = "Error, decode - data image " + " | " + query
		dataLogImages("", logData, ip, allHeader, NamaFile, "1", "1", errorMessage, errorMessage, c)

		c.Writer.Header().Add("Content-Type", "text/html")
		c.String(200, body)
		return

	}

	tempfile := os.Getenv("TEMPFILE")
	f, err := os.Create(tempfile + NamaFile)
	if err != nil {
		errorMessage = "Error, create file - data image " + " | " + query
		dataLogImages("", logData, ip, allHeader, NamaFile, "1", "1", errorMessage, errorMessage, c)
		c.Writer.Header().Add("Content-Type", "text/html")
		c.String(200, body)
		return
	}
	defer f.Close()

	if _, err := f.Write(dec); err != nil {
		errorMessage = "Error, Write to file - data image " + " | " + query
		dataLogImages("", logData, ip, allHeader, NamaFile, "1", "1", errorMessage, errorMessage, c)
		c.Writer.Header().Add("Content-Type", "text/html")
		c.String(200, body)
		return
	}
	if err := f.Sync(); err != nil {
		errorMessage = "Error, Sync - data image " + " | " + query
		dataLogImages("", logData, ip, allHeader, NamaFile, "1", "1", errorMessage, errorMessage, c)
		c.Writer.Header().Add("Content-Type", "text/html")
		c.String(200, body)
		return
	}

	dataLogImages("", logData, ip, allHeader, NamaFile, "0", "0", errorMessage, errorMessage, c)

	if strings.Contains(strings.ToUpper(NamaFile), "GIF") {
		c.Header("Content-Type", "image/gif")
	} else if strings.Contains(strings.ToUpper(NamaFile), "PNG") {
		c.Header("Content-Type", "image/png")
	} else {
		c.Header("Content-type", "image/jpeg")
	}

	c.File(tempfile + NamaFile)
	return

}

func dataLogImages(userId string, logData string, ip string, allHeader string, namaFile string, errorCode string, errorCodeReturn string, errorMessage string, errorMessageReturn string, c *gin.Context) {
	if errorCode != "0" {
		helper.SendLogError("", "IMAGES", errorMessage, "", "", errorCode, allHeader, "", "", ip, c)
	}
	returnDataGetImages(logData, errorCode, errorCodeReturn, errorMessage, errorMessageReturn)
	return
}

func returnDataGetImages(logData string, ErrorCode string, errorCodeReturn string, errorMessage string, errorMessageReturn string) {

	startTime := time.Now()

	rex := regexp.MustCompile(`\r?\n`)
	endTime := time.Now()
	codeError := "200"

	if errorMessage != "" {
		codeError = "500"
	}

	diff := endTime.Sub(startTime)

	logDataNew := rex.ReplaceAllString(logData+codeError+"~"+endTime.String()+"~"+diff.String()+"~"+errorMessage, "")
	log.Println(logDataNew)

	runtime.GC()

	return
}
