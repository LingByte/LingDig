package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type DNSRecord struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	TTL   uint32 `json:"ttl"`
	Value string `json:"value"`
}

type DigService struct {
	client *dns.Client
}

func NewDigService() *DigService {
	return &DigService{
		client: &dns.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (d *DigService) Query(domain, recordType, server string) ([]DNSRecord, error) {
	// 创建DNS查询消息
	msg := new(dns.Msg)

	// 获取记录类型
	qtype, ok := dns.StringToType[strings.ToUpper(recordType)]
	if !ok {
		return nil, fmt.Errorf("不支持的记录类型: %s", recordType)
	}

	msg.SetQuestion(dns.Fqdn(domain), qtype)
	msg.RecursionDesired = true

	// 执行查询
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, _, err := d.client.ExchangeContext(ctx, msg, server)
	if err != nil {
		return nil, fmt.Errorf("DNS查询失败: %v", err)
	}

	if response.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("DNS查询返回错误代码: %s", dns.RcodeToString[response.Rcode])
	}

	// 解析响应
	var records []DNSRecord

	// 处理Answer部分
	for _, rr := range response.Answer {
		record := d.parseRecord(rr)
		if record != nil {
			records = append(records, *record)
		}
	}

	return records, nil
}

func (d *DigService) parseRecord(rr dns.RR) *DNSRecord {
	header := rr.Header()
	record := &DNSRecord{
		Name: header.Name,
		Type: dns.TypeToString[header.Rrtype],
		TTL:  header.Ttl,
	}

	switch v := rr.(type) {
	case *dns.A:
		record.Value = v.A.String()
	case *dns.AAAA:
		record.Value = v.AAAA.String()
	case *dns.CNAME:
		record.Value = v.Target
	case *dns.MX:
		record.Value = fmt.Sprintf("%d %s", v.Preference, v.Mx)
	case *dns.NS:
		record.Value = v.Ns
	case *dns.TXT:
		record.Value = strings.Join(v.Txt, " ")
	case *dns.SOA:
		record.Value = fmt.Sprintf("%s %s %d %d %d %d %d",
			v.Ns, v.Mbox, v.Serial, v.Refresh, v.Retry, v.Expire, v.Minttl)
	case *dns.PTR:
		record.Value = v.Ptr
	case *dns.SRV:
		record.Value = fmt.Sprintf("%d %d %d %s", v.Priority, v.Weight, v.Port, v.Target)
	default:
		record.Value = rr.String()
	}

	return record
}
