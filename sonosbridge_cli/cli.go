package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/brotherlogic/goserver/utils"

	pb "github.com/brotherlogic/sonosbridge/proto"
)

func main() {
	ctx, cancel := utils.ManualContext("sonosbridge-cli", time.Minute*5)
	defer cancel()

	conn, err := utils.LFDialServer(ctx, "sonosbridge")
	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}
	client := pb.NewSonosBridgeServiceClient(conn)

	switch os.Args[1] {
	case "config":
		addFlags := flag.NewFlagSet("AddConfig", flag.ExitOnError)
		var key = addFlags.String("key", "", "Id of the record to add")
		var secret = addFlags.String("secret", "", "Cost of the record")
		var code = addFlags.String("code", "", "Code")

		if err := addFlags.Parse(os.Args[2:]); err == nil {
			_, err := client.SetConfig(ctx, &pb.SetConfigRequest{Client: *key, Secret: *secret, Code: *code})
			if err != nil {
				log.Fatalf("Bad set: %v", err)
			}
		}
	case "token":
		token, err := client.GetToken(ctx, &pb.GetTokenRequest{})
		if err != nil {
			log.Fatalf("Bad set: %v", err)
		}
		fmt.Printf("%v\n", token.GetToken())
	case "url":
		url, err := client.GetAuthUrl(ctx, &pb.GetAuthUrlRequest{})
		if err != nil {
			log.Fatalf("Bad set: %v", err)
		}
		fmt.Printf("%v\n", url.GetUrl())
	case "household":
		url, err := client.GetHousehold(ctx, &pb.GetHouseholdRequest{})
		if err != nil {
			log.Fatalf("Bad set: %v", err)
		}
		fmt.Printf("%v\n", url.GetHousehold())
	case "volume":
		url, err := client.GetVolume(ctx, &pb.GetVolumeRequest{Player: "Office 2"})
		if err != nil {
			log.Fatalf("Bad set: %v", err)
		}
		fmt.Printf("%v\n", url.GetVolume())
	case "svolume":
		url, err := client.SetVolume(ctx, &pb.SetVolumeRequest{Player: "Office 2", Volume: 20})
		if err != nil {
			log.Fatalf("Bad set: %v", err)
		}
		fmt.Printf("%v\n", url)
	}
}
