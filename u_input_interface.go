package main

import (
	"bytes"
	"encoding/binary"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"fmt"
	"os"
	"unsafe"

	"github.com/lunixbochs/struc"

	"github.com/kenshaw/evdev"
)

func toUInputName(name []byte) [uinputMaxNameSize]byte {
	var fixedSizeName [uinputMaxNameSize]byte
	copy(fixedSizeName[:], name)
	return fixedSizeName
}

func uInputDevToBytes(uiDev UinputUserDev) []byte {
	var buf bytes.Buffer
	_ = struc.PackWithOptions(&buf, &uiDev, &struc.Options{Order: binary.LittleEndian})
	return buf.Bytes()
}

func createDevice(f *os.File) (err error) {
	return ioctl(f.Fd(), UIDEVCREATE(), uintptr(0))
}

func create_u_input_touch_screen(width int32, height int32) *os.File {
	deviceFile, err := os.OpenFile("/dev/uinput", syscall.O_WRONLY|syscall.O_NONBLOCK, 0660)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	ioctl(deviceFile.Fd(), UISETEVBIT(), evKey)
	ioctl(deviceFile.Fd(), UISETKEYBIT(), 0x014a) //一个是BTN_TOUCH 一个不知道是啥
	ioctl(deviceFile.Fd(), UISETKEYBIT(), 0x003e) //是从手机直接copy出来的

	ioctl(deviceFile.Fd(), UISETEVBIT(), evAbs)
	ioctl(deviceFile.Fd(), UISETABSBIT(), absMtSlot)
	ioctl(deviceFile.Fd(), UISETABSBIT(), absMtTrackingId)

	ioctl(deviceFile.Fd(), UISETABSBIT(), absMtTouchMajor)
	ioctl(deviceFile.Fd(), UISETABSBIT(), absMtWidthMajor)
	ioctl(deviceFile.Fd(), UISETABSBIT(), absMtPositionX)
	ioctl(deviceFile.Fd(), UISETABSBIT(), absMtPositionY)

	ioctl(deviceFile.Fd(), UISETPROPBIT(), inputPropDirect)

	var absMin [absCnt]int32
	absMin[absMtPositionX] = 0
	absMin[absMtPositionY] = 0
	absMin[absMtTouchMajor] = 0
	absMin[absMtWidthMajor] = 0
	absMin[absMtSlot] = 0
	absMin[absMtTrackingId] = 0

	var absMax [absCnt]int32
	absMax[absMtPositionX] = width
	absMax[absMtPositionY] = height
	absMax[absMtTouchMajor] = 255
	absMax[absMtWidthMajor] = 0
	absMax[absMtSlot] = 255
	absMax[absMtTrackingId] = 65535

	uiDev := UinputUserDev{
		Name: toUInputName([]byte("v_touch_screen")),
		ID: InputID{
			BusType: 0,
			Vendor:  randUInt16Num(0x2000),
			Product: randUInt16Num(0x2000),
			Version: randUInt16Num(0x20),
		},
		EffectsMax: 0,
		AbsMax:     absMax,
		AbsMin:     absMin,
		AbsFuzz:    [absCnt]int32{},
		AbsFlat:    [absCnt]int32{},
	}
	deviceFile.Write(uInputDevToBytes(uiDev))
	createDevice(deviceFile)
	return deviceFile
}

