package main

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
)

var (
	chnIPs []TheIP
)

type Item struct {
	Platform string
	URL      string
	FileName string
}

func homePage(c *gin.Context) {
	gateway := c.Param("gateway")
	if gateway == "" {
		gateway = "auto"
	}
	platform := c.Param("platform")

	var items []Item
	switch platform {
	default:
		items = []Item{
			Item{Platform: "Linux", URL: "/f/" + gateway + "/linux/routes-up.sh", FileName: "routes-up.sh"},
			Item{Platform: "Linux", URL: "/f/" + gateway + "/linux/routes-down.sh", FileName: "routes-down.sh"},
			Item{Platform: "Linux", URL: "/f/" + gateway + "/linux/package.zip", FileName: "package.zip"},
			Item{Platform: "Android", URL: "/f/" + gateway + "/android/routes-up.sh", FileName: "routes-up.sh"},
			Item{Platform: "Android", URL: "/f/" + gateway + "/android/routes-down.sh", FileName: "routes-down.sh"},
			Item{Platform: "Android", URL: "/f/" + gateway + "/android/package.zip", FileName: "package.zip"},
			Item{Platform: "macOS", URL: "/f/" + gateway + "/mac/routes-up.sh", FileName: "routes-up.sh"},
			Item{Platform: "macOS", URL: "/f/" + gateway + "/mac/routes-down.sh", FileName: "routes-down.sh"},
			Item{Platform: "macOS", URL: "/f/" + gateway + "/mac/package.zip", FileName: "package.zip"},
			Item{Platform: "ChinaDNS", URL: "/f/" + gateway + "/chinadns/chnroute.txt", FileName: "chnroute.txt"},
			Item{Platform: "ChinaDNS", URL: "/f/" + gateway + "/chinadns/package.zip", FileName: "package.zip"},
			Item{Platform: "RouterOS", URL: "/f/" + gateway + "/routeros/routeros-address-list.rsc", FileName: "routeros-address-list.rsc"},
			Item{Platform: "RouterOS", URL: "/f/" + gateway + "/routeros/routeros.rsc", FileName: "routeros.rsc"},
			Item{Platform: "RouterOS", URL: "/f/" + gateway + "/routeros/package.zip", FileName: "package.zip"},
			Item{Platform: "Windows", URL: "/f/" + gateway + "/windows/cmroute.dll", FileName: "cmroute.dll"},
			Item{Platform: "Windows", URL: "/f/" + gateway + "/windows/routes-up.bat", FileName: "routes-up.bat"},
			Item{Platform: "Windows", URL: "/f/" + gateway + "/windows/routes-up.txt", FileName: "routes-up.txt"},
			Item{Platform: "Windows", URL: "/f/" + gateway + "/windows/routes-down.bat", FileName: "routes-down.bat"},
			Item{Platform: "Windows", URL: "/f/" + gateway + "/windows/routes-down.txt", FileName: "routes-down.txt"},
			Item{Platform: "Windows", URL: "/f/" + gateway + "/windows/package.zip", FileName: "package.zip"},
		}
		platform = "all"
	case "routeros":
		items = []Item{
			Item{Platform: "RouterOS", URL: "/f/" + gateway + "/routeros/routeros-address-list.rsc", FileName: "routeros-address-list.rsc"},
			Item{Platform: "RouterOS", URL: "/f/" + gateway + "/routeros/routeros.rsc", FileName: "routeros.rsc"},
			Item{Platform: "RouterOS", URL: "/f/" + gateway + "/routeros/package.zip", FileName: "package.zip"},
		}
	case "chinadns":
		items = []Item{
			Item{Platform: "ChinaDNS", URL: "/f/" + gateway + "/chinadns/chnroute.txt", FileName: "chnroute.txt"},
			Item{Platform: "ChinaDNS", URL: "/f/" + gateway + "/chinadns/package.zip", FileName: "package.zip"},
		}
	case "mac", "linux", "android":
		platformMap := map[string]string{
			"linux":   "Linux",
			"mac":     "macOS",
			"android": "Android",
		}
		items = []Item{
			Item{Platform: platformMap[platform], URL: "/f/" + gateway + "/" + platform + "/routes-up.sh", FileName: "routes-up.sh"},
			Item{Platform: platformMap[platform], URL: "/f/" + gateway + "/" + platform + "/routes-down.sh", FileName: "routes-down.sh"},
			Item{Platform: platformMap[platform], URL: "/f/" + gateway + "/" + platform + "/package.zip", FileName: "package.zip"},
		}
	case "windows":
		items = []Item{
			Item{Platform: "Windows", URL: "/f/" + gateway + "/windows/cmroute.dll", FileName: "cmroute.dll"},
			Item{Platform: "Windows", URL: "/f/" + gateway + "/windows/routes-up.bat", FileName: "routes-up.bat"},
			Item{Platform: "Windows", URL: "/f/" + gateway + "/windows/routes-up.txt", FileName: "routes-up.txt"},
			Item{Platform: "Windows", URL: "/f/" + gateway + "/windows/routes-down.bat", FileName: "routes-down.bat"},
			Item{Platform: "Windows", URL: "/f/" + gateway + "/windows/routes-down.txt", FileName: "routes-down.txt"},
			Item{Platform: "Windows", URL: "/f/" + gateway + "/windows/package.zip", FileName: "package.zip"},
		}
	}
	if gateway == "auto" {
		gateway = ""
	}
	c.HTML(http.StatusOK, "index.tpl", gin.H{
		"items":    items,
		"gateway":  gateway,
		"platform": platform,
	})
}

