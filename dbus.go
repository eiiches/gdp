package main

import (
	"github.com/godbus/dbus/v5"
)

// $ gdbus introspect --session --dest org.gnome.Mutter.DisplayConfig --object-path /org/gnome/Mutter/DisplayConfig
//
//  interface org.gnome.Mutter.DisplayConfig {
//    methods:
//      GetResources(out u serial,
//                   out a(uxiiiiiuaua{sv}) crtcs,
//                   out a(uxiausauaua{sv}) outputs,
//                   out a(uxuudu) modes,
//                   out i max_screen_width,
//                   out i max_screen_height);
//      ApplyConfiguration(in  u serial,
//                         in  b persistent,
//                         in  a(uiiiuaua{sv}) crtcs,
//                         in  a(ua{sv}) outputs);
//      ChangeBacklight(in  u serial,
//                      in  u output,
//                      in  i value,
//                      out i new_value);
//      GetCrtcGamma(in  u serial,
//                   in  u crtc,
//                   out aq red,
//                   out aq green,
//                   out aq blue);
//      SetCrtcGamma(in  u serial,
//                   in  u crtc,
//                   in  aq red,
//                   in  aq green,
//                   in  aq blue);
//      GetCurrentState(out u serial,
//                      out a((ssss)a(siiddada{sv})a{sv}) monitors,
//                      out a(iiduba(ssss)a{sv}) logical_monitors,
//                      out a{sv} properties);
//      ApplyMonitorsConfig(in  u serial,
//                          in  u method,
//                          in  a(iiduba(ssa{sv})) logical_monitors,
//                          in  a{sv} properties);
//      SetOutputCTM(in  u serial,
//                   in  u output,
//                   in  (ttttttttt) ctm);
//    signals:
//      MonitorsChanged();
//    properties:
//      readwrite i PowerSaveMode = 0;
//      readonly b PanelOrientationManaged = false;
//  };
//
// https://gitlab.gnome.org/GNOME/mutter/-/blob/main/data/dbus-interfaces/org.gnome.Mutter.DisplayConfig.xml
//
// meta_monitor_manager_handle_get_current_state()
// https://gitlab.gnome.org/GNOME/mutter/-/blob/7734d6f56b66be22986b4a69582099838229ac48/src/backends/meta-monitor-manager.c#L1859
//
// meta_monitor_manager_handle_apply_monitors_config()
// https://gitlab.gnome.org/GNOME/mutter/-/blob/7734d6f56b66be22986b4a69582099838229ac48/src/backends/meta-monitor-manager.c#L2557

type ConnectorAndMonitorId struct {
	Connector string `json:"connector"` // connector name (e.g. HDMI-1, DP-1, etc)
	Vendor    string `json:"vendor"`    // vendor name (e.g. VSC)
	Product   string `json:"product"`   // product name (e.g. VX2705-2KP)
	Serial    string `json:"serial"`    // product serial (e.g. W6Z213******)
}

func (this *ConnectorAndMonitorId) FromDbusValue(value interface{}) error {
	rawId := value.([]interface{})
	this.Connector = rawId[0].(string)
	this.Vendor = rawId[1].(string)
	this.Product = rawId[2].(string)
	this.Serial = rawId[3].(string)
	return nil
}

type Monitor struct {
	Id         *ConnectorAndMonitorId  `json:"id"`
	Modes      []*MonitorMode          `json:"modes"`      // available modes
	Properties map[string]dbus.Variant `json:"properties"` // optional properties, including: display-name, is-builtin, ...
}

func (this *Monitor) FromDbusValue(value interface{}) error {
	rawMonitor := value.([]interface{})

	id := &ConnectorAndMonitorId{}
	id.FromDbusValue(rawMonitor[0])
	this.Id = id

	modes := []*MonitorMode{}
	for _, rawMode := range rawMonitor[1].([][]interface{}) {
		mode := &MonitorMode{}
		mode.FromDbusValue(rawMode)
		modes = append(modes, mode)
	}
	this.Modes = modes

	this.Properties = rawMonitor[2].(map[string]dbus.Variant)

	return nil
}

type MonitorMode struct {
	Id              string                  `json:"id"`               // mode id
	Width           int32                   `json:"width"`            // width in physical pixels
	Height          int32                   `json:"height"`           // height in physical pixels
	RefreshRate     float64                 `json:"refresh_rate"`     // refresh rate
	PreferredScale  float64                 `json:"preferred_scale"`  // scale preferred as per calculations
	SupportedScales []float64               `json:"supported_scales"` // scales supported by this mode
	Properties      map[string]dbus.Variant `json:"properties"`       // optional properties, including: is-current, is-preferred, is-interlaced
}

func (this *MonitorMode) FromDbusValue(value interface{}) error {
	rawMode := value.([]interface{})
	this.Id = rawMode[0].(string)
	this.Width = rawMode[1].(int32)
	this.Height = rawMode[2].(int32)
	this.RefreshRate = rawMode[3].(float64)
	this.PreferredScale = rawMode[4].(float64)
	this.SupportedScales = rawMode[5].([]float64)
	this.Properties = rawMode[6].(map[string]dbus.Variant)
	return nil
}

