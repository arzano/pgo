package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/machinebox/graphql"
	. "github.com/logrusorgru/aurora"
	"log"
	"os"
	"soko-cli/pkg/models"
	"sort"
	"strconv"
	"strings"
	"time"
)


func showPackage(searchTerm string, first bool) {

	fmt.Println()
	fmt.Println("[ Results for search key : ", Bold(searchTerm), " ]")
	fmt.Println("Searching...")
	fmt.Println()

	gpackage, err := findPackage(searchTerm, first)
	if err != nil {
		fmt.Println()
		fmt.Println(err)
		fmt.Println()
	}

	fmt.Println("")
	fmt.Println("")
	fmt.Println(Underline(Bold(strings.Repeat(" ", 50-len(gpackage.Atom)/2) + strings.Repeat(" ", len(gpackage.Atom)) + strings.Repeat(" ", 50-len(gpackage.Atom)/2))))
	fmt.Println(strings.Repeat(" ", 50-len(gpackage.Atom)/2) + strings.Repeat(" ", len(gpackage.Atom)) + strings.Repeat(" ", 50-len(gpackage.Atom)/2))
	fmt.Println(Bold(strings.Repeat(" ", 50-len(gpackage.Atom)/2) + gpackage.Atom + strings.Repeat(" ", 50-len(gpackage.Atom)/2)))
	fmt.Println("")
	fmt.Println(strings.Repeat(" ", 50-len(gpackage.Description())/2) + gpackage.Description())
	fmt.Println(strings.Repeat(" ", 50-len(gpackage.Versions[0].Homepage[0])/2) + gpackage.Versions[0].Homepage[0])
	fmt.Println("")

	if showVersions {
		printVersions(gpackage.Versions)
	}

	if showMetadata {
		printMetadata(gpackage)
	}

	if showBugs {
		printBugs(gpackage.Bugs)
	}

	if showPullRequests {
		printPullRequests(gpackage.PullRequests)
	}

	if showQAreports {
		printQAReports(gpackage)
	}

	if showDependencies {
		printDependencies(gpackage)
	}

	if showChangelog {
		printChangelog(gpackage.Commits)
	}

	fmt.Println()
	fmt.Println(Underline(Bold(strings.Repeat(" ", 50-len(gpackage.Atom)/2) + strings.Repeat(" ", len(gpackage.Atom)) + strings.Repeat(" ", 50-len(gpackage.Atom)/2))))
	fmt.Println()
}

func printVersions(versions []*models.Version){
	fmt.Println(Underline(Bold(Green("Available Versions"))))
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].GreaterThan(*versions[j])
	})
	for _, version := range versions {
		fmt.Println(Bold("  " + version.Version + ": "), version.Keywords)
	}
	fmt.Println("")
}

func printMetadata(gpackage models.Package){
	fmt.Println(Underline(Bold(Green("Package Metadata"))))
	if gpackage.Longdescription != "" {
		fmt.Println(Bold("  Full description: "), gpackage.Longdescription)
	}
	var useflags []string
	var useExpands []string
	if gpackage.Versions[0].Version != "9999" {
		for _, useflag := range gpackage.Versions[0].Useflags {
			if !strings.HasPrefix(useflag, gpackage.Name + "_"){
				useflags = append(useflags, useflag)
			}else{
				useExpands = append(useExpands, useflag)
			}
		}
	} else if len(gpackage.Versions) > 1 {
		for _, useflag := range gpackage.Versions[0].Useflags {
			if !strings.HasPrefix(useflag, gpackage.Name + "_"){
				useflags = append(useflags, useflag)
			}else{
				useExpands = append(useExpands, useflag)
			}
		}
	}
	fmt.Println(Bold("  Useflags: "), strings.Join(useflags, ", "))
	fmt.Println(Bold("  Use Expands: "), strings.Join(useExpands, ", "))
	fmt.Println(Bold("  License: "), gpackage.Versions[0].License)
	fmt.Print(Bold("  Maintainers: "))
	for idx, maintainer := range gpackage.Maintainers {
		if idx < len(gpackage.Maintainers) - 1 {
			fmt.Print(maintainer.Name + ", ")
		} else {
			fmt.Print(maintainer.Name)
		}
	}
	fmt.Println()
	fmt.Println()
}

