package main

import (
	"encoding/json"
	"os"

	"github.com/trecnoc/nexus-resource"
	"github.com/trecnoc/nexus-resource/out"
	"github.com/trecnoc/nexus-resource/utils"
)

func main() {
	if len(os.Args) < 2 {
		utils.Sayf("usage: %s <sources directory>\n", os.Args[0])
		os.Exit(1)
	}

	var request out.Request
	inputRequest(&request)

	sourceDir := os.Args[1]

	client := nexusresource.NewNexusClient(request.Source.URL, request.Source.Username, request.Source.Password)

	command := out.NewCommand(os.Stderr, client)
	response, err := command.Run(sourceDir, request)
	if err != nil {
		utils.Fatal("running command", err)
	}

	outputResponse(response)
}

func inputRequest(request *out.Request) {
	if err := json.NewDecoder(os.Stdin).Decode(request); err != nil {
		utils.Fatal("reading request from stdin", err)
	}
}

func outputResponse(response out.Response) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		utils.Fatal("writing response to stdout", err)
	}
}
