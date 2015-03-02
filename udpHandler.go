package udphandler

import (
	"code.google.com/p/gopacket"
	"code.google.com/p/gopacket/layers"
	"fmt"
	//	"net"
	"sip_parser_lite"
	"strconv"
	"time"
)

type UdpHandler struct {
	In_chan chan gopacket.Packet
	Gw_ip   map[string]*Gateway
	Quit    chan string
}

type Gateway struct {
	in_work_state bool //если запущено то true
	Ip            string
	In_chan       chan Dmsg
	Pbx_ip        map[string]*Pbx
	Quit          chan string
}

type Pbx struct {
	SipGwip                string
	Host_ip                string
	In_chan                chan Dmsg
	Quit                   chan string
	Callid_state           map[string]string
	User_agent             string //д.б добавлен автоматически
	Caps60, Caps30, Caps15 float64
}

func New(in_chan chan gopacket.Packet) *UdpHandler {
	return &UdpHandler{In_chan: in_chan, Gw_ip: make(map[string]*Gateway), Quit: make(chan string)}
}

//добавляет но не запускает горутину для gateway
func (uh *UdpHandler) Add_gw_ip(ip string) {
	uh.Gw_ip[ip] = &Gateway{Ip: ip, In_chan: make(chan Dmsg, 5), Pbx_ip: make(map[string]*Pbx), Quit: make(chan string)}
}

//func (uh *UdpHandler) Add_gw_ip(ip string) {
//	uh.Gw_ip[ip] = &Gateway{Ip: ip, In_chan: make(chan Dmsg, 5), Quit: make(chan string)}
//	//defer uh.Del_gw_ip(ip)
//	go Gateway_Run(uh.Gw_ip[ip])
//}

//func (uh *UdpHandler) Del_gw_ip(ip string) {
//	if _, ok := uh.Gw_ip[ip]; ok {
//		close(uh.Gw_ip[ip].In_chan)
//		_ = <-uh.Gw_ip[ip].Quit
//		delete(uh.Gw_ip, ip)
//	}
//}

//func (g *Gateway) Add_pbx_ip(ip string) {
//	g.Pbx_ip[ip] = &Pbx{
//		Gw_ip:   g.Ip,
//		Host_ip: ip,
//		In_chan: make(chan *gopacket.Packet, 10),
//		Quit:    make(chan string),
//	}
//}

type Dmsg struct {
	Dir   string //can be "RECV", "SEND"
	SrcIP string
	DstIP string
	Msg   sip_parser_lite.Sip_msg
}

func Udp_handler_Run(uh *UdpHandler) {
	fmt.Println("udp processor started")
	for _, g := range uh.Gw_ip {
		g.in_work_state = true
		go Gateway_Run(g)
	}
	for packet := range uh.In_chan {
		//fmt.Println(&packet)
		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		ip4, _ := ipLayer.(*layers.IPv4)
		if g, ok := uh.Gw_ip[ip4.SrcIP.String()]; ok {
			udpLayer := packet.Layer(layers.LayerTypeUDP)
			udp, _ := udpLayer.(*layers.UDP)
			if msg, ok := sip_parser_lite.Get_sip_msg(string(udp.Payload)); ok {
				g.In_chan <- Dmsg{"SEND", ip4.SrcIP.String(), ip4.DstIP.String(), msg}
				//fmt.Println("udp", string(udp.Payload))
			}
		} else if g, ok := uh.Gw_ip[ip4.DstIP.String()]; ok {
			udpLayer := packet.Layer(layers.LayerTypeUDP)
			udp, _ := udpLayer.(*layers.UDP)
			//fmt.Println("udp", ip4.DstIP.String())
			if msg, ok := sip_parser_lite.Get_sip_msg(string(udp.Payload)); ok {
				g.In_chan <- Dmsg{"RECV", ip4.SrcIP.String(), ip4.DstIP.String(), msg}
				//fmt.Println("udp", string(udp.Payload))
			}
		} else {
			//fmt.Println("drop packet")
		}

	}
	//блокируемся до прихода сообщения о закрытии всех порожденных
	for _, g := range uh.Gw_ip {
		close(g.In_chan)
	}
	//дожидаемся пока они завершаться
	for ip, g := range uh.Gw_ip {
		fmt.Println("gate stop: ", ip, <-g.Quit)
	}
	fmt.Println("udp processor stopped")
	uh.Quit <- "UdpHandler stopped"
}

func newPbx(sipgwip string, hostip string) *Pbx {
	return &Pbx{SipGwip: sipgwip,
		Host_ip:      hostip,
		In_chan:      make(chan Dmsg, 3),
		Quit:         make(chan string),
		Callid_state: make(map[string]string),
		User_agent:   "",
	}
}

