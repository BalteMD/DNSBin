package dns

import (
	"context"
	db "dnsbin/db/sqlc"
	"dnsbin/ipwry"
	"dnsbin/notify"
	"dnsbin/util"
	"fmt"
	"log"
	"net"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
)

type DNSServer struct {
	config   util.Config
	store    db.Store
	telegram *notify.Telegram
}

func NewDNSServer(config util.Config, store db.Store, telegram *notify.Telegram) (*DNSServer, error) {
	dnsServer := &DNSServer{
		config:   config,
		store:    store,
		telegram: telegram,
	}

	return dnsServer, nil
}

// Start starts the DNS server and listens for incoming requests.
func (dnsServer *DNSServer) Start() error {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 53})
	if err != nil {
		return fmt.Errorf("failed to start DNS server: %w", err)
	}
	defer conn.Close()

	log.Println("DNS server started on port 53...")
	for {
		buf := make([]byte, 512)
		_, addr, _ := conn.ReadFromUDP(buf)
		var msg dnsmessage.Message
		if err := msg.Unpack(buf); err != nil {
			log.Printf("Error reading UDP packet: %v", err)
			continue
		}

		go dnsServer.handleRequest(addr, conn, msg)
	}
}

func (dnsServer *DNSServer) handleRequest(addr *net.UDPAddr, conn *net.UDPConn, msg dnsmessage.Message) {
	if len(msg.Questions) < 1 {
		log.Println("No questions in DNS request")
		return
	}

	question := msg.Questions[0]
	var (
		queryNameStr = strings.ToLower(question.Name.String())
		queryType    = question.Type
		queryName, _ = dnsmessage.NewName(queryNameStr)
		resource     dnsmessage.Resource
	)

	if strings.Contains(queryNameStr, dnsServer.config.DNSDomain) {
		location, _ := ipwry.Query(addr.IP.String())
		dnsLog, err := dnsServer.store.CreateDNSLog(context.Background(), db.CreateDNSLogParams{
			DnsQueryRecord: queryNameStr[:len(queryNameStr)-1],
			Type:           "DNS",
			Location:       location,
			IpAddress:      addr.IP.String(),
		})
		if err != nil {
			log.Printf("Failed to save DNS log: %v", err)
		}

		err = dnsServer.telegram.SendMarkdown("*🔥 Detected "+dnsLog.Type+" log*", notify.RenderDNSInfo(&notify.DNSLogInfo{
			Type:           dnsLog.Type,
			DnsQueryRecord: dnsLog.DnsQueryRecord,
			Location:       location,
			IpAddress:      dnsLog.IpAddress,
			Time:           dnsLog.CreatedAt,
		}))
		if err != nil {
			log.Printf("Failed to send markdown log: %v", err)
		}
	}

	switch queryType {
	case dnsmessage.TypeA:
		resource = NewAResource(queryName, [4]byte{127, 0, 0, 1})
	case dnsmessage.TypeTXT:
		if dnsServer.config.TXTValue != "" {
			resource = dnsmessage.Resource{
				Header: dnsmessage.ResourceHeader{
					Name:  queryName,
					Class: dnsmessage.ClassINET,
					TTL:   0,
				},
				Body: &dnsmessage.TXTResource{
					TXT: []string{dnsServer.config.TXTValue},
				},
			}
		} else {
			resource = dnsmessage.Resource{
				Header: dnsmessage.ResourceHeader{
					Name:  queryName,
					Class: dnsmessage.ClassINET,
					TTL:   0,
				},
				Body: &dnsmessage.TXTResource{
					TXT: []string{""},
				},
			}
		}
	default:
		resource = NewAResource(queryName, [4]byte{127, 0, 0, 1})
	}

	// send response
	msg.Response = true
	msg.Answers = append(msg.Answers, resource)
	Response(addr, conn, msg)
}

// Response return
func Response(addr *net.UDPAddr, conn *net.UDPConn, msg dnsmessage.Message) {
	packed, err := msg.Pack()
	if err != nil {
		log.Printf("Failed to pack DNS response: %v", err)
		return
	}
	if _, err := conn.WriteToUDP(packed, addr); err != nil {
		log.Printf("Failed to send DNS response: %v", err)
	}
}

func NewAResource(query dnsmessage.Name, a [4]byte) dnsmessage.Resource {
	return dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  query,
			Class: dnsmessage.ClassINET,
			TTL:   0,
		},
		Body: &dnsmessage.AResource{
			A: a,
		},
	}
}
