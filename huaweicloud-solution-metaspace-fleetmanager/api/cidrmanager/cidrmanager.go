// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// cidr管理模块
package cidrmanager

import (
	"crypto/rand"
	"fleetmanager/api/errors"
	"fleetmanager/db/dao"
	"fleetmanager/setting"
	"fmt"
	"github.com/google/uuid"
	"math"
	"math/big"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	NetMask32     = 32
	NetMask24     = 24
	NetMask16     = 16
	NetMask8      = 8
	IPv4len       = 4
	MaxRetryTimes = 10
)

var (
	ipnetFrom     *net.IPNet
	ipnetTo       *net.IPNet
	vpcNetmask    int
	subnetNetmask int
	allocatedMap  AllocatedMap
)

type AllocatedMap struct {
	sync.Mutex
	Map map[string]IpRangeMap
}

type IpRangeMap struct {
	sync.Mutex
	Map map[uint32]string
}

// Init: Cidr管理类初始化函数. 用于基于配置文件初始化可用的cidr范围. 如果报错说明配置错误
func Init() error {
	cidrFrom := setting.Config.Get(setting.FleetCidrFrom).ToString("")
	cidrTo := setting.Config.Get(setting.FleetCidrTo).ToString("")
	vpcNetmask = setting.Config.Get(setting.FleetCidrVpcNetmask).ToInt(0)
	subnetNetmask = setting.Config.Get(setting.FleetCidrSubnetNetmask).ToInt(0)
	// 判断子网掩码配置是否正确, subnet掩码应当在vpc掩码范围内, 且subnet掩码应当小于32
	if cidrFrom == "" || cidrTo == "" || vpcNetmask == 0 || subnetNetmask == 0 || subnetNetmask < vpcNetmask ||
		subnetNetmask >= NetMask32 {
		return errors.NewErrorF(errors.ServerInternalError, " service init error, invalid fleet vpc cidr config")
	}

	var err error
	_, ipnetFrom, err = net.ParseCIDR(cidrFrom + "/" + strconv.Itoa(vpcNetmask))
	if err != nil {
		return err
	}

	_, ipnetTo, err = net.ParseCIDR(cidrTo + "/" + strconv.Itoa(vpcNetmask))
	if err != nil {
		return err
	}

	allocatedMap = AllocatedMap{Map: make(map[string]IpRangeMap)}

	// 初始化allocatedMap
	if err := loadAllocatedCidr(); err != nil {
		return err
	}

	return nil
}

func ipToUInt32(ip net.IP) (sum uint32, err error) {
	bits := strings.Split(ip.String(), ".")
	if len(bits) != IPv4len {
		err = fmt.Errorf("invalid ip: %v", ip)
		return
	}

	b0, e1 := strconv.Atoi(bits[0])
	b1, e2 := strconv.Atoi(bits[1])
	b2, e3 := strconv.Atoi(bits[2])
	b3, e4 := strconv.Atoi(bits[3])
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
		err = fmt.Errorf("invalid ip: %v", ip)
		return
	}

	sum += uint32(b0) << NetMask24
	sum += uint32(b1) << NetMask16
	sum += uint32(b2) << NetMask8
	sum += uint32(b3)
	return
}

func uint32ToIp(intIp uint32) net.IP {
	var bytes [4]byte
	bytes[0] = byte(intIp & 0xFF)
	bytes[1] = byte((intIp >> 8) & 0xFF)
	bytes[2] = byte((intIp >> 16) & 0xFF)
	bytes[3] = byte((intIp >> 24) & 0xFF)
	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0])
}

func toIPNet(intIp uint32, mask int) (*net.IPNet, error) {
	ip := uint32ToIp(intIp)
	ipCidr := ip.String() + "/" + strconv.Itoa(mask)
	_, ipnet, e := net.ParseCIDR(ipCidr)
	return ipnet, e
}

func getAvailableIntIpRange(fromIp net.IP, toIp net.IP, mask int) (uint32, uint32, error) {
	fromInt, e := ipToUInt32(fromIp)
	if e != nil {
		return 0, 0, e
	}

	toInt, e := ipToUInt32(toIp)
	if e != nil {
		return 0, 0, e
	}

	return fromInt, (toInt - fromInt) / (1 << uint32(NetMask32-mask)), nil
}

func makeRange(min, max uint32) []uint32 {
	a := make([]uint32, max-min+1)
	for i := range a {
		a[i] = min + uint32(i)
	}

	return a
}

func calAllocatedIntIpRange(fromIp net.IP, toIp net.IP, mask int, destIp net.IP, destMask uint32) ([]uint32, error) {
	fromInt, availableRange, e := getAvailableIntIpRange(fromIp, toIp, mask)
	if e != nil {
		return []uint32{}, e
	}

	destIntBegin, e := ipToUInt32(destIp)
	if e != nil {
		return []uint32{}, e
	}

	destIntEnd := destIntBegin + (1 << (NetMask32 - destMask))
	maskIpNum := 1 << uint32(NetMask32-mask)
	rangeBegin := uint32(math.Floor(float64(destIntBegin-fromInt) / float64(maskIpNum)))
	rangeEnd := uint32(math.Ceil(float64(destIntEnd-fromInt) / float64(maskIpNum)))
	if rangeBegin > availableRange || rangeEnd < 0 {
		return []uint32{}, fmt.Errorf("no available ip range")
	}

	return makeRange(rangeBegin, rangeEnd), nil
}

func loadAllocatedCidr() error {
	fvc, e := dao.GetAllFleetVpcCidr()
	if e != nil {
		return e
	}

	for _, value := range fvc {
		if e := addFleetVpcCidrToMap(&value); e != nil {
			return e
		}
	}

	return nil
}

