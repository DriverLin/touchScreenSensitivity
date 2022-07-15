package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/kenshaw/evdev"
)

type TouchHandler struct {
	events                  chan *event_pack         //接收事件的channel
	touch_controll_channel  chan *touch_control_pack //发送触屏控制信号的channel
	view_id                 int32                    //视角的触摸ID
	allocated_id            []bool                   //10个触摸点分配情况
	screen_x                int32                    //屏幕宽度
	screen_y                int32                    //屏幕高度
	view_init_x             int32                    //初始化视角映射的x坐标
	view_init_y             int32                    //初始化视角映射的y坐标
	view_current_x          int32                    //当前视角映射的x坐标
	view_current_y          int32                    //当前视角映射的y坐标
	view_lock               sync.Mutex               //视角控制相关的锁 用于自动释放和控制相关
	totalMovedX             int32                    //视角移动的总x距离
	totalMovedY             int32                    //视角移动的总y距离
	auto_release_view_count int32                    //自动释放计时器 有视角移动则重置 否则100ms加一 超过1s 自动释放
}

const (
	TouchActionRequire int8 = 0
	TouchActionRelease int8 = 1
	TouchActionMove    int8 = 2
)

const (
	UInput_mouse_move  int8 = 0
	UInput_mouse_btn   int8 = 1
	UInput_mouse_wheel int8 = 2
	UInput_key_event   int8 = 3
)

const (
	DOWN int32 = 1
	UP   int32 = 0
)

const (
	Wheel_action_move    int8 = 1
	Wheel_action_release int8 = 0
)

var HAT_D_U map[string]([]int32) = map[string]([]int32){
	"0.5_1.0": []int32{1, DOWN},
	"0.5_0.0": []int32{0, DOWN},
	"1.0_0.5": []int32{1, UP},
	"0.0_0.5": []int32{0, UP},
}

var HAT0_KEYNAME map[string][]string = map[string][]string{
	"HAT0X": {"BTN_DPAD_LEFT", "BTN_DPAD_RIGHT"},
	"HAT0Y": {"BTN_DPAD_UP", "BTN_DPAD_DOWN"},
}

func rand_offset() int32 {
	return rand.Int31n(20) - 10
}

func InitTouchHandler(
	events chan *event_pack,
	touch_controller chan *touch_control_pack,
) *TouchHandler {
	width, height := get_wm_size()
	screenX, screenY := height, width
	viewInitX, viewInitY := screenX/2+100, screenY/2
	return &TouchHandler{
		events:                  events,
		touch_controll_channel:  touch_controller,
		view_id:                 -1,
		allocated_id:            make([]bool, 12),
		screen_x:                screenX,
		screen_y:                screenY,
		view_init_x:             viewInitX,
		view_init_y:             viewInitY,
		view_current_x:          viewInitX,
		view_current_y:          viewInitY,
		view_lock:               sync.Mutex{},
		totalMovedX:             0,
		totalMovedY:             0,
		auto_release_view_count: 0,
	}
}

func (self *TouchHandler) touch_require(x int32, y int32) int32 {
	for i, v := range self.allocated_id {
		if !v {
			self.allocated_id[i] = true
			self.send_touch_control_pack(TouchActionRequire, int32(i), x, y)
			return int32(i)
		}
	}
	return -1
}

func (self *TouchHandler) touch_release(id int32) int32 {
	if id != -1 {
		self.allocated_id[int(id)] = false
		self.send_touch_control_pack(TouchActionRelease, id, -1, -1)
	}
	return -1
}

func (self *TouchHandler) touch_move(id int32, x int32, y int32) {
	if id != -1 {
		self.send_touch_control_pack(TouchActionMove, id, x, y)
	}
}

func (self *TouchHandler) send_touch_control_pack(action int8, id int32, x int32, y int32) {
	self.touch_controll_channel <- &touch_control_pack{
		action:   action,
		id:       id,
		x:        x,
		y:        y,
		screen_x: self.screen_x,
		screen_y: self.screen_y,
	}
}

func (self *TouchHandler) auto_handel_view_release(timeout int) { //视角释放
	if timeout == 0 {
		return
	} else {
		for {
			select {
			case <-global_close_signal:
				return
			default:
				self.view_lock.Lock()
				if self.view_id != -1 {
					self.auto_release_view_count += 1
					if self.auto_release_view_count > int32(timeout/50) { //200ms不动 则释放
						self.auto_release_view_count = 0
						self.view_id = self.touch_release(self.view_id)
						// fmt.Printf("auto release view id:%d\n", self.view_id)
					}
				}
				self.view_lock.Unlock()
				time.Sleep(time.Duration(50) * time.Millisecond)
			}
		}
	}
}

