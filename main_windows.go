// CyberCircuit Gotham - native Windows desktop application.
// Built with Go + Win32 APIs; no third-party runtime or service required.
package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	WS_TABSTOP          = 0x00010000
	WS_BORDER           = 0x00800000
	ES_AUTOHSCROLL      = 0x0080
	ES_MULTILINE        = 0x0004
	ES_AUTOVSCROLL      = 0x0040
	ES_WANTRETURN       = 0x1000
	ES_PASSWORD         = 0x0020
	WS_VSCROLL          = 0x00200000
	WS_EX_CLIENTEDGE    = 0x00000200
	CW_USEDEFAULT       = 0x80000000
	SW_SHOW             = 5
	SW_HIDE             = 0
	SWP_NOZORDER        = 0x0004
	SWP_NOACTIVATE      = 0x0010
	WM_CREATE           = 0x0001
	WM_DESTROY          = 0x0002
	WM_SIZE             = 0x0005
	WM_PAINT            = 0x000F
	WM_ERASEBKGND       = 0x0014
	WM_TIMER            = 0x0113
	WM_LBUTTONUP        = 0x0202
	WM_CTLCOLOREDIT     = 0x0133
	WM_APP_REFRESH      = 0x8001
	COLOR_WINDOW        = 5
	TRANSPARENT         = 1
	PS_SOLID            = 0
	DT_LEFT             = 0x0000
	DT_TOP              = 0x0000
	DT_SINGLELINE       = 0x0020
	DT_VCENTER          = 0x0004
	DT_END_ELLIPSIS     = 0x8000
	FW_NORMAL           = 400
	FW_SEMIBOLD         = 600
	FW_BOLD             = 700
	IDC_ARROW           = 32512
	ID_SEARCH           = 1001
	ID_NODE             = 1002
	ID_LINK             = 1003
	ID_HOST             = 1004
	ID_NOTES            = 1005
	ID_CASE_TITLE       = 1006
	ID_CASE_YEAR        = 1007
	ID_API_TOKEN        = 1008
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	gdi32                   = syscall.NewLazyDLL("gdi32.dll")
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	shell32                 = syscall.NewLazyDLL("shell32.dll")
	advapi32                = syscall.NewLazyDLL("advapi32.dll")
	procRegisterClassEx     = user32.NewProc("RegisterClassExW")
	procCreateWindowEx      = user32.NewProc("CreateWindowExW")
	procDefWindowProc       = user32.NewProc("DefWindowProcW")
	procShowWindow          = user32.NewProc("ShowWindow")
	procUpdateWindow        = user32.NewProc("UpdateWindow")
	procGetMessage          = user32.NewProc("GetMessageW")
	procTranslateMessage    = user32.NewProc("TranslateMessage")
	procDispatchMessage     = user32.NewProc("DispatchMessageW")
	procPostQuitMessage     = user32.NewProc("PostQuitMessage")
	procBeginPaint          = user32.NewProc("BeginPaint")
	procEndPaint            = user32.NewProc("EndPaint")
	procInvalidateRect      = user32.NewProc("InvalidateRect")
	procGetClientRect       = user32.NewProc("GetClientRect")
	procSetTimer            = user32.NewProc("SetTimer")
	procSetWindowPos        = user32.NewProc("SetWindowPos")
	procSetWindowText       = user32.NewProc("SetWindowTextW")
	procGetWindowText       = user32.NewProc("GetWindowTextW")
	procGetWindowTextLength = user32.NewProc("GetWindowTextLengthW")
	procShowControl         = user32.NewProc("ShowWindow")
	procSetFocus            = user32.NewProc("SetFocus")
	procSendMessage         = user32.NewProc("SendMessageW")
	procLoadCursor          = user32.NewProc("LoadCursorW")
	procGetModuleHandle     = kernel32.NewProc("GetModuleHandleW")
	procCreateSolidBrush    = gdi32.NewProc("CreateSolidBrush")
	procDeleteObject        = gdi32.NewProc("DeleteObject")
	procCreatePen           = gdi32.NewProc("CreatePen")
	procSelectObject        = gdi32.NewProc("SelectObject")
	procRoundRect           = gdi32.NewProc("RoundRect")
	procRectangle           = gdi32.NewProc("Rectangle")
	procEllipse             = gdi32.NewProc("Ellipse")
	procMoveToEx            = gdi32.NewProc("MoveToEx")
	procLineTo              = gdi32.NewProc("LineTo")
	procSetTextColor        = gdi32.NewProc("SetTextColor")
	procSetBkMode           = gdi32.NewProc("SetBkMode")
	procTextOut             = gdi32.NewProc("TextOutW")
	procCreateFont          = gdi32.NewProc("CreateFontW")
	procGetSystemTimes      = kernel32.NewProc("GetSystemTimes")
	procGlobalMemoryStatus  = kernel32.NewProc("GlobalMemoryStatusEx")
	procGetDiskFreeSpace    = kernel32.NewProc("GetDiskFreeSpaceExW")
	procShellExecute        = shell32.NewProc("ShellExecuteW")
	procGetWindowLong       = user32.NewProc("GetWindowLongW")
	procGetStockObject      = gdi32.NewProc("GetStockObject")
)

type POINT struct{ X, Y int32 }
type RECT struct{ Left, Top, Right, Bottom int32 }
type MSG struct {
	Hwnd           syscall.Handle
	Message        uint32
	WParam, LParam uintptr
	Time           uint32
	Pt             POINT
}
type WNDCLASSEX struct {
	CbSize     uint32
	Style      uint32
	WndProc    uintptr
	CbClsExtra int32
	CbWndExtra int32
	Instance   syscall.Handle
	Icon       syscall.Handle
	Cursor     syscall.Handle
	Background syscall.Handle
	MenuName   *uint16
	ClassName  *uint16
	IconSm     syscall.Handle
}
type PAINTSTRUCT struct {
	Hdc       syscall.Handle
	Erase     int32
	Paint     RECT
	Restore   int32
	IncUpdate int32
	Reserved  [32]byte
}
type FILETIME struct{ Low, High uint32 }
type MEMORYSTATUSEX struct {
	Length                                                                                               uint32
	MemoryLoad                                                                                           uint32
	TotalPhys, AvailPhys, TotalPageFile, AvailPageFile, TotalVirtual, AvailVirtual, AvailExtendedVirtual uint64
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}
type CaseLog struct {
	Query string `json:"query"`
	When  string `json:"when"`
}
type Workspace struct {
	Nodes    []string  `json:"nodes"`
	Edges    []Edge    `json:"edges"`
	Searches []CaseLog `json:"searches"`
	Notes    string    `json:"notes"`
}
type Metrics struct {
	CPU       float64
	MemUsed   uint64
	MemTotal  uint64
	DiskUsed  uint64
	DiskTotal uint64
	LocalIP   string
	Adapter   string
	Online    bool
	Updated   time.Time
}
type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}
type LegalCase struct {
	CaseName          string   `json:"case_name"`
	DateFiled         string   `json:"date_filed"`
	Court             string   `json:"court"`
	CourtID           string   `json:"court_id"`
	DocketNumber      string   `json:"docket_number"`
	Citation          []string `json:"citation"`
	Snippet           string   `json:"snippet"`
	Syllabus          string   `json:"syllabus"`
	SuitNature        string   `json:"suit_nature"`
	Posture           string   `json:"posture"`
	ProceduralHistory string   `json:"procedural_history"`
	Status            string   `json:"status"`
	Judge             string   `json:"judge"`
	URL               string   `json:"url"`
	Location          GeoPoint `json:"location"`
	HasLocation       bool     `json:"has_location"`
}
type CourtListenerSearch struct {
	Count   int `json:"count"`
	Results []struct {
		CaseName          string   `json:"caseName"`
		CaseNameFull      string   `json:"caseNameFull"`
		DateFiled         string   `json:"dateFiled"`
		Court             string   `json:"court"`
		CourtID           string   `json:"court_id"`
		DocketNumber      string   `json:"docketNumber"`
		Citation          []string `json:"citation"`
		AbsoluteURL       string   `json:"absolute_url"`
		Snippet           string   `json:"snippet"`
		Syllabus          string   `json:"syllabus"`
		SuitNature        string   `json:"suitNature"`
		Posture           string   `json:"posture"`
		ProceduralHistory string   `json:"procedural_history"`
		Status            string   `json:"status"`
		Judge             string   `json:"judge"`
	} `json:"results"`
}

type AppState struct {
	mu                                      sync.RWMutex
	page                                    int
	workspace                               Workspace
	metrics                                 Metrics
	processes                               []string
	devices                                 []string
	deviceCount                             int
	hostResult                              string
	hostChecked                             string
	analystOutput                           []string
	cases                                   []LegalCase
	caseCount                               int
	caseStatus                              string
	lastQuery                               string
	mapStatus                               string
	status                                  string
	lastCPUIdle, lastCPUKernel, lastCPUUser uint64
	lastNetCheck                            time.Time
}

var (
	hwnd                                                                                           syscall.Handle
	hInst                                                                                          syscall.Handle
	editSearch, editNode, editLink, editHost, editNotes, editCaseTitle, editCaseYear, editAPIToken syscall.Handle
	editBrush                                                                                      syscall.Handle
	fontRegular, fontSmall, fontBold, fontLarge                                                    syscall.Handle
	app                                                                                            = &AppState{workspace: Workspace{Nodes: []string{}, Edges: []Edge{}, Searches: []CaseLog{}}, status: "LOCAL SYSTEM READY", cases: []LegalCase{}, caseStatus: "READY — search published opinions", mapStatus: "Map is available without search results; courthouse pins appear when locations are resolved."}
	appDataPath                                                                                    string
	className                                                                                      = syscall.StringToUTF16Ptr("CyberCircuitGothamWindow")
)

func rgb(r, g, b byte) uintptr { return uintptr(uint32(r) | uint32(g)<<8 | uint32(b)<<16) }
func utf(s string) *uint16     { p, _ := syscall.UTF16PtrFromString(s); return p }
func lowWord(v uintptr) int    { return int(int16(v & 0xffff)) }
func highWord(v uintptr) int   { return int(int16((v >> 16) & 0xffff)) }
func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
func pointIn(x, y, bx, by, bw, bh int) bool { return x >= bx && x <= bx+bw && y >= by && y <= by+bh }

