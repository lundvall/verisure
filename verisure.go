package verisure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"time"
)

var (
	mediaType = "application/json"
	apiURLs   = []string{
		"https://e-api01.verisure.com/xbn/2",
		"https://e-api02.verisure.com/xbn/2"}
)

// Overview generated
type Overview struct {
	AccountPermissions    AccountPermissions   `json:"accountPermissions"`
	ArmState              ArmState             `json:"armState"`
	ArmstateCompatible    bool                 `json:"armstateCompatible"`
	ControlPlugs          []ControlPlug        `json:"controlPlugs"`
	SmartPlugs            []SmartPlug          `json:"smartPlugs"`
	DoorLockStatusList    []interface{}        `json:"doorLockStatusList"`
	TotalSmsCount         int                  `json:"totalSmsCount"`
	ClimateValues         []ClimateValue       `json:"climateValues"`
	InstallationErrorList []interface{}        `json:"installationErrorList"`
	PendingChanges        int                  `json:"pendingChanges"`
	EthernetModeActive    bool                 `json:"ethernetModeActive"`
	EthernetConnectedNow  bool                 `json:"ethernetConnectedNow"`
	HeatPumps             []interface{}        `json:"heatPumps"`
	SmartCameras          []interface{}        `json:"smartCameras"`
	LatestEthernetStatus  LatestEthernetStatus `json:"latestEthernetStatus"`
	CustomerImageCameras  []interface{}        `json:"customerImageCameras"`
	BatteryProcess        BatteryProcess       `json:"batteryProcess"`
	UserTracking          UserTracking         `json:"userTracking"`
	EventCounts           []interface{}        `json:"eventCounts"`
	DoorWindow            DoorWindow           `json:"doorWindow"`
}

// AccountPermissions generated
type AccountPermissions struct {
	AccountPermissionsHash string `json:"accountPermissionsHash"`
}

// ArmState generated
type ArmState struct {
	StatusType string    `json:"statusType"`
	Date       time.Time `json:"date"`
	ChangedVia string    `json:"changedVia"`
}

// ControlPlug generated
type ControlPlug struct {
	DeviceID     string `json:"deviceId"`
	DeviceLabel  string `json:"deviceLabel"`
	Area         string `json:"area"`
	Profile      string `json:"profile"`
	CurrentState string `json:"currentState"`
	PendingState string `json:"pendingState"`
}

// SmartPlug generated
type SmartPlug struct {
	Icon         string `json:"icon"`
	IsHazardous  bool   `json:"isHazardous"`
	DeviceLabel  string `json:"deviceLabel"`
	Area         string `json:"area"`
	CurrentState string `json:"currentState"`
	PendingState string `json:"pendingState"`
}

// ClimateValue generated
type ClimateValue struct {
	DeviceLabel string    `json:"deviceLabel"`
	DeviceArea  string    `json:"deviceArea"`
	DeviceType  string    `json:"deviceType"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity,omitempty"`
	Time        time.Time `json:"time"`
}

// LatestEthernetStatus generated
type LatestEthernetStatus struct {
	LatestEthernetTestResult bool      `json:"latestEthernetTestResult"`
	TestDate                 time.Time `json:"testDate"`
	ProtectedArea            string    `json:"protectedArea"`
	DeviceLabel              string    `json:"deviceLabel"`
}

// BatteryProcess generated
type BatteryProcess struct {
	Active bool `json:"active"`
}

// UserTracking generated
type UserTracking struct {
	InstallationStatus string `json:"installationStatus"`
}

// DoorWindow generated
type DoorWindow struct {
	ReportState      bool               `json:"reportState"`
	DoorWindowDevice []DoorWindowDevice `json:"doorWindowDevice"`
}

// DoorWindowDevice generated
type DoorWindowDevice struct {
	DeviceLabel string    `json:"deviceLabel"`
	Area        string    `json:"area"`
	State       string    `json:"state"`
	Wired       bool      `json:"wired"`
	ReportTime  time.Time `json:"reportTime"`
}

// SmartPlugState command
type SmartPlugState struct {
	DeviceLabel string `json:"deviceLabel"`
	State       bool   `json:"state"`
}

type installation struct {
	GIID            string `json:"giid"`
	FirmwareVersion int    `json:"firmwareVersion"`
	RoutingGroup    string `json:"routingGroup"`
	Shard           int    `json:"shard"`
	Locale          string `json:"locale"`
	SignalFilterID  int    `json:"signalFilterId"`
	Deleted         bool   `json:"deleted"`
	CID             string `json:"cid"`
	Street          string `json:"street"`
	StreetNo1       string `json:"streetNo1"`
	StreetNo2       string `json:"streetNo2"`
	Alias           string `json:"alias"`
}

// Verisure app API client
type Verisure struct {
	baseURL       string
	client        http.Client
	installations []installation
}

// Login ...
func (v *Verisure) Login(ctx context.Context, username, password string) error {
	if err := v.tryURLs(ctx, username, password); err != nil {
		return err
	}

	return v.installation(ctx, username)
}

func (v *Verisure) tryURLs(ctx context.Context, username, password string) error {
	var err error
	for _, u := range apiURLs {
		v.baseURL = u
		if err = v.authenticate(ctx, username, password); err == nil {
			break
		}
	}
	return err
}

func (v *Verisure) authenticate(ctx context.Context, username, password string) error {
	req, err := newRequest(http.MethodPost, v.baseURL+"/cookie", nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth("CPE/"+username, password)

	res, err := v.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("login: %d %s", res.StatusCode, res.Status)
	}

	return nil
}

func (v *Verisure) installation(ctx context.Context, username string) error {
	url := fmt.Sprintf("%s/installation/search?email=%s", v.baseURL, username)
	req, err := newRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	res, err := v.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("installations: %d %s", res.StatusCode, res.Status)
	}

	return json.NewDecoder(res.Body).Decode(&v.installations)
}

// Logout ...
func (v *Verisure) Logout(ctx context.Context) error {
	req, err := http.NewRequest(http.MethodDelete, v.baseURL+"/cookie", nil)
	if err != nil {
		return err
	}

	res, err := v.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("logout: %d %s", res.StatusCode, res.Status)
	}

	return nil
}

// Overview ...
func (v *Verisure) Overview(ctx context.Context) (Overview, error) {
	var o Overview
	url := fmt.Sprintf("%s/installation/%s/overview", v.baseURL, v.installations[0].GIID)
	req, err := newRequest(http.MethodGet, url, nil)
	if err != nil {
		return o, err
	}

	res, err := v.client.Do(req.WithContext(ctx))
	if err != nil {
		return o, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return o, fmt.Errorf("overview: %d %s", res.StatusCode, res.Status)
	}

	err = json.NewDecoder(res.Body).Decode(&o)
	if err != nil {
		return o, err
	}

	return o, nil
}

// UpdateSmartplug ...
func (v *Verisure) UpdateSmartplug(ctx context.Context, updates []SmartPlugState) error {
	bs, err := json.Marshal(updates)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/installation/%s/smartplug/state", v.baseURL, v.installations[0].GIID)
	req, err := newRequest(http.MethodPost, url, bytes.NewReader(bs))
	if err != nil {
		return err
	}

	res, err := v.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("smartplug: %d %s", res.StatusCode, res.Status)
	}

	return nil
}

// New Verisure client
func New() Verisure {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}

	return Verisure{
		client:        http.Client{Jar: jar},
		installations: make([]installation, 0)}
}

func newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return req, err
	}

	req.Header.Add("Accept", mediaType)
	req.Header.Add("Content-Type", mediaType)

	return req, nil
}
