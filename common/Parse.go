package common

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func Parse(Info *HostInfo) {
	ParseUser()
	ParsePass(Info)
	ParseInput(Info)
	ParseScantype(Info)
}

func ParseUser() {
	if Username == "" && Userfile == "" {
		return
	}
	var Usernames []string
	if Username != "" {
		Usernames = strings.Split(Username, ",")
	}

	if Userfile != "" {
		users, err := Readfile(Userfile)
		if err == nil {
			for _, user := range users {
				if user != "" {
					Usernames = append(Usernames, user)
				}
			}
		}
	}

	Usernames = RemoveDuplicate(Usernames)
	for name := range Userdict {
		Userdict[name] = Usernames
	}
}

func ParsePass(Info *HostInfo) {
	var PwdList []string
	if Password != "" {
		passs := strings.Split(Password, ",")
		for _, pass := range passs {
			if pass != "" {
				PwdList = append(PwdList, pass)
			}
		}
		Passwords = PwdList
	}
	if Passfile != "" {
		passs, err := Readfile(Passfile)
		if err == nil {
			for _, pass := range passs {
				if pass != "" {
					PwdList = append(PwdList, pass)
				}
			}
			Passwords = PwdList
		}
	}
	if URL != "" {
		urls := strings.Split(URL, ",")
		TmpUrls := make(map[string]struct{})
		for _, url := range urls {
			if _, ok := TmpUrls[url]; !ok {
				TmpUrls[url] = struct{}{}
				if url != "" {
					Urls = append(Urls, url)
				}
			}
		}
	}
	if UrlFile != "" {
		urls, err := Readfile(UrlFile)
		if err == nil {
			TmpUrls := make(map[string]struct{})
			for _, url := range urls {
				if _, ok := TmpUrls[url]; !ok {
					TmpUrls[url] = struct{}{}
					if url != "" {
						Urls = append(Urls, url)
					}
				}
			}
		}
	}
	if PortFile != "" {
		ports, err := Readfile(PortFile)
		if err == nil {
			newport := ""
			for _, port := range ports {
				if port != "" {
					newport += port + ","
				}
			}
			Info.Ports = newport
		}
	}
}

func Readfile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Open %s error, %v\n", filename, err)
		os.Exit(0)
	}
	defer file.Close()
	var content []string
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text != "" {
			content = append(content, scanner.Text())
		}
	}
	return content, nil
}

func ParseInput(Info *HostInfo) {
	if Info.Host == "" && HostFile == "" && URL == "" && UrlFile == "" {
		fmt.Println("Host is none")
		flag.Usage()
		os.Exit(0)
	}

	if BruteThread <= 0 {
		BruteThread = 1
	}
	if TmpOutputfile != "" {
		if !strings.Contains(TmpOutputfile, "/") && !strings.Contains(TmpOutputfile, `\`) {
			Outputfile = getpath() + TmpOutputfile
		} else {
			Outputfile = TmpOutputfile
		}
	}
	if TmpSave == true {
		IsSave = false
	}

	if Info.Ports == DefaultPorts {
		Info.Ports += "," + Webport
	}

	if PortAdd != "" {
		if strings.HasSuffix(Info.Ports, ",") {
			Info.Ports += PortAdd
		} else {
			Info.Ports += "," + PortAdd
		}
	}

	if UserAdd != "" {
		user := strings.Split(UserAdd, ",")
		for a := range Userdict {
			Userdict[a] = append(Userdict[a], user...)
			Userdict[a] = RemoveDuplicate(Userdict[a])
		}
	}

	if PassAdd != "" {
		pass := strings.Split(PassAdd, ",")
		Passwords = append(Passwords, pass...)
		Passwords = RemoveDuplicate(Passwords)
	}
	if Socks5Proxy != "" && !strings.HasPrefix(Socks5Proxy, "socks5://") {
		if !strings.Contains(Socks5Proxy, ":") {
			Socks5Proxy = "socks5://127.0.0.1" + Socks5Proxy
		} else {
			Socks5Proxy = "socks5://" + Socks5Proxy
		}
	}
	if Socks5Proxy != "" {
		fmt.Println("Socks5Proxy:", Socks5Proxy)
		_, err := url.Parse(Socks5Proxy)
		if err != nil {
			fmt.Println("Socks5Proxy parse error:", err)
			os.Exit(0)
		}
		NoPing = true
	}
	if Proxy != "" {
		if Proxy == "1" {
			Proxy = "http://127.0.0.1:8080"
		} else if Proxy == "2" {
			Proxy = "socks5://127.0.0.1:1080"
		} else if !strings.Contains(Proxy, "://") {
			Proxy = "http://127.0.0.1:" + Proxy
		}
		fmt.Println("Proxy:", Proxy)
		if !strings.HasPrefix(Proxy, "socks") && !strings.HasPrefix(Proxy, "http") {
			fmt.Println("no support this proxy")
			os.Exit(0)
		}
		_, err := url.Parse(Proxy)
		if err != nil {
			fmt.Println("Proxy parse error:", err)
			os.Exit(0)
		}
	}

	if Hash != "" && len(Hash) != 32 {
		fmt.Println("[-] Hash is error,len(hash) must be 32")
		os.Exit(0)
	} else {
		var err error
		HashBytes, err = hex.DecodeString(Hash)
		if err != nil {
			fmt.Println("[-] Hash is error,hex decode error")
			os.Exit(0)
		}
	}
}

func ParseScantype(Info *HostInfo) {
	_, ok := PORTList[Scantype]
	if !ok {
		showmode()
	}
	if Scantype != "all" && Info.Ports == DefaultPorts+","+Webport {
		switch Scantype {
		case "wmiexec":
			Info.Ports = "135"
		case "wmiinfo":
			Info.Ports = "135"
		case "smbinfo":
			Info.Ports = "445"
		case "hostname":
			Info.Ports = "135,137,139,445"
		case "smb2":
			Info.Ports = "445"
		case "web":
			Info.Ports = Webport
		case "webonly":
			Info.Ports = Webport
		case "ms17010":
			Info.Ports = "445"
		case "cve20200796":
			Info.Ports = "445"
		case "portscan":
			Info.Ports = DefaultPorts + "," + Webport
		case "main":
			Info.Ports = DefaultPorts
		default:
			port, _ := PORTList[Scantype]
			Info.Ports = strconv.Itoa(port)
		}
		fmt.Println("-m ", Scantype, " start scan the port:", Info.Ports)
	}
}

func CheckErr(text string, err error, flag bool) {
	if err != nil {
		fmt.Println("Parse", text, "error: ", err.Error())
		if flag {
			if err != ParseIPErr {
				fmt.Println(ParseIPErr)
			}
			os.Exit(0)
		}
	}
}

func getpath() string {
	file, _ := exec.LookPath(os.Args[0])
	path1, _ := filepath.Abs(file)
	filename := filepath.Dir(path1)
	var path string
	if strings.Contains(filename, "/") {
		tmp := strings.Split(filename, `/`)
		path = strings.Join(tmp, `/`)
		path += `/`
	} else if strings.Contains(filename, `\`) {
		tmp := strings.Split(filename, `\`)
		path = strings.Join(tmp, `\`)
		path += `\`
	}
	return path
}

func showmode() {
	fmt.Println("The specified scan type does not exist")
	fmt.Println("-m")
	for name := range PORTList {
		fmt.Println("   [" + name + "]")
	}
	os.Exit(0)
}