func create_u_input_mouse_keyboard() *os.File {
	deviceFile, err := os.OpenFile("/dev/uinput", syscall.O_WRONLY|syscall.O_NONBLOCK, 0660)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	ioctl(deviceFile.Fd(), UISETEVBIT(), evSyn)
	ioctl(deviceFile.Fd(), UISETEVBIT(), evKey)
	ioctl(deviceFile.Fd(), UISETEVBIT(), evRel)
	ioctl(deviceFile.Fd(), UISETEVRELBIT(), relX)
	ioctl(deviceFile.Fd(), UISETEVRELBIT(), relY)
	ioctl(deviceFile.Fd(), UISETEVRELBIT(), relWheel)
	ioctl(deviceFile.Fd(), UISETEVRELBIT(), relHWheel)
	for i := 0x110; i < 0x117; i++ {
		ioctl(deviceFile.Fd(), UISETKEYBIT(), uintptr(i))
	}
	for i := 0; i < 256; i++ {
		ioctl(deviceFile.Fd(), UISETKEYBIT(), uintptr(i))
	}

	uiDev := UinputUserDev{
		Name: toUInputName([]byte("v_keyboard_mouse")),
		ID: InputID{
			BusType: 0,
			Vendor:  randUInt16Num(0x2000),
			Product: randUInt16Num(0x2000),
			Version: randUInt16Num(0x20),
		},
		EffectsMax: 0,
		AbsMax:     [absCnt]int32{},
		AbsMin:     [absCnt]int32{},
		AbsFuzz:    [absCnt]int32{},
		AbsFlat:    [absCnt]int32{},
	}
	deviceFile.Write(uInputDevToBytes(uiDev))
	createDevice(deviceFile)
	return deviceFile
}

func handel_u_input_mouse_keyboard(u_input chan *u_input_control_pack) {

	sizeofEvent := int(unsafe.Sizeof(evdev.Event{}))
	sendEvents := func(fd *os.File, events []*evdev.Event) {
		if fd == nil {
			fmt.Printf("fd is nil,pass %v\n", events)
			return
		}

		buf := make([]byte, sizeofEvent*len(events))
		for i, event := range events {
			copy(buf[i*sizeofEvent:], (*(*[1<<27 - 1]byte)(unsafe.Pointer(event)))[:sizeofEvent])
		}
		n, err := fd.Write(buf)
		if err != nil {
			fmt.Println(err, n)
		}
	}
	ev_sync := evdev.Event{Type: EV_SYN, Code: 0, Value: 0}
	fd := create_u_input_mouse_keyboard()
	for {
		write_events := make([]*evdev.Event, 0)
		select {
		case <-global_close_signal:
			return
		case pack := <-u_input:
			switch pack.action {
			case UInput_mouse_move:
				write_events = append(write_events, &evdev.Event{Type: EV_REL, Code: REL_X, Value: pack.arg1})
				write_events = append(write_events, &evdev.Event{Type: EV_REL, Code: REL_Y, Value: pack.arg2})
				write_events = append(write_events, &ev_sync)
				sendEvents(fd, write_events)
			case UInput_mouse_btn:
				write_events = append(write_events, &evdev.Event{Type: EV_KEY, Code: uint16(pack.arg1), Value: pack.arg2})
				write_events = append(write_events, &ev_sync)
				sendEvents(fd, write_events)
			case UInput_mouse_wheel:
				write_events = append(write_events, &evdev.Event{Type: EV_REL, Code: uint16(pack.arg1), Value: pack.arg2})
				write_events = append(write_events, &ev_sync)
				sendEvents(fd, write_events)
			case UInput_key_event:
				write_events = append(write_events, &evdev.Event{Type: EV_KEY, Code: uint16(pack.arg1), Value: pack.arg2})
				write_events = append(write_events, &ev_sync)
				sendEvents(fd, write_events)
			}
		}
	}
}

const (
	ABS_MT_POSITION_X  = 0x35
	ABS_MT_POSITION_Y  = 0x36
	ABS_MT_SLOT        = 0x2F
	ABS_MT_TRACKING_ID = 0x39
	EV_SYN             = 0x00
	EV_KEY             = 0x01
	EV_REL             = 0x02
	EV_ABS             = 0x03
	REL_X              = 0x00
	REL_Y              = 0x01
	REL_WHEEL          = 0x08
	REL_HWHEEL         = 0x06
	SYN_REPORT         = 0x00
	BTN_TOUCH          = 0x14A
)

func get_wm_size() (int32, int32) {
	cmd := exec.Command("sh", "-c", "wm size")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}
	wxh := string(out[15 : len(out)-1])
	res := strings.Split(wxh, "x")
	width, _ := strconv.Atoi(res[0])
	height, _ := strconv.Atoi(res[1])
	return int32(width), int32(height)
}