func createFont(height int32, weight int32, face string) syscall.Handle {
	p := utf(face)
	r, _, _ := procCreateFont.Call(uintptr(height), 0, 0, 0, uintptr(weight), 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(p)))
	return syscall.Handle(r)
}
func makeBrush(color uintptr) syscall.Handle {
	r, _, _ := procCreateSolidBrush.Call(color)
	return syscall.Handle(r)
}
func fill(hdc syscall.Handle, x, y, w, h int, color uintptr, radius int) {
	brush := makeBrush(color)
	old, _, _ := procSelectObject.Call(uintptr(hdc), uintptr(brush))
	pen, _, _ := procCreatePen.Call(PS_SOLID, 1, color)
	oldPen, _, _ := procSelectObject.Call(uintptr(hdc), pen)
	if radius > 0 {
		procRoundRect.Call(uintptr(hdc), uintptr(x), uintptr(y), uintptr(x+w), uintptr(y+h), uintptr(radius), uintptr(radius))
	} else {
		procRectangle.Call(uintptr(hdc), uintptr(x), uintptr(y), uintptr(x+w), uintptr(y+h))
	}
	procSelectObject.Call(uintptr(hdc), old)
	procSelectObject.Call(uintptr(hdc), oldPen)
	procDeleteObject.Call(uintptr(brush))
	procDeleteObject.Call(pen)
}
func line(hdc syscall.Handle, x1, y1, x2, y2 int, color uintptr, width int) {
	pen, _, _ := procCreatePen.Call(PS_SOLID, uintptr(width), color)
	old, _, _ := procSelectObject.Call(uintptr(hdc), pen)
	procMoveToEx.Call(uintptr(hdc), uintptr(x1), uintptr(y1), 0)
	procLineTo.Call(uintptr(hdc), uintptr(x2), uintptr(y2))
	procSelectObject.Call(uintptr(hdc), old)
	procDeleteObject.Call(pen)
}
func circle(hdc syscall.Handle, x, y, d int, color uintptr) {
	brush := makeBrush(color)
	old, _, _ := procSelectObject.Call(uintptr(hdc), uintptr(brush))
	pen, _, _ := procCreatePen.Call(PS_SOLID, 1, color)
	oldp, _, _ := procSelectObject.Call(uintptr(hdc), pen)
	procEllipse.Call(uintptr(hdc), uintptr(x), uintptr(y), uintptr(x+d), uintptr(y+d))
	procSelectObject.Call(uintptr(hdc), old)
	procSelectObject.Call(uintptr(hdc), oldp)
	procDeleteObject.Call(uintptr(brush))
	procDeleteObject.Call(pen)
}
func drawText(hdc syscall.Handle, x, y int, s string, color uintptr, font syscall.Handle) {
	if s == "" {
		return
	}
	old := uintptr(0)
	if font != 0 {
		old, _, _ = procSelectObject.Call(uintptr(hdc), uintptr(font))
	}
	procSetBkMode.Call(uintptr(hdc), TRANSPARENT)
	procSetTextColor.Call(uintptr(hdc), color)
	p := utf(s)
	n := len(syscall.StringToUTF16(s)) - 1
	procTextOut.Call(uintptr(hdc), uintptr(x), uintptr(y), uintptr(unsafe.Pointer(p)), uintptr(n))
	if font != 0 {
		procSelectObject.Call(uintptr(hdc), old)
	}
}
func button(hdc syscall.Handle, x, y, w, h int, label string, accent bool) {
	c := rgb(28, 39, 57)
	tc := rgb(187, 205, 225)
	if accent {
		c = rgb(0, 195, 180)
		tc = rgb(5, 18, 29)
	}
	fill(hdc, x, y, w, h, c, 9)
	drawText(hdc, x+14, y+9, label, tc, fontBold)
}
func card(hdc syscall.Handle, x, y, w, h int, title, value, sub string, accent uintptr) {
	fill(hdc, x, y, w, h, rgb(17, 27, 42), 12)
	fill(hdc, x+16, y+17, 4, h-34, accent, 4)
	drawText(hdc, x+32, y+19, strings.ToUpper(title), rgb(129, 153, 179), fontSmall)
	drawText(hdc, x+32, y+49, value, rgb(236, 244, 252), fontLarge)
	drawText(hdc, x+32, y+h-31, sub, rgb(105, 129, 153), fontSmall)
}
func sectionTitle(hdc syscall.Handle, x, y int, t, sub string) {
	drawText(hdc, x, y, t, rgb(238, 245, 252), fontBold)
	drawText(hdc, x, y+31, sub, rgb(124, 149, 174), fontSmall)
}
func drawApp(hdc syscall.Handle) {
	var rc RECT
	procGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rc)))
	W, H := int(rc.Right), int(rc.Bottom)
	fill(hdc, 0, 0, W, H, rgb(7, 13, 23), 0)
	fill(hdc, 0, 0, 238, H, rgb(10, 19, 32), 0)
	fill(hdc, 0, 0, 238, 84, rgb(12, 24, 39), 0)
	// Brand mark
	fill(hdc, 22, 20, 42, 42, rgb(0, 195, 180), 10)
	drawText(hdc, 31, 27, "CC", rgb(4, 19, 27), fontBold)
	drawText(hdc, 76, 20, "CYBERCIRCUIT", rgb(231, 242, 251), fontBold)
	drawText(hdc, 76, 43, "GOTHAM  /  INTELLIGENCE  v0.5.0", rgb(0, 205, 188), fontSmall)
	drawText(hdc, 25, 111, "WORKSPACE", rgb(89, 116, 144), fontSmall)
	nav := []string{"Overview", "Case Explorer", "Case Map", "Link Analysis", "Network Ops", "Device Monitor", "Analyst Lab"}
	navIcons := []string{"01", "02", "03", "04", "05", "06", "07"}
	for i, n := range nav {
		y := 143 + i*54
		if app.page == i {
			fill(hdc, 12, y-8, 212, 44, rgb(20, 48, 63), 8)
			fill(hdc, 12, y-8, 3, 44, rgb(0, 204, 187), 4)
		}
		c := rgb(107, 132, 158)
		if app.page == i {
			c = rgb(0, 215, 195)
		}
		drawText(hdc, 28, y+3, navIcons[i], c, fontSmall)
		nc := rgb(150, 171, 192)
		if app.page == i {
			nc = rgb(239, 248, 255)
		}
		drawText(hdc, 65, y+2, n, nc, fontRegular)
	}
	line(hdc, 22, 493, 215, 493, rgb(31, 48, 67), 1)
	drawText(hdc, 25, 515, "LOCAL-FIRST WORKSPACE", rgb(90, 116, 143), fontSmall)
	drawText(hdc, 25, 540, "No account required", rgb(124, 149, 174), fontSmall)
	drawText(hdc, 25, 562, "Case data stays on this PC", rgb(124, 149, 174), fontSmall)
	fill(hdc, 25, H-74, 188, 44, rgb(15, 31, 47), 8)
	circle(hdc, 39, H-58, 9, rgb(0, 209, 151))
	drawText(hdc, 57, H-62, "SYSTEM ONLINE", rgb(186, 210, 226), fontSmall)
	drawText(hdc, 57, H-45, "WINDOWS DESKTOP", rgb(89, 118, 145), fontSmall)
	// Top bar
	fill(hdc, 238, 0, W-238, 84, rgb(10, 18, 30), 0)
	drawText(hdc, 270, 22, "CYBERCIRCUIT  /  GOTHAM", rgb(110, 138, 165), fontSmall)
	pages := []string{"Command overview", "Public case explorer", "Courthouse map", "Relationship graph", "Network operations", "Connected device inventory", "Local intelligence analysis"}
	drawText(hdc, 270, 43, pages[clamp(app.page, 0, 5)], rgb(233, 242, 251), fontBold)
	fill(hdc, W-240, 25, 12, 12, rgb(0, 211, 151), 6)
	drawText(hdc, W-220, 22, "LOCAL SESSION", rgb(140, 171, 193), fontSmall)
	line(hdc, 238, 84, W, 84, rgb(28, 43, 60), 1)
	switch app.page {
	case 0:
		drawDashboard(hdc, W, H)
	case 1:
		drawIntel(hdc, W, H)
	case 2:
		drawCaseMap(hdc, W, H)
	case 3:
		drawGraph(hdc, W, H)
	case 4:
		drawNetwork(hdc, W, H)
	case 5:
		drawDevices(hdc, W, H)
	case 6:
		drawAnalyst(hdc, W, H)
	}
}
func drawDashboard(hdc syscall.Handle, W, H int) {
	m := getMetricsCopy()
	x0 := 270
	gap := 16
	usable := W - x0 - 34
	cw := (usable - 3*gap) / 4
	if cw < 170 {
		cw = 170
	}
	sectionTitle(hdc, x0, 112, "System overview", "Live telemetry from this Windows computer and its active network adapters")
	cpu := fmt.Sprintf("%.0f%%", m.CPU)
	mem := "--"
	disk := "--"
	if m.MemTotal > 0 {
		mem = fmt.Sprintf("%.0f%%", float64(m.MemUsed)*100/float64(m.MemTotal))
	}
	if m.DiskTotal > 0 {
		disk = fmt.Sprintf("%.0f%%", float64(m.DiskUsed)*100/float64(m.DiskTotal))
	}
	card(hdc, x0, 170, cw, 135, "CPU UTILISATION", cpu, "Processor activity", rgb(0, 201, 185))
	card(hdc, x0+cw+gap, 170, cw, 135, "MEMORY IN USE", mem, fmt.Sprintf("%.1f / %.1f GB", float64(m.MemUsed)/1073741824, float64(m.MemTotal)/1073741824), rgb(107, 145, 255))
	card(hdc, x0+2*(cw+gap), 170, cw, 135, "SYSTEM DRIVE", disk, "Used storage on Windows drive", rgb(185, 132, 255))
	ip := m.LocalIP
	if ip == "" {
		ip = "Not detected"
	}
	card(hdc, x0+3*(cw+gap), 170, cw, 135, "LOCAL IP ADDRESS", ip, "Private network interface", rgb(255, 174, 95))
	leftW := usable * 6 / 10
	y := 329
	fill(hdc, x0, y, leftW, 390, rgb(13, 23, 37), 12)
	drawText(hdc, x0+22, y+20, "LIVE SYSTEM SNAPSHOT", rgb(220, 235, 247), fontBold)
	drawText(hdc, x0+22, y+54, "HOST", rgb(96, 125, 153), fontSmall)
	drawText(hdc, x0+180, y+52, computerName(), rgb(192, 212, 228), fontRegular)
	drawText(hdc, x0+22, y+88, "OPERATING SYSTEM", rgb(96, 125, 153), fontSmall)
	drawText(hdc, x0+180, y+86, "Microsoft Windows", rgb(192, 212, 228), fontRegular)
	drawText(hdc, x0+22, y+122, "ARCHITECTURE", rgb(96, 125, 153), fontSmall)
	drawText(hdc, x0+180, y+120, runtime.GOARCH, rgb(192, 212, 228), fontRegular)
	drawText(hdc, x0+22, y+156, "ACTIVE ADAPTER", rgb(96, 125, 153), fontSmall)
	drawText(hdc, x0+180, y+154, trimTo(m.Adapter, 45), rgb(192, 212, 228), fontRegular)
	line(hdc, x0+22, y+190, x0+leftW-22, y+190, rgb(31, 49, 67), 1)
	drawText(hdc, x0+22, y+210, "RESOURCE LOAD", rgb(220, 235, 247), fontBold)
	drawBar(hdc, x0+22, y+250, leftW-44, "CPU", m.CPU, rgb(0, 205, 188))
	mr := 0.0
	if m.MemTotal > 0 {
		mr = float64(m.MemUsed) * 100 / float64(m.MemTotal)
	}
	drawBar(hdc, x0+22, y+302, leftW-44, "MEMORY", mr, rgb(112, 147, 255))
	dr := 0.0
	if m.DiskTotal > 0 {
		dr = float64(m.DiskUsed) * 100 / float64(m.DiskTotal)
	}
	drawBar(hdc, x0+22, y+354, leftW-44, "DISK", dr, rgb(190, 135, 255))
	rx := x0 + leftW + gap
	rw := usable - leftW - gap
	fill(hdc, rx, y, rw, 390, rgb(13, 23, 37), 12)
	drawText(hdc, rx+20, y+20, "QUICK ACTIONS", rgb(220, 235, 247), fontBold)
	button(hdc, rx+20, y+62, rw-40, 42, "Open public-source research", true)
	button(hdc, rx+20, y+119, rw-40, 42, "Inspect connected devices", false)
	button(hdc, rx+20, y+176, rw-40, 42, "Run local analysis", false)
	line(hdc, rx+20, y+238, rx+rw-20, y+238, rgb(31, 49, 67), 1)
	drawText(hdc, rx+20, y+258, "NETWORK REACHABILITY", rgb(96, 125, 153), fontSmall)
	c := rgb(0, 209, 151)
	s := "ONLINE"
	if !m.Online {
		c = rgb(245, 171, 85)
		s = "UNVERIFIED"
	}
	circle(hdc, rx+20, y+294, 10, c)
	drawText(hdc, rx+39, y+291, s, rgb(193, 213, 229), fontBold)
	drawText(hdc, rx+20, y+331, "Updates automatically every few seconds", rgb(105, 130, 154), fontSmall)
}
func drawBar(hdc syscall.Handle, x, y, w int, label string, val float64, color uintptr) {
	drawText(hdc, x, y, label, rgb(133, 158, 182), fontSmall)
	drawText(hdc, x+w-52, y, fmt.Sprintf("%.0f%%", val), rgb(195, 215, 231), fontSmall)
	fill(hdc, x, y+24, w, 7, rgb(30, 46, 63), 4)
	bw := int(float64(w) * val / 100)
	if bw < 0 {
		bw = 0
	}
	if bw > w {
		bw = w
	}
	if bw > 0 {
		fill(hdc, x, y+24, bw, 7, color, 4)
	}
}
func drawIntel(hdc syscall.Handle, W, H int) {
	x := 270
	sectionTitle(hdc, x, 112, "Public case explorer", "Search published U.S. court opinions from CourtListener. This is case-law research, not a criminal-record database.")
	drawText(hdc, x+2, 174, "TOPIC / OFFENCE / LEGAL ISSUE", rgb(127, 153, 178), fontSmall)
	drawText(hdc, x+2, 247, "OPTIONAL CASE TITLE OR CITATION", rgb(127, 153, 178), fontSmall)
	drawText(hdc, x+420, 247, "YEAR", rgb(127, 153, 178), fontSmall)
	drawText(hdc, x+570, 247, "COURTLISTENER API TOKEN (MASKED)", rgb(127, 153, 178), fontSmall)
	button(hdc, W-262, 203, 205, 43, "Search published cases", true)
	button(hdc, W-262, 267, 205, 38, "Open courthouse map", false)
	fill(hdc, x, 328, W-x-34, H-367, rgb(13, 23, 37), 12)
	drawText(hdc, x+22, 348, "MATCHING PUBLISHED OPINIONS", rgb(225, 238, 249), fontBold)
	app.mu.RLock()
	cases := append([]LegalCase{}, app.cases...)
	count := app.caseCount
	status := app.caseStatus
	q := app.lastQuery
	app.mu.RUnlock()
	drawText(hdc, x+22, 377, trimTo(fmt.Sprintf("%s   /   %d total matches   /   showing up to 7 results", status, count), 130), rgb(0, 204, 184), fontSmall)
	if len(cases) == 0 {
		drawText(hdc, x+22, 432, "Enter a topic such as fraud, homicide sentencing, theft appeal, or evidence admissibility.", rgb(122, 147, 170), fontRegular)
		drawText(hdc, x+22, 459, "Queries are sent to CourtListener. Do not enter private medical, biometric, or confidential information.", rgb(115, 138, 160), fontSmall)
	}
	for i, c := range cases {
		if i >= 7 {
			break
		}
		yy := 410 + i*61
		fill(hdc, x+14, yy-5, W-x-62, 54, rgb(17, 30, 46), 7)
		title := c.CaseName
		if title == "" {
			title = "Published court opinion"
		}
		drawText(hdc, x+26, yy, trimTo(title, 105), rgb(220, 237, 249), fontBold)
		meta := strings.TrimSpace(c.DateFiled + "   /   " + c.Court + "   /   " + strings.Join(c.Citation, " · ") + "   /   docket " + c.DocketNumber + "   /   " + c.Status + "   /   judge " + c.Judge)
		drawText(hdc, x+26, yy+19, trimTo(meta, 130), rgb(117, 147, 173), fontSmall)
		detail := strings.TrimSpace(c.Syllabus + "  " + c.SuitNature + "  " + c.Posture + "  " + c.ProceduralHistory)
		if detail == "" {
			detail = c.Snippet
		}
		drawText(hdc, x+26, yy+36, trimTo(stripMarkup(detail), 125), rgb(169, 190, 207), fontSmall)
	}
	drawText(hdc, x+22, H-54, "Requires free CourtListener API token. Queries are sent to CourtListener; do not enter private biometric or medical data.", rgb(92, 119, 145), fontSmall)
	_ = q
}
func drawCaseMap(hdc syscall.Handle, W, H int) {
	x := 270
	sectionTitle(hdc, x, 112, "Case map", "Interactive map plots the court location associated with each result — not the incident or a person's home.")
	button(hdc, x+2, 180, 210, 42, "Open interactive map", true)
	button(hdc, x+230, 180, 185, 42, "Get API token", false)
	app.mu.RLock()
	cases := append([]LegalCase{}, app.cases...)
	status := app.mapStatus
	app.mu.RUnlock()
	located := 0
	courts := map[string]bool{}
	for _, c := range cases {
		if c.HasLocation {
			located++
			courts[c.Court] = true
		}
	}
	fill(hdc, x, 244, W-x-34, 390, rgb(13, 23, 37), 12)
	drawText(hdc, x+24, 266, "COURTHOUSE LOCATION OVERVIEW", rgb(225, 238, 249), fontBold)
	drawText(hdc, x+24, 305, fmt.Sprintf("%d search results   /   %d geocoded court markers   /   %d unique courts", len(cases), located, len(courts)), rgb(0, 204, 184), fontRegular)
	drawText(hdc, x+24, 354, trimTo(status, 115), rgb(167, 190, 209), fontRegular)
	drawText(hdc, x+24, 415, "MAP DETAILS", rgb(225, 238, 249), fontBold)
	drawText(hdc, x+24, 450, "• Pins identify the courthouse named in the published opinion.", rgb(174, 197, 214), fontRegular)
	drawText(hdc, x+24, 480, "• Pins do not show where an offence happened or where anyone lives.", rgb(174, 197, 214), fontRegular)
	drawText(hdc, x+24, 510, "• Court locations are approximate and are resolved from court names.", rgb(174, 197, 214), fontRegular)
	drawText(hdc, x+24, 570, "Map tiles and geocoding require an internet connection. OpenStreetMap attribution appears on the map.", rgb(104, 133, 157), fontSmall)
	drawText(hdc, x+24, H-55, "Search results are optional. The map opens without pins when no courthouse coordinates are available.", rgb(92, 119, 145), fontSmall)
}
func drawGraph(hdc syscall.Handle, W, H int) {
	x := 270
	sectionTitle(hdc, x, 112, "Relationship graph", "Build a local map of entities and connections from information you are authorised to use.")
	gx, gy, gw, gh := x, 168, W-x-34, 420
	fill(hdc, gx, gy, gw, gh, rgb(11, 21, 34), 12)
	// subtle grid
	for xx := gx + 25; xx < gx+gw; xx += 40 {
		line(hdc, xx, gy+15, xx, gy+gh-15, rgb(18, 32, 48), 1)
	}
	for yy := gy + 25; yy < gy+gh; yy += 40 {
		line(hdc, gx+15, yy, gx+gw-15, yy, rgb(18, 32, 48), 1)
	}
	nodes, edges := getGraph()
	positions := map[string]POINT{}
	n := len(nodes)
	if n > 0 {
		cols := 4
		if n < 4 {
			cols = n
		}
		rows := (n + cols - 1) / cols
		usableW := gw - 120
		usableH := gh - 100
		for i, name := range nodes {
			col := i % cols
			row := i / cols
			px := gx + 60 + col*(usableW/maxInt(cols, 1)) + int((i%2)*11)
			py := gy + 65 + row*(usableH/maxInt(rows, 1)) + int((i%3)*9)
			positions[name] = POINT{int32(px), int32(py)}
		}
	}
	for _, e := range edges {
		a, oka := positions[e.From]
		b, okb := positions[e.To]
		if oka && okb {
			line(hdc, int(a.X)+50, int(a.Y)+18, int(b.X)+50, int(b.Y)+18, rgb(0, 117, 130), 2)
		}
	}
	for i, name := range nodes {
		p := positions[name]
		fill(hdc, int(p.X), int(p.Y), 100, 38, rgb(18, 48, 61), 9)
		drawText(hdc, int(p.X)+10, int(p.Y)+10, trimTo(name, 13), rgb(215, 242, 248), fontSmall)
		if i == 0 {
			circle(hdc, int(p.X)-7, int(p.Y)+12, 6, rgb(0, 208, 190))
		}
	}
	if len(nodes) == 0 {
		drawText(hdc, gx+32, gy+50, "Your graph is empty", rgb(185, 207, 224), fontBold)
		drawText(hdc, gx+32, gy+83, "Add a node below, then connect two nodes using their exact names.", rgb(111, 140, 166), fontRegular)
	}
	fill(hdc, x, 609, W-x-34, 150, rgb(13, 23, 37), 12)
	drawText(hdc, x+20, 628, "ADD ENTITY", rgb(131, 157, 180), fontSmall)
	button(hdc, x+360, 650, 120, 40, "Add node", true)
	drawText(hdc, x+525, 628, "CONNECT ENTITIES  (source -> target)", rgb(131, 157, 180), fontSmall)
	button(hdc, W-205, 650, 130, 40, "Create link", false)
	drawText(hdc, x+20, 713, fmt.Sprintf("%d nodes   /   %d connections   /   Stored locally", len(nodes), len(edges)), rgb(98, 126, 152), fontSmall)
}
func drawNetwork(hdc syscall.Handle, W, H int) {
	x := 270
	sectionTitle(hdc, x, 112, "Network operations", "Monitor this computer's network interfaces and test reachability to one host you specify.")
	fill(hdc, x, 171, W-x-34, 158, rgb(13, 23, 37), 12)
	drawText(hdc, x+22, 190, "HOST REACHABILITY CHECK", rgb(225, 238, 249), fontBold)
	drawText(hdc, x+22, 223, "Enter a hostname or IP address. This sends one standard ping request.", rgb(106, 133, 158), fontSmall)
	button(hdc, W-255, 264, 175, 42, "Ping host", true)
	_, hr := getHostState()
	if hr != "" {
		drawText(hdc, x+22, 294, trimTo(hr, 75), rgb(0, 211, 187), fontSmall)
	}
	fill(hdc, x, 348, W-x-34, H-390, rgb(13, 23, 37), 12)
	drawText(hdc, x+22, 369, "LOCAL NETWORK ADAPTERS", rgb(225, 238, 249), fontBold)
	drawText(hdc, x+22, 398, "Interfaces and private addresses reported by Windows", rgb(106, 133, 158), fontSmall)
	m := getMetricsCopy()
	drawText(hdc, x+22, 450, "PRIMARY ADAPTER", rgb(96, 125, 153), fontSmall)
	drawText(hdc, x+240, 448, trimTo(m.Adapter, 60), rgb(201, 220, 235), fontRegular)
	drawText(hdc, x+22, 492, "LOCAL ADDRESS", rgb(96, 125, 153), fontSmall)
	drawText(hdc, x+240, 490, m.LocalIP, rgb(201, 220, 235), fontRegular)
	drawText(hdc, x+22, 534, "INTERNET CHECK", rgb(96, 125, 153), fontSmall)
	status := "Not verified"
	col := rgb(242, 176, 85)
	if m.Online {
		status = "Reachable"
		col = rgb(0, 211, 151)
	}
	drawText(hdc, x+240, 532, status, col, fontRegular)
	drawText(hdc, x+22, 585, "Safety: this build checks reachability only. It does not scan ports or probe devices you do not own.", rgb(98, 126, 152), fontSmall)
}
func drawDevices(hdc syscall.Handle, W, H int) {
	x := 270
	sectionTitle(hdc, x, 112, "Connected device inventory", "Read-only inventory of devices Windows currently reports as connected to this computer.")
	button(hdc, W-242, 174, 185, 42, "Refresh inventory", true)
	fill(hdc, x, 235, W-x-34, H-278, rgb(13, 23, 37), 12)
	drawText(hdc, x+22, 255, "WINDOWS DEVICE ENUMERATION", rgb(225, 238, 249), fontBold)
	dev, count := getDevices()
	drawText(hdc, x+22, 290, fmt.Sprintf("%d connected device entries", count), rgb(0, 204, 184), fontSmall)
	if len(dev) == 0 {
		drawText(hdc, x+22, 340, "Select Refresh inventory to read the connected-device list.", rgb(130, 154, 177), fontRegular)
	}
	for i, d := range dev {
		if i >= 17 {
			break
		}
		yy := 330 + i*25
		circle(hdc, x+25, yy+5, 6, rgb(74, 153, 177))
		drawText(hdc, x+44, yy, trimTo(d, 105), rgb(184, 207, 224), fontSmall)
		line(hdc, x+22, yy+20, W-75, yy+20, rgb(25, 41, 58), 1)
	}
	drawText(hdc, x+22, H-56, "Some device details may require administrator permissions. No device settings are changed.", rgb(98, 126, 152), fontSmall)
}
func drawAnalyst(hdc syscall.Handle, W, H int) {
	x := 270
	sectionTitle(hdc, x, 112, "Local analyst lab", "Offline text analysis: extract public indicators and structure notes without sending data to a server.")
	fill(hdc, x, 170, W-x-34, 318, rgb(13, 23, 37), 12)
	drawText(hdc, x+22, 190, "CASE NOTES / SOURCE EXCERPTS", rgb(130, 156, 180), fontSmall)
	button(hdc, x+22, 436, 175, 39, "Analyse notes", true)
	drawText(hdc, x+220, 449, "Offline processing  /  no API key  /  no cloud upload", rgb(95, 124, 149), fontSmall)
	fill(hdc, x, 510, W-x-34, H-545, rgb(13, 23, 37), 12)
	drawText(hdc, x+22, 529, "ANALYSIS OUTPUT", rgb(225, 238, 249), fontBold)
	out := getAnalyst()
	if len(out) == 0 {
		drawText(hdc, x+22, 573, "Paste notes above, then select Analyse notes.", rgb(127, 151, 174), fontRegular)
	}
	for i, s := range out {
		if i > 8 {
			break
		}
		drawText(hdc, x+22, 570+i*25, trimTo(s, 120), rgb(185, 208, 225), fontSmall)
	}
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func trimTo(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	if n < 4 {
		return s[:n]
	}
	return s[:n-3] + "..."
}
func computerName() string { n, _ := os.Hostname(); return n }

func getMetricsCopy() Metrics { app.mu.RLock(); defer app.mu.RUnlock(); return app.metrics }
func getSearches() []CaseLog {
	app.mu.RLock()
	defer app.mu.RUnlock()
	out := append([]CaseLog{}, app.workspace.Searches...)
	return out
}
func getGraph() ([]string, []Edge) {
	app.mu.RLock()
	defer app.mu.RUnlock()
	return append([]string{}, app.workspace.Nodes...), append([]Edge{}, app.workspace.Edges...)
}
func getHostState() (string, string) {
	app.mu.RLock()
	defer app.mu.RUnlock()
	return app.hostChecked, app.hostResult
}
func getDevices() ([]string, int) {
	app.mu.RLock()
	defer app.mu.RUnlock()
	return append([]string{}, app.devices...), app.deviceCount
}
func getAnalyst() []string {
	app.mu.RLock()
	defer app.mu.RUnlock()
	return append([]string{}, app.analystOutput...)
}
func invalidate() {
	if hwnd != 0 {
		procInvalidateRect.Call(uintptr(hwnd), 0, 0)
	}
}
func setEditText(h syscall.Handle, s string) {
	if h != 0 {
		p := utf(s)
		procSetWindowText.Call(uintptr(h), uintptr(unsafe.Pointer(p)))
	}
}
func getEditText(h syscall.Handle) string {
	if h == 0 {
		return ""
	}
	n, _, _ := procGetWindowTextLength.Call(uintptr(h))
	if n == 0 {
		return ""
	}
	buf := make([]uint16, int(n)+1)
	procGetWindowText.Call(uintptr(h), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}
func showEdit(h syscall.Handle, x, y, w, he int, visible bool) {
	if h == 0 {
		return
	}
	if visible {
		procSetWindowPos.Call(uintptr(h), 0, uintptr(x), uintptr(y), uintptr(w), uintptr(he), SWP_NOZORDER|SWP_NOACTIVATE)
		procShowControl.Call(uintptr(h), SW_SHOW)
	} else {
		procShowControl.Call(uintptr(h), SW_HIDE)
	}
}
func layoutEdits() {
	var rc RECT
	procGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rc)))
	W := int(rc.Right)
	H := int(rc.Bottom)
	if app.page == 1 {
		showEdit(editSearch, 292, 202, W-610, 36, true)
		showEdit(editCaseTitle, 292, 273, 365, 34, true)
		showEdit(editCaseYear, 720, 273, 90, 34, true)
		showEdit(editAPIToken, 860, 273, W-1150, 34, true)
	} else {
		showEdit(editSearch, 0, 0, 0, 0, false)
		showEdit(editCaseTitle, 0, 0, 0, 0, false)
		showEdit(editCaseYear, 0, 0, 0, 0, false)
		showEdit(editAPIToken, 0, 0, 0, 0, false)
	}
	if app.page == 3 {
		showEdit(editNode, 292, 650, 340, 36, true)
		showEdit(editLink, 807, 650, W-1035, 36, true)
	} else {
		showEdit(editNode, 0, 0, 0, 0, false)
		showEdit(editLink, 0, 0, 0, 0, false)
	}
	if app.page == 4 {
		showEdit(editHost, 292, 264, W-590, 36, true)
	} else {
		showEdit(editHost, 0, 0, 0, 0, false)
	}
	if app.page == 6 {
		showEdit(editNotes, 292, 218, W-340, 205, true)
	} else {
		showEdit(editNotes, 0, 0, 0, 0, false)
	}
	_ = H
}
func createPasswordEdit(parent syscall.Handle, id int) syscall.Handle {
	style := uintptr(WS_CHILD | WS_TABSTOP | WS_BORDER | ES_AUTOHSCROLL | ES_PASSWORD)
	ex := uintptr(WS_EX_CLIENTEDGE)
	title := utf("")
	cls := utf("EDIT")
	r, _, _ := procCreateWindowEx.Call(ex, uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(title)), style, 0, 0, 100, 30, uintptr(parent), uintptr(id), uintptr(hInst), 0)
	h := syscall.Handle(r)
	if h != 0 {
		procSendMessage.Call(uintptr(h), 0x0030, uintptr(fontRegular), 1)
	}
	return h
}

