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

package migrationscripts

import (
	"time"

	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
)

var _ plugin.MigrationScript = (*addAuthoredDateToPrCommits20230415)(nil)

type prCommits202304 struct {
	AuthoredDate time.Time
}

func (prCommits202304) TableName() string {
	return "pull_request_commits"
}

type addAuthoredDateToPrCommits20230415 struct{}

func (script *addAuthoredDateToPrCommits20230415) Up(basicRes context.BasicRes) errors.Error {
	return basicRes.GetDal().AutoMigrate(&prCommits202304{})
}

func (*addAuthoredDateToPrCommits20230415) Version() uint64 {
	return 20230415141275
}

func (*addAuthoredDateToPrCommits20230415) Name() string {
	return "add authored_date for pull_request_commits"
}
