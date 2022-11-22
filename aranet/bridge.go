package aranet

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"
)

var (
	pinRegexp     = regexp.MustCompile("^\\d{8}$")
	setupIDRegexp = regexp.MustCompile("^[A-Z0-9]{4}$")
)

type Bridge struct {
	server *hap.Server

	Pin      string
	SetupURI string
}

func NewBridge(stateDir string, pin string, setupID string, accessories ...*Accessory) *Bridge {
	if !pinRegexp.Match([]byte(pin)) {
		panic(fmt.Sprintf("failed homekit setup: invalid pin %v does not match regex %v", pin, pinRegexp))
	}
	if !setupIDRegexp.Match([]byte(setupID)) {
		panic(fmt.Sprintf("failed homekit setup: invalid pin %v does not match regex %v", setupID, setupIDRegexp))
	}

	bridgeAcc := accessory.NewBridge(accessory.Info{Name: "Aranet4 Exporter", Manufacturer: "Ryan Souza", Model: "https://github.com/ryansouza/aranet4-exporter-go"})
	db := hap.NewFsStore(stateDir)

	as := []*accessory.A{}
	for _, acc := range accessories {
		as = append(as, acc.A)
	}

	server, err := hap.NewServer(db, bridgeAcc.A, as...)
	if err != nil {
		panic(fmt.Sprintf("failed homekit setup: %v\n", err))
	}

	server.Pin = pin
	server.SetupId = setupID

	return &Bridge{
		server: server,
		Pin:    pin[0:3] + "-" + pin[3:5] + "-" + pin[5:8],

		// 10 is sensor category, 2 is IP transport
		SetupURI: HomekitSetupURI(10, 2, server.Pin, server.SetupId),
	}
}

func (bridge *Bridge) Serve(ctx context.Context) error {
	return bridge.server.ListenAndServe(ctx)
}

func HomekitSetupURI(categoryId uint8, flag uint8, setupCode string, setupID string) string {
	version := uint64(0)
	reserved := uint64(0)

	setupCodeNumber, err := strconv.ParseUint(setupCode, 10, 64)
	if err != nil {
		panic("could not parse setup code as int: " + setupCode)
	}

	var payload uint64

	payload = payload | (version & 0x7)
	payload = payload << 4

	payload = payload | (reserved & 0xf)
	payload = payload << 8

	payload = payload | (uint64(categoryId) & 0xff)
	payload = payload << 4

	payload = payload | (uint64(flag) & 0xf)
	payload = payload << 27

	payload = payload | (setupCodeNumber & 0x7fffffff)
	payload36 := strconv.FormatUint(payload, 36)

	paddingZeroes := strings.Repeat("0", (9 - len(payload36)))

	return "X-HM://" + paddingZeroes + strings.ToUpper(payload36) + strings.ToUpper(setupID)
}