func createEdit(parent syscall.Handle, id int, multiline bool) syscall.Handle {
	style := uintptr(WS_CHILD | WS_TABSTOP | WS_BORDER | ES_AUTOHSCROLL)
	if multiline {
		style = uintptr(WS_CHILD | WS_TABSTOP | WS_BORDER | ES_MULTILINE | ES_AUTOVSCROLL | ES_WANTRETURN | WS_VSCROLL)
	}
	ex := uintptr(WS_EX_CLIENTEDGE)
	title := utf("")
	cls := utf("EDIT")
	r, _, _ := procCreateWindowEx.Call(ex, uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(title)), style, 0, 0, 100, 30, uintptr(parent), uintptr(id), uintptr(hInst), 0)
	h := syscall.Handle(r)
	if h != 0 {
		procSendMessage.Call(uintptr(h), 0x0030, uintptr(fontRegular), 1)
	}
	return h
}
func saveWorkspace() {
	app.mu.RLock()
	b, err := json.MarshalIndent(app.workspace, "", "  ")
	app.mu.RUnlock()
	if err != nil {
		return
	}
	_ = os.MkdirAll(appDataPath, 0700)
	_ = os.WriteFile(filepath.Join(appDataPath, "workspace.json"), b, 0600)
}
func loadWorkspace() {
	base := os.Getenv("APPDATA")
	if base == "" {
		base = os.TempDir()
	}
	appDataPath = filepath.Join(base, "CyberCircuitGotham")
	b, err := os.ReadFile(filepath.Join(appDataPath, "workspace.json"))
	if err == nil {
		var w Workspace
		if json.Unmarshal(b, &w) == nil {
			if w.Nodes == nil {
				w.Nodes = []string{}
			}
			if w.Edges == nil {
				w.Edges = []Edge{}
			}
			if w.Searches == nil {
				w.Searches = []CaseLog{}
			}
			app.workspace = w
		}
	}
}
func collectMetrics() {
	var m Metrics
	var idle, kernel, user FILETIME
	ret, _, _ := procGetSystemTimes.Call(uintptr(unsafe.Pointer(&idle)), uintptr(unsafe.Pointer(&kernel)), uintptr(unsafe.Pointer(&user)))
	if ret != 0 {
		id := uint64(idle.High)<<32 | uint64(idle.Low)
		k := uint64(kernel.High)<<32 | uint64(kernel.Low)
		u := uint64(user.High)<<32 | uint64(user.Low)
		app.mu.Lock()
		if app.lastCPUKernel != 0 {
			dt := (k - app.lastCPUKernel) + (u - app.lastCPUUser)
			di := id - app.lastCPUIdle
			if dt > 0 && dt >= di {
				m.CPU = float64(dt-di) * 100 / float64(dt)
			}
		}
		app.lastCPUIdle = id
		app.lastCPUKernel = k
		app.lastCPUUser = u
		app.mu.Unlock()
	}
	ms := MEMORYSTATUSEX{Length: uint32(unsafe.Sizeof(MEMORYSTATUSEX{}))}
	r, _, _ := procGlobalMemoryStatus.Call(uintptr(unsafe.Pointer(&ms)))
	if r != 0 {
		m.MemTotal = ms.TotalPhys
		m.MemUsed = ms.TotalPhys - ms.AvailPhys
	}
	drive := os.Getenv("SystemDrive")
	if drive == "" {
		drive = "C:"
	}
	dp := utf(drive + "\\")
	var avail, total, totalFree uint64
	r, _, _ = procGetDiskFreeSpace.Call(uintptr(unsafe.Pointer(dp)), uintptr(unsafe.Pointer(&avail)), uintptr(unsafe.Pointer(&total)), uintptr(unsafe.Pointer(&totalFree)))
	if r != 0 {
		m.DiskTotal = total
		m.DiskUsed = total - totalFree
	}
	ifs, _ := net.Interfaces()
	for _, it := range ifs {
		if it.Flags&net.FlagUp == 0 || it.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := it.Addrs()
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				m.LocalIP = ipnet.IP.String()
				m.Adapter = it.Name
				break
			}
		}
		if m.LocalIP != "" {
			break
		}
	}
	app.mu.Lock()
	old := app.metrics
	m.Online = old.Online
	m.Updated = time.Now()
	app.metrics = m
	app.mu.Unlock()
}
func testInternet() {
	ok := false
	c, err := net.DialTimeout("tcp", "1.1.1.1:443", 1800*time.Millisecond)
	if err == nil {
		ok = true
		_ = c.Close()
	}
	app.mu.Lock()
	app.metrics.Online = ok
	app.lastNetCheck = time.Now()
	app.mu.Unlock()
	invalidate()
}
func refreshProcesses() {
	cmd := exec.Command("tasklist", "/FO", "CSV", "/NH")
	out, err := cmd.Output()
	if err != nil {
		return
	}
	lines := []string{}
	r := csv.NewReader(strings.NewReader(string(out)))
	r.FieldsPerRecord = -1
	for i := 0; i < 8; i++ {
		row, e := r.Read()
		if e != nil {
			break
		}
		if len(row) > 0 {
			lines = append(lines, row[0])
		}
	}
	app.mu.Lock()
	app.processes = lines
	app.mu.Unlock()
}
func refreshDevices() {
	cmd := exec.Command("pnputil", "/enum-devices", "/connected")
	out, err := cmd.Output()
	if err != nil {
		app.mu.Lock()
		app.devices = []string{"Could not read device inventory. Try running on a supported Windows version."}
		app.deviceCount = 0
		app.mu.Unlock()
		invalidate()
		return
	}
	list := []string{}
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		s := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(strings.ToLower(s), "device description:") {
			v := strings.TrimSpace(strings.SplitN(s, ":", 2)[1])
			if v != "" {
				list = append(list, v)
			}
		}
	}
	app.mu.Lock()
	app.devices = list
	app.deviceCount = len(list)
	app.mu.Unlock()
	invalidate()
}
func validHost(h string) bool {
	h = strings.TrimSpace(h)
	if len(h) == 0 || len(h) > 253 {
		return false
	}
	if net.ParseIP(h) != nil {
		return true
	}
	if strings.ContainsAny(h, " /\\:@?#") {
		return false
	}
	for _, p := range strings.Split(h, ".") {
		if len(p) == 0 || len(p) > 63 {
			return false
		}
		for _, r := range p {
			if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-') {
				return false
			}
		}
	}
	return true
}
func pingHost(host string) {
	host = strings.TrimSpace(host)
	if !validHost(host) {
		app.mu.Lock()
		app.hostChecked = host
		app.hostResult = "Invalid hostname or IP address"
		app.mu.Unlock()
		invalidate()
		return
	}
	app.mu.Lock()
	app.hostChecked = host
	app.hostResult = "Checking reachability..."
	app.mu.Unlock()
	invalidate()
	go func() {
		cmd := exec.Command("ping", "-n", "1", "-w", "1400", host)
		start := time.Now()
		out, err := cmd.CombinedOutput()
		elapsed := time.Since(start).Milliseconds()
		res := "No reply received"
		if err == nil {
			res = fmt.Sprintf("Reachable  /  %d ms response time", elapsed)
		} else if strings.Contains(strings.ToLower(string(out)), "could not find host") {
			res = "DNS could not resolve this hostname"
		}
		app.mu.Lock()
		app.hostResult = res
		app.mu.Unlock()
		invalidate()
	}()
}

