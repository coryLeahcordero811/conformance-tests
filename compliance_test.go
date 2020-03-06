package main

import (
	"fmt"
	"net"
	"testing"
	"time"

	"gopkg.in/yaml.v2"
	"gotest.tools/v3/assert"
)

const (
	localhost      = "127.0.0.1"
	pingEntrypoint = "http://" + localhost + ":8080/ping"
	pingTargetUrl  = "target:8080/ping"
	pingUrl        = pingEntrypoint + "?address=" + pingTargetUrl

	volumefileEntrypoint = "http://" + localhost + ":8080/volumefile"
	volumeUrl            = volumefileEntrypoint + "?filename="

	udpEntrypoint = "http://" + localhost + ":8080/udp"

	scaleEntrypoint = "http://" + localhost + ":8080/scalechecker"
)

func TestSimpleLifecycle(t *testing.T) {
	h := TestHelper{
		T:       t,
		testDir: "simple_lifecycle",
	}
	h.TestUpDown(func() {
		time.Sleep(time.Second)
	})
}

func TestSimpleNetwork(t *testing.T) {
	h := TestHelper{
		T:       t,
		testDir: "simple_network",
		specRef: "Networks-top-level-element",
	}
	h.TestUpDown(func() {
		actual := h.getHttpBody(pingUrl)
		expected := jsonResponse("PONG FROM TARGET")
		h.Check(expected, actual)
	})
}

func TestSimpleNetworkFail(t *testing.T) {
	h := TestHelper{
		T:       t,
		testDir: "simple_network",
		specRef: "Networks-top-level-element",
	}
	h.TestUpDown(func() {
		actual := h.getHttpBody(pingEntrypoint + "?address=notatarget:8080/ping")
		expected := jsonResponse("Could not reach address: notatarget:8080/ping")
		h.Check(expected, actual)
	})
}

func TestDifferentNetworks(t *testing.T) {
	h := TestHelper{
		T:       t,
		testDir: "different_networks",
		specRef: "Networks-top-level-element",
	}
	h.TestUpDown(func() {
		actual := h.getHttpBody(pingUrl)
		expected := jsonResponse("Could not reach address: target:8080/ping")
		h.Check(expected, actual)
	})
}

func TestVolumeFile(t *testing.T) {
	h := TestHelper{
		T:       t,
		testDir: "simple_volume",
		specRef: "volumes-top-level-element",
	}
	h.TestUpDown(func() {
		actual := h.getHttpBody(volumeUrl + "test_volume.txt")
		expected := jsonResponse("MYVOLUME")
		h.Check(expected, actual)

	})
}

func TestSecretFile(t *testing.T) {
	h := TestHelper{
		T:       t,
		testDir: "simple_secretfile",
		specRef: "secrets-top-level-element",
	}
	h.TestUpDown(func() {
		actual := h.getHttpBody(volumeUrl + "test_secret.txt")
		expected := jsonResponse("MYSECRET")
		h.Check(expected, actual)
	})
}

func TestConfigFile(t *testing.T) {
	h := TestHelper{
		T:            t,
		testDir:      "simple_configfile",
		skipCommands: []string{"docker-composeV1"},
		specRef:      "configs-top-level-element",
	}
	h.TestUpDown(func() {
		actual := h.getHttpBody(volumeUrl + "test_config.txt")
		expected := jsonResponse("MYCONFIG")
		h.Check(expected, actual)
	})
}

func TestUdpPort(t *testing.T) {
	h := TestHelper{
		T:       t,
		testDir: "udp_port",
		specRef: "Networks-top-level-element",
	}
	h.TestUpDown(func() {
		udpValue := "myUdpvalue"

		ServerAddr, err := net.ResolveUDPAddr("udp", localhost+":10001")
		assert.NilError(h.T, err)
		LocalAddr, err := net.ResolveUDPAddr("udp", localhost+":0")
		assert.NilError(h.T, err)
		Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
		assert.NilError(h.T, err)
		defer Conn.Close()
		buf := []byte(fmt.Sprintf("{\"request\":%q}", udpValue))
		_, err = Conn.Write(buf)
		assert.NilError(h.T, err)

		actual := h.getHttpBody(udpEntrypoint)
		expected := jsonResponse(udpValue)
		h.Check(expected, actual)
	})
}

func TestScaling(t *testing.T) {
	h := TestHelper{
		T:            t,
		testDir:      "scaling",
		skipCommands: []string{"compose-ref"},
		specRef:      "Networks-top-level-element",
	}
	h.TestUpDown(func() {
		actual := h.getHttpBody(scaleEntrypoint)
		responseArray := Response{}
		err := yaml.Unmarshal([]byte(actual), &responseArray)
		assert.NilError(h.T, err)
		h.Check(responseArray.Response, "3")
	})
}
