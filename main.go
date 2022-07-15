package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/akamensky/argparse"
	"github.com/kenshaw/evdev"
)

type event_pack struct {
	//表示一个动作 由一系列event组成
	dev_name string
	events   []*evdev.Event
}

type touch_control_pack struct {
	//触屏控制信息
	action   int8
	id       int32
	x        int32
	y        int32
	screen_x int32
	screen_y int32
}

type u_input_control_pack struct {
	action int8
	arg1   int32
	arg2   int32
}

func create_event_reader(indexes map[int]bool) chan *event_pack {
	reader := func(event_reader chan *event_pack, index int) {
		fd, err := os.OpenFile(fmt.Sprintf("/dev/input/event%d", index), os.O_RDONLY, 0)
		if err != nil {
			log.Fatal(err)
		}
		d := evdev.Open(fd)
		defer d.Close()
		event_ch := d.Poll(context.Background())
		events := make([]*evdev.Event, 0)
		dev_name := d.Name()
		fmt.Printf("开始读取设备 : %s\n", dev_name)
		d.Lock()
		defer d.Unlock()
		for {
			select {
			case <-global_close_signal:
				fmt.Printf("释放设备 : %s \n ", dev_name)
				return
			case event := <-event_ch:
				if event.Type == evdev.SyncReport {
					pack := &event_pack{
						dev_name: dev_name,
						events:   events,
					}
					event_reader <- pack
					events = make([]*evdev.Event, 0)
				} else {
					events = append(events, &event.Event)
				}
			}
		}

	}
	event_reader := make(chan *event_pack)
	// for _, index := range indexes {
	// 	go reader(event_reader, index)
	// }
	for index, _ := range indexes {
		go reader(event_reader, index)
	}
	return event_reader
}

var global_close_signal = make(chan bool) //仅会在程序退出时关闭  不用于其他用途
var global_device_orientation int32 = 0

func get_device_orientation() int32 {
	cmd := "dumpsys input | grep -i surfaceorientation | awk '{ print $2 }'"
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return 0
	} else {
		result, err := strconv.Atoi(string(out[0:1]))
		if err != nil {
			return 0
		} else {
			return int32(result)
		}
	}
}

func listen_device_orientation() {
	for {
		select {
		case <-global_close_signal:
			return
		default:
			global_device_orientation = get_device_orientation()
			time.Sleep(time.Duration(1) * time.Second)
		}
	}
}

type dev_type uint8

const (
	type_mouse    = dev_type(0)
	type_keyboard = dev_type(1)
	type_joystick = dev_type(2)
	type_touch    = dev_type(3)
	type_unknown  = dev_type(4)
)

func check_dev_type(dev *evdev.Evdev) dev_type {
	abs := dev.AbsoluteTypes()
	key := dev.KeyTypes()
	rel := dev.RelativeTypes()
	_, MTPositionX := abs[evdev.AbsoluteMTPositionX]
	_, MTPositionY := abs[evdev.AbsoluteMTPositionY]
	_, MTSlot := abs[evdev.AbsoluteMTSlot]
	_, MTTrackingID := abs[evdev.AbsoluteMTTrackingID]
	if MTPositionX && MTPositionY && MTSlot && MTTrackingID {
		return type_touch //触屏检测这几个abs类型即可
	}
	_, RelX := rel[evdev.RelativeX]
	_, RelY := rel[evdev.RelativeY]
	_, HWheel := rel[evdev.RelativeHWheel]
	_, MouseLeft := key[evdev.BtnLeft]
	_, MouseRight := key[evdev.BtnRight]
	_, MouseMiddle := key[evdev.BtnMiddle]
	if RelX && RelY && HWheel && MouseLeft && MouseRight && MouseMiddle {
		return type_mouse //鼠标 检测XY 滚轮 左右中键
	}
	keyboard_keys := true
	for i := evdev.KeyEscape; i <= evdev.KeyScrollLock; i++ {
		_, ok := key[i]
		keyboard_keys = keyboard_keys && ok
	}
	if keyboard_keys {
		return type_keyboard //键盘 检测keycode(1-70)
	}

	axis_count := 0
	for i := evdev.AbsoluteX; i <= evdev.AbsoluteRZ; i++ {
		_, ok := abs[i]
		if ok {
			axis_count++
		}
	}
	LS_RS := axis_count >= 4

	key_count := 0
	for i := evdev.BtnA; i <= evdev.BtnZ; i++ {
		_, ok := key[i]
		if ok {
			key_count++
		}
	}
	A_B_X_Y := key_count >= 4

	if LS_RS && A_B_X_Y {
		return type_joystick //手柄 检测LS,RS A,B,X,Y
	}
	return type_unknown
}