func checkCidrDuplicated(allocatedIpRanges IpRangeMap, randInt uint32) bool {
	allocatedIpRanges.Lock()
	_, isAllocated := allocatedIpRanges.Map[randInt]
	allocatedIpRanges.Unlock()
	return isAllocated
}

func generateValidCidr(namespace string) (*net.IPNet, error) {
	if ipnetFrom == nil || ipnetTo == nil {
		return nil, fmt.Errorf("invalid ipnetFrom or ipnetTo")
	}

	// 获取总可分配范围
	fromInt, maxIpInt, e := getAvailableIntIpRange(ipnetFrom.IP, ipnetTo.IP, vpcNetmask)
	if e != nil {
		return nil, e
	}

	// 获取已经分配的地址
	allocatedIpRanges, ok := allocatedMap.Map[namespace]
	for {
		bigInt, err := rand.Int(rand.Reader, new(big.Int).SetUint64(uint64(maxIpInt)))
		if err != nil {
			return nil, err
		}
		randInt := uint32(bigInt.Uint64())
		if ok {
			if checkCidrDuplicated(allocatedIpRanges, randInt) {
				// 继续随机
				continue
			} else {
				return toIPNet(fromInt+uint32(randInt*(1<<uint32(NetMask32-vpcNetmask))), vpcNetmask)
			}
		} else {
			return toIPNet(fromInt+uint32(randInt*(1<<uint32(NetMask32-vpcNetmask))), vpcNetmask)
		}
	}
}

func generateFleetVpcCidr(namespace string, fleetId string) (*dao.FleetVpcCidr, error) {
	// 生成随机vpc cidr
	ipnet, e := generateValidCidr(namespace)
	if e != nil {
		return nil, e
	}

	ipCidr := ipnet.IP.String() + "/" + strconv.Itoa(vpcNetmask)

	// 构造数据库记录
	u, _ := uuid.NewUUID()
	vpcCidrId := u.String()
	fvc := dao.FleetVpcCidr{
		Id:         vpcCidrId,
		VpcCidr:    ipCidr,
		FleetId:    fleetId,
		Namespace:  namespace,
		CreateTime: time.Now(),
	}

	return &fvc, nil
}

func updateAllocatedIpRanges(fleetId string, allocatedIpRanges IpRangeMap, allocatedRange []uint32) {
	defer allocatedIpRanges.Unlock()
	allocatedIpRanges.Lock()

	for _, tmp := range allocatedRange {
		allocatedIpRanges.Map[tmp] = fleetId
	}
}

func updateAllocatedMap(namespace string, fleetId string, allocatedRange []uint32) {
	defer allocatedMap.Unlock()
	allocatedMap.Lock()

	allocatedIpRanges, ok := allocatedMap.Map[namespace]
	// 如果存在, 则直接插入
	if !ok {
		// 如果不存在, 则构造后插入
		allocatedIpRanges = IpRangeMap{Map: make(map[uint32]string)}
	}

	updateAllocatedIpRanges(fleetId, allocatedIpRanges, allocatedRange)
}

func addFleetVpcCidrToMap(cidr *dao.FleetVpcCidr) error {
	// 转换ip到range
	_, allocatedIp, e := net.ParseCIDR(cidr.VpcCidr)
	if e != nil {
		return e
	}

	allocatedMask, _ := allocatedIp.Mask.Size()
	if ipnetFrom == nil || ipnetTo == nil {
		return fmt.Errorf("invalid ipnetFrom or ipnetTo")
	}

	allocatedRange, e := calAllocatedIntIpRange(ipnetFrom.IP, ipnetTo.IP, vpcNetmask, allocatedIp.IP,
		uint32(allocatedMask))
	if e != nil {
		return e
	}

	updateAllocatedMap(cidr.Namespace, cidr.FleetId, allocatedRange)
	return nil
}

// CreateVpcCidr: 获取系统可用的vpc cidr, 会保证为每个fleet在namespace下获取不重复的vpc cidr
func CreateVpcCidr(namespace string, fleetId string) (string, error) {
	var fvc *dao.FleetVpcCidr
	var e error
	var retryTimes = 0
	for {
		fvc, e = generateFleetVpcCidr(namespace, fleetId)
		if e != nil {
			return "", e
		}
		if e = dao.InsertFleetVpcCidr(fvc); e != nil {
			// TODO(wangjun):判断错误类型, 重复插入才重试, 否增返回错误
			if e = addFleetVpcCidrToMap(fvc); e != nil {
				return "", e
			}
			retryTimes++
			if retryTimes > MaxRetryTimes {
				break
			}
			continue
		} else {
			if e = addFleetVpcCidrToMap(fvc); e != nil {
				return "", e
			}
			break
		}
	}

	if fvc == nil {
		return "", fmt.Errorf("no invalid vpc cidr")
	}

	return fvc.VpcCidr, nil
}

// CreateSubnetCidr: 基于vpcCidr获取可用的子网Cidr, 默认会根据配置文件配置获取子网的掩码, 并会获取第一个IP作为网关IP
func CreateSubnetCidr(vpcCidr string) (string, string, error) {
	// 基于vpc cidr生成subnet cidr
	_, ipnet, err := net.ParseCIDR(vpcCidr)
	if err != nil {
		return "", "", err
	}

	subnetCidr := ipnet.IP.String() + "/" + strconv.Itoa(subnetNetmask)
	strList := strings.Split(ipnet.IP.String(), ".")
	if len(strList) != IPv4len {
		return "", "", fmt.Errorf("invalid ip: %v", ipnet)
	}

	gatewayIp := strList[0] + "." + strList[1] + "." + strList[2] + "." + "1"
	return subnetCidr, gatewayIp, nil
}
