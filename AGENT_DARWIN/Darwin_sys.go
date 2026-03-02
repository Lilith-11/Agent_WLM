package main

import (
"fmt"
"os"
"path/filepath"
"runtime"
"time"

"math/rand"
"strconv"
"os/exec"
"strings")
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
	AgentID string `json:"agent_id"`
	Hostname string `json:"hostname"`
	Os string `json:"os"`
	Platform string `json:"platform"`
	Architecture string `json:"architecture"`
	Boottime time.Time `json:"boottime"`
	TotalRamGB uint64 `json:"total_ram_gb"`
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
 func getLogicalDrives()[]DiskInfo{
      if plat=="darwin"{
		cmd:=exec.Command("df","-kP")
		out,err:=cmd.Output()
		if err!=nil{
			return nil
		}
		lines:=strings.Split(string(out),"\n")
		var disks []DiskInfo
		for i:=1;i<len(lines);i++{
			line:=strings.TrimSpace(lines[i])
			if line==""{
				continue
			}
          fields:=strings.Fields(line)
		  if len(fields)<6{
			continue
		  }
		  totalKB,_:=strconv.ParseFloat(fields[1],64)
		  usedKB,_:=strconv.ParseFloat(fields[2],64)

		  freeKB,_:=strconv.ParseFloat(fields[3],64)
          totalGB:=totalKB/(1024*1024)
		  usedGB :=usedKB/(1024*1024)
		  freeGB:=freeKB/(1024*1024)
		  percent:=(usedGB/totalGB)*100
		  disks=append(disks,DiskInfo{
			Device: fields[0],
			FSType:"APFS/UNIX",
			Total:totalGB,
			Used:usedGB,
			Free:freeGB,
			Percent:percent,
		  })
		}
		return disks
	  }
	  return nil
 }
 func getHostInfo() (string,string){
	version:="unknown"
	if plat=="darwin"{
		out,err:=exec.Command("sw_vers","-productVersion").Output()
		if err==nil{
			version=strings.TrimSpace(string(out))
		}
	}
	return plat,version 
 }
 func GetBootTimeInfo()(time.Time,error){
      	if plat=="darwin"{
			out,err:=exec.Command("sysctl","-n","kern.boottime").Output()
			if err!=nil{
				return time.Time{},err
			}
			output:=string(out)
			start:=strings.Index(output,"sec =")
			if start ==-1{
				return time.Time{},fmt.Errorf("failed to parse boot time")
			}
			start+=6
			end:=strings.Index(output[start:],",")
			secStr:=output[start:start+end]
			sec,_:=strconv.ParseInt(strings.TrimSpace(secStr),10,64)
			return time.Unix(sec,0),nil
		}
		return time.Time{},fmt.Errorf("unsupported platfom %v",plat)
 }
 func TotalRamGb()(float64,error){
	if plat=="darwin"{
	 out,err:=exec.Command("sysctl","-n","hw.memsize").Output()
	 if err!=nil{
		return 0,fmt.Errorf("unable to get darwin totalramgb")
	 }
	 membytes, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	 if err!=nil{
		return 0,fmt.Errorf("unable to convet totalramgb output into foat64")
	 }
	 totalRAMgb:=membytes/(1024*1024*1024)
	 return totalRAMgb,nil}
	 return 0,fmt.Errorf("unsupported platform:%s",plat)
 }
func Manufacturer() (string, error) {
	if plat == "darwin" {
		cmd := exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
		out, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("couldn't get manufacturer: %v", err)
		}
		return fmt.Sprintf("Manufacturer: %s",
			strings.TrimSpace(string(out))), nil
	}
	return "", fmt.Errorf("couldn't parse CPU manufacturer output")
}
 func Model()(string,error){
   out,err:=exec.Command("sysctl","-n","machdep.cpu.brand_string").Output()
   if err!=nil{
	return "",fmt.Errorf("unable to get a Model")
   }
   return fmt.Sprintf("model:%v",strings.TrimSpace(string(out))),nil
 }
 func CPUSPEEDMHz()(string,error){
	if plat=="darwin"{
       cmd:=exec.Command("sysctl","-n","hw.cpufrequency")
	   out,err:=cmd.Output()
	   if err!=nil{
          return "",fmt.Errorf("problem to get an information of cpusepped in MHz")
	   }
	   hzStr:=strings.TrimSpace(string(out))
	   hz,err:=strconv.ParseUint(hzStr,10,64)
	   if err!=nil{
		return "",fmt.Errorf("failed to parse cpu frequency")
	   }
	   mhz:=hz/1000000
       return fmt.Sprintf("CPU_SPEED_MHz : %d",mhz),nil
	}
	return "",fmt.Errorf("couldn't find the cpu speed info")
 }
 func PHYSICAL_CORES()(string,error){
	cmd:=exec.Command("sysctl","-n","hw.physicalcpu")
	out,err:=cmd.Output()
	if err!=nil{
      return "",fmt.Errorf("unable to find the physical cores in mac")
	}
	return fmt.Sprintf("Physical cores :%s",strings.TrimSpace(string(out))),nil
 }
 func IsHyperThreaded()(string,error){
	Phyout, err := exec.Command("sysctl", "-n", "hw.physicalcpu").Output()
	if err!=nil{
		return "",fmt.Errorf("unable to find physical cpu in hyperthreading")
	}
	Logout,err:=exec.Command("sysctl","-n","hw.logicalcpu").Output()
	if err!=nil{
		return "",fmt.Errorf("unable to find the logical cores")
	}
	phy,_:=strconv.Atoi(strings.TrimSpace(string(Phyout)))
    log,_:=strconv.Atoi(strings.TrimSpace(string(Logout)))
	if log>phy{
		return "Hyperthreading: Enabled",nil
	}
	return "Hyperthreading: Not Enabled",nil
 }
 func GetCoreUsage()([]float64,float64,error){
     if plat=="darwin"{
		cmd:=exec.Command("top","-l","2","-n","0")
        out,err:=cmd.Output()
		if err!=nil{
			return nil,0,err
		}
		lines:=strings.Split(string(out),"\n")
      var usage float64
      for _,line:= range lines{
	
		if strings.Contains(line,"Cpu usage:"){
            	 fields:=strings.Fields(line)
		 for i,f := range fields{
			if strings.Contains(f,"%")&&i<len(fields)-1{
			 valstr:=strings.TrimSuffix(f,"%")
			val,_:=strconv.ParseFloat(valstr,64)
			if fields[i+1]=="user,"|| fields[i+1]=="sys,"{
				usage+=val
			}
		}
		}
	  }  
	 }
	return []float64{usage},usage,nil }
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
	if plat!="darwin"{
		fmt.Println("This program is designed to run on macOS (Darwin). Exiting.")
		return
	}
	rand.Seed(time.Now().UnixNano())
	homeDir,err:=os.UserHomeDir()
	if err!=nil{
		fmt.Println("error in getting home directory:", err)
	}
	filePath:=filepath.Join(homeDir,"agent.txt")
	f,err:=os.Create(filePath)
	if err!=nil{
		fmt.Println("unable to create file:", err)
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
	boottime,err:=GetBootTimeInfo()
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