type LogicalMonitor struct {
	X          int32                    `json:"x"`          // x position
	Y          int32                    `json:"y"`          // y position
	Scale      float64                  `json:"scale"`      // scale
	Transform  uint32                   `json:"transform"`  // transform (0: normal, 1: 90 deg, 2: 180 deg, 3: 270 deg, 4: flipped, 5: 90 deg flipped, 6: 180 deg flipped, 7: 270 deg flipped)
	Primary    bool                     `json:"primary"`    // true if this is the primary logical monitor
	Monitors   []*ConnectorAndMonitorId `json:"monitors"`   // monitors displaying this logical monitor
	Properties map[string]dbus.Variant  `json:"properties"` // possibly other properties
}

func (this *LogicalMonitor) FromDbusValue(value interface{}) error {
	rawLogicalMonitor := value.([]interface{})
	this.X = rawLogicalMonitor[0].(int32)
	this.Y = rawLogicalMonitor[1].(int32)
	this.Scale = rawLogicalMonitor[2].(float64)
	this.Transform = rawLogicalMonitor[3].(uint32)
	this.Primary = rawLogicalMonitor[4].(bool)

	monitorIds := []*ConnectorAndMonitorId{}
	for _, rawMonitorId := range rawLogicalMonitor[5].([][]interface{}) {
		monitorId := &ConnectorAndMonitorId{}
		monitorId.FromDbusValue(rawMonitorId)
		monitorIds = append(monitorIds, monitorId)
	}
	this.Monitors = monitorIds

	this.Properties = rawLogicalMonitor[6].(map[string]dbus.Variant)
	return nil
}

type ConnectorAndMode struct {
	Connector  string
	Mode       string
	Properties map[string]dbus.Variant
}

type LogicalMonitorRequest struct {
	X         int32               `json:"x"`         // x position
	Y         int32               `json:"y"`         // y position
	Scale     float64             `json:"scale"`     // scale
	Transform uint32              `json:"transform"` // transform (0: normal, 1: 90 deg, 2: 180 deg, 3: 270 deg, 4: flipped, 5: 90 deg flipped, 6: 180 deg flipped, 7: 270 deg flipped)
	Primary   bool                `json:"primary"`   // true if this is the primary logical monitor
	Monitors  []*ConnectorAndMode `json:"monitors"`  // monitors displaying this logical monitor
}

type GetCurrentStateRequest struct {
}

type GetCurrentStateResponse struct {
	Serial          uint32                  `json:"serial"`
	Monitors        []*Monitor              `json:"monitors"`
	LogicalMonitors []*LogicalMonitor       `json:"logical_monitors"`
	Properties      map[string]dbus.Variant `json:"properties"`
}

type GetCurrentStateRawResponse struct {
	Serial          uint32
	Monitors        [][]interface{}
	LogicalMonitors [][]interface{}
	Properties      map[string]dbus.Variant
}

func (this *GetCurrentStateRawResponse) ToGetCurrentStateResponse() (*GetCurrentStateResponse, error) {
	monitors := []*Monitor{}
	for _, rawMonitor := range this.Monitors {
		monitor := &Monitor{}
		monitor.FromDbusValue(rawMonitor)
		monitors = append(monitors, monitor)
	}

	logicalMonitors := []*LogicalMonitor{}
	for _, rawLogicalMonitor := range this.LogicalMonitors {
		logicalMonitor := &LogicalMonitor{}
		logicalMonitor.FromDbusValue(rawLogicalMonitor)
		logicalMonitors = append(logicalMonitors, logicalMonitor)
	}

	response := &GetCurrentStateResponse{
		Serial:          this.Serial,
		Monitors:        monitors,
		LogicalMonitors: logicalMonitors,
		Properties:      this.Properties,
	}

	return response, nil
}

type ApplyMonitorsConfigRequest struct {
	Serial          uint32
	Method          uint32
	LogicalMonitors []*LogicalMonitorRequest
	Properties      map[string]dbus.Variant
}

type ApplyMonitorsConfigResponse struct {
}

func GetCurrentState(conn *dbus.Conn, request *GetCurrentStateRequest) (*GetCurrentStateResponse, error) {
	rawResponse := &GetCurrentStateRawResponse{}

	err := conn.Object("org.gnome.Mutter.DisplayConfig", "/org/gnome/Mutter/DisplayConfig").
		Call("org.gnome.Mutter.DisplayConfig.GetCurrentState", 0).
		Store(&rawResponse.Serial, &rawResponse.Monitors, &rawResponse.LogicalMonitors, &rawResponse.Properties)
	if err != nil {
		return nil, err
	}

	response, err := rawResponse.ToGetCurrentStateResponse()
	if err != nil {
		return nil, err
	}

	return response, nil
}

func ApplyMonitorsConfig(conn *dbus.Conn, request *ApplyMonitorsConfigRequest) (*ApplyMonitorsConfigResponse, error) {
	err := conn.Object("org.gnome.Mutter.DisplayConfig", "/org/gnome/Mutter/DisplayConfig").
		Call("org.gnome.Mutter.DisplayConfig.ApplyMonitorsConfig", 0, request.Serial, request.Method, request.LogicalMonitors, request.Properties).
		Store()
	if err != nil {
		return nil, err
	}

	return &ApplyMonitorsConfigResponse{}, nil
}
