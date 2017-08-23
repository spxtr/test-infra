/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tide

import (
	"context"
	"fmt"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/shurcooL/githubql"

	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/kube"
)

func testPullsMatchList(t *testing.T, test string, actual []pullRequest, expected []int) {
	if len(actual) != len(expected) {
		t.Errorf("Wrong size for case %s. Got PRs %+v, wanted numbers %v.", test, actual, expected)
		return
	}
	for _, pr := range actual {
		var found bool
		n1 := int(pr.Number)
		for _, n2 := range expected {
			if n1 == n2 {
				found = true
			}
		}
		if !found {
			t.Errorf("For case %s, found PR %d but shouldn't have.", test, n1)
		}
	}
}

func TestAccumulate(t *testing.T) {
	presubmits := []string{"job1", "job2"}
	testPulls := []int{1, 2, 3, 4, 5, 6, 7}
	testPJs := []struct {
		prNumber int
		job      string
		state    kube.ProwJobState
	}{
		{2, "job1", kube.PendingState},
		{3, "job1", kube.PendingState},
		{3, "job2", kube.TriggeredState},
		{4, "job1", kube.FailureState},
		{4, "job2", kube.PendingState},
		{5, "job1", kube.PendingState},
		{5, "job2", kube.FailureState},
		{5, "job2", kube.PendingState},
		{6, "job1", kube.SuccessState},
		{6, "job2", kube.PendingState},
		{7, "job1", kube.SuccessState},
		{7, "job2", kube.SuccessState},
		{7, "job1", kube.FailureState},
	}
	var pulls []pullRequest
	for _, p := range testPulls {
		pulls = append(pulls, pullRequest{Number: githubql.Int(p)})
	}
	var pjs []kube.ProwJob
	for _, pj := range testPJs {
		pjs = append(pjs, kube.ProwJob{
			Spec:   kube.ProwJobSpec{Job: pj.job, Refs: kube.Refs{Pulls: []kube.Pull{{Number: pj.prNumber}}}},
			Status: kube.ProwJobStatus{State: pj.state},
		})
	}
	successes, pendings, nones := accumulate(presubmits, pulls, pjs)
	testPullsMatchList(t, "successes", successes, []int{7})
	testPullsMatchList(t, "pendings", pendings, []int{3, 5, 6})
	testPullsMatchList(t, "nones", nones, []int{1, 2, 4})
}

type fgc struct {
	refs map[string]string
}

func (f *fgc) GetRef(o, r, ref string) (string, error) {
	return f.refs[o+"/"+r+" "+ref], nil
}

func (f *fgc) Query(ctx context.Context, q interface{}, vars map[string]interface{}) error {
	return nil
}

func (f *fgc) Merge(org, repo string, number int, details github.MergeDetails) error {
	return nil
}

// TestDividePool ensures that subpools returned by dividePool satisfy a few
// important invariants.
func TestDividePool(t *testing.T) {
	testPulls := []struct {
		org    string
		repo   string
		number int
		branch string
	}{
		{
			org:    "k",
			repo:   "t-i",
			number: 5,
			branch: "master",
		},
		{
			org:    "k",
			repo:   "t-i",
			number: 6,
			branch: "master",
		},
		{
			org:    "k",
			repo:   "k",
			number: 123,
			branch: "master",
		},
		{
			org:    "k",
			repo:   "k",
			number: 1000,
			branch: "release-1.6",
		},
	}
	testPJs := []struct {
		jobType kube.ProwJobType
		org     string
		repo    string
		baseRef string
		baseSHA string
	}{
		{
			jobType: kube.PresubmitJob,
			org:     "k",
			repo:    "t-i",
			baseRef: "master",
			baseSHA: "123",
		},
		{
			jobType: kube.BatchJob,
			org:     "k",
			repo:    "t-i",
			baseRef: "master",
			baseSHA: "123",
		},
		{
			jobType: kube.PeriodicJob,
		},
		{
			jobType: kube.PresubmitJob,
			org:     "k",
			repo:    "t-i",
			baseRef: "patch",
			baseSHA: "123",
		},
		{
			jobType: kube.PresubmitJob,
			org:     "k",
			repo:    "t-i",
			baseRef: "master",
			baseSHA: "abc",
		},
		{
			jobType: kube.PresubmitJob,
			org:     "o",
			repo:    "t-i",
			baseRef: "master",
			baseSHA: "123",
		},
		{
			jobType: kube.PresubmitJob,
			org:     "k",
			repo:    "other",
			baseRef: "master",
			baseSHA: "123",
		},
	}
	fc := &fgc{
		refs: map[string]string{"k/t-i heads/master": "123"},
	}
	c := &Controller{
		log: logrus.NewEntry(logrus.StandardLogger()),
		ghc: fc,
	}
	var pulls []pullRequest
	for _, p := range testPulls {
		npr := pullRequest{Number: githubql.Int(p.number)}
		npr.BaseRef.Name = githubql.String(p.branch)
		npr.BaseRef.Prefix = "refs/heads/"
		npr.Repository.Name = githubql.String(p.repo)
		npr.Repository.Owner.Login = githubql.String(p.org)
		pulls = append(pulls, npr)
	}
	var pjs []kube.ProwJob
	for _, pj := range testPJs {
		pjs = append(pjs, kube.ProwJob{
			Spec: kube.ProwJobSpec{
				Type: pj.jobType,
				Refs: kube.Refs{
					Org:     pj.org,
					Repo:    pj.repo,
					BaseRef: pj.baseRef,
					BaseSHA: pj.baseSHA,
				},
			},
		})
	}
	sps, err := c.dividePool(pulls, pjs)
	if err != nil {
		t.Fatalf("Error dividing pool: %v", err)
	}
	if len(sps) == 0 {
		t.Error("No subpools.")
	}
	for _, sp := range sps {
		name := fmt.Sprintf("%s/%s %s", sp.org, sp.repo, sp.branch)
		sha := fc.refs[sp.org+"/"+sp.repo+" heads/"+sp.branch]
		if sp.sha != sha {
			t.Errorf("For subpool %s, got sha %s, expected %s.", name, sp.sha, sha)
		}
		if len(sp.prs) == 0 {
			t.Errorf("Subpool %s has no PRs.", name)
		}
		for _, pr := range sp.prs {
			if string(pr.Repository.Owner.Login) != sp.org || string(pr.Repository.Name) != sp.repo || string(pr.BaseRef.Name) != sp.branch {
				t.Errorf("PR in wrong subpool. Got PR %+v in subpool %s.", pr, name)
			}
		}
		for _, pj := range sp.pjs {
			if pj.Spec.Type != kube.PresubmitJob && pj.Spec.Type != kube.BatchJob {
				t.Errorf("PJ with bad type in subpool %s: %+v", name, pj)
			}
			if pj.Spec.Refs.Org != sp.org || pj.Spec.Refs.Repo != sp.repo || pj.Spec.Refs.BaseRef != sp.branch || pj.Spec.Refs.BaseSHA != sp.sha {
				t.Errorf("PJ in wrong subpool. Got PJ %+v in subpool %s.", pj, name)
			}
		}
	}
}
