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
	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
)

var _ plugin.MigrationScript = (*addAccountUsernameToPrComments20230414)(nil)

type prComments202304 struct {
	AccountUsername string `gorm:"index;type:varchar(255)"`
}

func (prComments202304) TableName() string {
	return "pull_request_comments"
}

type addAccountUsernameToPrComments20230414 struct{}

func (script *addAccountUsernameToPrComments20230414) Up(basicRes context.BasicRes) errors.Error {
	return basicRes.GetDal().AutoMigrate(&prComments202304{})
}

func (*addAccountUsernameToPrComments20230414) Version() uint64 {
	return 20230414275114
}

func (*addAccountUsernameToPrComments20230414) Name() string {
	return "add account_username for pull_request_comments"
}
