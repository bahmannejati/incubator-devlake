/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

import React from 'react';
import { Checkbox, FormGroup, InputGroup } from '@blueprintjs/core';

interface Props {
  transformation: any;
  setTransformation: React.Dispatch<React.SetStateAction<any>>;
}

export const Reviewer = ({ transformation, setTransformation }: Props) => {
  const handleAuthorExcludeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setTransformation({
      ...transformation,
      excludeAuthorAsFirstReviewer: e.target.checked,
    });
  };

  const handleBotsUsernameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setTransformation({
      ...transformation,
      excludedBotsAsFirstReviewer: e.target.value,
    });
  };

  return (
    <>
      <h3>Exclude reviewers</h3>
      <p>These configs are used to detect first review on pull requests to calculate Review Time metric for DORA.</p>
      <Checkbox
        label="Exclude author as first reviewer"
        checked={transformation.excludeAuthorAsFirstReviewer}
        onChange={handleAuthorExcludeChange}
      />
      <p>Exclude bots username as first reviewer</p>
      <FormGroup inline label="Usernames">
        <InputGroup
          placeholder="Comma separated usernames"
          value={transformation.excludedBotsAsFirstReviewer}
          onChange={handleBotsUsernameChange}
        />
      </FormGroup>
    </>
  );
};
