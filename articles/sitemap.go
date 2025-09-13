package articles

import (
	"fmt"
	"net/http"
	"os"

	"github.com/ikeikeikeike/go-sitemap-generator/v2/stm"
)

func HandleSitemap(w http.ResponseWriter, r *http.Request) {
	sm := stm.NewSitemap(1)
	sitemapDomain := os.Getenv("BASE_DOMAIN")
	if len(sitemapDomain) == 0 {
		http.Error(w, "Sitemap domain not set", http.StatusInternalServerError)
		return
	}
	sm.Create()

	fmt.Println("Generating sitemap for " + sitemapDomain)

	sm.SetDefaultHost(sitemapDomain)
	sm.Add(stm.URL{
		{"loc", "/"},
	})
	// sm.Add("https://" + sitemapDomain + "/search")
	xml := sm.Finalize().XMLContent()

	fmt.Println(xml)

	// w.Header().Set("Content-Type", "application/xml")
	// w.Write(xml)
}
