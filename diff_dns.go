package main

import (
	"fmt"
	"strings"
	"os"
	"encoding/csv"
	"log"
	"bytes"
	"os/exec"
	"time"
	"sync"
)

const ROUTINE_NUM  = 100

func main() {
	start := time.Now()
	count := 0

	var group sync.WaitGroup
	buf := make(chan int)

	file,err := os.Open("ipv6.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	sites, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	total := len(sites)
	interval := total / ROUTINE_NUM

	for i := 0; i <= ROUTINE_NUM ; i++  {
		group.Add(1)
		start := i * interval
		end := start + interval
		if end > total {
			end = total
		}
		go func() {
			defer group.Done()
			num := 0
			site_chunck := sites[start: end]
			fmt.Println(site_chunck)
			for _,website := range site_chunck {
				s4 := http_v4(website[0])
				s6 := http_v6(website[0])
				if strings.Compare(s4, s6) != 0 {
					num++
				}
			}
			buf <- num
		}()
	}

	go func() {
		group.Wait()
		close(buf)
	}()

	for n := range buf {
		count += n
	}

	end := time.Now()
	time := end.Sub(start)
	fmt.Printf("there are %d websites supporting differentiated services, and the total websites number is %d\n", count, total)
	fmt.Printf("cosuming time %s\n", time)
}


func http_v6(domain string) string {
	cmd := exec.Command("proxychains4", "curl", "-6", "-L", domain)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return out.String()
}

func http_v4(domain string) string {
	cmd := exec.Command("proxychains4", "curl", "-4", "-L", domain)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return out.String()
}