package main

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
)

func runTerminalSub(onLineConfigs []map[string]string) int {
	inputKey := ""
	var hostNum = len(onLineConfigs)
	var ipInfoList []IpInfo
	listStr := ""
	for i := 0; i < hostNum; i++ {
		item := onLineConfigs[i]
		port, _ := strconv.Atoi(item["port"])
		ipInfo := IpInfo{ip: item["ip"], user: item["user"], authMethod: item["authMethod"], authToken: item["authToken"], desc: item["desc"], port: port}
		ipInfoList = append(ipInfoList, ipInfo)
		if i < hostNum-1 {
			listStr += fmt.Sprintf("[%d]\t%s\t%s\t\t\n", i+1, ipInfo.ip, ipInfo.desc)
		} else {
			listStr += fmt.Sprintf("[%d]\t%s\t%s\t\t", i+1, ipInfo.ip, ipInfo.desc)
		}
	}
	//fmt.Println("[*]\tip\tdesc")
	for inputKey == "" {
		fmt.Println(listStr)
		fmt.Scanln(&inputKey)
		//fmt.Println("need input number")
		//fmt.Scanln(&inputKey)
		//fmt.Println(fmt.Sprintf("inputKey=[%s]", inputKey))
	}
	if inputKey == "exit" || inputKey == "q" {
		return 1
	}
	index, err := strconv.Atoi(inputKey)
	if err != nil {
		fmt.Println(listStr)
		println("todo parse inputKey ")
		return 0
	}
	if index > 0 && index <= hostNum {
		ipinfo := ipInfoList[index-1]
		fmt.Printf("connecting to[%s][%s]......\n", ipinfo.ip,ipinfo.desc)
		err = ipinfo.login()
		if err != nil {
			fmt.Println("login host error:", err)
		}
		fmt.Println("logout done")
		//_, err = os.Stdin.Write([]byte(""))
		//checkErr(err)
		//
		_, err = os.Stdin.Read([]byte(""))
		checkErr(err)
	} else {
		println("todo match inputKey")
		return 1
	}
	return 0
}

func runTerminal(jsonMap map[string][]map[string]string, indexIn int) int {
	keys := reflect.ValueOf(jsonMap).MapKeys()
	var env string
	fmt.Println("please input env : ", keys)
	if len(keys) == 1 && indexIn == 1 {
		env = keys[0].String()
		fmt.Println("env only one use:" + env)
	} else {
		fmt.Scanln(&env)
	}
	if env == "exit" || env == "q" {
		return 1
	}
	onLineConfigs := jsonMap[env]
	if onLineConfigs == nil {
		fmt.Println("env is error")
		return 0
	}
	var subTerminalStatus = 0
	for subTerminalStatus == 0 {
		subTerminalStatus = runTerminalSub(onLineConfigs)
		if subTerminalStatus != 0 {
			break
		}
	}
	return 0

}
func main() {
	jsonMap := loadConfig()
	if jsonMap == nil {
		os.Exit(1)
	}
	exitFlag := 0
	indexIn := 0
	for exitFlag == 0 {
		indexIn += 1
		exitFlag = runTerminal(jsonMap, indexIn)
		//if exitFlag == 1 {
		//	os.Exit(0)
		//}
	}
}
