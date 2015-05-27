package main

import (
	"flag"
	"fmt"
	"github.com/hoisie/mustache"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
)

func main() {

	var (
		help      = flag.Bool("h", false, "display usage")
		compose   = flag.String("c", "", "compose file")
		template  = flag.String("t", "", "template source file")
		targetdir = flag.String("o", "/etc/systemd/system", "target directory for unit files")
	)

	flag.Parse()

	if *help {
		println("Convert Docker compose files to systemd unit files.")
		println(os.Args[0], " [ OPTIONS ] ")
		flag.PrintDefaults()
		os.Exit(1)
	}

	filename, _ := filepath.Abs(*compose)
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	var whole map[string]interface{}
	yaml.Unmarshal(yamlFile, &whole)
	fmt.Printf("whole => %#+v\n", whole)

	for key, value := range whole {
		//fmt.Println("Key:", key, "Value:", value)
		//fmt.Println("V:", value.(map[interface{}]interface{})["ports"])

		// set name as part of the container config
		value.(map[interface{}]interface{})["name"] = key

		// unify types to be arrays
		for k, v := range value.(map[interface{}]interface{}) {
			//fmt.Println("Key:", k, "Value:", v)
			switch v.(type) {
			case string:
				//fmt.Println("string", v)
				value.(map[interface{}]interface{})[k] = [1]string{v.(string)}
			case int32, int64:
				//fmt.Println("int", v)
			case []interface{}:
				//fmt.Println("[]interface{}", v)
			default:
				fmt.Println("unknown")
				fmt.Println(reflect.TypeOf(value))
			}

		}

		if _, ok := value.(map[interface{}]interface{})["image"]; ok {

			data := mustache.RenderFile(*template, value.(map[interface{}]interface{}))
			println(data)
			fname := *targetdir + "/" + key + ".service"
			err := ioutil.WriteFile(fname, []byte(data), 0644)
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Println("Container definition doesn't contain 'image' property!", key, ":", value.(map[interface{}]interface{}))
		}

	}

}
