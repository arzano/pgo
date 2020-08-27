package main

import (
	"context"
	"github.com/machinebox/graphql"
	"log"
)

func main() {


	// create a client (safe to share across requests)
	client := graphql.NewClient("https://packages.gentoo.org/api/graphql/")

	// make a request
	req := graphql.NewRequest(`
    {
	  packages(Name: "gentoo-sources"){
		Atom,
		Maintainers {
		  Name
		}
	  }
	}
	`)

	// set any variables
	//req.Var("key", "value")

	// set header fields
	req.Header.Set("Cache-Control", "no-cache")

	// define a Context for the request
	ctx := context.Background()

	// run it and capture the response
	var respData ResponseStruct
	if err := client.Run(ctx, req, &respData); err != nil {
		log.Fatal(err)
	}
}

type ResponseStruct struct {
	Packages []Package
}

type Package struct {
	Atom string
	Maintainers []Maintainer
}

type Maintainer struct {
	Name string
}