var ipRegex = regexp.MustCompile(`\b(?:25[0-5]|2[0-4][0-9]|1?[0-9]?[0-9])(?:\.(?:25[0-5]|2[0-4][0-9]|1?[0-9]?[0-9])){3}\b`)
var emailRegex = regexp.MustCompile(`(?i)\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b`)
var urlRegex = regexp.MustCompile(`(?i)\bhttps?://[^\s<>()]+`)

func analyseNotes(s string) {
	ips := unique(ipRegex.FindAllString(s, -1))
	emails := unique(emailRegex.FindAllString(s, -1))
	urls := unique(urlRegex.FindAllString(s, -1))
	words := strings.Fields(strings.ToLower(regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, " ")))
	freq := map[string]int{}
	stop := map[string]bool{"the": true, "and": true, "for": true, "with": true, "that": true, "this": true, "from": true, "into": true, "are": true, "was": true, "were": true, "have": true, "has": true, "not": true, "but": true, "you": true, "your": true, "they": true, "their": true, "about": true, "http": true, "https": true, "www": true}
	for _, w := range words {
		if len(w) >= 4 && !stop[w] {
			freq[w]++
		}
	}
	type kv struct {
		k string
		v int
	}
	arr := []kv{}
	for k, v := range freq {
		arr = append(arr, kv{k, v})
	}
	sort.Slice(arr, func(i, j int) bool { return arr[i].v > arr[j].v })
	keys := []string{}
	for i := 0; i < len(arr) && i < 6; i++ {
		keys = append(keys, arr[i].k)
	}
	output := []string{fmt.Sprintf("INPUT: %d characters  /  %d words", len(s), len(words)), fmt.Sprintf("INDICATORS: %d IP addresses  /  %d email addresses  /  %d URLs", len(ips), len(emails), len(urls))}
	if len(ips) > 0 {
		output = append(output, "IP ADDRESSES: "+strings.Join(ips, ", "))
	}
	if len(emails) > 0 {
		output = append(output, "EMAILS: "+strings.Join(emails, ", "))
	}
	if len(urls) > 0 {
		output = append(output, "URLS: "+strings.Join(urls, ", "))
	}
	if len(keys) > 0 {
		output = append(output, "FREQUENT TERMS: "+strings.Join(keys, ", "))
	}
	output = append(output, "METHOD: local pattern extraction; results are leads to verify, not proof of a relationship.")
	app.mu.Lock()
	app.analystOutput = output
	app.workspace.Notes = s
	app.mu.Unlock()
	saveWorkspace()
	invalidate()
}
func unique(a []string) []string {
	m := map[string]bool{}
	o := []string{}
	for _, s := range a {
		if !m[s] {
			m[s] = true
			o = append(o, s)
		}
	}
	return o
}
func openSearch(q string) {
	q = strings.TrimSpace(q)
	if q == "" {
		return
	}
	u := "https://duckduckgo.com/?q=" + url.QueryEscape(q)
	pu := utf(u)
	procShellExecute.Call(uintptr(hwnd), uintptr(unsafe.Pointer(utf("open"))), uintptr(unsafe.Pointer(pu)), 0, 0, SW_SHOW)
	app.mu.Lock()
	app.workspace.Searches = append([]CaseLog{{Query: q, When: time.Now().Format("02 Jan 2006 15:04")}}, app.workspace.Searches...)
	if len(app.workspace.Searches) > 50 {
		app.workspace.Searches = app.workspace.Searches[:50]
	}
	app.mu.Unlock()
	saveWorkspace()
	invalidate()
}
func addNode(name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	app.mu.Lock()
	for _, n := range app.workspace.Nodes {
		if strings.EqualFold(n, name) {
			app.mu.Unlock()
			return
		}
	}
	if len(app.workspace.Nodes) >= 32 {
		app.mu.Unlock()
		return
	}
	app.workspace.Nodes = append(app.workspace.Nodes, name)
	app.mu.Unlock()
	saveWorkspace()
	invalidate()
}
func addLink(s string) {
	parts := strings.SplitN(s, "->", 2)
	if len(parts) != 2 {
		return
	}
	a := strings.TrimSpace(parts[0])
	b := strings.TrimSpace(parts[1])
	if a == "" || b == "" || strings.EqualFold(a, b) {
		return
	}
	app.mu.Lock()
	hasA, hasB := false, false
	for _, n := range app.workspace.Nodes {
		if strings.EqualFold(n, a) {
			hasA = true
		}
		if strings.EqualFold(n, b) {
			hasB = true
		}
	}
	if hasA && hasB {
		for _, e := range app.workspace.Edges {
			if strings.EqualFold(e.From, a) && strings.EqualFold(e.To, b) {
				app.mu.Unlock()
				return
			}
		}
		if len(app.workspace.Edges) < 64 {
			app.workspace.Edges = append(app.workspace.Edges, Edge{From: a, To: b})
		}
	}
	app.mu.Unlock()
	saveWorkspace()
	invalidate()
}

