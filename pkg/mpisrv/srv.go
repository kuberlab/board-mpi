package mpisrv

import (
	"net/http"
	"fmt"
	"log"
	"io/ioutil"
	"io"
	"bufio"
	"strings"
	"os/exec"
	"os"
)

func ServMPI(port int){
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		bargs,err := ioutil.ReadAll(r.Body)
		if err!=nil{
			errorResponse(w,fmt.Errorf("Failed read exec args: %s",err))
			os.Exit(1)
			return
		}
		err = execCmd(w,string(bargs))
		if err!=nil{
			errorResponse(w,err)
			os.Exit(1)
			return
		}
		os.Exit(0)
		w.WriteHeader(http.StatusOK)

	})
	fmt.Printf("Server: %d\n",port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d",port), nil))
}

func errorResponse(w http.ResponseWriter,err error){
	message := fmt.Sprintf("Failed remote: %v",err)
	fmt.Println(message)
	w.Write([]byte(message))
	w.WriteHeader(http.StatusInternalServerError)
}

func execCmd(w http.ResponseWriter, args string) error {
	fmt.Println("Exec: ",args)
	cmd := exec.Command("/bin/sh","-c",args)
	cout, err := cmd.StdoutPipe()
	if err != nil {
		return  err
	}
	cerr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	go read_state(w,outPair{in:cerr,name:"ERROR"},outPair{in:cout,name:"INFO"})
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err = <-done:
	}
	return err
}

type outPair struct {
	in io.ReadCloser
	name string
}
func read_state(w http.ResponseWriter,in ...outPair) {
	for i := range in {
		r := bufio.NewReader(in[i].in)
		name := in[i].name
		go func() {
			for line, err := r.ReadString('\n'); err == nil; line, err = r.ReadString('\n') {
				row := strings.TrimSuffix(line, "\n")
				row = fmt.Sprintln(name,":",row)
				fmt.Println(row)
				w.Write([]byte(row))
			}
		}()
	}
}