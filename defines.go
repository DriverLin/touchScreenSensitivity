package main

// https://github.com/kp7742/TouchSimulation/blob/main/UinputDefs.go

import (
	"syscall"
)

//---------------------------------EVCodes--------------------------------------//

//Ref: input-event-codes.h
const (
	evSyn           = 0x00
	evKey           = 0x01
	evRel           = 0x02
	evAbs           = 0x03
	evMsc           = 0x04
	relX            = 0x00
	relY            = 0x01
	relWheel        = 0x08
	relHWheel       = 0x06
	btnTouch        = 0x14a
	synReport       = 0
	synMtReport     = 2
	absMtSlot       = 0x2f
	absMtPositionX  = 0x35
	absMtPositionY  = 0x36
	absMtTrackingId = 0x39
	absMax          = 0x3f
	absCnt          = absMax + 1
	inputPropDirect = 0x01
	inputPropMax    = 0x1f
	inputPropCnt    = inputPropMax + 1
	//-------------------------------appends-------------------------------//
	absMtTouchMajor = 0x30
	absMtWidthMajor = 0x32
)

//---------------------------------IOCTL--------------------------------------//

//Ref: ioctl.h
const (
	iocNone  = 0x0
	iocWrite = 0x1
	iocRead  = 0x2

	iocNrbits   = 8
	iocTypebits = 8
	iocSizebits = 14
	iocNrshift  = 0

	iocTypeshift = iocNrshift + iocNrbits
	iocSizeshift = iocTypeshift + iocTypebits
	iocDirshift  = iocSizeshift + iocSizebits
)

func _IOC(dir int, t int, nr int, size int) int {
	return (dir << iocDirshift) | (t << iocTypeshift) |
		(nr << iocNrshift) | (size << iocSizeshift)
}

func _IOR(t int, nr int, size int) int {
	return _IOC(iocRead, t, nr, size)
}

func _IOW(t int, nr int, size int) int {
	return _IOC(iocWrite, t, nr, size)
}

//Ref: input.h
func EVIOCGNAME() int {
	return _IOC(iocRead, 'E', 0x06, uinputMaxNameSize)
}

func EVIOCGPROP() int {
	return _IOC(iocRead, 'E', 0x09, inputPropMax)
}

func EVIOCGABS(abs int) int {
	return _IOR('E', 0x40+abs, 24) //sizeof(struct input_absinfo)
}

func EVIOCGBIT(ev, len int) int {
	return _IOC(iocRead, 'E', 0x20+ev, len)
}

func EVIOCGRAB() int {
	return _IOW('E', 0x90, 4) //sizeof(int)
}

//Syscall
func ioctl(fd uintptr, name int, data uintptr) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(name), data)
	if err != 0 {
		return err
	}
	return nil
}

//---------------------------------Input--------------------------------------//

type InputID struct {
	BusType uint16
	Vendor  uint16
	Product uint16
	Version uint16
}

type AbsInfo struct {
	Value      int32
	Minimum    int32
	Maximum    int32
	Fuzz       int32
	Flat       int32
	Resolution int32
}

type InputEvent struct {
	Time  syscall.Timeval
	Type  uint16
	Code  uint16
	Value int32
}

//---------------------------------UInput--------------------------------------//

//Ref: uinput.h
const (
	uinputMaxNameSize = 80
)

type UinputUserDev struct {
	Name       [uinputMaxNameSize]byte
	ID         InputID
	EffectsMax uint32
	AbsMax     [absCnt]int32
	AbsMin     [absCnt]int32
	AbsFuzz    [absCnt]int32
	AbsFlat    [absCnt]int32
}

//Ref: uinput.h
func UISETEVBIT() int {
	return _IOW('U', 100, 4) //sizeof(int)
}

func UISETKEYBIT() int {
	return _IOW('U', 101, 4) //sizeof(int)
}

func UISETEVRELBIT() int {
	return _IOW('U', 102, 4) //sizeof(int)
}

func UISETABSBIT() int {
	return _IOW('U', 103, 4) //sizeof(int)
}

func UISETPROPBIT() int {
	return _IOW('U', 110, 4) //sizeof(int)
}

