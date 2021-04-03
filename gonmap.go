package gonmap

import (
	"errors"
	"fmt"
	"strings"
)

var NMAP *nmap

func Init() {
	fmt.Println("初始化了")
	NMAP_SERVICE_PROBES = strings.Replace(NMAP_SERVICE_PROBES, "${backquote}", "`", -1)
	NMAP = &nmap{
		exclude:     newPort(),
		probeGroup:  make(map[string]*probe),
		probeSort:   []string{},
		portMap:     make(map[int][]string),
		probeFilter: 0,
		target:      newTarget(),
		response:    nil,
		finger:      nil,
	}
	for i := 1; i <= 65535; i++ {
		NMAP.portMap[i] = []string{}
	}
	NMAP.loads(NMAP_SERVICE_PROBES)
}

func New() *nmap {
	n := &nmap{
		exclude:     newPort(),
		probeGroup:  make(map[string]*probe),
		probeSort:   []string{},
		probeFilter: 0,
		target:      nil,
		response:    nil,
		finger:      nil,
	}
	*n = *NMAP
	return n
}

type nmap struct {
	exclude *port

	probeGroup  map[string]*probe
	probeSort   []string
	probeFilter int
	portMap     map[int][]string

	target *target

	response *response
	finger   *finger
}

func (n *nmap) Scan(ip string, port int) *finger {
	n.target.host = ip
	n.target.port = port
	n.target.uri = fmt.Sprintf("%s:%d", ip, port)

	//fmt.Println(n.portMap[port])
	for _, requestName := range n.portMap[port] {
		fmt.Println("开始探测：", requestName, "权重为", n.probeGroup[requestName].rarity)
		data, err := n.probeGroup[requestName].scan(n.target)
		if err != nil {
			continue
		} else {
			//若存在返回包，则开始捕获指纹
			f := n.getFinger(data, requestName)
			//如果成功匹配指纹，则直接返回指纹
			if f != nil {
				return f
			}
		}
	}
	return nil
}

func (n *nmap) getFinger(data string, requestName string) *finger {
	f := n.probeGroup[requestName].match(data)
	if f == nil {
		if n.probeGroup[requestName].fallback != "" {
			return n.probeGroup["TCP_"+n.probeGroup[requestName].fallback].match(data)
		}
	}
	return f
}

func (n *nmap) loads(s string) {
	lines := strings.Split(s, "\n")
	probeArr := []string{}
	p := newProbe()
	for _, line := range lines {
		if !n.isCommand(line) {
			continue
		}
		commandName := line[:strings.Index(line, " ")]
		if commandName == "Exclude" {
			n.loadExclude(line)
			continue
		}
		if commandName == "Probe" {
			if len(probeArr) != 0 {
				p.loads(probeArr)
				n.pushProbe(p)
			}
			probeArr = []string{}
			p.Clean()
		}
		probeArr = append(probeArr, line)
	}
	p.loads(probeArr)
	n.pushProbe(p)
}

func (n *nmap) loadExclude(expr string) {
	var exclude = newPort()
	expr = expr[strings.Index(expr, " ")+1:]
	for _, s := range strings.Split(expr, ",") {
		if !exclude.Load(s) {
			panic(errors.New("exclude 语句格式错误"))
		}
	}
	n.exclude = exclude
}

func (n *nmap) pushProbe(p *probe) {
	PROBE := newProbe()
	*PROBE = *p
	//if p.ports.length == 0 && p.sslports.length == 0 {
	//	fmt.Println(p.request.name)
	//}
	n.probeSort = append(n.probeSort, p.request.name)
	n.probeGroup[p.request.name] = PROBE

	//建立端口扫描对应表，将根据端口号决定使用何种请求包
	if p.ports.length+p.sslports.length == 0 {
		p.ports.Fill()
		p.sslports.Fill()
	}
	for _, i := range p.ports.value {
		n.portMap[i] = append(n.portMap[i], p.request.name)
	}
	for _, i := range p.sslports.value {
		n.portMap[i] = append(n.portMap[i], p.request.name)
	}

}

func (n *nmap) isCommand(line string) bool {
	//删除注释行和空行
	if len(line) < 2 {
		return false
	}
	if line[:1] == "#" {
		return false
	}
	//删除异常命令
	commandName := line[:strings.Index(line, " ")]
	commandArr := []string{
		"Exclude", "Probe", "match", "softmatch", "ports", "sslports", "totalwaitms", "tcpwrappedms", "rarity", "fallback",
	}
	for _, item := range commandArr {
		if item == commandName {
			return true
		}
	}
	return false
}
