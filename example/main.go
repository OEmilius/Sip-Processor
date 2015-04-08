package main

import (
	//	"code.google.com/p/gopacket"
	"fmt"
	//"os"
	"os"
	"snif"
	"time"
	"udphandler"
)

func main() {
	var iface string = `\Device\NPF_{20CF0BD0-F556-410A-931C-3DE6AA54E9EE}`
	//var iface string = `\Device\NPF_{3230EF11-015C-45F9-836B-3227A3E2D8E5}`
	//my interface
	//var iface string = `\Device\NPF_{48086FEC-C3F3-4976-9104-253AA5089822}`
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
	//uh := udphandler.UdpHandler{In_chan: spr.UDP_chan, Gw_ip: make(map[string]*udphandler.Gateway), Quit: make(chan string)}
	uh := udphandler.New(spr.UDP_chan)
	fmt.Printf("%+v\n", uh)
	//go uh.Run()
	uh.Add_gw_ip("10.200.66.5")
	//uh.Add_gw_ip("10.200.73.100")
	//uh.Add_gw_ip("10.220.16.14")
	go udphandler.Udp_handler_Run(uh)
	//	uh.Gw_ip["10.220.16.13"] = &Gateway{Ip: ip, In_chan: make(chan Dmsg, 5), Quit: make(chan string)}
	//	uh.Gw_ip["10.220.16.13"] = &Gateway{Ip: ip, In_chan: make(chan Dmsg, 5), Quit: make(chan string)}
	//	uh.Gw_ip["10.220.16.13"] = &Gateway{Ip: ip, In_chan: make(chan Dmsg, 5), Quit: make(chan string)}

	//fmt.Println(<-uh.Quit)
	//go stop_server()
	Periodicaly_print("10.200.66.5", "10.200.76.120", uh)
	//Periodicaly_print("10.200.66.5", "10.100.100.104", uh)

}

func stop_server() {
	timer := time.NewTimer(20 * time.Second)
	if tn, ok := <-timer.C; ok {
		fmt.Println(tn)
		//		handle.Close()
		//		close(uh.In_chan)
		//		fmt.Println("ждем пока завершится горутина uh", <-uh.Quit)
		//		//надо бы по идее еще тормознуть separator
		//handle.Close()
		//fmt.Println(snif.Get_stats(handle))
		//close(uh.In_chan)
		os.Exit(0)
	}

}

func Periodicaly_print(gwip string, pbx_ip string, uh *udphandler.UdpHandler) {
	tick := time.NewTicker(30 * time.Second)
	for {
		if _, ok := <-tick.C; ok {
			fmt.Println(uh.GetPeerCalls(gwip, pbx_ip))
		}
	}
}
