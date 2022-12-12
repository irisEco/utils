package util

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

//以拥有 有效默认网关 的ip作为有效ip
//for all
func GetAdapterList() (*syscall.IpAdapterInfo, error) {
	b := make([]byte, 1000)
	l := uint32(len(b))
	a := (*syscall.IpAdapterInfo)(unsafe.Pointer(&b[0]))
	err := syscall.GetAdaptersInfo(a, &l)
	if err == syscall.ERROR_BUFFER_OVERFLOW {
		b = make([]byte, l)
		a = (*syscall.IpAdapterInfo)(unsafe.Pointer(&b[0]))
		err = syscall.GetAdaptersInfo(a, &l)
	}
	if err != nil {
		return nil, os.NewSyscallError("GetAdaptersInfo", err)
	}
	return a, nil
}

func LocalIp() (local_ip string, err error) {
	aList, err := GetAdapterList()
	if err != nil {
		return "", err
	}
	for ai := aList; ai != nil; ai = ai.Next {
		ipl := &ai.IpAddressList
		gwl := &ai.GatewayList
		for ; ipl != nil; ipl = ipl.Next {
			ip := ipl.IpAddress.String
			gw := gwl.IpAddress.String
			ip_str := string(ip[:])
			str := string(gw[:])
			ip_zero := strings.Index(ip_str, "0")
			gw_zero := strings.Index(str, "0")
			//过滤掉ip和默认网关以0开头的数据
			if ip_zero != 0 && gw_zero != 0 {
				local_ip = ip_str
				break
			}
		}
	}

	return local_ip, err
}

//for windows
//调用外部cmd，获取commend的输入，转为切片，按行匹配
func LocalIpWindows() string {
	var finalIp string
	cmd := exec.Command("cmd", "/c", "ipconfig")
	if out, err := cmd.StdoutPipe(); err != nil {
		fmt.Println(err)
	} else {
		defer out.Close()
		if err := cmd.Start(); err != nil {
			fmt.Println(err)
		}

		if opBytes, err := ioutil.ReadAll(out); err != nil {
			log.Fatal(err)
		} else {
			str := string(opBytes)
			var strs = strings.Split(str, "\r\n")
			if 0 != len(strs) {
				var havingFinalIp4 bool = false
				var cnt int = 0
				for index, value := range strs {
					vidx := strings.Index(value, "IPv4")
					if vidx != -1 && vidx != 0 {
						ip4lines := strings.Split(value, ":")
						if len(ip4lines) == 2 {
							cnt = index
							havingFinalIp4 = true
							ip4str := ip4lines[1]
							finalIp = strings.TrimSpace(ip4str)
						}

					}
					if havingFinalIp4 && index == cnt+2 {
						lindex := strings.Index(value, ":")
						if lindex != -1 {
							lines := strings.Split(value, ":")
							fip := lines[1]
							if strings.TrimSpace(fip) != "" {
								break
							}
						}
						havingFinalIp4 = false
						finalIp = ""
					}
				}
			}
		}
	}
	return finalIp
}