func Gateway_Run(g *Gateway) {
	fmt.Println(g.Ip, "gateway started")
	for dmsg := range g.In_chan {
		//fmt.Println(dmsg.Dir, "src ip", dmsg.SrcIP, "dst ip", dmsg.DstIP, dmsg.Msg.First_line)
		if dmsg.Dir == "SEND" {
			if _, ok := g.Pbx_ip[dmsg.DstIP]; ok {
				g.Pbx_ip[dmsg.DstIP].In_chan <- dmsg
			} else {
				pbx := newPbx(g.Ip, dmsg.DstIP)
				//fmt.Println("g.pbx_ip", g.Pbx_ip[dmsg.DstIP])
				g.Pbx_ip[dmsg.DstIP] = pbx
				go Pbx_Run(pbx)
				g.Pbx_ip[dmsg.DstIP].In_chan <- dmsg
			}
		} else if dmsg.Dir == "RECV" {
			if _, ok := g.Pbx_ip[dmsg.SrcIP]; ok {
				g.Pbx_ip[dmsg.SrcIP].In_chan <- dmsg
			} else {
				pbx := newPbx(g.Ip, dmsg.SrcIP)
				g.Pbx_ip[dmsg.SrcIP] = pbx
				go Pbx_Run(pbx)
				g.Pbx_ip[dmsg.SrcIP].In_chan <- dmsg
			}
		}
	}
	for _, pbx := range g.Pbx_ip {
		close(pbx.In_chan)
	}
	for ip, pbx := range g.Pbx_ip {
		fmt.Println("pbx stop", ip, <-pbx.Quit)

	}
	g.Quit <- "gateway stop"
}

func Pbx_Run(p *Pbx) {
	//fmt.Println(p.Host_ip, "started")
	var total60, total30, total15 float64
	go p.calc_caps60(&total60)
	go p.calc_caps15(&total15)
	for dmsg := range p.In_chan {
		//fmt.Println("PBX recived")
		//fmt.Println(dmsg.Dir, "src ip", dmsg.SrcIP, "dst ip", dmsg.DstIP, dmsg.Msg.First_line)
		if dmsg.Msg.Method_or_Code == "INVITE" {
			switch dmsg.Dir {
			case "SEND":
				p.Callid_state[dmsg.Msg.Call_id] = "SEND"
				//go p.calc_caps60(&total60)
			case "RECV":
				p.Callid_state[dmsg.Msg.Call_id] = "RECV"
				p.User_agent = dmsg.Msg.Headers["User-Agent"]
				total60, total30, total15 = total60+1, total30+1, total15+1
			}
			if p.Host_ip == "10.100.100.104" {
				fmt.Println("New call", dmsg.Dir, dmsg.Msg.Call_id, dmsg.Msg.First_line)
			}

			//fmt.Println("new call", p.Callid_state)
			continue // что бы дальше не проверять
		}
		if p.Host_ip == "10.100.100.104" {
			fmt.Println(dmsg.Msg.Sip_type, dmsg.Msg.Method_or_Code)
		}
		if dmsg.Msg.Sip_type == "REQUEST" {
			switch dmsg.Msg.Method_or_Code {
			case "BYE":
				delete(p.Callid_state, dmsg.Msg.Call_id)
				//fmt.Println("call finished", p.Callid_state)
				//fmt.Println("total calls:", len(p.Callid_state))
				if p.Host_ip == "10.100.100.104" {
					fmt.Println("BYE call finished", dmsg.Msg.Call_id)
				}

			case "CANCEL":
				delete(p.Callid_state, dmsg.Msg.Call_id)
				//fmt.Println("call finished", p.Callid_state)
				//fmt.Println("total calls:", len(p.Callid_state))
				if p.Host_ip == "10.100.100.104" {
					fmt.Println("CANCEL call finished", dmsg.Msg.Call_id)
				}

			}
		} else if code, err := strconv.Atoi(dmsg.Msg.Method_or_Code); err == nil {
			switch {
			case 100 < code && code < 199:
				continue
			case 200 == code: //&& dmsg.Msg.Headers["CSeq"] содержит INVITE
				//вызов в разговорном состоянии
			case 401 == code: // Anuthorized. будет 2 й invite
			case 300 < code && code < 699:
				delete(p.Callid_state, dmsg.Msg.Call_id)
				if p.Host_ip == "10.100.100.104" {
					fmt.Println(code, "call finished", dmsg.Msg.Call_id)
				}
				//fmt.Println(p.Host_ip, "total calls:", len(p.Callid_state))
			}

		}
		_ = dmsg
	}
	p.Quit <- "pbx stopped"
}

func (uh *UdpHandler) GetPeerCalls(gwip string, pbx_ip string) string {
	var in, out int
	for _, dir := range uh.Gw_ip[gwip].Pbx_ip[pbx_ip].Callid_state {
		if dir == "SEND" {
			out += 1
		} else {
			in += 1
		}
	}
	return fmt.Sprintf("in=%d, out=%d, total=%d", in, out, in+out)
}

func (p *Pbx) calc_caps60(total60 *float64) {
	tick60 := time.NewTicker(60 * time.Second)
	for {
		if _, ok := <-tick60.C; ok {
			p.Caps60 = *total60 / 60.0
			if p.Host_ip == "10.200.76.106" {
				fmt.Println("total60:", *total60, "caps60:", p.Caps60)
			}
			*total60 = 0.0
		}
	}
}

func (p *Pbx) calc_caps15(total15 *float64) {
	c := time.Tick(15 * time.Second)
	for _ = range c {
		p.Caps15 = *total15 / 15.0
		if p.Host_ip == "10.200.76.106" {
			fmt.Println("total15:", *total15, "caps15:", p.Caps15)
		}
		*total15 = 0.0
	}
}
