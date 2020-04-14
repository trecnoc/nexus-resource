package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/trecnoc/nexus-resource"
	"github.com/trecnoc/nexus-resource/check"
	"github.com/trecnoc/nexus-resource/utils"
)

func main() {
	var request check.Request
	inputRequest(&request)

	if request.Source.Debug {
		jsonString, _ := json.Marshal(request)
		ioutil.WriteFile("/tmp/concourse-nexus-request.json", jsonString, os.ModePerm)
	}

	client := nexusresource.NewNexusClient(request.Source.URL, request.Source.Username, request.Source.Password, request.Source.Debug)

	command := check.NewCommand(client)
	response, err := command.Run(request)
	if err != nil {
		utils.Fatal("running command", err)
	}

	outputResponse(response)
}

func inputRequest(request *check.Request) {
	if err := json.NewDecoder(os.Stdin).Decode(request); err != nil {
		utils.Fatal("reading request from stdin", err)
	}
}

func outputResponse(response check.Response) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		utils.Fatal("writing response to stdout", err)
	}
}