func (self *TouchHandler) handel_view_move(offset_x int32, offset_y int32) { //视角移动
	self.view_lock.Lock()
	self.auto_release_view_count = 0
	self.totalMovedX += offset_x
	self.totalMovedY += offset_y
	fmt.Printf("totalMove [%4d,%4d]\n", self.totalMovedX, self.totalMovedY)
	if self.view_id == -1 {
		self.view_current_x = self.view_init_x
		self.view_current_y = self.view_init_y
		self.view_id = self.touch_require(self.view_current_x, self.view_current_y)
	}
	self.view_current_x += offset_x
	self.view_current_y += offset_y

	// if self.view_current_x <= 0 || self.view_current_x >= self.screen_x || self.view_current_y <= 0 || self.view_current_y >= self.screen_y {
	// 	self.view_current_x = self.view_init_x
	// 	self.view_current_y = self.view_init_y
	// 	tmp_view_id := self.touch_require(self.view_current_x, self.view_current_y)
	// 	self.view_current_x += offset_x
	// 	self.view_current_y += offset_y
	// 	self.touch_move(tmp_view_id, self.view_current_x, self.view_current_y)
	// 	self.touch_release(self.view_id)
	// 	self.view_id = tmp_view_id
	// } else {
	// 	self.touch_move(self.view_id, self.view_current_x, self.view_current_y)
	// }

	self.touch_move(self.view_id, self.view_current_x, self.view_current_y)

	self.view_lock.Unlock()
}

func (self *TouchHandler) handel_key_up_down(key_name string, up_down int32, dev_name string) {
	// fmt.Printf("handel_key_up_down: %s %d %s\n", key_name, up_down, dev_name)
	if up_down == DOWN {
		switch key_name {
		case "BTN_LEFT":
			self.handel_view_move(-1, 0)
		case "BTN_RIGHT":
			self.handel_view_move(1, 0)
		case "BTN_MIDDLE":
			self.totalMovedX = 0
			self.totalMovedY = 0
			fmt.Println("totalMove cleared")
		}
	}
}

func (self *TouchHandler) handel_key_events(events []*evdev.Event, dev_name string) {
	for _, event := range events {
		self.handel_key_up_down(GetKeyName(event.Code), event.Value, dev_name)
	}
}

func (self *TouchHandler) mix_touch(touch_events chan *event_pack) {
	id_2_vid := make([]int32, 10) //硬件ID到虚拟ID的映射
	var last_id int32 = 0
	pos_s := make([][]int32, 10)
	for i := 0; i < 10; i++ {
		pos_s[i] = make([]int32, 2)
	}
	id_stause := make([]bool, 10)
	for i := 0; i < 10; i++ {
		id_stause[i] = false
	}

	translate_xy := func(x, y int32) (int32, int32) { //根据设备方向 将eventX的坐标系转换为标准坐标系
		switch global_device_orientation { //
		case 0: //normal
			return x, y
		case 1: //left side down
			return y, self.screen_y - x
		case 2: //up side down
			return self.screen_y - x, self.screen_x - y
		case 3: //right side down
			return self.screen_x - y, x
		default:
			return x, y
		}
	}

	for {
		copy_pos_s := make([][]int32, 10)
		copy(copy_pos_s, pos_s)
		copy_id_stause := make([]bool, 10)
		copy(copy_id_stause, id_stause)
		select {
		case <-global_close_signal:
			return
		case event_pack := <-touch_events:
			for _, event := range event_pack.events {
				switch event.Code {
				case ABS_MT_POSITION_X:
					pos_s[last_id] = []int32{event.Value, pos_s[last_id][1]}
				case ABS_MT_POSITION_Y:
					pos_s[last_id] = []int32{pos_s[last_id][0], event.Value}
				case ABS_MT_TRACKING_ID:
					if event.Value == -1 {
						id_stause[last_id] = false
					} else {
						id_stause[last_id] = true
					}
				case ABS_MT_SLOT:
					last_id = event.Value
				}
			}
			for i := 0; i < 10; i++ {
				if copy_id_stause[i] != id_stause[i] {
					if id_stause[i] { //false -> true 申请
						x, y := translate_xy(pos_s[i][0], pos_s[i][1])
						id_2_vid[i] = self.touch_require(x, y)
					} else {
						self.touch_release(id_2_vid[i])
					}
				} else {
					if pos_s[i][0] != copy_pos_s[i][0] || pos_s[i][1] != copy_pos_s[i][1] {
						x, y := translate_xy(pos_s[i][0], pos_s[i][1])
						self.touch_move(id_2_vid[i], x, y)
					}
				}
			}

		}
	}
}

func (self *TouchHandler) handel_event() {
	for {
		key_events := make([]*evdev.Event, 0)
		var x int32 = 0
		var y int32 = 0
		select {
		case <-global_close_signal:
			return
		case event_pack := <-self.events:
			for _, event := range event_pack.events {
				switch event.Type {
				case evdev.EventKey:
					key_events = append(key_events, event)
				case evdev.EventRelative:
					switch event.Code {
					case uint16(evdev.RelativeX):
						x = event.Value
					case uint16(evdev.RelativeY):
						y = event.Value
					}
				}
			}
			self.handel_key_events(key_events, event_pack.dev_name)
			if x != 0 || y != 0 {
				self.handel_view_move(x, y)
			}
		}
	}
}
