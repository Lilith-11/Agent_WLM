package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)
type memoryStatusEx struct{
	dwLength uint32
    deMemoryLoad uint32
    ullTotalphys uint64
	ullAvailPhys uint64
	ullTotalPageFile uint64
	ullAvailPageFile uint64
	ullTotalVirtual uint64
	ullAvailVirtual uint64
	ullAvailExtendedVirtual uint64 
}
var plat string=runtime.GOOS
type AgentInfo struct{
	AgentID string
	Hostname string
	os string
	Platform string
	Architecture string
	Boottime time.Time
	TotalRamGB uint64
}
type DiskInfo struct{
	Device string
	FSType string
	Total float64
	Used float64
	Free float64
	Percent float64
}
 type AgentOutput struct{
	AgentDisks []DiskInfo
 }

func getLogicalDrives()[]DiskInfo {
	switch plat{
	case "windows":
	kernek32:=syscall.NewLazyDLL("kernel32.dll")
	getLogicalDrives:=kernek32.NewProc("GetLogicalDrives")
	getDiskFreeSpaceEx:=kernek32.NewProc("GetDiskFreeSpaceExW")
	getDriveType:=kernek32.NewProc("GetDriveTypeW")
	ret,_,_:=getLogicalDrives.Call()
	bitmask:=uint32(ret)
	var disks []DiskInfo
	for i:=0;i<26;i++{
		if bitmask&(1<<uint(i))!=0{
			driveLetter:=string('A'+rune(i))+":"
			drivePath,_:=syscall.UTF16PtrFromString(driveLetter+ "\\")
			dt,_,_:=getDriveType.Call(uintptr(unsafe.Pointer(drivePath)))
			if dt !=3{continue}
			var freeBytes,totalBytes,totalFreeBytes uint64
			r,_,_:=getDiskFreeSpaceEx.Call(
				uintptr(unsafe.Pointer(drivePath)),
				uintptr(unsafe.Pointer(&freeBytes)),
				uintptr(unsafe.Pointer(&totalBytes)),
				uintptr (unsafe.Pointer(&totalFreeBytes)),
			)
			if r!=0{
				totalGB:=float64(totalBytes)/(1024*1024*1024)
				freeGB:=float64(totalFreeBytes)/(1024*1024*1024)
				usedGB:=totalGB-freeGB
				percent:=(usedGB/totalGB)*100
				disks=append(disks,DiskInfo{
					Device:driveLetter,
					FSType:"NTFS",
					Total:totalGB,
					Used:usedGB,
                    Free :freeGB,
	                Percent:percent,
				})
			}
		}
	}
	return disks
	case "linux":
		data:=exec.Command("df","-kP")
	     out,err:= data.Output()
		 if err!=nil{
			return nil
		 }
		 lines:= strings.Split(string(out),"\n")
		 var disks []DiskInfo
		 for i:=1;i<len(lines);i++{
			line:=strings.TrimSpace(lines[i])
			if line ==""{
				continue
			}
			field:=strings.Fields(line)
			if len(field)<6{
				continue
			}
			totalKB,_:=strconv.ParseFloat(field[1],64)
			usedKB,_:=strconv.ParseFloat(field[2],64)
			freeKB,_:=strconv.ParseFloat(field[3],64)
			totalGB:=totalKB/(1024*1024)
			usedGB := usedKB / (1024 * 1024)
		    freeGB := freeKB / (1024 * 1024)
			percent:=(usedGB/totalGB)*100
			disks=append(disks,DiskInfo{
				Device:field[0],
				FSType:"LinuxFS",
				Total:totalGB,
				Used:usedGB,
				Free:freeGB,
				Percent:percent,
      
			})

		 }
		 return disks
 case "darwin":

	cmd := exec.Command("df", "-kP")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(out), "\n")
	var disks []DiskInfo

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		totalKB, _ := strconv.ParseFloat(fields[1], 64)
		usedKB, _ := strconv.ParseFloat(fields[2], 64)
		freeKB, _ := strconv.ParseFloat(fields[3], 64)

		totalGB := totalKB / (1024 * 1024)
		usedGB := usedKB / (1024 * 1024)
		freeGB := freeKB / (1024 * 1024)

		percent := (usedGB / totalGB) * 100

		disks = append(disks, DiskInfo{
			Device:  fields[0],
			FSType:  "APFS/Unix",
			Total:   totalGB,
			Used:    usedGB,
			Free:    freeGB,
			Percent: percent,
		})
	}

	return disks
}
return nil
}
func getHostInfo() (string, string) {
	platform := runtime.GOOS
	version := "unknown"
	switch platform {
case "linux":
		out, err := exec.Command("sh", "-c", ". /etc/os-release && echo $PRETTY_NAME").Output()
		if err == nil {
			version = strings.TrimSpace(string(out))
		}
	case "darwin":
		out, err := exec.Command("sw_vers", "-productVersion").Output()
		if err == nil {
			version = strings.TrimSpace(string(out))
		}
	case "windows":
		out, err := exec.Command("cmd", "/c", "ver").Output()
		if err == nil {
			version = strings.TrimSpace(string(out))
		}
	case "macOS":	
		out, err := exec.Command("sw_vers", "-productVersion").Output()
		if err == nil {
			version = strings.TrimSpace(string(out))
		}
	}

	return platform, version
}
func GetBoottimeInfo()(time.Time,error){
    
	switch plat{
	case "windows":
    kernal32:=syscall.NewLazyDLL("kernel32.dll")
	getTickCount64:=kernal32.NewProc("GetTickCount64")
	ret,_,_:=getTickCount64.Call()
	
	if ret==0{
		return time.Time{},errors.New("failed to get uptime")
	}
	uptimeMillis:=int64(ret)

	bootTime:=time.Now().Add(-time.Duration(uptimeMillis)*time.Millisecond)
	return bootTime,nil
	case "linux":
	out, err := exec.Command("sh", "-c", "uptime -s").Output()
	if err != nil {
		return time.Time{}, err
	}
	bootTime, err := time.Parse("2006-01-02 15:04:05", strings.TrimSpace(string(out)))
	if err != nil {
		return time.Time{}, err
	}
	return bootTime, nil
	case "darwin","macOS":
		out,err:=exec.Command("sysctl","-n","kern.boottime").Output()
		if err!=nil{
			return time.Time{},err
		}
		output:=string(out)
		start:=strings.Index(output,"sec =")
		if start==-1{
         return time.Time{},fmt.Errorf("failed to parse bootime")
		}
		start+=6
		end:=strings.Index(output[start:],",")
		secStr:=output[start:start+end]
		sec,_:=strconv.ParseInt(strings.TrimSpace(secStr),10,64)
		return time.Unix(sec,0),nil
	default:
		return time.Time{}, fmt.Errorf("unsupported platform: %s", plat)
	}
}
func TotalRamGb()(float64,error){	

	switch plat{
	case "windows":
	
	var mem memoryStatusEx
	mem.dwLength=uint32(unsafe.Sizeof(mem))
	kernel32:=syscall.NewLazyDLL("kernel32.dll")
	proc:=kernel32.NewProc("GlobalMemoryStatusEx")
	ret,_,err:=proc.Call(uintptr(unsafe.Pointer(&mem)))
	if ret==0{
		return 0,err
	}
	totalRAMgb:=float64(mem.ullTotalphys)/(1024*1024*1024)
	totalAvailRam:=float64(mem.ullAvailPhys)/(1024*1024*1024)
	fmt.Printf("Total RAM: %.2f GB, Available RAM: %.2f GB\n", totalRAMgb, totalAvailRam)
	return totalRAMgb,nil
case "linux":
	data,err:=os.ReadFile("/proc/meminfo")
	if err!=nil{
		return 0,err
	}
	lines:=strings.Split(string(data),"\n")
	for _,line:=range lines{
		if strings.HasPrefix(line,"MemTotal:"){
			fields:=strings.Fields(line)
			memKB,err:=strconv.ParseFloat(fields[1],64)
			if err!=nil{
				return 0,err
			}
			totalRAMgb:=memKB/(1024*1024)
			fmt.Printf("Total RAM %.2f GB\n",totalRAMgb)
			return totalRAMgb,nil
		}
	}
case "darwin","macOS":
	out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return 0, err
	}

	memBytes, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, err
	}

	totalRAMgb := memBytes / (1024 * 1024 * 1024)
	fmt.Printf("Total RAM: %.2f GB\n", totalRAMgb)
	return totalRAMgb, nil


	default:	
		return 0,fmt.Errorf("unsupported platform: %s", plat)
	}
	return 0,fmt.Errorf("os type not found")
}

