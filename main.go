package main

import (
	"fmt"
	Admin "lippoland/admin"
	Web "lippoland/web"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}
}

func main() {
	router := gin.Default()
	router.Use(CORS())

	apiVersion := "/api/v1/"
	adminEndpoint := apiVersion + "admin/"
	webEndpoint := apiVersion + "web/"

	// ---------- admin ----------
	router.POST(adminEndpoint+"Login", Admin.Login)
	router.POST(adminEndpoint+"Awards", Admin.Awards)
	router.POST(adminEndpoint+"AwardsYear", Admin.AwardsYear)
	router.POST(adminEndpoint+"CategoryDevelopments", Admin.CategoryDevelopments)
	router.POST(adminEndpoint+"CompanyOverview", Admin.CompanyOverview)
	router.POST(adminEndpoint+"CoreValues", Admin.CoreValues)
	router.POST(adminEndpoint+"DevelopmentSection", Admin.DevelopmentSection)
	router.POST(adminEndpoint+"HeaderLogo", Admin.HeaderLogo)
	router.POST(adminEndpoint+"HeaderMenu", Admin.HeaderMenu)
	router.POST(adminEndpoint+"HeroHome", Admin.HeroHome)
	router.POST(adminEndpoint+"HeaderWaNumber", Admin.HeaderWhatsappNumber)
	router.POST(adminEndpoint+"LeadershipInitiative", Admin.LeadershipInitiative)
	router.POST(adminEndpoint+"LeftMenu", Admin.LeftMenu)
	router.POST(adminEndpoint+"MasterProduk", Admin.MasterProduk)
	router.POST(adminEndpoint+"ProductHome", Admin.ProductHome)
	router.POST(adminEndpoint+"Statistic", Admin.Statistic)
	router.POST(adminEndpoint+"UpcomingProject", Admin.UpcomingProject)
	router.POST(adminEndpoint+"VisionMision", Admin.VisionMision)
	router.GET(adminEndpoint+"/Images/:NamaFile", Admin.Images)
	// ---------- end of admin ----------

	// ---------- web ----------
	router.POST(webEndpoint+"AboutUsDesc", Web.AboutUsDesc)
	router.POST(webEndpoint+"AboutUsHero", Web.AboutUsHero)
	router.POST(webEndpoint+"CoreValues", Web.CoreValues)
	router.POST(webEndpoint+"HeaderLogo", Web.HeaderLogo)
	router.POST(webEndpoint+"LeaderInitiative", Web.LeaderInitiative)
	router.POST(webEndpoint+"ListAward", Web.ListAward)
	router.POST(webEndpoint+"ListCategoryDevHome", Web.ListCategoryDevHome)
	router.POST(webEndpoint+"ListHeaderMenu", Web.ListHeaderMenu)
	router.POST(webEndpoint+"ListHeroHome", Web.ListHeroHome)
	router.POST(webEndpoint+"ListImageDevHome", Web.ListImageDevHome)
	router.POST(webEndpoint+"ListProductHome", Web.ListProductHome)
	router.POST(webEndpoint+"ListStats", Web.ListStats)
	router.POST(webEndpoint+"UpcomingProject", Web.UpcomingProject)
	router.POST(webEndpoint+"VisiMisi", Web.VisiMisi)
	// ---------- end of web ----------

	PORT := os.Getenv("PORT")

	router.Run(":" + PORT)
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Signature, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "SAMEORIGIN")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
