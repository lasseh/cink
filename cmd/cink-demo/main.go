package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/lasseh/cink/highlighter"
)

const sampleConfig = `!
hostname core-router-01
!
interface GigabitEthernet0/0/0
 description Uplink to ISP
 ip address 203.0.113.1 255.255.255.252
 no shutdown
!
interface GigabitEthernet0/0/1
 description Server LAN
 ip address 10.0.1.1 255.255.255.0
 switchport mode access
 switchport access vlan 100
 spanning-tree portfast
 no shutdown
!
interface Loopback0
 ip address 10.255.255.1 255.255.255.255
!
interface Vlan100
 description Management VLAN
 ip address 10.100.0.1 255.255.255.0
 no shutdown
!
router ospf 1
 router-id 10.255.255.1
 network 10.0.0.0 0.0.0.255 area 0
 network 10.255.255.1 0.0.0.0 area 0
 passive-interface default
 no passive-interface GigabitEthernet0/0/0
!
router bgp 65001
 bgp router-id 10.255.255.1
 bgp log-neighbor-changes
 neighbor 203.0.113.2 remote-as 65000
 neighbor 203.0.113.2 description ISP Transit Peer
 !
 address-family ipv4 unicast
  network 10.0.0.0 mask 255.255.0.0
  neighbor 203.0.113.2 activate
  neighbor 203.0.113.2 route-map ISP-IN in
  neighbor 203.0.113.2 route-map ISP-OUT out
 exit-address-family
!
ip access-list extended PROTECT
 permit tcp 10.0.0.0 0.0.255.255 any eq 22
 permit tcp 10.0.0.0 0.0.255.255 any eq 443
 permit icmp any any
 deny   ip any any log
!
ip prefix-list DEFAULT-ONLY seq 10 permit 0.0.0.0/0
ip prefix-list DEFAULT-ONLY seq 20 deny 0.0.0.0/0 le 32
!
route-map ISP-IN permit 10
 match ip address prefix-list DEFAULT-ONLY
!
route-map ISP-OUT permit 10
 match ip address prefix-list ADVERTISE
!
ip route 0.0.0.0 0.0.0.0 203.0.113.2 name Default-to-ISP
!
logging host 10.0.0.100
logging trap informational
!
ntp server 10.0.0.1
ntp server 10.0.0.2
!
snmp-server community public RO
snmp-server location "Main Data Center, Rack 42"
snmp-server contact noc@example.com
!
line con 0
 exec-timeout 5 0
 logging synchronous
!
line vty 0 15
 access-class MGMT-ACCESS in
 transport input ssh
 login local
!
banner motd ^
*** WARNING: Authorized access only ***
^
!
end
`

const sampleBGPSummary = `BGP router identifier 10.255.255.1, local AS number 65001
BGP table version is 12345, main routing table version 12345
4 network entries using 992 bytes of memory
6 path entries using 480 bytes of memory

Neighbor        V           AS MsgRcvd MsgSent   TblVer  InQ OutQ Up/Down  State/PfxRcd
203.0.113.2     4        65000   12345   12340    12345    0    0 1w2d     150
10.0.0.2        4        65001    8234    8230    12345    0    0 3d12:30  2500
192.168.1.1     4        65002     100     105    12345    0   15 00:05:30 Active
172.16.0.1      4        65003       0       0        0    0    0 2w1d     Idle
10.0.0.5        4        65004    5000    4998    12345    0    0 12:45:00 established
`

const sampleOSPFNeighbors = `Neighbor ID     Pri   State           Dead Time   Address         Interface
10.255.255.2    128   FULL/DR         00:00:35    10.0.0.2        GigabitEthernet0/0/0
10.255.255.3    128   FULL/BDR        00:00:38    10.0.0.6        GigabitEthernet0/0/1
10.255.255.4      1   2WAY/DROTHER    00:00:32    10.0.0.10       Port-channel1
10.255.255.5    128   INIT/-          00:00:40    10.0.0.14       GigabitEthernet0/0/2
10.255.255.6    128   EXSTART/DR      00:00:37    10.0.0.18       TenGigabitEthernet1/0/0
0.0.0.0           0   DOWN/-          00:00:00    172.16.0.2      Tunnel0
`

const sampleInterfaceBrief = `Interface                  IP-Address      OK? Method Status                Protocol
GigabitEthernet0/0/0       203.0.113.1     YES manual up                    up
GigabitEthernet0/0/1       10.0.1.1        YES manual up                    up
GigabitEthernet0/0/2       unassigned      YES unset  administratively down down
TenGigabitEthernet1/0/0    10.0.0.1        YES manual up                    up
Loopback0                  10.255.255.1    YES manual up                    up
Vlan100                    10.100.0.1      YES manual up                    up
Port-channel1              172.16.0.1      YES manual up                    up
Tunnel0                    192.168.100.1   YES manual up                    down
Null0                      unassigned      YES unset  up                    up
`

