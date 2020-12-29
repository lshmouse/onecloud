package models

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"yunion.io/x/pkg/errors"
	"yunion.io/x/pkg/util/netutils"

	"yunion.io/x/onecloud/pkg/multicloud/esxi"
)

func TestParseAndSuggest(t *testing.T) {
	param, err := structureTestData()
	if err != nil {
		t.Fatalf(err.Error())
	}
	out := CloudaccountManager.parseAndSuggestSingleWire(param)
	for _, net := range out.CAWireNets[0].GuestSuggestedNetworks {
		ips, _ := netutils.NewIPV4Addr(net.GuestIpStart)
		ipe, _ := netutils.NewIPV4Addr(net.GuestIpEnd)
		ipn := ips.NetAddr(24)
		if ips-ipn != 1 || ipe-ipn != 254 {
			t.Fatalf("wrong ip range '%s-%s' of suggest networks", net.GuestIpStart, net.GuestIpEnd)
		}
	}
}

func structureTestData() (sParseAndSuggest, error) {
	ret := sParseAndSuggest{}
	vlanips := make([]sVlanIPs, 0, 1)
	err := json.Unmarshal([]byte(_testData), &vlanips)
	if err != nil {
		return ret, errors.Wrap(err, "message")
	}
	ninfo := sNetworkInfo{
		prefix: "test",
	}
	hip, _ := netutils.NewIPV4Addr("10.2.1.23")
	ninfo.HostIps = map[string]netutils.IPV4Addr{
		"host1": hip,
		"host2": hip + 2,
		"host3": hip + 3,
	}
	ninfo.VlanIps = make(map[int32][]netutils.IPV4Addr)
	ninfo.IPPool = esxi.NewIPPool()
	for _, vlanip := range vlanips {
		for _, vlan := range vlanip.Vlans {
			ips, _ := transferIps(vlan.Ips)
			ninfo.VlanIps[vlan.Id] = append(ninfo.VlanIps[vlan.Id], ips...)
			for _, ip := range ips {
				ninfo.IPPool.Insert(ip, esxi.SIPProc{
					VlanId: vlan.Id,
				})
				ninfo.VMs = append(ninfo.VMs, esxi.SSimpleVM{
					Name: fmt.Sprintf("%s%d", "vm", len(ninfo.VMs)),
					IPVlans: []esxi.SIPVlan{
						{
							IP:     ip,
							VlanId: vlan.Id,
						},
					},
				})
			}
		}
	}
	for _, ips := range ninfo.VlanIps {
		sort.Slice(ips, func(i, j int) bool {
			return ips[i] < ips[j]
		})
	}

	return sParseAndSuggest{
		NInfos: []sNetworkInfo{
			ninfo,
		},
		AccountName: "test",
		ZoneIds: []string{
			"zone_test",
		},
		Wires:         []SWire{},
		Networks:      [][]SNetwork{},
		ExistedIpPool: newIPPool(),
	}, nil
}

func transferIps(ips []string) ([]netutils.IPV4Addr, error) {
	nips := make([]netutils.IPV4Addr, len(ips))
	for i := range ips {
		nip, err := netutils.NewIPV4Addr(ips[i])
		if err != nil {
			return nil, err
		}
		nips[i] = nip
	}
	return nips, nil
}

type sVlanIP struct {
	Id  int32
	Ips []string
}

type sVlanIPs struct {
	Vlans []sVlanIP
}

