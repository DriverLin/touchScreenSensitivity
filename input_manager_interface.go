package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
)

func handel_touch_using_input_manager(control_ch chan *touch_control_pack) {
	socket, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 61068,
	})
	if err != nil {
		fmt.Printf("连接input_manager失败 : %s\n", err.Error())
		os.Exit(3)
	}
	defer socket.Close()
	for {
		select {
		case <-global_close_signal:
			return
		case control_data := <-control_ch:
			// fmt.Printf("控制设备 : %v\n", control_data)
			action := byte(control_data.action)
			id := byte(control_data.id & 0xff)
			x := make([]byte, 4)
			y := make([]byte, 4)
			binary.LittleEndian.PutUint32(x, uint32(control_data.x))
			binary.LittleEndian.PutUint32(y, uint32(control_data.y))
			// fmt.Printf("接收到控制命令 : %v\n", control_data)
			// fmt.Printf("write bytes : %v\n", []byte{action, id, x[0], x[1], x[2], x[3], y[0], y[1], y[2], y[3]})
			socket.Write([]byte{action, id, x[0], x[1], x[2], x[3], y[0], y[1], y[2], y[3]})
		}
	}
}
