package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

// Forward สร้าง reverse proxy ไปยัง upstream URL
func Forward(target string) gin.HandlerFunc {
	targetURL, err := url.Parse(target)
	if err != nil {
		panic("invalid upstream url: " + target)
	}

	rp := httputil.NewSingleHostReverseProxy(targetURL)

	// custom error handler
	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error":"upstream service unavailable"}`))
	}

	return func(c *gin.Context) {
		rp.ServeHTTP(c.Writer, c.Request)
	}
}
