package ip2region

import (
	"strings"

	"github.com/aaabigfish/gopkg/net/ip"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

type IpInfo struct {
	Country   string
	CountryId int64
	Region    string
	Province  string
	City      string
	ISP       string
}

type Ip2Region struct {
	region *xdb.Searcher
}

func New(path string) (*Ip2Region, error) {
	// 1、从 dbPath 加载整个 xdb 到内存
	cBuff, err := xdb.LoadContentFromFile(path)
	if err != nil {
		return nil, err
	}

	// 2、用全局的 cBuff 创建完全基于内存的查询对象。
	searcher, err := xdb.NewWithBuffer(cBuff)
	if err != nil {
		return nil, err
	}

	return &Ip2Region{region: searcher}, nil
}

func ParseIpInfo(regionStr string) *IpInfo {
	lineSlice := strings.Split(regionStr, "|")
	ipInfo := IpInfo{}
	length := len(lineSlice)
	if length < 5 {
		for i := 0; i <= 5-length; i++ {
			lineSlice = append(lineSlice, "")
		}
	}

	ipInfo.Country = lineSlice[0]
	ipInfo.Region = lineSlice[1]
	ipInfo.Province = lineSlice[2]
	ipInfo.City = lineSlice[3]
	ipInfo.ISP = lineSlice[4]

	country := ipInfo.Country
	if ipInfo.Country == "中国" {
		if ipInfo.Province == "香港" || ipInfo.Province == "台湾省" || ipInfo.Province == "澳门" {
			if ipInfo.Province == "台湾省" {
				ipInfo.Province = "台湾"
			}
			country = ipInfo.Province
		}
	}

	if id, ok := ip.Country[country]; ok {
		ipInfo.CountryId = id
	}

	return &ipInfo
}

func (s *Ip2Region) GetIpInfo(ipStr string) (*IpInfo, error) {
	regionStr, err := s.region.SearchByStr(ipStr)
	if err != nil {
		return nil, err
	}

	return ParseIpInfo(regionStr), nil
}
