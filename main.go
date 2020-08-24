package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

type tfShowJson struct {
	Values struct {
		RootModule struct {
			Resources []struct {
				Name   string `json:"name"`
				Values struct {
					Image       string `json:"image"`
					ImageName   string `json:"image_name"`
					Name        string `json:"name"`
					Ipv4Address string `json:"ipv4_address"`
					PublicIp    string `json:"public_ip"`
					Network     []struct {
						FixedIpV4 string `json:"fixed_ip_v4"`
					} `json:"network"`
				} `json:"values"`
			} `json:"resources"`
		} `json:"root_module"`
	} `json:"values"`
}

var distro = regexp.MustCompile(`\A\w+`)

func main() {
	tfRaw, errRA := ioutil.ReadAll(os.Stdin)
	if errRA != nil {
		fmt.Fprintf(os.Stderr, "ioutil.ReadAll(os.Stdin): %s\n", errRA.Error())
		os.Exit(1)
		return
	}

	var tfJson tfShowJson
	if errJU := json.Unmarshal(tfRaw, &tfJson); errJU != nil {
		fmt.Fprintf(os.Stderr, "json.Unmarshal(ioutil.ReadAll(os.Stdin)): %s\n", errJU.Error())
		os.Exit(1)
		return
	}

	for _, resource := range tfJson.Values.RootModule.Resources {
		out := &bytes.Buffer{}
		ok := false

		if resource.Values.Name != "" {
			fmt.Fprint(out, resource.Values.Name)
		}

		if len(resource.Values.Network) > 0 && resource.Values.Network[0].FixedIpV4 != "" {
			fmt.Fprint(out, " ansible_host=")
			fmt.Fprint(out, resource.Values.Network[0].FixedIpV4)
			ok = true
		}

		if resource.Values.Ipv4Address != "" {
			fmt.Fprint(out, " ansible_host=")
			fmt.Fprint(out, resource.Values.Ipv4Address)
			ok = true
		}

		if resource.Values.PublicIp != "" {
			fmt.Fprint(out, " ansible_host=")
			fmt.Fprint(out, resource.Values.PublicIp)
			ok = true
		}

		if match := distro.FindString(resource.Values.Image); match != "" {
			fmt.Fprint(out, " ansible_user=")
			fmt.Fprint(out, strings.ToLower(match))
		}

		if match := distro.FindString(resource.Values.ImageName); match != "" {
			fmt.Fprint(out, " ansible_user=")
			fmt.Fprint(out, strings.ToLower(match))
		}

		out.Write([]byte{'\n'})

		if ok {
			io.Copy(os.Stdout, out)
		}
	}
}
