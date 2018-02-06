package main

import (
	"bufio"
	"fmt"
	"github.com/kuberlab/board-mpi/pkg/mpisrv"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"strconv"
	"time"
)

func main() {
	args := os.Args
	port := 8080
	if len(args)>1 {
		for i := range args {
			if args[i] == "-p" {
				if v,err := strconv.Atoi(args[i+1]);err==nil {
					port = v
					args = args[i + 1:]
					break
				} else{
					panic(err)
				}
			}
		}
	}
	if len(args) < 2 {
		mpisrv.ServMPI(port)
	}
	/*fmt.Println("---------------ARGS---------------")
	for i,a := range os.Args{
		fmt.Printf("%d: %s\n",i,a)
	}
	fmt.Println("----------------------------------")
	fmt.Println("---------------ENVS---------------")
	for i,a := range os.Environ(){
		fmt.Printf("%d: %s\n",i,a)
	}
	fmt.Println("----------------------------------")*/
	host := args[1]
	execArgs := strings.Join(args[2:], " ")
	fmt.Printf("Connection to %s:%d\n",host,port)
	var resp *http.Response
	var err error
	for{
		resp, err = http.Post(fmt.Sprintf("http://%s:%d", host, port), "text/plain", strings.NewReader(execArgs))
		if err != nil {
			fmt.Printf("Failed connect to worker: %v\n",err)
			time.Sleep(5*time.Second)
		} else {
			break
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		message, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		panic(fmt.Sprintf("Bad response from board mpi: %s\n", string(message)))
	}
	r := bufio.NewReader(resp.Body)
	for line, err := r.ReadString('\n'); err == nil; line, err = r.ReadString('\n') {
		row := strings.TrimSuffix(line, "\n")
		fmt.Println(row)
	}
}