var _testData = `[{"id":"VmwareDistributedVirtualSwitch:dvs-208","vlans":[{"id":156,"ips":["10.129.56.3","10.129.56.7","10.129.56.104","10.129.56.105","10.129.56.106","10.129.56.107","10.129.56.108","10.129.56.109","10.129.56.115"]},{"id":182,"ips":["10.129.82.9","10.129.82.32","10.129.82.176"]},{"id":199,"ips":["10.129.99.57","10.129.99.88"]},{"id":135,"ips":["10.129.35.7"]},{"id":125,"ips":["10.129.25.42"]},{"id":150,"ips":["10.129.50.101"]},{"id":107,"ips":["10.129.7.103"]},{"id":166,"ips":["10.129.66.131","10.129.66.132","10.129.66.133","10.129.66.134","10.129.66.135","10.129.66.136","10.129.66.141","10.129.66.142","10.129.66.143"]},{"id":128,"ips":["10.129.28.9","10.129.28.10","10.129.28.107"]},{"id":272,"ips":["10.129.172.112"]},{"id":204,"ips":["10.129.104.61","10.129.104.117","10.129.104.118","10.129.104.153"]},{"id":113,"ips":["10.129.13.104","10.129.13.105","10.129.13.106","10.129.13.107","10.129.13.108","10.129.13.109"]},{"id":183,"ips":["10.129.83.1","10.129.83.2"]},{"id":181,"ips":["10.129.81.94","10.129.81.99","10.129.81.183","10.129.81.184"]},{"id":124,"ips":["10.129.24.5"]},{"id":152,"ips":["10.129.52.33","10.129.52.111"]}]},{"id":"VmwareDistributedVirtualSwitch:dvs-211","vlans":[{"id":273,"ips":["10.129.173.1","10.129.173.2","10.129.173.11","10.129.173.22","10.129.173.23","10.129.173.24","10.129.173.25","10.129.173.27","10.129.173.31","10.129.173.32","10.129.173.33","10.129.173.34","10.129.173.35","10.129.173.41","10.129.173.42","10.129.173.52","10.129.173.101"]},{"id":109,"ips":["10.129.9.104","10.129.9.108","10.129.9.166","10.129.9.172","10.129.9.173"]},{"id":143,"ips":["10.129.43.4","10.129.43.31"]},{"id":158,"ips":["10.129.58.104","10.129.58.105"]},{"id":131,"ips":["10.129.31.7"]},{"id":104,"ips":["10.129.4.101","10.129.4.102","10.129.4.103","10.129.4.104"]},{"id":133,"ips":["10.129.33.109"]},{"id":103,"ips":["10.129.3.78"]},{"id":176,"ips":["10.129.76.3"]},{"id":281,"ips":["10.129.181.5"]},{"id":157,"ips":["10.129.57.2"]},{"id":204,"ips":["10.129.104.152"]},{"id":272,"ips":["10.129.172.1","10.129.172.2","10.129.172.3","10.129.172.4","10.129.172.5","10.129.172.6","10.129.172.7","10.129.172.8","10.129.172.10","10.129.172.21","10.129.172.31","10.129.172.32","10.129.172.41","10.129.172.42","10.129.172.51","10.129.172.111"]},{"id":128,"ips":["10.129.28.8","10.129.28.64","10.129.28.68","10.129.28.74","10.129.28.101","10.129.28.102","10.129.28.103","10.129.28.104","10.129.28.105","10.129.28.124","10.129.28.132","10.129.28.152","10.129.28.153","10.129.28.159","10.129.28.165","10.129.28.166","192.168.122.1"]},{"id":183,"ips":["10.129.83.3","10.129.83.4"]},{"id":110,"ips":["10.129.10.65","10.129.10.66","10.129.10.67","10.129.10.75","10.129.10.76","10.129.10.77"]},{"id":271,"ips":["10.129.171.3"]},{"id":136,"ips":["10.129.36.86","10.129.36.87","10.129.36.88","10.129.36.89","10.129.36.97","10.129.36.98","10.129.36.99","10.129.36.161","10.129.36.162","10.129.36.163"]},{"id":149,"ips":["10.129.49.103","10.129.49.104","10.129.49.105","10.129.49.142","10.129.49.171","192.168.122.1","192.168.122.1"]},{"id":182,"ips":["10.129.82.118","10.129.82.119","10.129.82.143","10.129.82.155"]},{"id":154,"ips":["10.129.54.171"]},{"id":124,"ips":["10.129.24.8"]},{"id":168,"ips":["10.153.99.100"]},{"id":192,"ips":["10.129.92.8"]},{"id":245,"ips":["10.129.145.73","10.129.145.75"]},{"id":185,"ips":["10.129.85.51","10.129.85.52","10.129.85.71","10.129.85.72"]},{"id":179,"ips":["10.129.79.15","10.129.79.16"]},{"id":181,"ips":["10.129.81.8","10.129.81.95","10.129.81.96","10.129.81.97","10.129.81.98","10.129.81.108","10.129.81.122","10.129.81.135","10.129.81.136","10.129.81.162","10.129.81.171","10.129.81.172","10.129.81.173","10.129.81.181"]},{"id":175,"ips":["10.129.75.103","10.129.75.104","10.129.75.105","10.129.75.106","10.129.75.138","10.129.75.142","10.129.75.151","10.129.75.152","10.129.75.171","10.129.75.172"]},{"id":270,"ips":["10.129.170.3","10.129.170.4","10.129.170.5"]},{"id":187,"ips":["10.129.87.103","10.129.87.131","10.129.87.132"]},{"id":260,"ips":["10.129.160.1"]},{"id":107,"ips":["10.129.7.144","10.129.7.145"]},{"id":130,"ips":["10.129.30.106","10.129.30.107","10.129.30.108","10.129.30.131"]},{"id":123,"ips":["10.129.23.5","10.129.23.6","10.129.23.7","10.129.23.103","10.129.23.104","10.129.23.105","10.129.23.106","10.129.23.107","10.129.23.108","10.129.23.161","10.129.23.162","10.129.23.163"]},{"id":280,"ips":["10.129.180.1","10.129.180.2","10.129.180.11","10.129.180.21","10.129.180.41","10.129.180.42","10.129.180.51","10.129.180.52","10.129.180.161","10.129.180.162"]},{"id":186,"ips":["10.129.86.103","10.129.86.104"]},{"id":240,"ips":["10.129.140.33","10.129.140.43"]},{"id":120,"ips":["10.129.20.101"]},{"id":243,"ips":["10.129.143.1","10.129.143.2","10.129.143.3","10.129.143.4","10.129.143.11","10.129.143.12","10.129.143.22","10.129.143.31","10.129.143.33","10.129.143.34","10.129.143.35","10.129.143.36","10.129.143.38"]},{"id":125,"ips":["10.129.25.22","10.129.25.102","10.129.25.122"]},{"id":141,"ips":["10.129.41.3","10.129.41.4","10.129.41.5","10.129.41.109"]},{"id":146,"ips":["10.129.46.101"]},{"id":150,"ips":["10.129.50.3","10.129.50.4","10.129.50.109"]},{"id":135,"ips":["10.129.35.20","10.129.35.171","10.129.35.172","10.129.35.173"]},{"id":200,"ips":["10.129.100.13"]}]},{"id":"VmwareDistributedVirtualSwitch:dvs-2217","vlans":[{"id":250,"ips":["10.129.150.3","10.129.150.4","10.129.150.5","10.129.150.14","10.129.150.15","10.129.150.21"]},{"id":127,"ips":["10.129.27.106","10.129.27.122","10.129.27.161"]},{"id":149,"ips":["10.129.49.12","10.129.49.21","10.129.49.77","10.129.49.78","10.129.49.79","10.129.49.80","10.129.49.136","10.129.49.149","10.129.49.174","10.129.49.175","10.129.49.181","10.129.49.182","10.129.49.183","10.129.49.185","10.129.49.186"]},{"id":280,"ips":["10.129.180.12","10.129.180.31"]},{"id":146,"ips":["10.129.46.109"]},{"id":124,"ips":["10.129.24.104","10.129.24.105","10.129.24.141","10.129.24.142","10.129.24.144"]},{"id":168,"ips":["10.129.68.33","10.129.68.34"]},{"id":240,"ips":["10.129.140.71","10.129.140.72","10.129.140.78","10.129.140.79","10.129.140.81","10.129.140.82"]},{"id":172,"ips":["10.129.72.123"]},{"id":137,"ips":["10.129.37.12"]},{"id":141,"ips":["10.129.41.113","10.129.41.114"]},{"id":263,"ips":["10.129.163.171","10.129.163.172","10.129.163.173","10.129.163.174","10.129.163.175","10.129.163.176","10.129.163.177","10.129.163.178","10.129.163.191","10.129.163.192","10.129.163.193","10.129.163.194","10.129.163.195","10.129.163.196","10.129.163.197","10.129.163.198"]},{"id":175,"ips":["10.129.75.12","10.129.75.13","10.129.75.14","10.129.75.15","10.129.75.16","10.129.75.17","10.129.75.18","10.129.75.21","10.129.75.22","10.129.75.41","10.129.75.42","10.129.75.51","10.129.75.52","10.129.75.53","10.129.75.54","10.129.75.55","10.129.75.56","10.129.75.57","10.129.75.58","10.129.75.81","10.129.75.82","10.129.75.83","10.129.75.84","10.129.75.85","10.129.75.86","10.129.75.87","10.129.75.88","10.129.75.101","10.129.75.102","10.129.75.126","10.129.75.163","10.129.75.164","10.129.75.165"]},{"id":181,"ips":["10.129.81.36","10.129.81.55","10.129.81.76","10.129.81.112","10.129.81.125","10.129.81.126","10.129.81.143","10.129.81.155","10.129.81.156","10.129.81.157","10.129.81.158","10.129.81.159","10.129.81.165","10.129.81.166"]},{"id":150,"ips":["10.129.50.31","10.129.50.54","10.129.50.113","10.129.50.114","10.129.50.142","10.129.50.161","10.129.50.162","10.129.50.163"]},{"id":219,"ips":["10.129.119.112","10.129.119.113","10.129.119.118"]},{"id":135,"ips":["10.129.35.32","10.129.35.33","10.129.35.41"]},{"id":140,"ips":["10.129.40.30","10.129.40.31","10.129.40.32","10.129.40.33","10.129.40.34","10.129.40.35"]},{"id":117,"ips":["10.129.17.56","10.129.17.57"]},{"id":122,"ips":["10.129.22.2","10.129.22.3","10.129.22.5","10.129.22.6","10.129.22.31","10.129.22.32","10.129.22.33","10.129.22.34","10.129.22.35","10.129.22.39","10.129.22.45"]},{"id":158,"ips":["10.129.58.11","10.129.58.41","10.129.58.42","10.129.58.75","10.129.58.76"]},{"id":272,"ips":["10.129.172.11"]},{"id":152,"ips":["10.129.52.31","10.129.52.32","10.129.52.64"]},{"id":271,"ips":["10.129.171.31","10.129.171.32"]},{"id":260,"ips":["10.129.160.101","10.129.160.131","10.129.160.146","10.129.160.147"]},{"id":108,"ips":["10.129.8.6","10.129.8.10","10.129.8.16","10.129.8.17","10.129.8.20","10.129.8.166","10.129.8.167","10.129.8.168"]},{"id":110,"ips":["10.129.10.47","10.129.10.49","10.129.10.104","10.129.10.105","10.129.10.133","10.129.10.162","10.129.10.164"]},{"id":154,"ips":["10.129.54.55","10.129.54.56","10.129.54.57","10.129.54.58","10.129.54.59","10.129.54.75","10.129.54.76","10.129.54.77","10.129.54.78","10.129.54.79","10.129.54.172"]},{"id":130,"ips":["10.129.30.111","10.129.30.121","10.129.30.123","10.129.30.124"]},{"id":243,"ips":["10.129.143.51","10.129.143.52"]},{"id":183,"ips":["10.129.83.126","10.129.83.127","10.129.83.128","10.129.83.136","10.129.83.138"]},{"id":157,"ips":["10.129.57.11","10.129.57.12","10.129.57.82"]},{"id":186,"ips":["10.129.86.2"]},{"id":200,"ips":["10.129.100.7","10.129.100.48"]},{"id":133,"ips":["10.129.33.24","10.129.33.41","10.129.33.42","10.129.33.43","10.129.33.44","10.129.33.124","10.129.33.125","10.129.33.126","10.129.33.127","10.129.33.128","10.129.33.129","10.129.33.131","10.129.33.141","10.129.33.142","10.129.33.143","10.129.33.144","10.129.33.146"]},{"id":161,"ips":["10.129.61.26"]},{"id":270,"ips":["10.129.170.6","10.129.170.7","10.129.170.8","10.129.170.11","10.129.170.12","10.129.170.13","10.129.170.14","10.129.170.16","10.129.170.21","10.129.170.22","10.129.170.23","10.129.170.24","10.129.170.31","10.129.170.32","10.129.170.33","10.129.170.34","10.129.170.41","10.129.170.42","10.129.170.43","10.129.170.44","10.129.170.164","10.129.170.165","10.129.170.166"]},{"id":148,"ips":["10.129.48.118","10.129.48.161"]},{"id":128,"ips":["10.129.28.12","10.129.28.21","10.129.28.53","10.129.28.54","10.129.28.108","10.129.28.154","10.129.28.174","10.129.28.175","10.129.28.176"]},{"id":125,"ips":["10.129.25.121"]},{"id":102,"ips":["10.129.2.113","10.129.2.114"]},{"id":156,"ips":["10.129.56.32","10.129.56.34","10.129.56.162","10.129.56.163","10.129.56.164"]},{"id":104,"ips":["10.129.4.28","10.129.4.77","10.129.4.78","10.129.4.79","10.129.4.86","10.129.4.92","10.129.4.164","10.129.4.165","10.129.4.166","10.129.4.167","10.129.4.170","10.129.4.171","10.129.4.172","10.129.4.173","10.129.4.174","10.129.4.175","10.129.4.176","10.129.4.177","10.129.4.178","10.129.4.179"]},{"id":109,"ips":["10.129.9.113","10.129.9.114","10.129.9.133","10.129.9.134","10.129.9.135","10.129.9.136"]},{"id":182,"ips":["10.129.82.184","10.129.82.185"]},{"id":176,"ips":["10.129.76.41","10.129.76.61"]},{"id":107,"ips":["10.129.7.111","10.129.7.112","10.129.7.113","10.129.7.117","10.129.7.118","10.129.7.131","10.129.7.133","10.129.7.137","10.129.7.142","10.129.7.143"]},{"id":160,"ips":["10.129.60.11"]},{"id":132,"ips":["10.129.32.56","10.129.32.89"]},{"id":204,"ips":["10.129.104.101","10.129.104.111","10.129.104.112","10.129.104.113","10.129.104.151"]}]},{"id":"VmwareDistributedVirtualSwitch:dvs-3208","vlans":[{"id":241,"ips":["10.129.141.111"]},{"id":145,"ips":["10.129.45.34"]},{"id":128,"ips":["10.129.28.51","10.129.28.52","10.129.28.110"]},{"id":272,"ips":["10.129.172.43","10.129.172.52","10.129.172.85","10.129.172.86"]},{"id":176,"ips":["10.129.76.62"]},{"id":153,"ips":["10.129.53.11"]},{"id":172,"ips":["10.129.72.124"]},{"id":141,"ips":["10.129.41.11","10.129.41.13","10.129.41.14","10.129.41.31","10.129.41.41","10.129.41.51","10.129.41.61","10.129.41.122"]},{"id":204,"ips":["10.129.104.115","10.129.104.116","10.129.104.122","10.129.104.171"]},{"id":134,"ips":["10.129.34.26"]},{"id":130,"ips":["10.129.30.112"]},{"id":161,"ips":["10.129.61.27","10.129.61.71","10.129.61.74"]},{"id":185,"ips":["10.129.85.9","10.129.85.65","10.129.85.69"]},{"id":124,"ips":["10.129.24.106","10.129.24.107","10.129.24.143"]},{"id":110,"ips":["10.129.10.46","10.129.10.115","10.129.10.116","10.129.10.131","10.129.10.132"]},{"id":123,"ips":["10.129.23.8","10.129.23.63","10.129.23.71"]},{"id":126,"ips":["10.129.26.121","10.129.26.122","10.129.26.123","10.129.26.124","10.129.26.131","10.129.26.132","10.129.26.133","10.129.26.134","10.129.26.135","10.129.26.136","10.129.26.137","10.129.26.138","10.129.26.139","10.129.26.140"]},{"id":135,"ips":["10.129.35.10","10.129.35.21","10.129.35.24","10.129.35.61","10.129.35.122"]},{"id":127,"ips":["10.129.27.42","10.129.27.121"]},{"id":219,"ips":["10.129.119.156"]},{"id":271,"ips":["10.129.171.21"]},{"id":144,"ips":["10.129.44.21","10.129.44.22"]},{"id":150,"ips":["10.129.50.51"]},{"id":183,"ips":["10.129.83.31","10.129.83.33","10.129.83.123"]},{"id":109,"ips":["10.129.9.115","10.129.9.116"]},{"id":152,"ips":["10.129.52.63"]},{"id":143,"ips":["10.129.43.22"]},{"id":151,"ips":["10.129.51.62"]},{"id":244,"ips":["10.129.144.22","10.129.144.41","10.129.144.42","10.129.144.161","10.129.144.162"]},{"id":270,"ips":["10.129.170.101","10.129.170.102","10.129.170.167","10.129.170.168","10.129.170.169"]},{"id":104,"ips":["10.129.4.1","10.129.4.2","10.129.4.3","10.129.4.4","10.129.4.5","10.129.4.6","10.129.4.7","10.129.4.35","10.129.4.36","10.129.4.37","10.129.4.38","10.129.4.91","10.129.4.93","10.129.4.94","10.129.4.95","10.129.4.96","10.129.4.97","10.129.4.98","10.129.4.105","10.129.4.106","10.129.4.168","10.129.4.169"]},{"id":138,"ips":["10.129.38.93","10.129.38.102"]},{"id":107,"ips":["10.129.7.114","10.129.7.115","10.129.7.146"]},{"id":157,"ips":["10.129.57.22","10.129.57.51"]},{"id":106,"ips":["10.129.6.15","10.129.6.17"]},{"id":174,"ips":["10.129.74.21"]},{"id":182,"ips":["10.129.82.13","10.129.82.124","10.129.82.139","10.129.82.144"]},{"id":250,"ips":["10.129.150.13"]},{"id":180,"ips":["10.129.80.115","10.129.80.116","10.129.80.117","10.129.80.118"]},{"id":187,"ips":["10.129.87.106","10.129.87.135"]},{"id":136,"ips":["10.129.36.40","10.129.36.45","10.129.36.81","10.129.36.112"]},{"id":102,"ips":["10.129.2.33"]},{"id":175,"ips":["10.129.75.11","10.129.75.131"]},{"id":156,"ips":["10.129.56.63","10.129.56.64"]},{"id":122,"ips":["10.129.22.15","10.129.22.17","10.129.22.18"]},{"id":133,"ips":["10.129.33.9","10.129.33.15","10.129.33.45","10.129.33.46","10.129.33.51","10.129.33.108","10.129.33.122","10.129.33.123","10.129.33.145"]},{"id":108,"ips":["10.129.8.18","10.129.8.19","10.129.8.46","10.129.8.56","10.129.8.104","10.129.8.105","10.129.8.113","10.129.8.114","10.129.8.115","10.129.8.116","10.129.8.117","10.129.8.118","10.129.8.164","10.129.8.165","10.129.8.169"]},{"id":166,"ips":["10.129.66.113"]},{"id":113,"ips":["10.129.13.13","10.129.13.143"]},{"id":125,"ips":["10.129.25.64","10.129.25.71","10.129.25.104","10.129.25.105"]},{"id":263,"ips":["10.129.163.181","10.129.163.182","10.129.163.183","10.129.163.184","10.129.163.185","10.129.163.186","10.129.163.187","10.129.163.188","10.129.163.201","10.129.163.202","10.129.163.203","10.129.163.204","10.129.163.205","10.129.163.206","10.129.163.207","10.129.163.208"]},{"id":280,"ips":["10.129.180.5","10.129.180.6","10.129.180.7","10.129.180.8","10.129.180.9","10.129.180.71","10.129.180.72","10.129.180.73","10.129.180.74","10.129.180.75","10.129.180.163"]},{"id":149,"ips":["10.129.49.7","10.129.49.8","10.129.49.9","10.129.49.24","10.129.49.25","10.129.49.26","10.129.49.27","10.129.49.28","10.129.49.29","10.129.49.33","10.129.49.34","10.129.49.35","10.129.49.36","10.129.49.37","10.129.49.38","10.129.49.39","10.129.49.44","10.129.49.45","10.129.49.46","10.129.49.47","10.129.49.48","10.129.49.49","10.129.49.54","10.129.49.55","10.129.49.56","10.129.49.57","10.129.49.58","10.129.49.59","10.129.49.63","10.129.49.64","10.129.49.65","10.129.49.66","10.129.49.67","10.129.49.68","10.129.49.73","10.129.49.133","10.129.49.148","10.129.49.163","10.129.49.164","10.129.49.165","10.129.49.184"]},{"id":246,"ips":["10.129.146.11","10.129.146.12","10.129.146.13","10.129.146.14","10.129.146.15"]},{"id":132,"ips":["10.129.32.81"]},{"id":243,"ips":["10.129.143.25","10.129.143.32","10.129.143.37","10.129.143.42"]},{"id":116,"ips":["10.129.16.24"]},{"id":181,"ips":["10.129.81.16","10.129.81.21","10.129.81.25","10.129.81.26","10.129.81.27","10.129.81.28","10.129.81.35","10.129.81.46","10.129.81.56","10.129.81.62","10.129.81.65","10.129.81.66","10.129.81.67","10.129.81.68","10.129.81.77","10.129.81.93","10.129.81.121","10.129.81.127","10.129.81.141","10.129.81.145","10.129.81.147","10.129.81.148","10.129.81.164"]}]},{"id":"VmwareDistributedVirtualSwitch:dvs-4153","vlans":[{"id":250,"ips":["10.129.150.1"]},{"id":204,"ips":["10.129.104.102","10.129.104.114"]},{"id":125,"ips":["10.129.25.65"]},{"id":1413,"ips":["10.132.51.1","10.132.51.2","10.132.51.3","10.132.51.4","10.132.51.5"]},{"id":1424,"ips":["10.132.92.1","10.132.92.2","10.132.92.3"]},{"id":123,"ips":["10.129.23.121","10.129.23.122","10.129.23.123","10.129.23.124","10.129.23.125","10.129.23.126","10.129.23.127","10.129.23.128","10.129.23.131","10.129.23.132","10.129.23.133","10.129.23.134","10.129.23.135"]},{"id":116,"ips":["10.129.16.65","10.129.16.66"]},{"id":270,"ips":["10.129.170.113"]},{"id":280,"ips":["10.129.180.22"]},{"id":1505,"ips":["10.133.16.1","10.133.16.2","10.133.16.3","10.133.16.4","10.133.16.5","10.133.16.6","10.133.16.7","10.133.16.8","10.133.16.9","10.133.16.10","10.133.16.11","10.133.16.12","10.133.16.13","10.133.16.14","10.133.16.15","10.133.16.16","10.133.16.17","10.133.16.18","10.133.16.19","10.133.16.20","10.133.16.21","10.133.16.22","10.133.19.1","10.133.19.2","10.133.19.3"]},{"id":1405,"ips":["10.132.16.6","10.132.16.7","10.132.16.8","10.132.16.9","10.132.16.11","10.132.16.12","10.132.16.21","10.132.16.22","10.132.16.23","10.132.16.24","10.132.16.25","10.132.16.26","10.132.16.27","10.132.16.28","10.132.16.29","10.132.16.30","10.132.16.31","10.132.19.1","10.132.19.2","10.132.19.3","10.132.19.4"]},{"id":128,"ips":["10.129.28.81","10.129.28.125"]},{"id":260,"ips":["10.129.160.66"]}]},{"id":"VmwareDistributedVirtualSwitch:dvs-4221","vlans":[{"id":1413,"ips":["10.132.48.1","10.132.48.2","10.132.48.3","10.132.48.4","10.132.48.5","10.132.48.6","10.132.48.7","10.132.48.8","10.132.48.9","10.132.48.10","10.132.48.13","10.132.48.14","10.132.48.15","10.132.48.16","10.132.48.17","10.132.48.18","10.132.48.19","10.132.48.20","10.132.48.21","10.132.48.22","10.132.48.23","10.132.48.24","10.132.48.25","10.132.48.26","10.132.48.27","10.132.48.28","10.132.48.29","10.132.48.30","10.132.48.31","10.132.48.32","10.132.48.33","10.132.48.34","10.132.48.35","10.132.48.36","10.132.51.6","10.132.51.7","10.132.51.8","10.132.51.9","10.132.51.10","10.132.51.11","10.132.51.12","10.132.51.13","10.132.51.14","10.132.51.15","10.132.51.16","10.132.51.17","10.132.51.18","10.132.51.19","10.132.51.20","10.132.51.21","10.132.51.22","10.132.51.23","10.132.51.24","10.132.51.25","10.132.51.26","10.132.51.27","10.132.51.28","10.132.51.29","10.132.51.30","10.132.51.31","10.132.51.32","10.132.51.33","10.132.51.34","10.132.51.35","10.132.51.36","10.132.51.37","10.132.51.38","10.132.51.39"]},{"id":1421,"ips":["10.132.80.1","10.132.80.2","10.132.80.3","10.132.83.1","10.132.83.2","10.132.83.3","10.132.83.4","10.132.83.5","10.132.83.6","10.132.83.7"]},{"id":1521,"ips":["10.133.80.1","10.133.80.2","10.133.80.3"]},{"id":1513,"ips":["10.133.48.11","10.133.48.12"]}]},{"id":"VmwareDistributedVirtualSwitch:dvs-520","vlans":[{"id":116,"ips":["10.129.16.16"]},{"id":270,"ips":["10.129.170.15","10.129.170.17","10.129.170.18","10.129.170.111","10.129.170.112","10.129.170.114","10.129.170.115","10.129.170.116","10.129.170.118"]},{"id":103,"ips":["10.129.3.113","10.129.3.114","10.129.3.117","10.129.3.118","10.129.3.119","10.129.3.157","10.129.3.158","10.129.3.159"]},{"id":158,"ips":["10.129.58.101","10.129.58.111","10.129.58.112","10.129.58.113","10.129.58.114","10.129.58.121","10.129.58.122"]},{"id":204,"ips":["10.129.104.110","10.129.104.119","10.129.104.120","10.129.104.121","10.129.104.141","10.129.104.142","10.129.104.143"]},{"id":240,"ips":["10.129.140.23","10.129.140.24","10.129.140.34","10.129.140.44"]},{"id":110,"ips":["10.129.10.48","10.129.10.163"]},{"id":104,"ips":["10.129.4.25","10.129.4.26","10.129.4.27"]},{"id":156,"ips":["10.129.56.75","10.129.56.121"]},{"id":187,"ips":["10.129.87.121","10.129.87.122","10.129.87.123"]},{"id":136,"ips":["10.129.36.80","10.129.36.82","10.129.36.83","10.129.36.84","10.129.36.85"]},{"id":182,"ips":["10.129.82.114","10.129.82.123","10.129.82.147","10.129.82.148","10.129.82.156","10.129.82.157","10.129.82.158","10.129.82.175","10.129.82.177"]},{"id":281,"ips":["10.129.181.111"]},{"id":133,"ips":["10.129.33.25"]},{"id":175,"ips":["10.129.75.47","10.129.75.125","10.129.75.132","10.129.75.139","10.129.75.141"]},{"id":128,"ips":["10.129.28.63","10.129.28.67","10.129.28.73","10.129.28.106","10.129.28.131","10.129.28.133","10.129.28.141","10.129.28.142","10.129.28.151","10.129.28.171","10.129.28.172","10.129.28.173"]},{"id":145,"ips":["10.129.45.31","10.129.45.32","10.129.45.33","10.129.45.35","10.129.45.36"]},{"id":199,"ips":["10.129.99.152"]},{"id":109,"ips":["10.129.9.147","10.129.9.148","10.129.9.164"]},{"id":243,"ips":["10.129.143.21","10.129.143.41"]},{"id":176,"ips":["10.129.76.42"]},{"id":280,"ips":["10.129.180.32"]},{"id":164,"ips":["10.129.64.121","10.129.64.122","10.129.64.123","10.129.64.124"]},{"id":185,"ips":["10.129.85.53","10.129.85.61","10.129.85.62","10.129.85.63","10.129.85.64","10.129.85.66","10.129.85.67","10.129.85.68"]},{"id":154,"ips":["10.129.54.173","10.129.54.174","10.129.54.175"]},{"id":135,"ips":["10.129.35.31","10.129.35.34","10.129.35.35","10.129.35.36"]},{"id":146,"ips":["10.129.46.31"]},{"id":132,"ips":["10.129.32.43","10.129.32.44"]},{"id":124,"ips":["10.129.24.15","10.129.24.16"]},{"id":152,"ips":["10.129.52.5"]},{"id":181,"ips":["10.129.81.49","10.129.81.50","10.129.81.111","10.129.81.123","10.129.81.128","10.129.81.130","10.129.81.131","10.129.81.133","10.129.81.140","10.129.81.142","10.129.81.144","10.129.81.146","10.129.81.175","10.129.81.176"]},{"id":183,"ips":["10.129.83.5","10.129.83.6","10.129.83.32","10.129.83.34","10.129.83.124","10.129.83.135","10.129.83.137"]},{"id":272,"ips":["10.129.172.9","10.129.172.12","10.129.172.44","10.129.172.53","10.129.172.87","10.129.172.88","10.129.172.89"]},{"id":141,"ips":["10.129.41.6","10.129.41.7","10.129.41.42"]},{"id":108,"ips":["10.129.8.5","10.129.8.15","10.129.8.35"]},{"id":143,"ips":["10.129.43.32"]},{"id":130,"ips":["10.129.30.122"]},{"id":245,"ips":["10.129.145.71","10.129.145.72"]},{"id":113,"ips":["10.129.13.71","10.129.13.72","10.129.13.75","10.129.13.76","10.129.13.79","10.129.13.151"]},{"id":273,"ips":["10.129.173.51","10.129.173.61","10.129.173.62"]},{"id":107,"ips":["10.129.7.116","10.129.7.132","10.129.7.134","10.129.7.141"]},{"id":149,"ips":["10.129.49.131","10.129.49.132","10.129.49.134","10.129.49.135","10.129.49.137","10.129.49.138","10.129.49.152","10.129.49.162","10.129.49.172","10.129.49.173","10.129.49.176"]},{"id":260,"ips":["10.129.160.44"]},{"id":125,"ips":["10.129.25.43","10.129.25.67","10.129.25.73","10.129.25.103"]}]}]`
