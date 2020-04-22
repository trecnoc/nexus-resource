package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/trecnoc/nexus-resource"
	"github.com/trecnoc/nexus-resource/in"
	"github.com/trecnoc/nexus-resource/utils"
)

func main() {
	if len(os.Args) < 2 {
		utils.Sayf("usage: %s <dest directory>\n", os.Args[0])
		os.Exit(1)
	}

	destinationDir := os.Args[1]

	var request in.Request
	inputRequest(&request)

	if request.Source.Debug {
		jsonString, _ := json.Marshal(request)
		ioutil.WriteFile("/tmp/concourse-nexus-request.json", jsonString, os.ModePerm)
	}

	client := nexusresource.NewNexusClient(request.Source.URL, request.Source.Username, request.Source.Password, request.Source.Timeout, request.Source.Debug)

	command := in.NewCommand(client)
	response, err := command.Run(destinationDir, request)
	if err != nil {
		utils.Fatal("running command", err)
	}

	outputResponse(response)
}

func inputRequest(request *in.Request) {
	if err := json.NewDecoder(os.Stdin).Decode(request); err != nil {
		utils.Fatal("reading request from stdin", err)
	}
}

func outputResponse(response in.Response) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		utils.Fatal("writing response to stdout", err)
	}
}
