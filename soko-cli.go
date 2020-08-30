package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	. "github.com/logrusorgru/aurora"
	"github.com/machinebox/graphql"
	"log"
	"os"
	"soko-cli/pkg/models"
	"sort"
	"strconv"
	"strings"
)

func main() {

	var first = flag.Bool("f", false, "Show the first search result only")
	var searchTerm = flag.String("s", "", "Search Query for a package")
	flag.Parse()

	if searchTerm == nil || *searchTerm == "" {
		flag.PrintDefaults()
		return
	}

	fmt.Println()
	fmt.Println("[ Results for search key : ", Bold(*searchTerm), " ]")
	fmt.Println("Searching...")
	fmt.Println()

	// create a client (safe to share across requests)
	client := graphql.NewClient("https://packages.gentoo.org/api/graphql/")

	// make a request
	resultSize := "10"
	if *first {
		resultSize = "1"
	}

	req := graphql.NewRequest(`
    {
	  packageSearch(searchTerm: "` + *searchTerm + `", resultSize: ` + resultSize + `){
		Atom,
		Versions {
		  Description,
          Homepage,
          Version,
          License,
          Keywords,
          Useflags,
          PkgCheckResults {
            Class,
            Message
          }
		},
        Longdescription,        
        Maintainers {
          Name
        },
        Bugs {
          Id,
          Summary,
        },
        PullRequests {
          Id,
          Title,
        },
        PkgCheckResults {
          Class,
          Message
        }
        Commits {
          Id,
          CommitterName,
          Message,
        },
        ReverseDependencies {
          ReverseDependencyAtom,
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
	var respData struct {
		PackageSearch []models.Package
	}
	if err := client.Run(ctx, req, &respData); err != nil {
		log.Fatal(err)
	}

	var gpackage models.Package

	if len(respData.PackageSearch) == 1 {
		gpackage = respData.PackageSearch[0]
	} else {
		for idx, gpackage := range respData.PackageSearch {
			fmt.Println(Bold(Green("[" + strconv.Itoa(idx) + "] ")), Bold(gpackage.Atom))
			fmt.Println("      ", Green("Homepage:      "), strings.Join(gpackage.Versions[0].Homepage, ", "))
			fmt.Println("      ", Green("Description:   "), gpackage.Versions[0].Description)
			fmt.Println("      ", Green("License:       "), gpackage.Versions[0].License)
			fmt.Println()

			if idx >= 10 {
				break
			}
		}

		fmt.Println("[ Applications found : ", Bold(strconv.Itoa(len(respData.PackageSearch))), " ]")
		fmt.Println()

		reader := bufio.NewReader(os.Stdin)
		fmt.Print(Bold("Which package have you been looking for? "), "[", Bold(Green("0-" + strconv.Itoa(min(10-1, len(respData.PackageSearch)-1)))), "] ")
		text, _ := reader.ReadString('\n')

		selectedIdx, err := strconv.Atoi(strings.ReplaceAll(text, "\n", ""))

		if err != nil || selectedIdx < 0 || selectedIdx > min(10, len(respData.PackageSearch)-1) {
			fmt.Println()
			fmt.Println("Invalid selection. Aborting...")
			return
		}

		gpackage = respData.PackageSearch[selectedIdx]
	}


	fmt.Println("")
	fmt.Println("")
	fmt.Println("----------------------------------- " + gpackage.Atom + " -----------------------------------")
	fmt.Println("")
	fmt.Println("Available Versions")
	for _, version := range gpackage.Versions {
		fmt.Println("  " + version.Version + ": " + version.Keywords)
	}

	fmt.Println("")
	fmt.Println("Package Metadata")
	if gpackage.Longdescription != "" {
		fmt.Println("  Full description: " + gpackage.Longdescription)
	}
	useflags := ""
	if gpackage.Versions[0].Version != "9999" {
		useflags = strings.Join(gpackage.Versions[0].Useflags, ", ")
	} else if len(gpackage.Versions) > 1 {
		useflags = strings.Join(gpackage.Versions[1].Useflags, ", ")
	}
	fmt.Println("  Useflags: " + useflags)
	fmt.Println("  License: " + gpackage.Versions[0].License)
	fmt.Print("  Maintainers: ")
	for idx, maintainer := range gpackage.Maintainers {
		if idx < len(gpackage.Maintainers) - 1 {
			fmt.Print(maintainer.Name + ", ")
		} else {
			fmt.Print(maintainer.Name)
		}
	}
	fmt.Println()

	if len(gpackage.Bugs) > 0 {
		fmt.Println("")
		fmt.Println("Bugs")
		for _, bug := range gpackage.Bugs {
			fmt.Println("  - " + bug.Id + ": " + bug.Summary)
		}
	}

	if len(gpackage.PullRequests) > 0 {
		fmt.Println("")
		fmt.Println("Pull Requests")
		for _, pr := range gpackage.PullRequests {
			fmt.Println("  - " + pr.Id + ": " + pr.Title)
		}
	}

	qareportsfound := false
	for _, version := range gpackage.Versions {
		if len(version.PkgCheckResults) > 0 {
			qareportsfound = true
		}
	}
	if len(gpackage.PkgCheckResults) > 0 || qareportsfound {
		fmt.Println("")
		fmt.Println("QA Report")
		if len(gpackage.PkgCheckResults) > 0 {
			fmt.Println("  - All Versions: ")
		}
		for _, qareport := range gpackage.PkgCheckResults {
			fmt.Println("    - " + qareport.Class + ": " + qareport.Message)
		}
		for _, version := range gpackage.Versions {
			if len(version.PkgCheckResults) > 0 {
				fmt.Println("  - " + version.Version + ": ")
				for _, qareport := range version.PkgCheckResults {
					fmt.Println("    - " + qareport.Class + ": " + qareport.Message)
				}
			}
		}
	}

	if len(gpackage.ReverseDependencies) > 0 {
		fmt.Println("")
		fmt.Println("Reverse Dependencies")
		var revDeps []string
		for _, revDep := range gpackage.ReverseDependencies {
			revDeps = append(revDeps, revDep.ReverseDependencyAtom)
		}
		revDeps = Deduplicate(revDeps)
		fmt.Println("  - " + strings.Join(revDeps, ", "))
	}

	fmt.Println("")
	fmt.Println("Changelog")
	for idx, commit := range gpackage.Commits {
		fmt.Println("  - " + commit.Id + ": " + commit.Message + " (" + commit.CommitterName + ")")
		if idx == 15-1 {
			break
		}
	}

	fmt.Println()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Deduplicate(items []string) []string {
	if items != nil && len(items) > 1 {
		sort.Strings(items)
		j := 0
		for i := 1; i < len(items); i++ {
			if items[j] == items[i] {
				continue
			}
			j++
			// preserve the original data
			// in[i], in[j] = in[j], in[i]
			// only set what is required
			items[j] = items[i]
		}
		result := items[:j+1]
		return result
	} else {
		return items
	}
}