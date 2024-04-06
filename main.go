package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type issue struct {
	Title  string
	Author struct {
		Login string
	}

	/*Assignees struct {
		Email string
	}*/
}

type pair struct {
	key  string
	Data int
}

func main() {
	fmt.Println("hello shurcool")

	log.Default().Println("Program started at ", time.Now())

	// creating auth client
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ""},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	// query struct for shurcooL library to parse the result
	var q struct {
		Repository struct {
			Issues struct {
				Nodes    []issue
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"issues(first: 100, states: [CLOSED], after: $startCursor)"`
		} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
	}

	// values of the inner varriables mentioned in struct
	varriables := map[string]interface{}{
		"repositoryOwner": githubv4.String("quarkusio"),
		"repositoryName":  githubv4.String("quarkus"),
		"startCursor":     (*githubv4.String)(nil),
	}

	// place holder for each issue result
	var allIssues []issue

	// loop to process and paginate requests
	for {
		err := client.Query(context.Background(), &q, varriables)
		if err != nil {
			log.Fatalln("error during getting issues, err= ", err)
		}
		allIssues = append(allIssues, q.Repository.Issues.Nodes...)
		if !q.Repository.Issues.PageInfo.HasNextPage {
			break
		}
		varriables["startCursor"] = *githubv4.NewString((q.Repository.Issues.PageInfo.EndCursor))
	}

	log.Default().Println("making requests is over at ", time.Now())

	//couinting issues for each author
	m := make(map[string]int)
	for _, i := range allIssues {
		//fmt.Println(t, issue_to_string(i))
		if i.Author.Login == "" {
			m["Empty"]++
			continue
		}

		m[i.Author.Login]++
	}

	//processing for sorting
	pairs := map_to_pairs(m)
	pairs = sort_pairs(pairs)

	// writing the sorted pairs to the q2.csv file
	file, err3 := os.Create("q2.csv")
	if err3 != nil {
		log.Fatalln("error creating file, err= ", err3)
	}

	for t := range pairs {
		file.WriteString(fmt.Sprintf("%s,%d\n", pairs[t].key, pairs[t].Data))
	}

	log.Default().Println("Program ended at ", time.Now())
}

/*
*
* UTIL FUNCTIONS FOR PROCESSING
*
 */

// function make issue into a string
func issue_to_string(issue issue) string {

	return fmt.Sprintf("Title= %s, author= %s\n",
		issue.Title,
		issue.Author.Login)
}

// function to create pairs
func map_to_pairs(m map[string]int) []pair {

	pairs := make([]pair, 0)

	for e, k := range m {
		pairs = append(pairs, pair{
			key:  e,
			Data: k,
		})
	}

	return pairs
}

// function to sort the pairs
func sort_pairs(pairs []pair) []pair {

	res := make([]pair, 0)

	for range pairs {
		i := get_max_index(pairs)
		res = append(res, pair{
			Data: pairs[i].Data,
			key:  pairs[i].key,
		})
		pairs[i].Data = -1
	}

	return res
}

// helper for pair sorting
func get_max_index(pairs []pair) int {

	max := 0
	index := 0

	for t := range pairs {
		if pairs[t].Data > max {
			max = pairs[t].Data
			index = t
		}
	}

	return index
}