func stripMarkup(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return strings.TrimSpace(html.UnescapeString(re.ReplaceAllString(s, "")))
}
func searchPublishedCases(topic, caseTitle, year, token string) {
	topic = strings.TrimSpace(topic)
	caseTitle = strings.TrimSpace(caseTitle)
	year = strings.TrimSpace(year)
	token = strings.TrimSpace(token)
	if token == "" {
		app.mu.Lock()
		app.caseStatus = "Paste your free CourtListener API token in the masked token field."
		app.mu.Unlock()
		invalidate()
		return
	}
	if topic == "" && caseTitle == "" {
		app.mu.Lock()
		app.caseStatus = "Enter a topic or case title"
		app.mu.Unlock()
		invalidate()
		return
	}
	q := topic
	if caseTitle != "" {
		safe := strings.NewReplacer("\"", " ", "(", " ", ")", " ").Replace(caseTitle)
		if q != "" {
			q += " AND "
		}
		q += "(caseName:\"" + strings.TrimSpace(safe) + "\" OR citation:\"" + strings.TrimSpace(safe) + "\")"
	}
	if year != "" {
		if len(year) != 4 {
			app.mu.Lock()
			app.caseStatus = "Year must use four digits (for example 2018)"
			app.mu.Unlock()
			invalidate()
			return
		}
		for _, r := range year {
			if r < '0' || r > '9' {
				app.mu.Lock()
				app.caseStatus = "Year must use four digits (for example 2018)"
				app.mu.Unlock()
				invalidate()
				return
			}
		}
		if q != "" {
			q += " AND "
		}
		q += "dateFiled:[" + year + "-01-01 TO " + year + "-12-31]"
	}
	app.mu.Lock()
	app.caseStatus = "Searching CourtListener public case-law archive..."
	app.lastQuery = q
	app.cases = []LegalCase{}
	app.caseCount = 0
	app.mapStatus = "Waiting for search results."
	app.mu.Unlock()
	invalidate()
	go func(query string) {
		endpoint := "https://www.courtlistener.com/api/rest/v4/search/?type=o&highlight=on&q=" + url.QueryEscape(query)
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			setCaseError(err.Error())
			return
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "CyberCircuitGotham/0.5.0 (Windows desktop public case research)")
		req.Header.Set("Authorization", "Token "+token)
		client := &http.Client{Timeout: 18 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			setCaseError("Network request failed: " + err.Error())
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			setCaseError(fmt.Sprintf("CourtListener returned HTTP %d. Check your connection or service availability.", resp.StatusCode))
			return
		}
		var payload CourtListenerSearch
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			setCaseError("Could not read the CourtListener response: " + err.Error())
			return
		}
		out := make([]LegalCase, 0, len(payload.Results))
		for _, r := range payload.Results {
			name := r.CaseName
			if name == "" {
				name = r.CaseNameFull
			}
			rel := r.AbsoluteURL
			if strings.HasPrefix(rel, "/") {
				rel = "https://www.courtlistener.com" + rel
			}
			if !strings.HasPrefix(rel, "https://www.courtlistener.com/") {
				rel = ""
			}
			out = append(out, LegalCase{CaseName: name, DateFiled: r.DateFiled, Court: r.Court, CourtID: r.CourtID, DocketNumber: r.DocketNumber, Citation: r.Citation, Snippet: stripMarkup(r.Snippet), Syllabus: stripMarkup(r.Syllabus), SuitNature: r.SuitNature, Posture: r.Posture, ProceduralHistory: r.ProceduralHistory, Status: r.Status, Judge: r.Judge, URL: rel})
		}
		app.mu.Lock()
		app.cases = out
		app.caseCount = payload.Count
		app.caseStatus = "Search complete — published opinions"
		app.mapStatus = "Select Open interactive map to resolve courthouse locations. New lookups are rate-limited and cached locally."
		app.mu.Unlock()
		// Keep the local query history, but do not store details about individual people.
		app.mu.Lock()
		historyLabel := "Public case-law search"
		if caseTitle != "" {
			historyLabel = "Case title/citation search"
		} else if year != "" {
			historyLabel = "Year-filtered case search"
		}
		app.workspace.Searches = append([]CaseLog{{Query: historyLabel, When: time.Now().Format("02 Jan 2006 15:04")}}, app.workspace.Searches...)
		if len(app.workspace.Searches) > 50 {
			app.workspace.Searches = app.workspace.Searches[:50]
		}
		app.mu.Unlock()
		saveWorkspace()
		invalidate()
	}(q)
}
func setCaseError(message string) {
	app.mu.Lock()
	app.caseStatus = trimTo(message, 130)
	app.cases = []LegalCase{}
	app.caseCount = 0
	app.mapStatus = "No locations available."
	app.mu.Unlock()
	invalidate()
}
func openCaseResult(idx int) {
	app.mu.RLock()
	if idx < 0 || idx >= len(app.cases) {
		app.mu.RUnlock()
		return
	}
	u := app.cases[idx].URL
	app.mu.RUnlock()
	if u != "" {
		pu := utf(u)
		verb := utf("open")
		procShellExecute.Call(uintptr(hwnd), uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(pu)), 0, 0, SW_SHOW)
	}
}
func geoCachePath() string { return filepath.Join(appDataPath, "court_geo_cache.json") }
func loadGeoCache() map[string]GeoPoint {
	m := map[string]GeoPoint{}
	b, e := os.ReadFile(geoCachePath())
	if e == nil {
		_ = json.Unmarshal(b, &m)
	}
	return m
}
func saveGeoCache(m map[string]GeoPoint) {
	b, e := json.MarshalIndent(m, "", "  ")
	if e == nil {
		_ = os.WriteFile(geoCachePath(), b, 0600)
	}
}