func UIDEVCREATE() int {
	return _IOC(iocNone, 'U', 1, 0)
}

func UIDEVDESTROY() int {
	return _IOC(iocNone, 'U', 2, 0)
}

var keycode_2_friendly_name map[uint16]string = map[uint16]string{
	0:     "KEY_RESERVED",
	1:     "KEY_ESC",
	2:     "KEY_1",
	3:     "KEY_2",
	4:     "KEY_3",
	5:     "KEY_4",
	6:     "KEY_5",
	7:     "KEY_6",
	8:     "KEY_7",
	9:     "KEY_8",
	10:    "KEY_9",
	11:    "KEY_0",
	12:    "KEY_MINUS",
	13:    "KEY_EQUAL",
	14:    "KEY_BACKSPACE",
	15:    "KEY_TAB",
	16:    "KEY_Q",
	17:    "KEY_W",
	18:    "KEY_E",
	19:    "KEY_R",
	20:    "KEY_T",
	21:    "KEY_Y",
	22:    "KEY_U",
	23:    "KEY_I",
	24:    "KEY_O",
	25:    "KEY_P",
	26:    "KEY_LEFTBRACE",
	27:    "KEY_RIGHTBRACE",
	28:    "KEY_ENTER",
	29:    "KEY_LEFTCTRL",
	30:    "KEY_A",
	31:    "KEY_S",
	32:    "KEY_D",
	33:    "KEY_F",
	34:    "KEY_G",
	35:    "KEY_H",
	36:    "KEY_J",
	37:    "KEY_K",
	38:    "KEY_L",
	39:    "KEY_SEMICOLON",
	40:    "KEY_APOSTROPHE",
	41:    "KEY_GRAVE",
	42:    "KEY_LEFTSHIFT",
	43:    "KEY_BACKSLASH",
	44:    "KEY_Z",
	45:    "KEY_X",
	46:    "KEY_C",
	47:    "KEY_V",
	48:    "KEY_B",
	49:    "KEY_N",
	50:    "KEY_M",
	51:    "KEY_COMMA",
	52:    "KEY_DOT",
	53:    "KEY_SLASH",
	54:    "KEY_RIGHTSHIFT",
	55:    "KEY_KPASTERISK",
	56:    "KEY_LEFTALT",
	57:    "KEY_SPACE",
	58:    "KEY_CAPSLOCK",
	59:    "KEY_F1",
	60:    "KEY_F2",
	61:    "KEY_F3",
	62:    "KEY_F4",
	63:    "KEY_F5",
	64:    "KEY_F6",
	65:    "KEY_F7",
	66:    "KEY_F8",
	67:    "KEY_F9",
	68:    "KEY_F10",
	69:    "KEY_NUMLOCK",
	70:    "KEY_SCROLLLOCK",
	71:    "KEY_KP7",
	72:    "KEY_KP8",
	73:    "KEY_KP9",
	74:    "KEY_KPMINUS",
	75:    "KEY_KP4",
	76:    "KEY_KP5",
	77:    "KEY_KP6",
	78:    "KEY_KPPLUS",
	79:    "KEY_KP1",
	80:    "KEY_KP2",
	81:    "KEY_KP3",
	82:    "KEY_KP0",
	83:    "KEY_KPDOT",
	85:    "KEY_ZENKAKUHANKAKU",
	86:    "KEY_102ND",
	87:    "KEY_F11",
	88:    "KEY_F12",
	89:    "KEY_RO",
	90:    "KEY_KATAKANA",
	91:    "KEY_HIRAGANA",
	92:    "KEY_HENKAN",
	93:    "KEY_KATAKANAHIRAGANA",
	94:    "KEY_MUHENKAN",
	95:    "KEY_KPJPCOMMA",
	96:    "KEY_KPENTER",
	97:    "KEY_RIGHTCTRL",
	98:    "KEY_KPSLASH",
	99:    "KEY_SYSRQ",
	100:   "KEY_RIGHTALT",
	101:   "KEY_LINEFEED",
	102:   "KEY_HOME",
	103:   "KEY_UP",
	104:   "KEY_PAGEUP",
	105:   "KEY_LEFT",
	106:   "KEY_RIGHT",
	107:   "KEY_END",
	108:   "KEY_DOWN",
	109:   "KEY_PAGEDOWN",
	110:   "KEY_INSERT",
	111:   "KEY_DELETE",
	112:   "KEY_MACRO",
	113:   "KEY_MUTE",
	114:   "KEY_VOLUMEDOWN",
	115:   "KEY_VOLUMEUP",
	116:   "KEY_POWER",
	117:   "KEY_KPEQUAL",
	118:   "KEY_KPPLUSMINUS",
	119:   "KEY_PAUSE",
	120:   "KEY_SCALE",
	121:   "KEY_KPCOMMA",
	122:   "KEY_HANGEUL",
	123:   "KEY_HANJA",
	124:   "KEY_YEN",
	125:   "KEY_LEFTMETA",
	126:   "KEY_RIGHTMETA",
	127:   "KEY_COMPOSE",
	128:   "KEY_STOP",
	129:   "KEY_AGAIN",
	130:   "KEY_PROPS",
	131:   "KEY_UNDO",
	132:   "KEY_FRONT",
	133:   "KEY_COPY",
	134:   "KEY_OPEN",
	135:   "KEY_PASTE",
	136:   "KEY_FIND",
	137:   "KEY_CUT",
	138:   "KEY_HELP",
	139:   "KEY_MENU",
	140:   "KEY_CALC",
	141:   "KEY_SETUP",
	142:   "KEY_SLEEP",
	143:   "KEY_WAKEUP",
	144:   "KEY_FILE",
	145:   "KEY_SENDFILE",
	146:   "KEY_DELETEFILE",
	147:   "KEY_XFER",
	148:   "KEY_PROG1",
	149:   "KEY_PROG2",
	150:   "KEY_WWW",
	151:   "KEY_MSDOS",
	152:   "KEY_SCREENLOCK",
	153:   "KEY_DIRECTION",
	154:   "KEY_CYCLEWINDOWS",
	155:   "KEY_MAIL",
	156:   "KEY_BOOKMARKS",
	157:   "KEY_COMPUTER",
	158:   "KEY_BACK",
	159:   "KEY_FORWARD",
	160:   "KEY_CLOSECD",
	161:   "KEY_EJECTCD",
	162:   "KEY_EJECTCLOSECD",
	163:   "KEY_NEXTSONG",
	164:   "KEY_PLAYPAUSE",
	165:   "KEY_PREVIOUSSONG",
	166:   "KEY_STOPCD",
	167:   "KEY_RECORD",
	168:   "KEY_REWIND",
	169:   "KEY_PHONE",
	170:   "KEY_ISO",
	171:   "KEY_CONFIG",
	172:   "KEY_HOMEPAGE",
	173:   "KEY_REFRESH",
	174:   "KEY_EXIT",
	175:   "KEY_MOVE",
	176:   "KEY_EDIT",
	177:   "KEY_SCROLLUP",
	178:   "KEY_SCROLLDOWN",
	179:   "KEY_KPLEFTPAREN",
	180:   "KEY_KPRIGHTPAREN",
	181:   "KEY_NEW",
	182:   "KEY_REDO",
	183:   "KEY_F13",
	184:   "KEY_F14",
	185:   "KEY_F15",
	186:   "KEY_F16",
	187:   "KEY_F17",
	188:   "KEY_F18",
	189:   "KEY_F19",
	190:   "KEY_F20",
	191:   "KEY_F21",
	192:   "KEY_F22",
	193:   "KEY_F23",
	194:   "KEY_F24",
	200:   "KEY_PLAYCD",
	201:   "KEY_PAUSECD",
	202:   "KEY_PROG3",
	203:   "KEY_PROG4",
	204:   "KEY_DASHBOARD",
	205:   "KEY_SUSPEND",
	206:   "KEY_CLOSE",
	207:   "KEY_PLAY",
	208:   "KEY_FASTFORWARD",
	209:   "KEY_BASSBOOST",
	210:   "KEY_PRINT",
	211:   "KEY_HP",
	212:   "KEY_CAMERA",
	213:   "KEY_SOUND",
	214:   "KEY_QUESTION",
	215:   "KEY_EMAIL",
	216:   "KEY_CHAT",
	217:   "KEY_SEARCH",
	218:   "KEY_CONNECT",
	219:   "KEY_FINANCE",
	220:   "KEY_SPORT",
	221:   "KEY_SHOP",
	222:   "KEY_ALTERASE",
	223:   "KEY_CANCEL",
	224:   "KEY_BRIGHTNESSDOWN",
	225:   "KEY_BRIGHTNESSUP",
	226:   "KEY_MEDIA",
	227:   "KEY_SWITCHVIDEOMODE",
	228:   "KEY_KBDILLUMTOGGLE",
	229:   "KEY_KBDILLUMDOWN",
	230:   "KEY_KBDILLUMUP",
	231:   "KEY_SEND",
	232:   "KEY_REPLY",
	233:   "KEY_FORWARDMAIL",
	234:   "KEY_SAVE",
	235:   "KEY_DOCUMENTS",
	236:   "KEY_BATTERY",
	237:   "KEY_BLUETOOTH",
	238:   "KEY_WLAN",
	239:   "KEY_UWB",
	240:   "KEY_UNKNOWN",
	241:   "KEY_VIDEO_NEXT",
	242:   "KEY_VIDEO_PREV",
	243:   "KEY_BRIGHTNESS_CYCLE",
	244:   "KEY_BRIGHTNESS_ZERO",
	245:   "KEY_DISPLAY_OFF",
	246:   "KEY_WWAN",
	247:   "KEY_RFKILL",
	248:   "KEY_MICMUTE",
	0x110: "BTN_LEFT",
	0x111: "BTN_RIGHT",
	0x112: "BTN_MIDDLE",
	0x113: "BTN_SIDE",
	0x114: "BTN_EXTRA",
	0x115: "BTN_FORWARD",
	0x116: "BTN_BACK",
	0x117: "BTN_TASK",
}

