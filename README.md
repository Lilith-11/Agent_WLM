Agent Software
1.This Project is go based system monitoring agent that collects the windows system metrices,linux system metrices ,Darwin system metrices
  ~displays them in terminal
  ~stores in text file(agent.txt)
  ~ automatically stores in the home folder
2.The agent gathers 
  ~system information
  ~CPU details
  ~RAM details 
  ~Disk Usage 
  ~bootime
  ~Hyperthreading status
3.Features 

 "System Information"
  1.AgentId (Randomly generated)
  2.Hostname
  3.os & Version
  4.Platform
  5.Architecture
  6.Boot time

 "RAM Information"
  1.Total RAM(GB)
  2.Available RAM(GB)

 "CPU Information"
  1.Manufacturer
  2.Model 
  3.CPU speed (MHz)
  4.Physical cores
  5.Logical Cores
  6.Hypertgreading status
  7.Usage Per Core
  8.Total core Usage

"Disk Information"
  1.Drive Letter
  2.File System Type
  3.Total Size
  4.Used Space 
  5.Used Percentage
  6.Free Sapce GB
________________________________________________________________________________________________
"Windows system Monitoring agent"
 FOR WINDOWS :I call the following Windows API to collect the information

-GetLogicalDrives   //retrives the a bitmask representing availbale disk drives
-GetDiskFreeSpaceExW  //gets information about disk space 
-GetDriveTypeW     //Identifying drive categories
-GetTicketCount64  //retruns the number of milliseconds since the syatem started
-GlobalMemoryStatusEx  //monitoring RAM usage

Build Instructions for Windows
->$env:GOOS="windows" 
->$env:GOARCH="amd64"
->go build -o agent.exe agentprogram.go
________________________________________________________________________________________________

"linux system monitoring agent"
FOR LINUX : I read the data directly from 
 
-/proc/stat  //CPU usage like user,system ,idletime -boottime
-/proc/meminfo //provides memory information like TOtal RAM ,Free RAM ,Available memory
-/proc/cpuinfo //provides the detailed CPU information
-/proc/mounts //Lists all mounted filesystems
-/etc/os-release  //provides operating system indentification data
 
 Build Instructions for Linux
 ->$env:GOOS="linux"
 ->$env:GOARCH="amd64"
 ->go build -o agentLinux linux_sys.go
________________________________________________________________________________________________

"MacOs system monitoring Agent"
For MAC:I collect the Information by executing the following commands

-sw_vers -productVersion //gives os version
-sysctl -n kern.boottime //gives system boot timestamp
-sysctl -n hw.memsize    //gives total RAM in bytes and then converted to Gb
-sysctl -n machdep.cpu.vendor  //gives CPU manufacturer
-sysctl -n machdep.cpu.brand_string  //gives cpu model name
-sysctl -n hw.cpufrequency  //gives CPU frequency in Hz 
-sysctl -n hw.physicalcpu   //gives number of physical cores 
-sysctl -n hw.logicalcpu    //gives number of logiacal cores
-top -l 2 -n 0 //gives cpu usage snapshot
-df -kp     //gives disk usage info

 Build Instructions for Linux
 ->$env:GOOS="darwin"
 ->$env:GOARCH="amd64" //for intel use amd64 (or)
 ->$env:GOARCH="amr64"  //for apple silicon use arm64
 ->go build -o agentDarwin Darwin_sys.go
