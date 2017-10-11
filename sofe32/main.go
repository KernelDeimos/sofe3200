package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/docopt/docopt-go"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

const UUIDFile = "/tmp/classact_uuid"

type DocoptArgs map[string]interface{}

func main() {
	usage := `sofe32 Command

Usage:
	sofe32 get  <name> [--host=<host>] [--debug]
	sofe32 send <name> [--host=<host>] [--debug]
	sofe32 uuid [--host=<host>] [--debug]

Options:
	-h --help    Show this help page
	--version    Show version`

	args, _ := docopt.Parse(usage, nil, true, "SOFE32 1.0.0", false)

	if args["--debug"].(bool) {
		logrus.SetLevel(logrus.DebugLevel)
	}

	ins := Instance{}

	if args["--host"] == nil {
		ins.host = os.Getenv("SOFEHOST")
		if ins.host == "" {
			ins.host = "http://localhost:8080"
		}
	} else {
		ins.host = args["--host"].(string)
	}

	ins.setup()

	if args["uuid"].(bool) {
		logrus.Info("Your UUID is " + ins.uuid)
		return
	}

	if args["get"].(bool) {
		ins.doGet(args["<name>"].(string))
		return
	}

	if args["send"].(bool) {
		ins.doSend(args["<name>"].(string))
		return
	}
}

type Instance struct {
	host string
	uuid string
}

func (ins *Instance) setup() {
	// Check if UUID file already exists
	data, err := ioutil.ReadFile(UUIDFile)

	// If file doesn't exist (or other error), fetch new UUID
	if err != nil {
		route := ins.host + "/uuid"
		resp, err := http.Get(route)
		if err != nil {
			logrus.Fatal("Could not get UUID: ", err.Error())
		}

		defer resp.Body.Close()
		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			logrus.Fatal("Could not get UUID: ", err.Error())
		}

		// Create file to store UUID
		err = ioutil.WriteFile(UUIDFile, data, 0644)
		if err != nil {
			logrus.Warn("Could not save UUID: ", err.Error())
		}
	}

	ins.uuid = string(data)
	logrus.Debug(fmt.Sprintf("Have UUID: %s", ins.uuid))

}

func (ins *Instance) doGet(name string) {
	logrus.Debug("Getting: " + name)

	route := ins.host + "/recv/" + ins.uuid + "/" + name
	resp, err := http.Get(route)
	if err != nil {
		logrus.Fatal("Could not get data: ", err.Error())
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Fatal("Could not get data: ", err.Error())
	}

	fmt.Print(string(data))
}

func (ins *Instance) doSend(name string) {
	reader := bufio.NewReader(os.Stdin)
	var input string
	for {
		inputPart, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		input += inputPart
	}

	// Also output input to stdout
	color.Magenta("===== Your Input =====")
	fmt.Println(input)

	route := ins.host + "/send/" + ins.uuid + "/" + name
	body := url.Values{}
	body.Add("data", input)
	resp, err := http.Post(
		route, "application/x-www-form-urlencoded",
		strings.NewReader(body.Encode()),
	)
	if err != nil {
		logrus.Error("Could not send data: " + err.Error())
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Fatal("Could not read response after sending data: " + err.Error())
	}
	if resp.StatusCode == http.StatusBadRequest {
		logrus.Error(string(data))
	} else if resp.StatusCode != http.StatusOK {
		logrus.Error("The server responded with an unexpected status code.")
	}
	color.Magenta("===== Server Response =====")
	fmt.Println(string(data))

}
