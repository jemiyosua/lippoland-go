package helper

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func CreateImageUrl(method string, filenameOld string, filename string, base64Data string, db *sql.DB, c *gin.Context) (string, string, string) {

	randomString, _ := GenerateRandomString(4)
	dateTime := GetDate("ymdhis")
	imageId := dateTime + randomString
	filenameReplace := imageId + "_" + filename
	query := ""

	if method == "UPDATE" {
		query = fmt.Sprintf("UPDATE lippo_images SET filename = '%s', base64data = '%s' WHERE filename = '%s'", filenameReplace, base64Data, filenameOld)
	} else if method == "INSERT" {
		query = fmt.Sprintf("INSERT INTO lippo_images (filename, Base64Data) VALUES ('%s','%s');", filenameReplace, base64Data)
	}

	_, err := db.Exec(query)
	if err != nil {
		errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
		fmt.Println(errorMessage)
		return "", "1", errorMessage
	}

	return filenameReplace, "0", ""

}