// courtSearchPlace converts CourtListener's compact court IDs into a useful
// courthouse search phrase. Search results commonly return IDs such as ca9,
// nysd or scotus; sending those slugs straight to Nominatim produces zero hits.
func courtSearchPlace(court, courtID string) string {
	id := strings.ToLower(strings.TrimSpace(courtID))
	places := map[string]string{
		"scotus": "United States Supreme Court, Washington, DC",
		"cadc":   "E. Barrett Prettyman United States Courthouse, Washington, DC",
		"ca1":    "John Joseph Moakley United States Courthouse, Boston, Massachusetts",
		"ca2":    "Thurgood Marshall United States Courthouse, New York, New York",
		"ca3":    "James A. Byrne United States Courthouse, Philadelphia, Pennsylvania",
		"ca4":    "Lewis F. Powell Jr. United States Courthouse, Richmond, Virginia",
		"ca5":    "John Minor Wisdom United States Court of Appeals Building, New Orleans, Louisiana",
		"ca6":    "Potter Stewart United States Courthouse, Cincinnati, Ohio",
		"ca7":    "Everett McKinley Dirksen United States Courthouse, Chicago, Illinois",
		"ca8":    "Thomas F. Eagleton United States Courthouse, St. Louis, Missouri",
		"ca9":    "James R. Browning United States Courthouse, San Francisco, California",
		"ca10":   "Byron White United States Courthouse, Denver, Colorado",
		"ca11":   "Elbert P. Tuttle United States Court of Appeals Building, Atlanta, Georgia",
		"cafc":   "Howard T. Markey National Courts Building, Washington, DC",
		"cand":   "United States District Court, San Francisco, California",
		"caed":   "Robert T. Matsui United States Courthouse, Sacramento, California",
		"cacd":   "United States District Court, Los Angeles, California",
		"casd":   "United States District Court, San Diego, California",
		"nysd":   "Daniel Patrick Moynihan United States Courthouse, New York, New York",
		"nyed":   "Theodore Roosevelt United States Courthouse, Brooklyn, New York",
		"nynd":   "James T. Foley United States Courthouse, Albany, New York",
		"nywd":   "United States District Court, Buffalo, New York",
		"njd":    "Martin Luther King Jr. Federal Building, Newark, New Jersey",
		"dcd":    "E. Barrett Prettyman United States Courthouse, Washington, DC",
		"txsd":   "Bob Casey United States Courthouse, Houston, Texas",
		"txnd":   "Eldon B. Mahon United States Courthouse, Fort Worth, Texas",
		"txed":   "John H. Hannah United States Courthouse, Tyler, Texas",
		"txwd":   "Hipolito F. Garcia Federal Building, San Antonio, Texas",
		"flsd":   "Wilkie D. Ferguson Jr. United States Courthouse, Miami, Florida",
		"flmd":   "Sam M. Gibbons United States Courthouse, Tampa, Florida",
		"flnd":   "United States District Court, Pensacola, Florida",
		"ilnd":   "Everett McKinley Dirksen United States Courthouse, Chicago, Illinois",
		"mad":    "John Joseph Moakley United States Courthouse, Boston, Massachusetts",
		"paed":   "James A. Byrne United States Courthouse, Philadelphia, Pennsylvania",
		"mied":   "Theodore Levin United States Courthouse, Detroit, Michigan",
		"cod":    "Byron G. Rogers United States Courthouse, Denver, Colorado",
	}
	if place := places[id]; place != "" {
		return place
	}
	q := strings.TrimSpace(court)
	if q == "" || strings.EqualFold(q, id) {
		return ""
	}
	// Court names from CourtListener can be long. Add a courthouse hint so
	// the geocoder searches for a building rather than a generic legal phrase.
	return q + " courthouse, United States"
}