func handel_touch_using_vTouch(control_ch chan *touch_control_pack) {

	sizeofEvent := int(unsafe.Sizeof(evdev.Event{}))
	sendEvents := func(fd *os.File, events []*evdev.Event) {
		buf := make([]byte, sizeofEvent*len(events))
		for i, event := range events {
			copy(buf[i*sizeofEvent:], (*(*[1<<27 - 1]byte)(unsafe.Pointer(event)))[:sizeofEvent])
		}
		n, err := fd.Write(buf)
		if err != nil {
			fmt.Println(err, n)
		}
	}
	rot_xy := func(pack *touch_control_pack) (int32, int32) { //根据方向旋转坐标
		switch global_device_orientation {
		case 0:
			return pack.x, pack.y
		case 1:
			return pack.screen_y - pack.y, pack.x
		case 2:
			return pack.screen_x - pack.x, pack.screen_y - pack.y
		case 3:
			return pack.y, pack.screen_x - pack.x
		default:
			return pack.x, pack.y
		}
	}
	ev_sync := evdev.Event{Type: EV_SYN, Code: 0, Value: 0}
	var count int32 = 0    //BTN_TOUCH 申请时为1 则按下 释放时为0 则松开
	var last_id int32 = -1 //ABS_MT_SLOT last_id每次动作后修改 如果不等则额外发送MT_SLOT事件
	w, h := get_wm_size()
	fmt.Printf("已创建虚拟触屏 : %vx%v\n", w, h)
	fd := create_u_input_touch_screen(w, h)
	defer fd.Close()
	for {
		write_events := make([]*evdev.Event, 0)
		select {
		case <-global_close_signal:
			return
		case control_data := <-control_ch:

			if control_data.id == -1 { //在任何正常情况下 这里是拿不到ID=-1的控制包的因此可以直接丢弃
				continue
			}
			if control_data.action == TouchActionRequire {
				last_id = control_data.id
				write_events = append(write_events, &evdev.Event{Type: EV_ABS, Code: ABS_MT_SLOT, Value: control_data.id})
				write_events = append(write_events, &evdev.Event{Type: EV_ABS, Code: ABS_MT_TRACKING_ID, Value: control_data.id})
				count += 1
				if count == 1 {
					write_events = append(write_events, &evdev.Event{Type: EV_KEY, Code: BTN_TOUCH, Value: DOWN})
				}
				x, y := rot_xy(control_data)
				write_events = append(write_events, &evdev.Event{Type: EV_ABS, Code: ABS_MT_POSITION_X, Value: x})
				write_events = append(write_events, &evdev.Event{Type: EV_ABS, Code: ABS_MT_POSITION_Y, Value: y})
				write_events = append(write_events, &ev_sync)
				sendEvents(fd, write_events)
			} else if control_data.action == TouchActionRelease {
				if last_id != control_data.id {
					last_id = control_data.id
					write_events = append(write_events, &evdev.Event{Type: EV_ABS, Code: ABS_MT_SLOT, Value: control_data.id})
				}
				write_events = append(write_events, &evdev.Event{Type: EV_ABS, Code: ABS_MT_TRACKING_ID, Value: -1})
				count -= 1
				if count == 0 {
					write_events = append(write_events, &evdev.Event{Type: EV_KEY, Code: BTN_TOUCH, Value: UP})
				}
				write_events = append(write_events, &ev_sync)
				sendEvents(fd, write_events)
			} else if control_data.action == TouchActionMove {
				if last_id != control_data.id {
					last_id = control_data.id
					write_events = append(write_events, &evdev.Event{Type: EV_ABS, Code: ABS_MT_SLOT, Value: control_data.id})
				}
				x, y := rot_xy(control_data)
				write_events = append(write_events, &evdev.Event{Type: EV_ABS, Code: ABS_MT_POSITION_X, Value: x})
				write_events = append(write_events, &evdev.Event{Type: EV_ABS, Code: ABS_MT_POSITION_Y, Value: y})
				write_events = append(write_events, &ev_sync)
				sendEvents(fd, write_events)
			}
		}
	}
}