var friendly_name_2_keycode map[string]uint16 = map[string]uint16{
	"KEY_RESERVED":         0,
	"KEY_ESC":              1,
	"KEY_1":                2,
	"KEY_2":                3,
	"KEY_3":                4,
	"KEY_4":                5,
	"KEY_5":                6,
	"KEY_6":                7,
	"KEY_7":                8,
	"KEY_8":                9,
	"KEY_9":                10,
	"KEY_0":                11,
	"KEY_MINUS":            12,
	"KEY_EQUAL":            13,
	"KEY_BACKSPACE":        14,
	"KEY_TAB":              15,
	"KEY_Q":                16,
	"KEY_W":                17,
	"KEY_E":                18,
	"KEY_R":                19,
	"KEY_T":                20,
	"KEY_Y":                21,
	"KEY_U":                22,
	"KEY_I":                23,
	"KEY_O":                24,
	"KEY_P":                25,
	"KEY_LEFTBRACE":        26,
	"KEY_RIGHTBRACE":       27,
	"KEY_ENTER":            28,
	"KEY_LEFTCTRL":         29,
	"KEY_A":                30,
	"KEY_S":                31,
	"KEY_D":                32,
	"KEY_F":                33,
	"KEY_G":                34,
	"KEY_H":                35,
	"KEY_J":                36,
	"KEY_K":                37,
	"KEY_L":                38,
	"KEY_SEMICOLON":        39,
	"KEY_APOSTROPHE":       40,
	"KEY_GRAVE":            41,
	"KEY_LEFTSHIFT":        42,
	"KEY_BACKSLASH":        43,
	"KEY_Z":                44,
	"KEY_X":                45,
	"KEY_C":                46,
	"KEY_V":                47,
	"KEY_B":                48,
	"KEY_N":                49,
	"KEY_M":                50,
	"KEY_COMMA":            51,
	"KEY_DOT":              52,
	"KEY_SLASH":            53,
	"KEY_RIGHTSHIFT":       54,
	"KEY_KPASTERISK":       55,
	"KEY_LEFTALT":          56,
	"KEY_SPACE":            57,
	"KEY_CAPSLOCK":         58,
	"KEY_F1":               59,
	"KEY_F2":               60,
	"KEY_F3":               61,
	"KEY_F4":               62,
	"KEY_F5":               63,
	"KEY_F6":               64,
	"KEY_F7":               65,
	"KEY_F8":               66,
	"KEY_F9":               67,
	"KEY_F10":              68,
	"KEY_NUMLOCK":          69,
	"KEY_SCROLLLOCK":       70,
	"KEY_KP7":              71,
	"KEY_KP8":              72,
	"KEY_KP9":              73,
	"KEY_KPMINUS":          74,
	"KEY_KP4":              75,
	"KEY_KP5":              76,
	"KEY_KP6":              77,
	"KEY_KPPLUS":           78,
	"KEY_KP1":              79,
	"KEY_KP2":              80,
	"KEY_KP3":              81,
	"KEY_KP0":              82,
	"KEY_KPDOT":            83,
	"KEY_ZENKAKUHANKAKU":   85,
	"KEY_102ND":            86,
	"KEY_F11":              87,
	"KEY_F12":              88,
	"KEY_RO":               89,
	"KEY_KATAKANA":         90,
	"KEY_HIRAGANA":         91,
	"KEY_HENKAN":           92,
	"KEY_KATAKANAHIRAGANA": 93,
	"KEY_MUHENKAN":         94,
	"KEY_KPJPCOMMA":        95,
	"KEY_KPENTER":          96,
	"KEY_RIGHTCTRL":        97,
	"KEY_KPSLASH":          98,
	"KEY_SYSRQ":            99,
	"KEY_RIGHTALT":         100,
	"KEY_LINEFEED":         101,
	"KEY_HOME":             102,
	"KEY_UP":               103,
	"KEY_PAGEUP":           104,
	"KEY_LEFT":             105,
	"KEY_RIGHT":            106,
	"KEY_END":              107,
	"KEY_DOWN":             108,
	"KEY_PAGEDOWN":         109,
	"KEY_INSERT":           110,
	"KEY_DELETE":           111,
	"KEY_MACRO":            112,
	"KEY_MUTE":             113,
	"KEY_VOLUMEDOWN":       114,
	"KEY_VOLUMEUP":         115,
	"KEY_POWER":            116,
	"KEY_KPEQUAL":          117,
	"KEY_KPPLUSMINUS":      118,
	"KEY_PAUSE":            119,
	"KEY_SCALE":            120,
	"KEY_KPCOMMA":          121,
	"KEY_HANGEUL":          122,
	"KEY_HANGUEL":          122,
	"KEY_HANJA":            123,
	"KEY_YEN":              124,
	"KEY_LEFTMETA":         125,
	"KEY_RIGHTMETA":        126,
	"KEY_COMPOSE":          127,
	"KEY_STOP":             128,
	"KEY_AGAIN":            129,
	"KEY_PROPS":            130,
	"KEY_UNDO":             131,
	"KEY_FRONT":            132,
	"KEY_COPY":             133,
	"KEY_OPEN":             134,
	"KEY_PASTE":            135,
	"KEY_FIND":             136,
	"KEY_CUT":              137,
	"KEY_HELP":             138,
	"KEY_MENU":             139,
	"KEY_CALC":             140,
	"KEY_SETUP":            141,
	"KEY_SLEEP":            142,
	"KEY_WAKEUP":           143,
	"KEY_FILE":             144,
	"KEY_SENDFILE":         145,
	"KEY_DELETEFILE":       146,
	"KEY_XFER":             147,
	"KEY_PROG1":            148,
	"KEY_PROG2":            149,
	"KEY_WWW":              150,
	"KEY_MSDOS":            151,
	"KEY_COFFEE":           152,
	"KEY_SCREENLOCK":       152,
	"KEY_ROTATE_DISPLAY":   153,
	"KEY_DIRECTION":        153,
	"KEY_CYCLEWINDOWS":     154,
	"KEY_MAIL":             155,
	"KEY_BOOKMARKS":        156,
	"KEY_COMPUTER":         157,
	"KEY_BACK":             158,
	"KEY_FORWARD":          159,
	"KEY_CLOSECD":          160,
	"KEY_EJECTCD":          161,
	"KEY_EJECTCLOSECD":     162,
	"KEY_NEXTSONG":         163,
	"KEY_PLAYPAUSE":        164,
	"KEY_PREVIOUSSONG":     165,
	"KEY_STOPCD":           166,
	"KEY_RECORD":           167,
	"KEY_REWIND":           168,
	"KEY_PHONE":            169,
	"KEY_ISO":              170,
	"KEY_CONFIG":           171,
	"KEY_HOMEPAGE":         172,
	"KEY_REFRESH":          173,
	"KEY_EXIT":             174,
	"KEY_MOVE":             175,
	"KEY_EDIT":             176,
	"KEY_SCROLLUP":         177,
	"KEY_SCROLLDOWN":       178,
	"KEY_KPLEFTPAREN":      179,
	"KEY_KPRIGHTPAREN":     180,
	"KEY_NEW":              181,
	"KEY_REDO":             182,
	"KEY_F13":              183,
	"KEY_F14":              184,
	"KEY_F15":              185,
	"KEY_F16":              186,
	"KEY_F17":              187,
	"KEY_F18":              188,
	"KEY_F19":              189,
	"KEY_F20":              190,
	"KEY_F21":              191,
	"KEY_F22":              192,
	"KEY_F23":              193,
	"KEY_F24":              194,
	"KEY_PLAYCD":           200,
	"KEY_PAUSECD":          201,
	"KEY_PROG3":            202,
	"KEY_PROG4":            203,
	"KEY_DASHBOARD":        204,
	"KEY_SUSPEND":          205,
	"KEY_CLOSE":            206,
	"KEY_PLAY":             207,
	"KEY_FASTFORWARD":      208,
	"KEY_BASSBOOST":        209,
	"KEY_PRINT":            210,
	"KEY_HP":               211,
	"KEY_CAMERA":           212,
	"KEY_SOUND":            213,
	"KEY_QUESTION":         214,
	"KEY_EMAIL":            215,
	"KEY_CHAT":             216,
	"KEY_SEARCH":           217,
	"KEY_CONNECT":          218,
	"KEY_FINANCE":          219,
	"KEY_SPORT":            220,
	"KEY_SHOP":             221,
	"KEY_ALTERASE":         222,
	"KEY_CANCEL":           223,
	"KEY_BRIGHTNESSDOWN":   224,
	"KEY_BRIGHTNESSUP":     225,
	"KEY_MEDIA":            226,
	"KEY_SWITCHVIDEOMODE":  227,
	"KEY_KBDILLUMTOGGLE":   228,
	"KEY_KBDILLUMDOWN":     229,
	"KEY_KBDILLUMUP":       230,
	"KEY_SEND":             231,
	"KEY_REPLY":            232,
	"KEY_FORWARDMAIL":      233,
	"KEY_SAVE":             234,
	"KEY_DOCUMENTS":        235,
	"KEY_BATTERY":          236,
	"KEY_BLUETOOTH":        237,
	"KEY_WLAN":             238,
	"KEY_UWB":              239,
	"KEY_UNKNOWN":          240,
	"KEY_VIDEO_NEXT":       241,
	"KEY_VIDEO_PREV":       242,
	"KEY_BRIGHTNESS_CYCLE": 243,
	"KEY_BRIGHTNESS_AUTO":  244,
	"KEY_BRIGHTNESS_ZERO":  244,
	"KEY_DISPLAY_OFF":      245,
	"KEY_WWAN":             246,
	"KEY_WIMAX":            246,
	"KEY_RFKILL":           247,
	"KEY_MICMUTE":          248,
	"BTN_LEFT":             0x110,
	"BTN_RIGHT":            0x111,
	"BTN_MIDDLE":           0x112,
	"BTN_SIDE":             0x113,
	"BTN_EXTRA":            0x114,
	"BTN_FORWARD":          0x115,
	"BTN_BACK":             0x116,
	"BTN_TASK":             0x117,
}

func GetKeyName(keycode interface{}) string {
	switch keycode.(type) {
	case string:
		return keycode.(string)
	case uint16:
		if friendly_name, ok := keycode_2_friendly_name[keycode.(uint16)]; ok {
			return friendly_name
		} else {
			return ""
		}
	default:
		return ""
	}
}