func geocodeCourt(court, courtID string) (GeoPoint, bool) {
	q := courtSearchPlace(court, courtID)
	if q == "" {
		return GeoPoint{}, false
	}
	endpoint := "https://nominatim.openstreetmap.org/search?format=jsonv2&limit=1&q=" + url.QueryEscape(q)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return GeoPoint{}, false
	}
	req.Header.Set("User-Agent", "CyberCircuitGotham/0.5.0 (public courthouse map)")
	req.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return GeoPoint{}, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return GeoPoint{}, false
	}
	var rows []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	if json.NewDecoder(resp.Body).Decode(&rows) != nil || len(rows) == 0 {
		return GeoPoint{}, false
	}
	var g GeoPoint
	if _, err := fmt.Sscanf(rows[0].Lat, "%f", &g.Lat); err != nil {
		return GeoPoint{}, false
	}
	if _, err := fmt.Sscanf(rows[0].Lon, "%f", &g.Lon); err != nil {
		return GeoPoint{}, false
	}
	if g.Lat < -90 || g.Lat > 90 || g.Lon < -180 || g.Lon > 180 {
		return GeoPoint{}, false
	}
	return g, true
}

func geocodeCases() {
	app.mu.RLock()
	cases := append([]LegalCase{}, app.cases...)
	app.mu.RUnlock()
	if len(cases) == 0 {
		return
	}
	cache := loadGeoCache()
	newQueries := 0
	seenCourts := map[string]bool{}
	for i := range cases {
		if cases[i].Court == "" && cases[i].CourtID == "" {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(cases[i].CourtID))
		if key == "" {
			key = strings.ToLower(strings.TrimSpace(cases[i].Court))
		}
		if seenCourts[key] {
			if g, ok := cache[key]; ok {
				cases[i].Location = g
				cases[i].HasLocation = true
			}
			continue
		}
		seenCourts[key] = true
		g, ok := cache[key]
		if !ok {
			if newQueries >= 8 {
				continue
			}
			if newQueries > 0 {
				time.Sleep(1100 * time.Millisecond)
			}
			newQueries++
			g, ok = geocodeCourt(cases[i].Court, cases[i].CourtID)
			if ok {
				cache[key] = g
				saveGeoCache(cache)
			}
		}
		if ok {
			cases[i].Location = g
			cases[i].HasLocation = true
		}
	}
	app.mu.Lock()
	app.cases = cases
	located := 0
	for _, c := range cases {
		if c.HasLocation {
			located++
		}
	}
	if located == 0 {
		app.mapStatus = "No courthouse locations resolved. Check internet access to nominatim.openstreetmap.org; unsupported court IDs are skipped rather than placing misleading pins."
	} else {
		app.mapStatus = fmt.Sprintf("Mapped %d of %d results. Pins represent approximate courthouse locations, not incident locations.", located, len(cases))
	}
	app.mu.Unlock()
	invalidate()
}

func openCourtListenerTokenPage() {
	verb := utf("open")
	u := utf("https://www.courtlistener.com/profile/api-token/")
	procShellExecute.Call(uintptr(hwnd), uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(u)), 0, 0, SW_SHOW)
}

func prepareAndOpenCaseMap() {
	app.mu.RLock()
	hasCases := len(app.cases) > 0
	app.mu.RUnlock()
	if hasCases {
		geocodeCases()
	}
	openCaseMap()
}

func openCaseMap() {
	app.mu.RLock()
	cases := append([]LegalCase{}, app.cases...)
	app.mu.RUnlock()
	type mapItem struct {
		Title   string  `json:"title"`
		Court   string  `json:"court"`
		Date    string  `json:"date"`
		URL     string  `json:"url"`
		Snippet string  `json:"snippet"`
		Lat     float64 `json:"lat"`
		Lon     float64 `json:"lon"`
	}
	items := []mapItem{}
	for _, c := range cases {
		if c.HasLocation {
			items = append(items, mapItem{Title: c.CaseName, Court: c.Court, Date: c.DateFiled, URL: c.URL, Snippet: trimTo(c.Syllabus+" "+c.Snippet, 400), Lat: c.Location.Lat, Lon: c.Location.Lon})
		}
	}
	// V0.5.0: Use OpenFreeMap vector tiles via MapLibre, avoiding direct requests to OSMF raster tiles.
	// The map always opens, even when no courthouse coordinates were
	// resolved. The default view is a general US map; it is never presented as
	// a case location and no synthetic pins are added.
	data, _ := json.Marshal(items)
	pinCount := len(items)
	emptyNotice := ""
	if pinCount == 0 {
		emptyNotice = `<div class="empty"><strong>NO COURTHOUSE PINS YET</strong><br>This is a general US map, not a map of case locations. Search results are still available in Gotham. Pins will appear here only when a courthouse location is successfully resolved.</div>`
	}
	htmlDoc := `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>CyberCircuit Gotham — Case Map</title>
<link href="https://unpkg.com/maplibre-gl@5.6.0/dist/maplibre-gl.css" rel="stylesheet">
<style>
html,body,#map{height:100%;width:100%;margin:0;background:#07111e;font-family:Segoe UI,Arial,sans-serif;overflow:hidden}
.banner{position:absolute;z-index:5;top:12px;left:60px;max-width:460px;padding:12px 16px;color:#dceaf7;background:#0c1b2b;border:1px solid #1a5360;border-radius:10px;box-shadow:0 4px 18px #0008;pointer-events:none}
.banner b{color:#00d2bd}.empty{margin-top:10px;padding:10px 12px;background:#18283a;border:1px solid #36516b;border-radius:8px;color:#c7d8e8;font-size:13px;line-height:1.5}.empty strong{color:#00d2bd}
.attribution-note{position:absolute;z-index:4;bottom:8px;left:10px;padding:5px 8px;border-radius:5px;background:#0c1b2bdc;color:#c7d8e8;font-size:11px}
.attribution-note a{color:#6fded2}.map-error{position:absolute;z-index:8;right:14px;bottom:42px;max-width:340px;padding:12px 14px;border-radius:8px;background:#351d25;border:1px solid #9a5261;color:#ffe4e8;font-size:13px;line-height:1.45;display:none}
.maplibregl-popup-content{background:#102237;color:#e1edf7;border:1px solid #31536c;border-radius:8px;font:13px Segoe UI,Arial;padding:12px;min-width:180px}.maplibregl-popup-content a{color:#00cdbc}.maplibregl-popup-close-button{color:#e1edf7;font-size:18px}.popup-title{font-weight:700;color:#fff;margin-bottom:5px}.popup-meta{color:#b9cde0;margin-bottom:6px}.popup-link{color:#00cdbc}
.fallback{height:100%;width:100%;border:0;background:#07111e}
</style></head><body>
<div id="map"></div>
<div class="banner"><b>CYBERCIRCUIT GOTHAM / CASE MAP</b><br>Markers show the courthouse associated with a published opinion — not the incident location or a person's home. Map powered by OpenFreeMap; data © OpenStreetMap contributors.` + emptyNotice + `</div>
<div class="attribution-note"><a href="https://openfreemap.org" target="_blank" rel="noopener">OpenFreeMap</a> · © OpenMapTiles · Data from <a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noopener">OpenStreetMap</a></div>
<div id="mapError" class="map-error"></div>
<script src="https://unpkg.com/maplibre-gl@5.6.0/dist/maplibre-gl.js" onerror="showFallbackMap()"></script>
<script>
const cases=` + string(data) + `;
let fallbackShown=false;
function showMapError(message){const el=document.getElementById('mapError');if(el){el.textContent=message;el.style.display='block';}}
function showFallbackMap(){
  if(fallbackShown)return;fallbackShown=true;
  const box=document.getElementById('map');
  if(!box)return;
  box.innerHTML='<iframe class="fallback" title="OpenStreetMap" src="https://www.openstreetmap.org/export/embed.html?bbox=-125%2C24%2C-66%2C50&layer=mapnik" referrerpolicy="strict-origin-when-cross-origin"></iframe>';
  showMapError('The primary map service could not load. Showing the OpenStreetMap website embed instead. Case pins may not be available in fallback mode.');
}
function escapeHTML(value){return String(value||'').replace(/[&<>\"]/g,ch=>({'&':'&amp;','<':'&lt;','>':'&gt;','\"':'&quot;'}[ch]));}
function startCaseMap(){
  if(fallbackShown)return;
  if(!window.maplibregl){showFallbackMap();return;}
  const map=new maplibregl.Map({container:'map',style:'https://tiles.openfreemap.org/styles/liberty',center:[-98.35,39.5],zoom:3.7,attributionControl:true,renderWorldCopies:false});
  map.addControl(new maplibregl.NavigationControl({visualizePitch:true}),'top-right');
  let styleReady=false;let errorCount=0;
  map.on('load',()=>{
    styleReady=true;
    if(!cases.length)return;
    const features=cases.map((c,i)=>({type:'Feature',id:i,geometry:{type:'Point',coordinates:[Number(c.lon),Number(c.lat)]},properties:{title:String(c.title||'Published opinion'),court:String(c.court||'Court not specified'),date:String(c.date||''),url:String(c.url||''),snippet:String(c.snippet||'')}}));
    map.addSource('court-cases',{type:'geojson',data:{type:'FeatureCollection',features}});
    map.addLayer({id:'court-pin-halo',type:'circle',source:'court-cases',paint:{'circle-radius':13,'circle-color':'#00d2bd','circle-opacity':0.22}});
    map.addLayer({id:'court-pins',type:'circle',source:'court-cases',paint:{'circle-radius':7,'circle-color':'#00d2bd','circle-stroke-width':2,'circle-stroke-color':'#eafffc'}});
    map.on('mouseenter','court-pins',()=>{map.getCanvas().style.cursor='pointer';});
    map.on('mouseleave','court-pins',()=>{map.getCanvas().style.cursor='';});
    map.on('click','court-pins',e=>{
      const f=e.features&&e.features[0];if(!f)return;const p=f.properties||{};const node=document.createElement('div');
      const title=document.createElement('div');title.className='popup-title';title.textContent=p.title||'Published opinion';node.appendChild(title);
      const meta=document.createElement('div');meta.className='popup-meta';meta.textContent=[p.court,p.date].filter(Boolean).join(' · ');node.appendChild(meta);
      if(p.snippet){const desc=document.createElement('div');desc.textContent=p.snippet;desc.style.marginBottom='8px';node.appendChild(desc);}
      if(p.url){const a=document.createElement('a');a.className='popup-link';a.href=p.url;a.target='_blank';a.rel='noopener noreferrer';a.textContent='Open published opinion';node.appendChild(a);}
      new maplibregl.Popup({closeButton:true,closeOnClick:true,maxWidth:'320px'}).setLngLat(e.lngLat).setDOMContent(node).addTo(map);
    });
    const bounds=new maplibregl.LngLatBounds();features.forEach(f=>bounds.extend(f.geometry.coordinates));
    if(features.length>1)map.fitBounds(bounds,{padding:70,maxZoom:9});else if(features.length===1)map.easeTo({center:features[0].geometry.coordinates,zoom:10});
  });
  map.on('error',e=>{
    errorCount++;
    const msg=e&&e.error&&e.error.message?e.error.message:'Map provider request failed';
    showMapError('Map provider reported an error: '+msg+'. Check your internet connection or try the OpenStreetMap fallback link.');
    if(errorCount>=6&&!styleReady)showFallbackMap();
  });
}
if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',()=>setTimeout(startCaseMap,0));else startCaseMap();
setTimeout(()=>{if(!window.maplibregl&&!fallbackShown)showFallbackMap();},4500);
</script></body></html>`
	path := filepath.Join(appDataPath, "case-map.html")
	if err := os.WriteFile(path, []byte(htmlDoc), 0600); err != nil {
		app.mu.Lock()
		app.mapStatus = "Could not write map file: " + err.Error()
		app.mu.Unlock()
		invalidate()
		return
	}
	if err := launchMapWindow(path); err != nil {
		app.mu.Lock()
		app.mapStatus = "Map saved, but could not launch a separate map window: " + trimTo(err.Error(), 90)
		app.mu.Unlock()
		invalidate()
		return
	}
	app.mu.Lock()
	if pinCount == 0 {
		app.mapStatus = "Interactive map opened in a separate window (general map view; 0 courthouse pins)."
	} else {
		app.mapStatus = fmt.Sprintf("Interactive map opened in a separate window (%d courthouse pins).", pinCount)
	}
	app.mu.Unlock()
	invalidate()
}

