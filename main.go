package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/admin100/util/console"
	"github.com/corpix/uarand"
)

var (
	start int
	end   int

	tokens  []string
	threads int

	valid   int
	invalid int
	locked  int
	unknown int

	running int
)

func main() {
	clear()
	wg := sync.WaitGroup{}

	f, err := os.Open("tokens.txt")

	if err != nil {
		fmt.Printf("[\x1b[38;5;63m%s\x1b[0m] Please put your tokens inside \x1b[38;5;63mtokens.txt\x1b[0m\n", time.Now().Format("15:04:05"))
		return
	}

	s := bufio.NewScanner(f)

	for s.Scan() {
		tokens = append(tokens, s.Text())
	}

	console.SetConsoleTitle("[Discord Token Checker]")
	fmt.Printf("[\x1b[38;5;63m%s\x1b[0m] Threads\x1b[38;5;63m>\x1b[0m ", time.Now().Format("15:04:05"))
	fmt.Scanln(&threads)
	fmt.Println()

	goroutines := make(chan struct{}, threads)
	start = int(time.Now().Unix())
	go title_worker()

	for i := 0; i < len(tokens); i++ {
		wg.Add(1)
		go func(token string) {
			defer wg.Done()
			goroutines <- struct{}{}
			check(token)
			<-goroutines
		}(tokens[i])
	}

	wg.Wait()
	end = int(time.Now().Unix())

	fmt.Println()
	fmt.Println("Results:")
	fmt.Println("   Valid\x1b[38;5;63m:\x1b[0m  ", valid)
	fmt.Println("   Locked\x1b[38;5;63m:\x1b[0m ", locked)
	fmt.Println("   Invalid\x1b[38;5;63m:\x1b[0m", invalid)
	fmt.Println("   Unknown\x1b[38;5;63m:\x1b[0m", unknown)
	fmt.Printf("   Took\x1b[38;5;63m:\x1b[0m    %ds\n", (end - start))
	time.Sleep(3 * time.Millisecond)
}

func title_worker() {
	for {
		console.SetConsoleTitle(fmt.Sprintf("[Discord Token Checker] Valid: %d | Locked: %d | Invalid: %d | Unknown: %d", valid, locked, invalid, unknown))
	}
}

func hide_token(token string) string {
	return string(token[:32]) + "**************************"
}

func check(token string) {
	running++
	hidden_token := hide_token(token)

	req, _ := http.NewRequest("GET", "https://discord.com/api/v9/users/@me/library", nil)
	req.Header.Set("User-Agent", uarand.GetRandom())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Authorization", token)

	client := &http.Client{}
	resp, _ := client.Do(req)
	resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Printf("[\x1b[38;5;63m%s\x1b[0m] Valid (\x1b[38;5;63m%s\x1b[0m)\n", time.Now().Format("15:04:05"), hidden_token)
		valid++

		file, _ := os.OpenFile("valid.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		file.Write([]byte(token + "\n"))

	} else if resp.StatusCode == 401 {
		fmt.Printf("[\x1b[38;5;63m%s\x1b[0m] Invalid (\x1b[38;5;63m%s\x1b[0m)\n", time.Now().Format("15:04:05"), hidden_token)
		invalid++
	} else if resp.StatusCode == 403 {
		fmt.Printf("[\x1b[38;5;63m%s\x1b[0m] Locked (\x1b[38;5;63m%s\x1b[0m)\n", time.Now().Format("15:04:05"), hidden_token)
		locked++
	} else {
		fmt.Printf("[\x1b[38;5;63m%s\x1b[0m] Unknown (\x1b[38;5;63m%s\x1b[0m)\n", time.Now().Format("15:04:05"), hidden_token)
		unknown++
	}

	running--
}

func clear() {
	c := exec.Command("cmd", "/c", "cls")
	c.Stdout = os.Stdout
	c.Run()
}