func Manufacturer() (string,error){
  switch plat{
	case "windows":
	cmd:=exec.Command("wmic","cpu","get","Manufacturer")
	out,_:=cmd.Output()
	lines:=strings.Split(string(out),"\n")

	if len(lines)>1{
		return fmt.Sprintf("Manufacturer : %v",strings.TrimSpace(lines[1])),nil
	}
case "linux":
	data,err:=os.ReadFile("/proc/cpuinfo")
	if err!=nil{
		return "",err
	}
	lines:=strings.Split(string(data),"\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "vendor_id") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
					return fmt.Sprintf("Manufacturer: %s",
						strings.TrimSpace(parts[1])), nil
				}
			}
		}
		return "",fmt.Errorf("vendor_id not Found")
	case  "darwin":
			cmd := exec.Command("sysctl", "-n", "machdep.cpu.vendor")
		out, err := cmd.Output()
		if err!=nil{
			return "",err
		}
		return fmt.Sprintf("Manufacturer:%s",strings.TrimSpace(string(out))),nil

   default:
	 return "",fmt.Errorf("couldn't Parse CPU Manufacturer Output")
  }
  return "",fmt.Errorf("couldn't Parse CPU Manufacturer Output")
}
func Model()(string,error){
	switch runtime.GOOS{
	case "windows":
	cmd:=exec.Command("wmic","cpu","get","Name")
	out,err:=cmd.Output()	
	if err!=nil{
		return "",err
	}
	lines:=strings.Split(string(out),"\n")
	for _,line := range lines{
		line:=strings.TrimSpace(line)
	if line !="" && !strings.Contains(line,"Name"){
		
		return fmt.Sprintf("Model : %s",line),nil
	}}
	return "",fmt.Errorf("couldn't Parese the Cpu Model")
case "linux":
	data,err:=os.ReadFile("/proc/cpuinfo")
	if err!=nil{
		return "",fmt.Errorf("unable to linux cpumodel device %v",err)
	}
	lines:=strings.Split(string(data),"\n")
	for _,line := range lines{
	if strings.HasPrefix(line,"model name"){
    parts :=strings.Split(line,":")
    if len(parts)>1{
	return fmt.Sprintf("Model : %s",strings.TrimSpace(parts[1])),nil
}
}
}
 return "",fmt.Errorf("couldn't Parese the Cpu Model")
case "darwin":
	cmd:=exec.Command("sysctl","-n","machdep.cpu.brand_string")
	out ,err:=cmd.Output()
	if err!=nil{
		return "",fmt.Errorf("unable to get macos cpumodel:%v",err)

	}
	return fmt.Sprintf("Model :  %s",strings.TrimSpace(string(out))),nil

}
return "",fmt.Errorf("couldn't Parese the Cpu Model")

}