// launchMapWindow prefers Edge or Chrome app mode so the map opens in its own
// standalone window instead of becoming another tab in the user's browser.
func launchMapWindow(path string) error {
	abs := filepath.ToSlash(path)
	fileURL := "file:///" + strings.TrimLeft(abs, "/")
	candidates := []string{}
	if p, err := exec.LookPath("msedge.exe"); err == nil {
		candidates = append(candidates, p)
	}
	if p, err := exec.LookPath("chrome.exe"); err == nil {
		candidates = append(candidates, p)
	}
	for _, root := range []string{os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)"), os.Getenv("LOCALAPPDATA")} {
		if root == "" {
			continue
		}
		candidates = append(candidates,
			filepath.Join(root, "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(root, "Google", "Chrome", "Application", "chrome.exe"))
	}
	seen := map[string]bool{}
	for _, browser := range candidates {
		if browser == "" || seen[strings.ToLower(browser)] {
			continue
		}
		seen[strings.ToLower(browser)] = true
		if _, err := os.Stat(browser); err != nil {
			continue
		}
		cmd := exec.Command(browser, "--app="+fileURL, "--new-window")
		if err := cmd.Start(); err == nil {
			return nil
		}
	}
	// Last resort: ask Windows to open the HTML file with the default browser.
	verb := utf("open")
	pp := utf(path)
	r, _, callErr := procShellExecute.Call(uintptr(hwnd), uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(pp)), 0, 0, SW_SHOW)
	if r > 32 {
		return nil
	}
	if callErr != syscall.Errno(0) {
		return callErr
	}
	return fmt.Errorf("Windows could not start a browser for the map")
}

func wndProc(h syscall.Handle, msg uint32, w, l uintptr) uintptr {
	switch msg {
	case WM_CREATE:
		hwnd = h
		editBrush = makeBrush(rgb(17, 29, 44))
		fontRegular = createFont(-18, FW_NORMAL, "Segoe UI")
		fontSmall = createFont(-14, FW_NORMAL, "Segoe UI")
		fontBold = createFont(-18, FW_SEMIBOLD, "Segoe UI")
		fontLarge = createFont(-31, FW_BOLD, "Segoe UI")
		editSearch = createEdit(h, ID_SEARCH, false)
		editNode = createEdit(h, ID_NODE, false)
		editLink = createEdit(h, ID_LINK, false)
		editHost = createEdit(h, ID_HOST, false)
		editNotes = createEdit(h, ID_NOTES, true)
		editCaseTitle = createEdit(h, ID_CASE_TITLE, false)
		editCaseYear = createEdit(h, ID_CASE_YEAR, false)
		editAPIToken = createPasswordEdit(h, ID_API_TOKEN)
		layoutEdits()
		procSetTimer.Call(uintptr(h), 1, 2000, 0)
		go func() { collectMetrics(); refreshProcesses(); refreshDevices(); testInternet(); invalidate() }()
		return 0
	case WM_SIZE:
		layoutEdits()
		invalidate()
		return 0
	case WM_ERASEBKGND:
		return 1
	case WM_CTLCOLOREDIT:
		hdc := syscall.Handle(w)
		procSetTextColor.Call(uintptr(hdc), rgb(218, 234, 246))
		procSetBkMode.Call(uintptr(hdc), TRANSPARENT)
		return uintptr(editBrush)
	case WM_PAINT:
		var ps PAINTSTRUCT
		hdc, _, _ := procBeginPaint.Call(uintptr(h), uintptr(unsafe.Pointer(&ps)))
		drawApp(syscall.Handle(hdc))
		procEndPaint.Call(uintptr(h), uintptr(unsafe.Pointer(&ps)))
		return 0
	case WM_TIMER:
		collectMetrics()
		app.mu.RLock()
		last := app.lastNetCheck
		app.mu.RUnlock()
		if time.Since(last) > 15*time.Second {
			go testInternet()
		}
		invalidate()
		return 0
	case WM_LBUTTONUP:
		x, y := lowWord(l), highWord(l)
		var rc RECT
		procGetClientRect.Call(uintptr(h), uintptr(unsafe.Pointer(&rc)))
		W := int(rc.Right)
		if x < 238 && y >= 135 && y < 135+7*54 {
			idx := (y - 135) / 54
			if idx >= 0 && idx < 7 {
				app.page = idx
				layoutEdits()
				invalidate()
				if idx == 5 {
					go refreshDevices()
				}
				return 0
			}
		}
		switch app.page {
		case 0:
			usable := W - 270 - 34
			cw := (usable - 48) / 4
			leftW := usable * 6 / 10
			rx := 270 + leftW + 16
			yy := 329
			if pointIn(x, y, rx+20, yy+62, usable-leftW-56, 42) {
				app.page = 1
				layoutEdits()
				invalidate()
				return 0
			}
			if pointIn(x, y, rx+20, yy+119, usable-leftW-56, 42) {
				app.page = 5
				layoutEdits()
				go refreshDevices()
				invalidate()
				return 0
			}
			if pointIn(x, y, rx+20, yy+176, usable-leftW-56, 42) {
				app.page = 6
				layoutEdits()
				invalidate()
				return 0
			}
			_ = cw
		case 1:
			if pointIn(x, y, W-262, 203, 205, 43) {
				go searchPublishedCases(getEditText(editSearch), getEditText(editCaseTitle), getEditText(editCaseYear), getEditText(editAPIToken))
			}
			if pointIn(x, y, W-262, 267, 205, 38) {
				app.page = 2
				layoutEdits()
				go prepareAndOpenCaseMap()
				invalidate()
			}
			app.mu.RLock()
			resultCount := len(app.cases)
			app.mu.RUnlock()
			if y >= 405 && resultCount > 0 {
				idx := (y - 405) / 61
				if idx >= 0 && idx < resultCount && idx < 7 {
					openCaseResult(idx)
				}
			}
		case 2:
			if pointIn(x, y, 272, 180, 210, 42) {
				go prepareAndOpenCaseMap()
			}
			if pointIn(x, y, 500, 180, 185, 42) {
				openCourtListenerTokenPage()
			}
		case 3:
			if pointIn(x, y, 270+360, 650, 120, 40) {
				addNode(getEditText(editNode))
				setEditText(editNode, "")
			}
			if pointIn(x, y, W-205, 650, 130, 40) {
				addLink(getEditText(editLink))
				setEditText(editLink, "")
			}
		case 4:
			if pointIn(x, y, W-255, 264, 175, 42) {
				pingHost(getEditText(editHost))
			}
		case 5:
			if pointIn(x, y, W-242, 174, 185, 42) {
				go refreshDevices()
			}
		case 6:
			if pointIn(x, y, 292, 436, 175, 39) {
				analyseNotes(getEditText(editNotes))
			}
		}
		return 0
	case WM_DESTROY:
		saveWorkspace()
		procPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := procDefWindowProc.Call(uintptr(h), uintptr(msg), w, l)
	return r
}

func main() {
	runtime.LockOSThread()
	loadWorkspace()
	h, _, _ := procGetModuleHandle.Call(0)
	hInst = syscall.Handle(h)
	cursor, _, _ := procLoadCursor.Call(0, uintptr(IDC_ARROW))
	wc := WNDCLASSEX{CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})), Style: 0, WndProc: syscall.NewCallback(wndProc), Instance: hInst, Cursor: syscall.Handle(cursor), Background: makeBrush(rgb(7, 13, 23)), ClassName: className}
	procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	title := utf("CyberCircuit Gotham 0.5.0 | Intelligence Workspace")
	r, _, _ := procCreateWindowEx.Call(0, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(title)), WS_OVERLAPPEDWINDOW|WS_VISIBLE, CW_USEDEFAULT, CW_USEDEFAULT, 1440, 940, 0, 0, uintptr(hInst), 0)
	if r == 0 {
		return
	}
	hwnd = syscall.Handle(r)
	procShowWindow.Call(r, SW_SHOW)
	procUpdateWindow.Call(r)
	var msg MSG
	for {
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}
