import os
import subprocess
import time

import requests


def run_command(command):
    command = command.split()
    result = subprocess.run(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    output = result.stdout.decode("utf-8")
    return output


# Monitor control
# 01 Display on
# 04 Display off

usb_connected = False
windows_bare_metal = False

# Let's not immediately reboot the machine
is_first_run = True

# Let's not run actions twice
current_state = ""

while True:

    response = requests.get("http://state.serpents.friedmicro-lab.org")
    if response.ok:
        body = response.json()
        if "den_light" in body:
            light_state = body["den_light"]
            if light_state == "off":
                is_first_run = True
                current_state = ""
                requests.post(
                    "http://serpent.friedmicro-lab.org:8123/api/webhook/-8nHLkTeXXh2i2xZGMsXYfu47"
                )
                time.sleep(60)
                run_command("ddcutil setvcp D6 04 --display 2")
                continue

    time.sleep(10)
    usb_output = run_command("lsusb")
    if not "Razer" in usb_output:
        usb_connected = False
    else:
        usb_connected = True
    windows_bare_metal = os.path.exists("/tmp/windows-bare-metal")
    # Thin Client
    if usb_connected and windows_bare_metal == False and current_state != "thin_client":
        print("Running as thin client")
        current_state = "thin_client"
        if not is_first_run:
            requests.post(
                "http://serpent.friedmicro-lab.org:8123/api/webhook/-8nHLkTeXXh2i2xZGMsXYfu47"
            )
            time.sleep(60)
        run_command("ddcutil setvcp D6 01 --display 1")
        requests.post(
            "http://serpent.friedmicro-lab.org:8123/api/webhook/-VxyHGjPzcz9dv7_Vevuytvc9"
        )
        time.sleep(60)
    # Windows bare metal
    elif (
        usb_connected == False
        and windows_bare_metal
        and current_state != "windows_bare_metal"
    ):
        print("Running as Windows bare metal")
        current_state = "windows_bare_metal"
        run_command("ddcutil setvcp D6 04 --display 2")
        # run_command("killall Hyprland")
        requests.post(
            "http://serpent.friedmicro-lab.org:8123/api/webhook/-8nHLkTeXXh2i2xZGMsXYfu47"
        )
        time.sleep(60)
        requests.post(
            "http://serpent.friedmicro-lab.org:8123/api/webhook/-2vo8dhzah-34sBY9-q_cGP8K"
        )
    # Work KVM
    elif (
        usb_connected == False
        and windows_bare_metal == False
        and current_state != "work_kvm"
    ):
        print("Running as work KVM")
        current_state = "work_kvm"
        run_command("ddcutil setvcp D6 04 --display 2")
        # run_command("killall Hyprland")
        requests.post(
            "http://serpent.friedmicro-lab.org:8123/api/webhook/-8nHLkTeXXh2i2xZGMsXYfu47"
        )
        time.sleep(60)
        requests.post(
            "http://serpent.friedmicro-lab.org:8123/api/webhook/-2vo8dhzah-34sBY9-q_cGP8K"
        )
    is_first_run = False
