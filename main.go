package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	stateAPIURL        = "http://state.serpents.friedmicro-lab.org"
	webhookDenLight    = "http://serpent.friedmicro-lab.org:8123/api/webhook/-8nHLkTeXXh2i2xZGMsXYfu47"
	webhookThinClient  = "http://serpent.friedmicro-lab.org:8123/api/webhook/-VxyHGjPzcz9dv7_Vevuytvc9"
	webhookWindowsBare = "http://serpent.friedmicro-lab.org:8123/api/webhook/-2vo8dhzah-34sBY9-q_cGP8K"

	windowsBareMetalFile = "/tmp/windows-bare-metal"
	pollInterval         = 10 * time.Second
	delayForHardware     = 60 * time.Second

	displayOn  = "01"
	displayOff = "04"
)

type currentMode string

const (
	modeUnknown          currentMode = ""
	modeThinClient       currentMode = "thin_client"
	modeWindowsBareMetal currentMode = "windows_bare_metal"
	modeWorkKVM          currentMode = "work_kvm"
)

var (
	httpClient         = &http.Client{Timeout: 30 * time.Second}
	previousLightState = true
	currentModeState   = modeUnknown
	firstRun           = true
)

func runCommand(name string, args ...string) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("command %q failed: %v (output: %q)\n", strings.Join(append([]string{name}, args...), " "), err, strings.TrimSpace(string(out)))
	}
}

type lightStateResult struct {
	DenLight *string `json:"den_light"`
}

func getLightState() (lightOff, ok bool, err error) {
	resp, err := httpClient.Get(stateAPIURL)
	if err != nil {
		return false, false, fmt.Errorf("GET %s: %w", stateAPIURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, false, nil
	}

	var raw lightStateResult
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1024)).Decode(&raw); err != nil {
		return false, false, fmt.Errorf("parse state JSON: %w", err)
	}

	if raw.DenLight == nil {
		return false, true, nil
	}
	return *raw.DenLight == "off", true, nil
}

func postWebhook(url string) error {
	resp, err := httpClient.Post(url, "", nil)
	if err != nil {
		return fmt.Errorf("webhook POST %s: %w", url, err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook POST %s returned status %d", url, resp.StatusCode)
	}
	return nil
}

func postWebhookLogged(url string) error {
	if err := postWebhook(url); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}
	return nil
}

func displaysOff() {
	runCommand("killall", "Hyprland")
	fmt.Println("Lights are off updating state")
	currentModeState = modeUnknown
	runCommand("ddcutil", "setvcp", "D6", displayOff, "--display", "2")
	postWebhookLogged(webhookDenLight)
	time.Sleep(delayForHardware)
}

func handleHomeAutomations() bool {
	lightOff, lightOK, err := getLightState()
	if err != nil {
		fmt.Printf("Error on Time API, possibly a network disconnection...logging for future reference, %v\n", err)
	}

	if lightOK {
		switch {
		case lightOff && previousLightState:
			firstRun = true
			displaysOff()
			previousLightState = false
			return true
		case lightOff:
			return true
		default:
			previousLightState = true
		}
	}
	return false
}

func isUSBConnected() bool {
	usbOut, err := exec.Command("lsusb").CombinedOutput()
	if err != nil {
		fmt.Printf("lsusb failed: %v\n", err)
		os.Exit(43)
	}
	return strings.Contains(string(usbOut), "Razer")
}

func killWindowSystem() {
	runCommand("killall", "Hyprland")
}

func thinClient() {
	fmt.Println("Running as thin client")
	currentModeState = modeThinClient
	if !firstRun {
		killWindowSystem()
		if err := postWebhookLogged(webhookDenLight); err == nil {
			time.Sleep(delayForHardware)
		}
	}
	runCommand("ddcutil", "setvcp", "D6", displayOn, "--display", "1")
	if err := postWebhookLogged(webhookThinClient); err == nil {
		time.Sleep(delayForHardware)
	}
}

func windowsBareMetal() {
	fmt.Println("Running as Windows bare metal")
	currentModeState = modeWindowsBareMetal
	runCommand("ddcutil", "setvcp", "D6", displayOff, "--display", "2")
	killWindowSystem()
	if err := postWebhookLogged(webhookDenLight); err == nil {
		time.Sleep(delayForHardware)
	}
	postWebhookLogged(webhookWindowsBare)
}

func workKVM() {
	fmt.Println("Running as work KVM")
	currentModeState = modeWorkKVM
	runCommand("ddcutil", "setvcp", "D6", displayOff, "--display", "2")
	killWindowSystem()
	if err := postWebhookLogged(webhookDenLight); err == nil {
		time.Sleep(delayForHardware)
	}
	postWebhookLogged(webhookWindowsBare)
}

func main() {
	fmt.Println("screen-daemon starting")
	for {
		time.Sleep(pollInterval)
		if handleHomeAutomations() {
			continue
		}

		usbConnected := isUSBConnected()

		_, markerErr := os.Stat(windowsBareMetalFile)
		windowsBareMetalActive := markerErr == nil

		if usbConnected && !windowsBareMetalActive && currentModeState != modeThinClient {
			thinClient()
		} else if !usbConnected && windowsBareMetalActive && currentModeState != modeWindowsBareMetal {
			windowsBareMetal()
		} else if !usbConnected && !windowsBareMetalActive && currentModeState != modeWorkKVM {
			workKVM()
		}

		firstRun = false
	}
}
