package articles

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Wlczak/blogfinity/database"
	"github.com/Wlczak/blogfinity/database/models"
	"github.com/Wlczak/blogfinity/logger"
	"github.com/ikeikeikeike/go-sitemap-generator/v2/stm"
)

func HandleSitemap(w http.ResponseWriter, r *http.Request) {
	zap := logger.GetLogger()

	sm := stm.NewSitemap(1)
	sitemapDomain := os.Getenv("BASE_DOMAIN")
	if len(sitemapDomain) == 0 {
		http.Error(w, "Sitemap domain not set", http.StatusInternalServerError)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		zap.Error(err.Error())
	}

	sm.SetDefaultHost(sitemapDomain)

	sm.Create()

	sm.Add(stm.URL{
		{"loc", "/"},
		{"changefreq", "monthly"},
		{"priority", "1.0"},
	})
	sm.Add(stm.URL{
		{"loc", "/search"},
		{"changefreq", "monthly"},
		{"priority", "0.9"},
	})
	articles := models.GetArticles(db, 500)
	for article := range articles {
		sm.Add(stm.URL{
			{"loc", "/article/" + fmt.Sprint(articles[article].ID)},
			{"changefreq", "monthly"},
			{"priority", "0.6"},
		})
	}

	xml := sm.XMLContent()

	w.Header().Set("Content-Type", "application/xml")
	w.Write(xml)
}