func getFile(c *gin.Context) {
	gateway := c.Param("gateway")
	if gateway == "" {
		c.String(http.StatusBadRequest, "bad gateway request")
		return
	}
	platform := c.Param("platform")
	if platform == "" {
		c.String(http.StatusBadRequest, "bad platform request")
		return
	}
	file := c.Param("file")
	if file == "" {
		c.String(http.StatusBadRequest, "bad file request")
		return
	}

	path := filepath.Join("templates", platform, file)
	ext := strings.ToLower(filepath.Ext(file))
	if ext == ".dll" {
		c.File(path)
		return
	}

	if ext == ".txt" || ext == ".rsc" || ext == ".sh" || ext == ".bat" {
		if gateway == "auto" {
			gateway = "$gateway"
		}
		data := Generate(path, chnIPs, gateway)
		c.Data(http.StatusOK, "text/plain", data)
		return
	}
	if file == "package.zip" {
		var data []byte
		switch platform {
		case "windows":
			data = packWin(gateway)
		case "android":
			data = packAndroid(gateway)
		case "linux":
			data = packLinux(gateway)
		case "mac":
			data = packMac(gateway)
		case "chinadns":
			data = packChinaDNS(gateway)
		case "routeros":
			data = packRouterOS(gateway)
		}

		c.Data(http.StatusOK, "application/application/x-zip-compressed", data)
		return
	}
}

// ZipItem represent an item of zip package
type ZipItem struct {
	Name       string
	RawContent []byte
	FilePath   string
}