func get_possible_device_indexes() map[int]dev_type {
	// fmt.Printf("检测设备...\n")
	files, _ := ioutil.ReadDir("/dev/input")
	result := make(map[int]dev_type)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if len(file.Name()) <= 5 {
			continue
		}
		if file.Name()[:5] != "event" {
			continue
		}
		index, _ := strconv.Atoi(file.Name()[5:])
		fd, err := os.OpenFile(fmt.Sprintf("/dev/input/%s", file.Name()), os.O_RDONLY, 0)
		if err != nil {
			log.Fatal(err)
		}
		d := evdev.Open(fd)
		defer d.Close()
		devType := check_dev_type(d)
		if devType != type_unknown {
			result[index] = devType
		}
	}
	return result
}

func get_dev_name_by_index(index int) string {
	fd, err := os.OpenFile(fmt.Sprintf("/dev/input/event%d", index), os.O_RDONLY, 0)
	if err != nil {
		return "read name error"
	}
	d := evdev.Open(fd)
	defer d.Close()
	return d.Name()
}

func main() {
	parser := argparse.NewParser("touchScreenSensitivity", " ")

	var stepValueArg *int = parser.Int("s", "step", &argparse.Options{
		Required: false,
		Help:     "输入模拟滑动每次滑动的像素点",
		Default:  24,
	})

	var delayMs *int = parser.Int("d", "delay", &argparse.Options{
		Required: false,
		Help:     "每次滑动的延迟时间，单位ms",
		Default:  16,
	})

	var autoReleaseTimeout *int = parser.Int("r", "release", &argparse.Options{
		Required: false,
		Help:     "自动释放触摸点时间，单位ms",
		Default:  1500,
	})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	auto_detect_result := get_possible_device_indexes()
	devTypeFriendlyName := map[dev_type]string{
		type_mouse:    "鼠标",
		type_keyboard: "键盘",
		type_joystick: "手柄",
		type_touch:    "触屏",
		type_unknown:  "未知",
	}
	for index, devType := range auto_detect_result {
		devName := get_dev_name_by_index(index)
		fmt.Printf("检测到设备 %s(/dev/input/event%d) : %s\n", devName, index, devTypeFriendlyName[devType])
	}

	touchScreen := make(map[int]bool)
	eventSet := make(map[int]bool)
	for k, v := range auto_detect_result {
		if v == type_keyboard || v == type_mouse {
			eventSet[k] = true
		}
		if v == type_touch {
			touchScreen[k] = true
		}
	}

	events_ch := create_event_reader(eventSet)

	touch_control_ch := make(chan *touch_control_pack)
	touch_event_ch := create_event_reader(touchScreen)

	handeler := InitTouchHandler(events_ch, touch_control_ch)
	go handeler.handel_event()
	go handeler.mix_touch(touch_event_ch)
	go handeler.auto_handel_view_release(*autoReleaseTimeout)
	go listen_device_orientation()
	// go handel_touch_using_vTouch(touch_control_ch)
	go handel_touch_using_input_manager(touch_control_ch)

	scanner := bufio.NewScanner(os.Stdin)

	var stepValue int32 = int32(*stepValueArg)
	var sleepMS int32 = int32(*delayMs)
	fmt.Printf("滑动步长: %d, 滑动延迟: %dms , 自动释放超时: %dms\n", stepValue, sleepMS, *autoReleaseTimeout)

	for scanner.Scan() {
		intVal, err := strconv.ParseInt(scanner.Text(), 10, 32)
		if err != nil {
			fmt.Printf("输入错误:%s\n", err.Error())
		} else {
			handeler.handel_view_move(0, 0)
			time.Sleep(time.Millisecond * 16)
			if intVal > 0 {
				steps := int32(intVal) / stepValue
				for i := 0; i < int(steps); i++ {
					handeler.handel_view_move(stepValue, 0)
					time.Sleep(time.Millisecond * time.Duration(sleepMS))
				}
				handeler.handel_view_move(int32(intVal)%stepValue, 0)
			} else {
				intVal = int64(math.Abs(float64(intVal)))
				steps := int32(intVal) / stepValue
				for i := 0; i < int(steps); i++ {
					handeler.handel_view_move(-stepValue, 0)
					time.Sleep(time.Millisecond * time.Duration(sleepMS))
				}
				handeler.handel_view_move(-int32(intVal)%stepValue, 0)
			}
		}
	}
	close(global_close_signal)
	fmt.Println("已停止")
	time.Sleep(time.Millisecond * 40)
}