func CPUSPEEDMHz()(string,error){
	switch plat{
	case "windows":
	cmd,err:=exec.Command("wmic","cpu","get","MaxClockSpeed").Output()
	if err!=nil{
		return "",fmt.Errorf("there is problem in getting cpu speed in MHz")
	}
	lines:=strings.Split(string(cmd),"\n")
	if len(lines)>1{
		return fmt.Sprintf("CPU_SPEED_MHz : %s",strings.TrimSpace(lines[1])),nil
	}
	return "",fmt.Errorf("couldn't find the speed info")
case "linux":
	data,err:=os.ReadFile("/proc/cpuinfo")
	if err!=nil{
		return "",fmt.Errorf("unable to read cpu info : %v",err)
	}
	lines:=strings.Split(string(data),"\n")
	for _,line := range lines {
		if strings.HasPrefix(line,"cpu MHz"){
			parts:=strings.Split(line,":")
			if len(parts)>1{
				return fmt.Sprintf("CPU_SPEED_MHZ :%s",strings.TrimSpace(parts[1])),nil
			}
		}
	}
	return "", fmt.Errorf("couldn't find cpu speed info")
case "darwin":
	cmd:=exec.Command("sysctl","-n","hw.cpufrequency")
	out,err:=cmd.Output()
	if err!=nil{
		return "",fmt.Errorf("unable to get cpu frequency %v",err)
	}
	hzStr:=strings.TrimSpace(string(out))
	hz,err:=strconv.ParseUint(hzStr,10,64)
	if err !=nil {
		return "",fmt.Errorf("failed to parse cpu frequency")
	}
	mhz:=hz/1000000
	return fmt.Sprintf("CPU_SPEED_MHz : %d", mhz), nil
}
	return "",fmt.Errorf("couldn't find the cpu speed info")
}
func PHYSICAL_CORES()(string,error){
	switch plat{
	case "windows":
		cmd:=exec.Command("wmic","cpu","get","NumberOfCores")
		out,err:=cmd.Output()
        if err!=nil{
			return "",err
		}
		lines:=strings.Split(string(out),"\n")
		if len(lines)>1{
			return fmt.Sprintf("Physical cores:%v",lines[1]),nil
		}
	case "linux":
		data,err:=os.ReadFile("/proc/cpuinfo")
		if err!=nil{
			return "",fmt.Errorf("unable to read cpuinfo: %v",err)
		}
		lines:=strings.Split(string(data),"\n")
		var cores int
		for _,line := range lines{
			if strings.HasPrefix(line,"cpu cores"){
				parts:=strings.Split(line,":")
				cores,_=strconv.Atoi(strings.TrimSpace(parts[1]))
			}
		}
		return fmt.Sprintf("Physical cores:%d",cores),nil
	case "darwin":
		cmd:=exec.Command("sysctl","-n","hw.physicalcpu")
		out,err:=cmd.Output()
		if err!=nil{
			return "",fmt.Errorf("unable to get physical cores: %v",err)
		}
		return fmt.Sprintf("Physical cores:%s",strings.TrimSpace(string(out))),nil
	}
		return "",fmt.Errorf("Couldn't find the physical cores")
}
func IsHyperThreaded()(string,error){
	switch plat{
	case "windows":
      out,err:=exec.Command("wmic","cpu","get","NumberOfCores,NumberOfLogicalProcessors").Output()
	  if err!=nil{
		return "",err
	  }
	  lines:=strings.Split(string(out),"\n")
	  fields:=strings.Fields(lines[1])
	  if len(fields)<2{
		return "",fmt.Errorf("Couldn't Parse the core counts")

	  }
	  phy,_:=strconv.Atoi(fields[0])
	  log,_:=strconv.Atoi(fields[1])
	  if log>phy{
		return "Hyperthreading : enabled",nil 
	  }
	  return "Hyperthreading :not enabled",nil 	
	case "linux":
		data,err:=os.ReadFile("/proc/cpuinfo")
		if err!=nil{
			return "",err
		}
		lines:=strings.Split(string(data),"\n")
		var cores int 
		var siblings int
		for _,line := range lines{
			if strings.HasPrefix(line,"cpu cores"){
				parts:=strings.Split(line,":")
				cores,_=strconv.Atoi(strings.TrimSpace(parts[1]))
			}
			if strings.HasPrefix(line,"siblings"){
				parts:=strings.Split(line,":")
                siblings,_=strconv.Atoi(strings.TrimSpace(parts[1]))
			}
		}
		if siblings>cores  && cores!=0{
			return "Hyperthreading : enabled",nil
		}
		return "Hyperthreading :Not enabled",nil
	case "darwin":
		phyOut,err:=exec.Command("sysctl","-n","hw.physicalcpu").Output()
			if err!=nil{
				return "",err
			}
		logout,err:=exec.Command("sysctl", "-n", "hw.logicalcpu").Output()
		if err!=nil{
			return "",nil
		}
		phy,_:=strconv.Atoi(strings.TrimSpace(string(phyOut)))
		log,_:=strconv.Atoi(strings.TrimSpace(string(logout)))
		if log>phy{
			return "Hyperthreading : enabled",nil
		}
		return "Hyperthreading :not enabled",nil
	}
	  return "",fmt.Errorf("Hyperthreading : not enabled")
}
func GetCoreUsage()([]float64,float64,error){
	switch plat{
	case "windows":
	cmd:=exec.Command("typeperf",`\Processor(*)\% Processor Time`,"-sc","1")
	out,err:=cmd.Output()
	if err!=nil{
		return nil,0,err
	}
	lines:=strings.Split(string(out),"\n")
	if len(lines)<3{
		return nil,0,fmt.Errorf("unexpected output")
	}
	dataline:=lines[2]
	parts:=strings.Split(dataline,",")
	var usages []float64
	for i:=1;i<len(parts);i++{
		valstr:=strings.Trim(parts[i],"\"")
		val,_:=strconv.ParseFloat(valstr,64)
        usages=append(usages,val)
	}
	var sum float64
	for _, v := range usages {
		sum += v
	}
	var average float64 = sum / float64(len(usages))
	return usages,average,nil
case "linux":

	readStat := func() ([][]uint64, error) {
		data, err := os.ReadFile("/proc/stat")
		if err != nil {
			return nil, err
		}

		var stats [][]uint64
		lines := strings.Split(string(data), "\n")

		for _, line := range lines {
			if strings.HasPrefix(line, "cpu") && line[3] != ' ' {
				fields := strings.Fields(line)
				var values []uint64
				for i := 1; i < len(fields); i++ {
					val, _ := strconv.ParseUint(fields[i], 10, 64)
					values = append(values, val)
				}
				stats = append(stats, values)
			}
		}
		return stats, nil
	}

	stat1, err := readStat()
	if err != nil {
		return nil, 0, err
	}

	time.Sleep(200 * time.Millisecond)

	stat2, err := readStat()
	if err != nil {
		return nil, 0, err
	}

	var usages []float64
	var totalAvg float64

	for i := range stat1 {
		idle1 := stat1[i][3]
		total1 := sum(stat1[i])

		idle2 := stat2[i][3]
		total2 := sum(stat2[i])

		deltaTotal := float64(total2 - total1)
		deltaIdle := float64(idle2 - idle1)

		usage := (deltaTotal - deltaIdle) / deltaTotal * 100
		usages = append(usages, usage)
		totalAvg += usage
	}

	avg := totalAvg / float64(len(usages))
	return usages, avg, nil
case "darwin":
	cmd:=exec.Command("top","-l","2","-n","0")
	out,err:=cmd.Output()
	if err !=nil{
		return nil,0,err
	}
	var usage float64
	lines:=strings.Split(string(out),"\n")
	for _,line := range lines{
	if strings.Contains(line,"CPU usage"){
		fields:=strings.Fields(line)
		for i,f := range fields{
			if strings.Contains(f,"%")&& i<len(fields)-1{
				valstr:=strings.TrimSuffix(f,"%")
				val,_:=strconv.ParseFloat(valstr,64)
				if fields[i+1]=="user," || fields[i+1]=="sys,"{
                   usage +=val
				}
			}
		}
	}
}
    return []float64{usage},usage,nil
}
	return nil,0,fmt.Errorf("unsupported platform: %s", plat)
	
}
func sum(vals []uint64)uint64{
	var s uint64
	for _,v := range vals{
		s+=v
	}
	return s
}
func main(){
	   rand.Seed(time.Now().UnixNano())
		homeDir, err := os.UserHomeDir()
if err != nil {
    fmt.Println("Error getting home directory:", err)
    return
}

filePath := filepath.Join(homeDir,"agent.txt")
fmt.Println("Creating file at:", filePath)

f, err := os.Create(filePath)
if err != nil {
    fmt.Println("Error creating file:", err)
    return
}
defer f.Close()
	Agentid:=rand.Intn(1000)+1
	hostname,err:=os.Hostname()
	p,v:=getHostInfo()
	if err!=nil{
		fmt.Println("there is some error in getting hostname")
		return
	}
	
	fmt.Println("AgentID:", Agentid)
	fmt.Println("hostname:", hostname)
	fmt.Println("os",p,   v)
	fmt.Println("platform:",p)
	boottime,err:=GetBoottimeInfo()
    if err!=nil{
		fmt.Println("there is some error fetching windows bootime ",err)
		return
	}
	fmt.Println("boottime:", boottime)
	architecture:=runtime.GOARCH
	fmt.Println("architecture:", architecture)
	totalramgb,err:=TotalRamGb()
	if err!=nil{
		fmt.Println("Error getting total RAM:", err)
	}
	fmt.Println("totalRamGB:", totalramgb)
	manufacturer,err:=Manufacturer()
	if err!=nil{
		fmt.Println("error",err)
		return 
	}
	fmt.Println(manufacturer)
	ModelName,err:=Model()
	if err!=nil{
		fmt.Println("error :",err)
	}
	fmt.Println(ModelName)
	CPUSpeed,err:=CPUSPEEDMHz()
	if err!=nil{
		fmt.Println("error",err)
		return
	}
	fmt.Println(CPUSpeed)
	Physical_cores,err:=PHYSICAL_CORES()
	if err!=nil{
		fmt.Println("error : ",err)
		return
	}
	fmt.Println(Physical_cores)
	fmt.Println("Logical cores :",runtime.NumCPU())
	Hyperthreading,err:=IsHyperThreaded()	
	if err!=nil{
		fmt.Println("error : ",err)
		return
	}
	fmt.Println(Hyperthreading) 
	UsagePerCore,Toatlcoreusage,err:=GetCoreUsage()
	if err!=nil{
		fmt.Println("error : ",err)
		return
	}
	fmt.Println("Usage Per Core",UsagePerCore)
	fmt.Println("Total Core Usage:", Toatlcoreusage)
	fmt.Println("Disk  Info:")
	DiskInfo:=getLogicalDrives()
	for _,val := range DiskInfo{
		fmt.Printf("%+v\n",val)
	}


systemdetails(f, fmt.Sprintf("AgentID: %d", Agentid))
systemdetails(f, fmt.Sprintf("Hostname: %s", hostname))
systemdetails(f, fmt.Sprintf("OS: %s %s", p, v))
systemdetails(f, fmt.Sprintf("Platform: %s", p))
systemdetails(f, fmt.Sprintf("Boot time: %s", boottime))
systemdetails(f, fmt.Sprintf("Architecture: %s", architecture))
systemdetails(f, fmt.Sprintf("Total RAM GB: %.2f", totalramgb))
systemdetails(f, manufacturer)
systemdetails(f, ModelName)
systemdetails(f, CPUSpeed)
systemdetails(f, Physical_cores)
systemdetails(f, Hyperthreading)
systemdetails(f, fmt.Sprintf("Usage Per Core: %v", UsagePerCore))
systemdetails(f, fmt.Sprintf("Total Core Usage: %.2f", Toatlcoreusage))
systemdetails(f, "Disk Info:")
for _,va := range DiskInfo{
	systemdetails(f,fmt.Sprintf("%+v",va))
}
}
func systemdetails(f *os.File,s string){
	f.WriteString(s+"\n")
}