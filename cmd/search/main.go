package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/tzuhsitseng/gitlab-search/internal/helpers"
	"github.com/tzuhsitseng/gitlab-search/internal/services"
)

const (
	DelayCallSeconds = 6
	MaxSearchResults = 10
)

func main() {
	var url string
	var token string
	var keyword string

	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Perform a thorough search of your GitLab projects",
		Long:  "Perform a thorough search of your GitLab projects",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Your url: [%s]\n", url)
			fmt.Printf("Your token: [%s]\n", token)
			fmt.Printf("Your keyword: [%s]\n", keyword)

			svc, err := services.NewGitLabService(url, token)
			if err != nil {
				log.Fatalf("failed to create gitlab client: %v", err)
			}

			// get groups
			groupIDs, err := getGroups(svc)
			if err != nil {
				log.Fatalf("failed to get groups: %v", err)
			}

			// get projects
			projects, err := getProjects(svc, groupIDs)
			if err != nil {
				log.Fatalf("failed to get projects: %v", err)
			}

			// do search
			for _, p := range projects {
				blobs, err := svc.Search(p.ID, keyword, MaxSearchResults+1)
				if err != nil {
					log.Fatalf("failed to search keyword: %v", err)
				}

				printResults(p.Name, blobs)
				time.Sleep(DelayCallSeconds * time.Second)
			}
		},
	}
	searchCmd.Flags().StringVarP(&url, "url", "u", "", "gitlab url")
	searchCmd.Flags().StringVarP(&token, "token", "t", "", "personal access token")
	searchCmd.Flags().StringVarP(&keyword, "keyword", "k", "", "search keyword")

	rootCmd := &cobra.Command{Use: "gs"}
	rootCmd.AddCommand(searchCmd)
	rootCmd.Execute()
}

func printResults(projectName string, blobs []*services.Blob) {
	if len(blobs) > 0 {
		var size string
		var comment string

		if len(blobs) > MaxSearchResults {
			size = fmt.Sprintf("%d+", MaxSearchResults)
			comment = fmt.Sprintf("(only show %d results)", MaxSearchResults)
		} else {
			size = strconv.Itoa(len(blobs))
		}

		fmt.Printf("🔍 Project [%s] has [%s] results %s\n\n", projectName, size, comment)

		for i := 0; i < helpers.Min(MaxSearchResults, len(blobs)); i++ {
			b := blobs[i]
			fmt.Printf("👉 %s\n\n", b.Path)
			fmt.Printf("```#L%d\n", b.Line)
			fmt.Printf("%s\n", strings.Trim(strings.Replace(b.Data, "\t", "  ", -1), "\n"))
			fmt.Printf("```\n\n")
		}
	} else {
		fmt.Printf("🔍 Project [%s] has no code results\n\n", projectName)
	}
}

func getProjects(svc services.GitLabSvc, groupIDs []int) ([]*services.Project, error) {
	res := make([]*services.Project, 0)
	for _, gid := range groupIDs {
		projects, err := svc.GetProjects(gid)
		if err != nil {
			return nil, err
		}
		res = append(res, projects...)
	}
	fmt.Printf("There are [%d] projects\n", len(res))
	return res, nil
}

func getGroups(svc services.GitLabSvc) ([]int, error) {
	groupIDs, err := svc.GetGroups()
	if err != nil {
		return nil, err
	}
	fmt.Printf("There are [%d] groups\n", len(groupIDs))
	return groupIDs, nil
}
