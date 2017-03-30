package main

import (
	"net"
	"log"
	"fmt"
	"strings"
	"os"
	"bufio"
	"io"
	"time"
	"sync"
	"encoding/csv"
)

type site struct {
	domain string
	ipv4 string
	ipv6 string
}

const FILEPATH  = "./top-1m.csv"
const DOMAIN_NUM  = 1000000
const ROUTINE_NUM  = 1000

func main()  {
	start := time.Now()

	var str []string
	var sites []site
	var group sync.WaitGroup
	buf := make(chan int)

	str = readFile(FILEPATH, str)

	for i := 0; i < ROUTINE_NUM ; i++ {
		start := i * DOMAIN_NUM / ROUTINE_NUM
		end := start + DOMAIN_NUM / ROUTINE_NUM
		group.Add(1)

		go func () {
			str_chunck := str[start: end]
			defer group.Done()
			var count  = 0
			if len(str_chunck) > 0 {
				fmt.Println("ok")
			}
			for _, s := range str_chunck {
				s = strings.TrimSpace(s)
				website, ok := parseDNS(s)
				if ok {
					count++
					if website.ipv4 != "" {
						sites = append(sites, website)
					}
				}
			}
			buf <- count
		}()
	}

	go func() {
		group.Wait()
		close(buf)
	}()

	total := 0
	for num := range buf {
		total += num
	}

	writeFile(sites)
	end := time.Now()
	delta := end.Sub(start)
	fmt.Printf("cosuming time %s\n", delta)
	fmt.Printf("there are %d websites supporting IPV6\n",total)
}

func readFile(filename string, strs []string) []string{
	f,err := os.Open(filename)
	defer f.Close()

	if err != nil {
		log.Fatal(err)
	}
	buf := bufio.NewReader(f)
	for {
		line,err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return strs
			}
			log.Fatal(err)
		}
		strs = append(strs, strings.Split(line, ",")[1])
	}
	return strs
}

/**
	write ipv6-supported domain into file ipv6.csv
 */
func writeFile(websites []site)  {
	cvsFile, err := os.Create("ipv6.csv")
	if err != nil {
		panic(err)
	}
	defer cvsFile.Close()

	writer := csv.NewWriter(cvsFile)
	for _,web := range websites {
		line := []string{web.domain, web.ipv4, web.ipv6}
		err := writer.Write(line)
		if err != nil {
			panic(err)
		}
	}
	writer.Flush()
}
/**
	return a struct site, if the web support ipv6
 */
func parseDNS(s string) (site, bool) {
	str,err := net.LookupHost(s) // first parse domain without
	var result site
	if err != nil {
		return result, false
	}
	for _,ip := range str {
		if isIPV4(ip) {
			result.ipv4 = ip
			continue
		}
		if isIPV6(ip) {
			result.ipv6 = ip
			continue
		}
	}
	if result.ipv6 != ""{
		result.domain = s
		return result, true
	} else {
		return result, false
	}
}

func isIPV6(str string) bool {
	if strings.Contains(str,":") {
		return true
	}
	return false
}

func isIPV4(str string) bool {
	if strings.Contains(str, ".") {
		return true
	}
	return false
}



