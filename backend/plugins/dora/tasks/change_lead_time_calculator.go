/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tasks

import (
	"fmt"
	"reflect"
	"time"

	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models/domainlayer/code"
	"github.com/apache/incubator-devlake/core/models/domainlayer/crossdomain"
	"github.com/apache/incubator-devlake/core/models/domainlayer/devops"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
)

var workHourStart = 10
var workHourEnd = 20

func isWeekend(t time.Time) bool {
	return t.Weekday() == time.Friday || t.Weekday() == time.Thursday
}
func isNonWorkingHour(t time.Time) bool {
	hour := t.Hour()
	return hour < workHourStart || hour > workHourEnd
}

func calculateTime(start time.Time, end time.Time) time.Duration {

	var duration time.Duration
	period := 24 * time.Hour

	var iterationStartDate = time.Date(start.Year(), start.Month(), start.Day(), workHourStart, 0, 0, 0, start.Location())
	var iterationEndDate = time.Date(end.Year(), end.Month(), end.Day(), workHourEnd, 0, 0, 0, end.Location())

	fmt.Println(start.Weekday(), end.Weekday())
	for current := iterationStartDate; current.Before(iterationEndDate); current = current.Add(period) {

		if isWeekend(current) {
			continue
		}

		var startDate time.Time
		var endDate time.Time

		if current.Before(start) {
			startDate = start
		} else {
			startDate = current
		}

		var currentEndDate = current.Add(time.Duration(workHourEnd-workHourStart) * time.Hour)
		if currentEndDate.After(end) {
			endDate = end
		} else {
			endDate = currentEndDate
		}

		duration += endDate.Sub(startDate)

	}

	return duration
}
func CalculateChangeLeadTime(taskCtx plugin.SubTaskContext) errors.Error {
	db := taskCtx.GetDal()
	logger := taskCtx.GetLogger()
	data := taskCtx.GetData().(*DoraTaskData)
	// construct a list of tuple[task, oldPipelineCommitSha, newPipelineCommitSha, taskFinishedDate]
	deploymentClause := []dal.Clause{
		dal.Select(`ct.id as task_id, cpc.commit_sha as new_deploy_commit_sha,
			ct.finished_date as task_finished_date, cpc.repo_id as repo_id`),
		dal.From(`cicd_tasks ct`),
		dal.Join(`left join cicd_pipeline_commits cpc on ct.pipeline_id = cpc.pipeline_id`),
		dal.Join(`left join project_mapping pm on pm.row_id = ct.cicd_scope_id`),
		dal.Where(`ct.environment = ? and ct.type = ? and ct.result = ? and pm.project_name = ? and pm.table = ?`,
			devops.PRODUCTION, devops.DEPLOYMENT, devops.SUCCESS, data.Options.ProjectName, "cicd_scopes"),
		dal.Orderby(`cpc.repo_id, ct.started_date `),
	}
	deploymentDiffPairs := make([]deploymentPair, 0)
	err := db.All(&deploymentDiffPairs, deploymentClause...)
	if err != nil {
		return err
	}
	// deploymentDiffPairs[i-1].NewDeployCommitSha is deploymentDiffPairs[i].OldDeployCommitSha
	oldDeployCommitSha := ""
	lastRepoId := ""
	for i := 0; i < len(deploymentDiffPairs); i++ {
		// if two deployments belong to different repo, let's skip
		if lastRepoId == deploymentDiffPairs[i].RepoId {
			deploymentDiffPairs[i].OldDeployCommitSha = oldDeployCommitSha
		} else {
			lastRepoId = deploymentDiffPairs[i].RepoId
		}
		oldDeployCommitSha = deploymentDiffPairs[i].NewDeployCommitSha
	}

	// get prs by repo project_name
	clauses := []dal.Clause{
		dal.From(&code.PullRequest{}),
		dal.Join(`left join project_mapping pm on pm.row_id = pull_requests.base_repo_id`),
		dal.Where("pull_requests.merged_date IS NOT NULL and pm.project_name = ? and pm.table = ?", data.Options.ProjectName, "repos"),
	}
	cursor, err := db.Cursor(clauses...)
	if err != nil {
		return err
	}
	defer cursor.Close()

	converter, err := api.NewDataConverter(api.DataConverterArgs{
		RawDataSubTaskArgs: api.RawDataSubTaskArgs{
			Ctx: taskCtx,
			Params: DoraApiParams{
				ProjectName: data.Options.ProjectName,
			},
			Table: "pull_requests",
		},
		BatchSize:    100,
		InputRowType: reflect.TypeOf(code.PullRequest{}),
		Input:        cursor,
		Convert: func(inputRow interface{}) ([]interface{}, errors.Error) {
			pr := inputRow.(*code.PullRequest)
			firstPrCommit, err := getFirstPrCommit(pr.Id, db)
			if err != nil {
				return nil, err
			}
			projectPrMetric := &crossdomain.ProjectPrMetric{}
			projectPrMetric.Id = pr.Id
			projectPrMetric.ProjectName = data.Options.ProjectName
			if err != nil {
				return nil, err
			}

			if firstPrCommit != nil {
				codingTime := int64(calculateTime(firstPrCommit.AuthoredDate, pr.CreatedDate).Seconds())

				if codingTime/60 == 0 && codingTime%60 > 0 {
					codingTime = 1
				} else {
					codingTime = codingTime / 60
				}
				projectPrMetric.PrCodingTime = processNegativeValue(codingTime)
				projectPrMetric.FirstCommitSha = firstPrCommit.CommitSha
			}
			firstReview, err := getFirstReview(
				pr.Id,
				pr.AuthorId,
				db,
				data.Options.TransformationRules.ExcludedBotsAsFirstReviewer,
			)
			if err != nil {
				return nil, err
			}
			// clauses filter by merged_date IS NOT NULL, so MergedDate must be not nil.
			prDuring := processNegativeValue(int64(calculateTime(pr.CreatedDate, *pr.MergedDate).Minutes()))
			if firstReview != nil && int64(pr.MergedDate.Sub(firstReview.CreatedDate).Seconds()) > 0 {
				projectPrMetric.PrPickupTime = processNegativeValue(int64(calculateTime(pr.CreatedDate, firstReview.CreatedDate).Minutes()))
				projectPrMetric.PrReviewTime = processNegativeValue(int64(calculateTime(firstReview.CreatedDate, *pr.MergedDate).Minutes()))
				projectPrMetric.FirstReviewId = firstReview.Id
			} else {
				projectPrMetric.PrReviewTime = prDuring
			}
			deployment, err := getDeployment(pr.MergeCommitSha, pr.BaseRepoId, deploymentDiffPairs, db)
			if err != nil {
				return nil, err
			}
			if deployment != nil && deployment.TaskFinishedDate != nil {
				timespan := deployment.TaskFinishedDate.Sub(*pr.MergedDate)
				projectPrMetric.PrDeployTime = processNegativeValue(int64(timespan.Minutes()))
				projectPrMetric.DeploymentId = deployment.TaskId
			} else {
				logger.Debug("deploy time of pr %v is nil\n", pr.PullRequestKey)
			}
			projectPrMetric.PrCycleTime = nil
			var result int64
			if projectPrMetric.PrCodingTime != nil {
				result += *projectPrMetric.PrCodingTime
			}
			if prDuring != nil {
				result += *prDuring
			}
			if projectPrMetric.PrDeployTime != nil {
				result += *projectPrMetric.PrDeployTime
			}
			if result > 0 {
				projectPrMetric.PrCycleTime = &result
			}
			return []interface{}{projectPrMetric}, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}

func getFirstPrCommit(prId string, db dal.Dal) (*code.PullRequestCommit, errors.Error) {
	prCommit := &code.PullRequestCommit{}
	prCommitClauses := []dal.Clause{
		dal.From(&code.PullRequestCommit{}),
		dal.Where("pull_request_commits.pull_request_id = ?", prId),
		dal.Orderby("pull_request_commits.authored_date ASC"),
	}
	err := db.First(prCommit, prCommitClauses...)
	if db.IsErrorNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return prCommit, nil
}

func getFirstReview(
	prId string,
	prCreator string,
	db dal.Dal,
	excludedBotsAsFirstReviewer string,
) (*code.PullRequestComment, errors.Error) {
	review := &code.PullRequestComment{}
	commentClauses := []dal.Clause{
		dal.From(&code.PullRequestComment{}),
		dal.Where("pull_request_id = ? and account_id != ? and account_username NOT IN (?)", prId, prCreator, excludedBotsAsFirstReviewer),
		dal.Orderby("created_date ASC"),
	}
	err := db.First(review, commentClauses...)
	if db.IsErrorNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return review, nil
}

func getDeployment(mergeSha string, repoId string, deploymentPairList []deploymentPair, db dal.Dal) (*deploymentPair, errors.Error) {
	// ignore environment at this point because detecting it by name is obviously not engouh
	// take https://github.com/apache/incubator-devlake/actions/workflows/build.yml for example
	// one can not distingush testing/production by looking at the job name solely.
	commitDiff := &code.CommitsDiff{}
	// find if tuple[merge_sha, new_commit_sha, old_commit_sha] exist in commits_diffs, if yes, return pair.FinishedDate
	for _, pair := range deploymentPairList {
		if repoId != pair.RepoId {
			continue
		}
		err := db.First(commitDiff, dal.Where(`commit_sha = ? and new_commit_sha = ? and old_commit_sha = ?`,
			mergeSha, pair.NewDeployCommitSha, pair.OldDeployCommitSha))
		if err == nil {
			return &pair, nil
		}
		if db.IsErrorNotFound(err) {
			continue
		}
		if err != nil {
			return nil, err
		}

	}
	return nil, nil
}

func processNegativeValue(v int64) *int64 {
	if v > 0 {
		return &v
	} else {
		return nil
	}
}

var CalculateChangeLeadTimeMeta = plugin.SubTaskMeta{
	Name:             "calculateChangeLeadTime",
	EntryPoint:       CalculateChangeLeadTime,
	EnabledByDefault: true,
	Description:      "Calculate change lead time",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_CICD, plugin.DOMAIN_TYPE_CODE},
}

type deploymentPair struct {
	TaskId             string
	RepoId             string
	NewDeployCommitSha string
	OldDeployCommitSha string
	TaskFinishedDate   *time.Time
}
