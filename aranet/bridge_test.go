package aranet

import (
	"context"
	"testing"
	"time"

	"github.com/brutella/hap/accessory"
)

func TestHomekitURI(t *testing.T) {
	assertURI := func(uri, expected string) {
		t.Run("generate uri "+expected, func(t *testing.T) {
			if uri != expected {
				t.Fatalf("expected " + uri + " to equal " + expected)
			}
		})
	}

	assertURI(HomekitSetupURI(8, 2, "84131633", ""), "X-HM://0081YCYEP")
	assertURI(HomekitSetupURI(8, 2, "84131633", "3QYT"), "X-HM://0081YCYEP3QYT")
	assertURI(HomekitSetupURI(8, 4, "84131633", "3QYT"), "X-HM://0086E6GJ53QYT")
}

func TestBridgeSetupURI(t *testing.T) {
	acc := NewAranetAccessory(accessory.Info{})
	bridge := NewBridge(t.TempDir(), "12344321", "RNDM", acc)

	if bridge.SetupURI != "X-HM://009ZSQCXTRNDM" {
		t.Fatalf("expected setup URI to encode correctly.\nthis depends on sensor category, ip transport, pin, setup id.\ngot: %v", bridge.SetupURI)
	}

	if bridge.Pin != "123-44-321" {
		t.Fatalf("expected bridge pin to be human readable with dashes, got: %v", bridge.Pin)
	}
}

func TestBridgeServes(t *testing.T) {
	acc := NewAranetAccessory(accessory.Info{})
	bridge := NewBridge(t.TempDir(), "12344321", "RNDM", acc)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		bridge.Serve(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	cancel()
}
