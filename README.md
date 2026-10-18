# screen-daemon
A daemon written to control monitors for a 3 port KVM switch based upon what system is active. You probably don't need this, but if you have a system with 4 monitors and a 3 switch KVM that only supports dual monitors here's an example.

Candidly, this is also uploaded here just because I keep getting asked why.

## Why
I have a 3 port KVM on my desk which includes systems for: a work laptop, a remote KVM over ethernet, and the thin client at my desk. Additionally there are 4 monitors connected to this desk\setup and the KVM only supports 2 monitors. You could argue to purchase a 4 monitor KVM which is fair however I could not find one that: supported 3 systems, mix of hdmi and displayport (I use both of these to have more monitors connected), and was controllable over RF.

As for the KVM over ethernet...that machine has 3 operating systems installed...I do not trust the remote access software will work entirely and you can't use remote access on a Windows box that has software running that includes an anticheat. Sadly due to limitations of this tech the physical hardware only allows a single monitor which is plugged directly into the KVM.

As for the work laptop most of the machines I have had issued to me typically are only going to support two monitors so simply enough I connect this to the two slots on the KVM.

As for the thin client I have 4 monitors and the KVM as mentioned previously only supports 2 monitors. This means that I need to have my side monitors plugged in directly to thin client and bypass the KVM. This here is also why I have this application in the first place because I want these monitors off if I am using a different device.

Lastly if I want to turn off my monitors when I leave the room I need an API for control my monitors over the network and preferably avoid the use of a smart plug where possible (this is why ddcutil is used).

## How this works

For determining if the monitors should be off we poll an external API (ex https://github.com/friedmicro/automation-state) and flip off the monitors if the API returns the correct state. For determining if the thin client is connected we check for the USB mouse connected to the device. For determining if the KVM in the next room is in use we check for a a file (`/tmp/windows-bare-metal`). If neither of these conditions are true then we know that the work laptop is in use. For each of these respective states we then toggle the monitor states accordingly.

## Run\build
This is highly specific to me, but if you would like to change the code\experiment with it please update the URLs (the current webhooks are only accessible locally on my network) and build the project with `go build`.
