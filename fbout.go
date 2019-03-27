package main

import "github.com/fluent/fluent-bit-go/output"
import (
	"bufio"
	"strings"
	"bytes"
	"fmt"
	"unsafe"
	"C"
	"context"
	"firebase.google.com/go"
	"google.golang.org/api/option"
)

var (
		db string = ""
		key string = ""
		dir string = ""
		child string = ""
	)

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return output.FLBPluginRegister(ctx, "fbout", "First Blood!")
}

//export FLBPluginInit
// (fluentbit will call this)
// ctx (context) pointer to fluentbit context (state/ c code)
func FLBPluginInit(ctx unsafe.Pointer) int {
	// Example to retrieve an optional configuration parameter
	db = output.FLBPluginConfigKey(ctx, "Db")
	key = output.FLBPluginConfigKey(ctx, "Key")
	dir = output.FLBPluginConfigKey(ctx, "Dir")
	child = output.FLBPluginConfigKey(ctx, "Child")
	fmt.Printf("[flb-go] databaseURL = '%s'\n", db)
	fmt.Printf("[flb-go] json Key Location = '%s'\n", key)
	fmt.Printf("[flb-go] database Directory = '%s'\n", dir)
	fmt.Printf("[flb-go] child device Name = '%s'\n", child)
	return output.FLB_OK
}

//export FLBPluginFlush
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
	var count int
	var ret int
	var ts interface{}
	var buf bytes.Buffer
	var record map[interface{}]interface{}
	
	ctx := context.Background()
	conf := &firebase.Config{
		DatabaseURL: db,
	}
	opt := option.WithCredentialsFile(key)
	app, _ := firebase.NewApp(ctx, conf, opt)
	client, _ := app.Database(ctx)
	ref := client.NewRef(dir)
	postsRef := ref.Child(child)
	
	type Post struct {
		    Time string `json:"Time,omitempty"`
			State  string `json:"State,omitempty"`
			}
	
	// Create Fluent Bit decoder
	dec := output.NewDecoder(data, int(length))

	// Iterate Records
	count = 0
	for {
		// Extract Record
		ret, ts, record = output.GetRecord(dec)
		if ret != 0 {
			break
		}
		// Print record keys and values
		timestamp := ts.(output.FLBTime)
		buffer := bufio.NewWriter(&buf)
		for k, v := range record {
			fmt.Fprintf(buffer,"\"%s\": %s",k,v)
		}
		buffer.Flush()
		s := buf.String()
		p := strings.Split(s,",")
		for j := range p {
			p[j] = strings.TrimPrefix(p[j],"\"exec\": ")
			postsRef.Push(ctx, &Post{
				Time: timestamp.String(),
				State: p[j],
			});
		}
		count++
	}
	return output.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit() int {
	return output.FLB_OK
}

func main() {
}