const sampleShowVersion = `Cisco IOS XE Software, Version 17.06.01
Cisco IOS Software [Bengaluru], ASR1000 Software (X86_64_LINUX_IOSD-UNIVERSALK9-M), Version 17.6.1, RELEASE SOFTWARE (fc2)
Technical Support: http://www.cisco.com/techsupport
Copyright (c) 1986-2024 by Cisco Systems, Inc.
Compiled Thu 20-Jun-24 12:15 by mcpre

cisco ASR1001-X (1NG) processor with 3670989K/6147K bytes of memory.
Processor board ID FXS2012Q3VH
2 Gigabit Ethernet interfaces
2 Ten Gigabit Ethernet interfaces
32768K bytes of non-volatile configuration memory.
8388608K bytes of physical memory.

Configuration register is 0x2102
`

const sampleMACTable = `          Mac Address Table
-------------------------------------------

Vlan    Mac Address       Type        Ports
----    -----------       --------    -----
 100    0011.2233.4455    DYNAMIC     Gi0/0/1
 100    aabb.ccdd.eeff    DYNAMIC     Gi0/0/2
 200    1122.3344.5566    STATIC      Po1
 All    0100.0ccc.cccc    STATIC      CPU
`

func main() {
	var (
		themeName  string
		showAll    bool
		showOutput bool
	)

	flag.StringVar(&themeName, "theme", "default", "Theme: default, solarized, monokai, nord, etc.")
	flag.StringVar(&themeName, "t", "default", "Theme (shorthand)")
	flag.BoolVar(&showAll, "all", false, "Show all themes")
	flag.BoolVar(&showAll, "a", false, "Show all themes (shorthand)")
	flag.BoolVar(&showOutput, "show", false, "Show 'show' command output demo")
	flag.BoolVar(&showOutput, "o", false, "Show command output demo (shorthand)")

	flag.Parse()

	if showAll {
		showAllThemes()
		return
	}

	if showOutput {
		showShowOutputDemo(themeName)
		return
	}

	theme := highlighter.ThemeByName(strings.ToLower(themeName))
	hl := highlighter.NewWithTheme(theme)

	fmt.Printf("\n=== Cisco IOS Syntax Highlighting Demo (Theme: %s) ===\n\n", themeName)
	fmt.Println(hl.HighlightForced(sampleConfig))
}

func showAllThemes() {
	themes := []struct {
		name  string
		theme *highlighter.Theme
	}{
		{"tokyonight (default)", highlighter.TokyoNightTheme()},
		{"vibrant", highlighter.VibrantTheme()},
		{"solarized", highlighter.SolarizedDarkTheme()},
		{"monokai", highlighter.MonokaiTheme()},
		{"nord", highlighter.NordTheme()},
		{"catppuccin", highlighter.CatppuccinMochaTheme()},
		{"dracula", highlighter.DraculaTheme()},
		{"gruvbox", highlighter.GruvboxDarkTheme()},
		{"onedark", highlighter.OneDarkTheme()},
	}

	sample := `!
hostname router-01
!
interface GigabitEthernet0/0/0
 description Uplink to ISP
 ip address 192.168.1.1 255.255.255.0
 no shutdown
!
router bgp 65001
 neighbor 10.0.0.1 remote-as 65000
!
ip access-list extended PROTECT
 permit tcp 10.0.0.0 0.0.255.255 any eq 22
 deny   ip any any log
!
`

	for _, t := range themes {
		hl := highlighter.NewWithTheme(t.theme)
		fmt.Printf("\n=== Theme: %s ===\n", t.name)
		fmt.Println(hl.HighlightForced(sample))
	}
}

func showShowOutputDemo(themeName string) {
	theme := highlighter.ThemeByName(strings.ToLower(themeName))
	hl := highlighter.NewWithTheme(theme)

	fmt.Printf("\n=== Cisco Show Output Highlighting Demo (Theme: %s) ===\n", themeName)

	fmt.Println("\n--- show ip bgp summary ---")
	fmt.Println(hl.HighlightShowOutput(sampleBGPSummary))

	fmt.Println("\n--- show ip ospf neighbor ---")
	fmt.Println(hl.HighlightShowOutput(sampleOSPFNeighbors))

	fmt.Println("\n--- show ip interface brief ---")
	fmt.Println(hl.HighlightShowOutput(sampleInterfaceBrief))

	fmt.Println("\n--- show version ---")
	fmt.Println(hl.HighlightShowOutput(sampleShowVersion))

	fmt.Println("\n--- show mac address-table ---")
	fmt.Println(hl.HighlightShowOutput(sampleMACTable))
}
