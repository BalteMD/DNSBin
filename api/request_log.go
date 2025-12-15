package api

import (
	db "dnsbin/db/sqlc"
	"dnsbin/ipwry"
	"dnsbin/notify"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (server *Server) httpRequestLog(ctx *gin.Context) {
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

	dnsLog, err := server.store.CreateDNSLog(ctx, db.CreateDNSLogParams{
		Type:           "HTTP",
		DnsQueryRecord: ctx.Request.URL.Path,
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
			"ip":       dnsLog.IpAddress,
			"location": dnsLog.Location,
		},
	})
}
