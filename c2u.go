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
	"strings"
)

const service_unit = `
[Unit]
Description=Start, stop and restart {{name}} container
Requires=docker.service {{#links}}{{.}} {{/links}}
After=docker.service {{#links}}{{.}} {{/links}}

[Service]
Restart=always

ExecStartPre=-/usr/bin/docker rm -f %n

ExecStart=/usr/bin/docker run \
    {{#environment}}-e {{{.}}} {{/environment}}\
    {{#links}}--link {{.}}:{{.}} {{/links}} \
    {{#ports}}-p {{.}} {{/ports}} \
    {{#env_file}}--env-file={{{.}}} {{/env_file}} \
    --rm --name %n \
    {{#image}}{{{.}}}{{/image}} {{#command}}{{{.}}}{{/command}}

ExecStop=-/usr/bin/docker stop %n

[Install]
WantedBy=multi-user.target
`
const oneshot_unit = `
[Unit]
Description=Start a oneshot for {{name}} container

[Service]
Type=oneshot

ExecStart=/usr/bin/docker exec \
    {{#exec}}{{{.}}}{{/exec}} {{#command}}{{{.}}}{{/command}}
`

func main() {

	var (
		help      = flag.Bool("h", false, "display usage")
		compose   = flag.String("c", "", "compose file")
		template  = flag.String("t", "", "template source file or one of [int:service_unit | int:oneshot_unit]")
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

		if _, ok := value.(map[interface{}]interface{})["build"]; ok {

			fmt.Println("Container build is not supported!", key, ":", value.(map[interface{}]interface{}))
		} else {

			var data string
			if strings.Contains(*template, "int:service_unit") {

				data = mustache.Render(service_unit, value.(map[interface{}]interface{}))
			} else if strings.Contains(*template, "int:oneshot_unit") {

				data = mustache.Render(oneshot_unit, value.(map[interface{}]interface{}))
			} else {

				data = mustache.RenderFile(*template, value.(map[interface{}]interface{}))
			}
			println(data)
			fname := *targetdir + "/" + key
			err := ioutil.WriteFile(fname, []byte(data), 0644)
			if err != nil {
				panic(err)
			}
		}

	}

}
