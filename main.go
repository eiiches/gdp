package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/godbus/dbus/v5"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

type State struct {
	Monitors        []*Monitor              `json:"monitors"`         // @monitors returned by GetCurrentState()
	LogicalMonitors []*LogicalMonitor       `json:"logical_monitors"` // @logical_monitors returned by GetCurrentState()
	Properties      map[string]dbus.Variant `json:"properties"`       // @properties returned by GetCurrentState()
}

func stateToRequest(resp *GetCurrentStateResponse) (*ApplyMonitorsConfigRequest, error) {

	findMonitor := func(connectorAndMonitorId *ConnectorAndMonitorId) *Monitor {
		for _, monitor := range resp.Monitors {
			if *monitor.Id == *connectorAndMonitorId {
				return monitor
			}
		}
		return nil
	}

	findCurrentMode := func(monitor *Monitor) *MonitorMode {
		for _, mode := range monitor.Modes {
			if _, ok := mode.Properties["is-current"]; ok {
				return mode
			}
		}
		return nil
	}

	monitorRequests := []*LogicalMonitorRequest{}
	for _, monitor := range resp.LogicalMonitors {

		connectorAndModes := []*ConnectorAndMode{}
		for _, connectorAndMonitorId := range monitor.Monitors {
			monitor := findMonitor(connectorAndMonitorId)
			mode := findCurrentMode(monitor)
			connectorAndMode := &ConnectorAndMode{
				Connector: connectorAndMonitorId.Connector,
				Mode:      mode.Id,
			}
			connectorAndModes = append(connectorAndModes, connectorAndMode)
		}

		monitorRequest := &LogicalMonitorRequest{
			X:         monitor.X,
			Y:         monitor.Y,
			Scale:     monitor.Scale,
			Transform: monitor.Transform,
			Primary:   monitor.Primary,
			Monitors:  connectorAndModes,
		}

		monitorRequests = append(monitorRequests, monitorRequest)
	}

	return &ApplyMonitorsConfigRequest{
		Serial:          resp.Serial,
		Method:          2,
		LogicalMonitors: monitorRequests,
		Properties:      map[string]dbus.Variant{},
	}, nil
}

func saveProfile(name string) error {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return errors.Wrapf(err, "dbus.ConnectSessionBus() failed")
	}
	defer conn.Close()

	resp, err := GetCurrentState(conn, &GetCurrentStateRequest{})
	if err != nil {
		return errors.Wrapf(err, "GetCurrentState(conn) failed")
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		return errors.Wrapf(err, "json.Marshal(currentState) failed")
	}

	storage, err := NewLocalStorage()
	if err != nil {
		return errors.Wrapf(err, "NewLocalStorage() failed")
	}

	if err := storage.Store(name, bytes); err != nil {
		return errors.Wrapf(err, "storage.Store(name, bytes) failed")
	}

	return nil
}

func switchProfile(name string) error {
	storage, err := NewLocalStorage()
	if err != nil {
		return errors.Wrapf(err, "NewLocalStorage() failed")
	}

	data, err := storage.Load(name)
	if err != nil {
		return errors.Wrapf(err, "storage.Load(name) failed")
	}

	desiredState := &GetCurrentStateResponse{}
	if err := json.Unmarshal(data, desiredState); err != nil {
		return errors.Wrapf(err, "json.Unmarshal(data, desiredState) failed")
	}

	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return errors.Wrapf(err, "dbus.ConnectSessionBus() failed")
	}
	defer conn.Close()

	resp, err := GetCurrentState(conn, &GetCurrentStateRequest{})
	if err != nil {
		return errors.Wrapf(err, "GetCurrentState(conn) failed")
	}

	req, err := stateToRequest(desiredState)
	if err != nil {
		return errors.Wrapf(err, "stateToRequest(conn) failed")
	}
	req.Serial = resp.Serial
	if _, err := ApplyMonitorsConfig(conn, req); err != nil {
		return errors.Wrapf(err, "ApplyMonitorsConfig(conn, ...) failed")
	}
	return nil
}

func deleteProfile(name string) error {
	storage, err := NewLocalStorage()
	if err != nil {
		return errors.Wrapf(err, "NewLocalStorage() failed")
	}

	if err := storage.Delete(name); err != nil {
		return errors.Wrapf(err, "storage.Delete() failed")
	}
	return nil
}

func listProfiles() error {
	storage, err := NewLocalStorage()
	if err != nil {
		return errors.Wrapf(err, "NewLocalStorage() failed")
	}

	names, err := storage.List()
	if err != nil {
		return errors.Wrapf(err, "storage.List() failed")
	}

	for _, name := range names {
		fmt.Printf("%s\n", name)
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:  "gdp",
		Usage: "manage display profiles and switch between them",
		Commands: []*cli.Command{
			{
				Name:  "delete",
				Usage: "delete a profile",
				Action: func(c *cli.Context) error {
					if c.NArg() != 1 {
						log.Fatal("name is required")
					}
					name := c.Args().First()
					return deleteProfile(name)
				},
			},
			{
				Name:  "list",
				Usage: "list profiles",
				Action: func(c *cli.Context) error {
					return listProfiles()
				},
			},
			{
				Name:  "save",
				Usage: "save a new profile",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "overwrite",
						Value: false,
						Usage: "overwrite existing profile (FIXME: currently, this option is not implemented and always true)",
					},
				},
				Action: func(c *cli.Context) error {
					if c.NArg() != 1 {
						log.Fatal("name is required")
					}
					name := c.Args().First()
					return saveProfile(name)
				},
			},
			{
				Name:    "switch",
				Aliases: []string{"s"},
				Usage:   "switch to a saved profile",
				Action: func(c *cli.Context) error {
					if c.NArg() != 1 {
						log.Fatal("name is required")
					}
					name := c.Args().First()
					return switchProfile(name)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
