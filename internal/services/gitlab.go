package services

import (
	"github.com/xanzy/go-gitlab"
)

type GitLabSvc interface {
	GetGroups() ([]int, error)
	GetProjects(groupID int) ([]*Project, error)
	Search(projectID int, keyword string, amount int) ([]*Blob, error)
}

type Project struct {
	ID       int
	Name     string
	Archived bool
}

type Blob struct {
	Path string
	Data string
	Line int
}

type gitLabSvc struct {
	gc *gitlab.Client
}

func NewGitLabService(url, token string) (GitLabSvc, error) {
	gc, err := gitlab.NewClient(token, gitlab.WithBaseURL(url))
	if err != nil {
		return nil, err
	}
	return &gitLabSvc{gc}, nil
}

func (r *gitLabSvc) GetGroups() ([]int, error) {
	groups, _, err := r.gc.Groups.ListGroups(&gitlab.ListGroupsOptions{})
	if err != nil {
		return nil, err
	}

	res := make([]int, 0, len(groups))
	for _, g := range groups {
		res = append(res, g.ID)
	}
	return res, nil
}

func (r *gitLabSvc) GetProjects(groupID int) ([]*Project, error) {
	projects, _, err := r.gc.Groups.ListGroupProjects(groupID, &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 150, // FIXME
		},
	})
	if err != nil {
		return nil, err
	}

	res := make([]*Project, 0, len(projects))
	for _, p := range projects {
		res = append(res, &Project{
			ID:       p.ID,
			Name:     p.NameWithNamespace,
			Archived: p.Archived,
		})
	}
	return res, nil
}

func (r *gitLabSvc) Search(projectID int, keyword string, amount int) ([]*Blob, error) {
	blobs, _, err := r.gc.Search.BlobsByProject(projectID, keyword, &gitlab.SearchOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: amount,
		},
	})
	if err != nil {
		return nil, err
	}

	res := make([]*Blob, 0, len(blobs))
	for _, b := range blobs {
		res = append(res, &Blob{
			Path: b.Filename,
			Data: b.Data,
			Line: b.Startline,
		})
	}
	return res, nil
}
