package udphandler

import (
	//	"code.google.com/p/gopacket"
	"fmt"
	//"os"
	"snif"
	"testing"
	"time"
)

//func TestNew_Sip_proc(t *testing.T) {
//	in_chan := make(chan *gopacket.Packet)
//	uh := NewUdpHandler(&in_chan)
//	uh.Add_gw_ip("10.220.0.2")
//	uh.Del_gw_ip("8.8.8.8")
//	fmt.Printf("%+v\n", uh)
//	if _, ok := uh.Gw_ip["10.220.0.2"]; !ok {
//		t.Error("no added ip")
//	}
//	//fmt.Println(sip_p.Local_ip["10.220.0.2"])
//}

func TestPbxRun(t *testing.T) {

}

func TestUdpHandlerRun(t *testing.T) {
	//var iface string = `\Device\NPF_{20CF0BD0-F556-410A-931C-3DE6AA54E9EE}`
	//var iface string = `\Device\NPF_{3230EF11-015C-45F9-836B-3227A3E2D8E5}`
	//my interface
	var iface string = `\Device\NPF_{48086FEC-C3F3-4976-9104-253AA5089822}`
	var snaplen int = 1500
	var promisc bool = true
	var bpfFilter string = ""
	handle := snif.Start_snif(iface, snaplen, promisc, bpfFilter)
	//defer handle.Close()
	spr := snif.NewTransportSeparator(handle)
	spr.New_udp_chan(20)
	go spr.Run()
	//fmt.Println("some data from spr.udp_chan", <-spr.UDP_chan)
	//uh := NewUdpHandler(&spr.UDP_chan)
	//uh := UdpHandler{In_chan: spr.UDP_chan, Gw_ip: make(map[string]*Gateway), Quit: make(chan string)}
	uh := *New(spr.UDP_chan)
	fmt.Printf("%+v\n", uh)
	//go uh.Run()
	uh.Add_gw_ip("10.220.16.13")
	uh.Add_gw_ip("10.200.73.100")
	uh.Add_gw_ip("10.220.16.14")
	go Udp_handler_Run(&uh)
	//go uh.Run()
	//	uh.Gw_ip["10.220.16.13"] = &Gateway{Ip: ip, In_chan: make(chan Dmsg, 5), Quit: make(chan string)}
	//	uh.Gw_ip["10.220.16.13"] = &Gateway{Ip: ip, In_chan: make(chan Dmsg, 5), Quit: make(chan string)}
	//	uh.Gw_ip["10.220.16.13"] = &Gateway{Ip: ip, In_chan: make(chan Dmsg, 5), Quit: make(chan string)}

	//fmt.Println(<-uh.Quit)
	timer := time.NewTimer(20 * time.Second)
	if tn, ok := <-timer.C; ok {
		fmt.Println(tn)
		close(uh.In_chan)
		handle.Close()
		fmt.Println("ждем пока завершится горутина uh", <-uh.Quit)
		//надо бы по идее еще тормознуть separator
		//handle.Close()
		//fmt.Println(snif.Get_stats(handle))
		//close(uh.In_chan)
		//os.Exit(0)
	}

}