func pack(items []ZipItem) []byte {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive.
	w := zip.NewWriter(buf)

	// Add some files to the archive.
	for _, file := range items {
		f, err := w.Create(file.Name)
		if err != nil {
			log.Fatal(err)
		}
		if file.RawContent != nil {
			_, err = f.Write(file.RawContent)
			if err != nil {
				log.Fatal(err)
			}
		}
		if file.FilePath != "" {
			fd, err := os.Open(file.FilePath)
			if err != nil {
				log.Fatal(err)
			}
			b, err := ioutil.ReadAll(fd)
			if err != nil {
				log.Fatal(err)
			}
			fd.Close()

			_, err = f.Write(b)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Make sure to check the error on Close.
	err := w.Close()
	if err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}

func packRouterOS(gateway string) []byte {
	items := []ZipItem{
		{Name: "routeros-address-list.rsc", RawContent: Generate(filepath.Join("templates", "routeros", "routeros-address-list.rsc"), chnIPs, gateway)},
		{Name: "routeros.rsc", RawContent: Generate(filepath.Join("templates", "routeros", "routeros.rsc"), chnIPs, gateway)},
	}
	return pack(items)
}

func packChinaDNS(gateway string) []byte {
	items := []ZipItem{
		{Name: "chnroute.txt", RawContent: Generate(filepath.Join("templates", "chinadns", "chnroute.txt"), chnIPs, gateway)},
	}
	return pack(items)
}

func packAndroid(gateway string) []byte {
	items := []ZipItem{
		{Name: "routes-up.sh", RawContent: Generate(filepath.Join("templates", "android", "routes-up.sh"), chnIPs, gateway)},
		{Name: "routes-down.sh", RawContent: Generate(filepath.Join("templates", "android", "routes-down.sh"), chnIPs, gateway)},
	}
	return pack(items)
}

func packLinux(gateway string) []byte {
	items := []ZipItem{
		{Name: "routes-up.sh", RawContent: Generate(filepath.Join("templates", "linux", "routes-up.sh"), chnIPs, gateway)},
		{Name: "routes-down.sh", RawContent: Generate(filepath.Join("templates", "linux", "routes-down.sh"), chnIPs, gateway)},
	}
	return pack(items)
}

func packMac(gateway string) []byte {
	items := []ZipItem{
		{Name: "routes-up.sh", RawContent: Generate(filepath.Join("templates", "mac", "routes-up.sh"), chnIPs, gateway)},
		{Name: "routes-down.sh", RawContent: Generate(filepath.Join("templates", "mac", "routes-down.sh"), chnIPs, gateway)},
	}
	return pack(items)
}

func packWin(gateway string) []byte {
	items := []ZipItem{
		{Name: "cmroute.dll", RawContent: nil, FilePath: filepath.Join("templates", "windows", "cmroute.dll")},
		{Name: "routes-up.bat", RawContent: Generate(filepath.Join("templates", "windows", "routes-up.bat"), chnIPs, gateway)},
		{Name: "routes-down.bat", RawContent: nil, FilePath: filepath.Join("templates", "windows", "routes-down.bat")},
		{Name: "routes-up.txt", RawContent: Generate(filepath.Join("templates", "windows", "routes-up.txt"), chnIPs, gateway)},
		{Name: "routes-down.txt", RawContent: Generate(filepath.Join("templates", "windows", "routes-down.txt"), chnIPs, gateway)},
	}
	return pack(items)
}

func main() {
	addr := ":8089"
	if bind, ok := os.LookupEnv("BIND"); ok {
		addr = bind
	}

	httpsEnabled := false
	if https, ok := os.LookupEnv("HTTPS_ENABLED"); ok {
		httpsEnabled, _ = strconv.ParseBool(https)
	}

	r := gin.Default()
	r.LoadHTMLGlob("templates/*.tpl")
	r.GET("/", homePage)
	r.GET("/l/:gateway", homePage)
	r.GET("/l/:gateway/:platform", homePage)
	r.GET("/f/:gateway/:platform/:file", getFile)

	chnIPs = FetchIps()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	apnic, err := filepath.Abs("apnic.txt")
	if err != nil {
		log.Fatal(err)
	}
	err = watcher.Add(apnic)
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					chnIPs = FetchIps()
				}
			case err := <-watcher.Errors:
				if err != nil {
					log.Println("error:", err)
				}
			}
		}
	}()

	if httpsEnabled {
		var certFile, keyFile string
		if cert, ok := os.LookupEnv("HTTPS_CERT"); ok {
			certFile = cert
		}
		if key, ok := os.LookupEnv("HTTPS_KEY"); ok {
			keyFile = key
		}
		if certFile != "" && keyFile != "" {
			log.Fatal(r.RunTLS(addr, certFile, keyFile))
		}
	}
	log.Fatal(r.Run(addr))
}
