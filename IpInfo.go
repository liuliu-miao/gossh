package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	terminal "golang.org/x/crypto/ssh/terminal"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"os/user"
	"runtime"
	"syscall"
	//"golang.org/x/crypto/ssh/terminal"
)

type IpInfo struct {
	ip         string
	user       string
	authMethod string
	authToken  string
	port       int
	desc       string

	auth []ssh.AuthMethod
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
func (info *IpInfo) genAuth() {
	info.auth = make([]ssh.AuthMethod, 0)
	if info.authMethod == "password" {
		authNew := ssh.Password(info.authToken)
		info.auth = append(info.auth, authNew)
	}
	if info.authMethod == "pem" {
		priKey, _ := ioutil.ReadFile(info.authToken)
		authNew, _ := ssh.ParsePrivateKey(priKey)
		info.auth = append(info.auth, ssh.PublicKeys(authNew))
	}

}
func refreshWindowSize(session *ssh.Session) {
	go func() {
		// 监听窗口变更事件
		sigwinchCh := make(chan os.Signal, 1)
		signal.Notify(sigwinchCh, syscall.SIGINT)

		fd := int(os.Stdin.Fd())
		termWidth, termHeight, err := terminal.GetSize(fd)
		if err != nil {
			fmt.Println(err)
		}

		for {
			select {
			// 阻塞读取
			case sigwinch := <-sigwinchCh:
				if sigwinch == nil {
					return
				}
				currTermWidth, currTermHeight, err := terminal.GetSize(fd)

				// 判断一下窗口尺寸是否有改变
				if currTermHeight == termHeight && currTermWidth == termWidth {
					continue
				}
				// 更新远端大小
				fmt.Println("window-size change")
				session.WindowChange(currTermHeight, currTermWidth)
				if err != nil {
					fmt.Printf("Unable to send window-change reqest: %s.", err)
					continue
				}
				termWidth, termHeight = currTermWidth, currTermHeight

			}
		}
	}()
}
func createSession(client *ssh.Client) error {
	session, err := client.NewSession()
	checkErr(err)

	defer session.Close()

	//当ssh连接建立过后, 我们就可以通过这个连接建立一个回话, 在回话上和远程主机通信。
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	//stdin, err := session.StdinPipe()
	//checkErr(err)
	//stdout, outerr := session.StdoutPipe()
	//checkErr(err)
	//stderr, errerr := session.StderrPipe()
	//checkErr(err)
	//if outerr == io.EOF || errerr == io.EOF {
	//	fmt.Println("EOF-ERROR")
	//}
	//go io.Copy(os.Stderr, stderr)
	//go io.Copy(os.Stdout, stdout)

	//go func() {
	//	buf := make([]byte, 128)
	//	for {
	//		n, err := os.Stdin.Read(buf)
	//		checkErr(err)
	//		if err == io.EOF {
	//			session.Close()
	//		}
	//		if n > 0 {
	//			tempStr := string(buf)
	//			tempStr = strings.ReplaceAll(tempStr,"EOF","")
	//			tempStr = strings.ReplaceAll(tempStr,"\n\n","\n")
	//			_, err = stdin.Write([]byte(tempStr))
	//			fmt.Printf("inputStr=[%s]", tempStr)
	//			if err == io.EOF {
	//				session.Close()
	//			}
	//			checkErr(err)
	//		}
	//	}
	//}()

	//modes := ssh.TerminalModes{
	//	ssh.ECHO:          1,
	//	ssh.ECHOCTL:       0,
	//	ssh.TTY_OP_ISPEED: 14400,
	//	ssh.TTY_OP_OSPEED: 14400,
	//}
	modes := ssh.TerminalModes{}
	termFd := int(os.Stdin.Fd())
	w, h, _ := terminal.GetSize(termFd)
	termState, _ := terminal.MakeRaw(termFd)
	// TODO 主动从服务器exit,会导致空指针.
	defer terminal.Restore(termFd, termState)

	systemType := runtime.GOOS
	termType := os.Getenv("TERM")
	if termType == "" {
		termType = "xterm-256color"
	}
	if systemType == "windows" {
		termType = "xterm"
	}
	err = session.RequestPty(termType, h, w, modes)
	if err != nil {
		log.Fatalln(err)
	}
	err = session.Shell()

	refreshWindowSize(session)

	if err != nil {
		log.Fatalln(err)
	}
	err = session.Wait()
	if err != nil {
		log.Fatalln(err)
	}
	return nil

}
func (info *IpInfo) login() error {
	info.genAuth()
	var address string = fmt.Sprintf("%s:%d", info.ip, info.port)
	client, err := ssh.Dial("tcp", address, &ssh.ClientConfig{
		User:            info.user,
		Auth:            info.auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	//defer client.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return createSession(client)
}

func loadConfig() map[string][]map[string]string {
	currentUser, err := user.Current()

	if err != nil {
		fmt.Println(err)
	}
	homeDir := currentUser.HomeDir
	configPath := homeDir + "/config/"

	os.MkdirAll(configPath, fs.ModeDir|fs.ModePerm)
	configPath = configPath + "remote_host.json"
	_, err = os.Stat(configPath)
	if err != nil {
		fmt.Println("file no found create default template")
		content := "{\n  \"test\": [\n    {\n      \"ip\": \"10.0.3.151\",\n      \"user\": \"root\",\n      \"port\": \"22\",\n      \"authMethod\": \"password\",\n      \"authToken\": \"1q2w3e4r5t\"\n    }\n  ]\n}"
		ioutil.WriteFile(configPath, []byte(content), fs.ModePerm)
	}
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Println(err)
		fmt.Println("config no found,please create it ,name=remote_host.json")
		return nil
	}

	var v map[string][]map[string]string
	err2 := json.Unmarshal(data, &v)
	if err2 != nil {
		fmt.Println(err2)
		return nil
	}

	return v
}