func printBugs(bugs []*models.Bug){
	if len(bugs) > 0 {
		fmt.Println(Underline(Bold(Green("Bugs"))))
		for _, bug := range bugs {
			fmt.Println("  " + bug.Id + ": " + bug.Summary)
		}
		fmt.Println()
	}
}

func printPullRequests(pullRequests []*models.GithubPullRequest){
	if len(pullRequests) > 0 {
		fmt.Println(Underline(Bold(Green("Pull Requests"))))
		for _, pr := range pullRequests {
			fmt.Println("  " + pr.Id + ": " + pr.Title + " (" + pr.Author + ")")
		}
		fmt.Println()
	}
}

func printQAReports(gpackage models.Package){
	qareportsfound := false
	for _, version := range gpackage.Versions {
		if len(version.PkgCheckResults) > 0 {
			qareportsfound = true
		}
	}
	if len(gpackage.PkgCheckResults) > 0 || qareportsfound {
		fmt.Println(Underline(Bold(Green("QA Report"))))
		if len(gpackage.PkgCheckResults) > 0 {
			fmt.Println(Bold("  All Versions: "))
		}
		for _, qareport := range gpackage.PkgCheckResults {
			fmt.Println("    - " + qareport.Class + ": " + qareport.Message)
		}
		for _, version := range gpackage.Versions {
			if len(version.PkgCheckResults) > 0 {
				fmt.Println(Bold("  " + version.Version + ": "))
				for _, qareport := range version.PkgCheckResults {
					fmt.Println("    - " + qareport.Class + ": " + qareport.Message)
				}
			}
		}
		fmt.Println("")
	}
}

func printDependencies(gpackage models.Package){
	if len(gpackage.ReverseDependencies) > 0 {
		fmt.Println(Underline(Bold(Green("Reverse Dependencies"))))
		var revDeps []string
		for _, revDep := range gpackage.ReverseDependencies {
			revDeps = append(revDeps, revDep.ReverseDependencyAtom)
		}
		revDeps = Deduplicate(revDeps)
		for _, revDep := range revDeps {
			fmt.Println("  - " + revDep)
		}
	}
	fmt.Println("")
}

func printChangelog(commits []*models.Commit){
	fmt.Println(Underline(Bold(Green("Changelog"))))

	sort.Slice(commits, func(i, j int) bool {
		return commits[i].PrecedingCommits > commits[j].PrecedingCommits
	})

	for idx, commit := range commits {
		fmt.Println("  " + commit.CommitterDate.Format(time.RFC822) + ", " + commit.Id[:7] + ": " + commit.Message + " (" + commit.CommitterName + ")")
		if idx == 15-1 {
			break
		}
	}
}

func findPackage(searchTerm string, first bool) (models.Package, error) {
	// create a client (safe to share across requests)
	client := graphql.NewClient("https://packages.gentoo.org/api/graphql/")

	// make a request
	resultSize := "10"
	if first {
		resultSize = "1"
	}

	req := graphql.NewRequest(buildSearchQuery(searchTerm, resultSize))

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
			return models.Package{}, errors.New("Invalid selection. Aborting...")
		}

		gpackage = respData.PackageSearch[selectedIdx]
	}

	return gpackage, nil
}

func buildSearchQuery(searchTerm, resultSize string) string {
	return `
    {
	  packageSearch(searchTerm: "` + searchTerm + `", resultSize: ` + resultSize + `){
		Name,
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
          Author,
        },
        PkgCheckResults {
          Class,
          Message
        }
        Commits {
          Id,
          CommitterName,
          Message,
          PrecedingCommits,
          CommitterDate,
        },
        ReverseDependencies {
          ReverseDependencyAtom,
        }
	  }
	}
	`
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

