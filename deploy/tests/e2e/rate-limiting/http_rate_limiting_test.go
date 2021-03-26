// Copyright 2019 HAProxy Technologies LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build e2e_parallel

package ratelimiting

import (
	"fmt"

	"github.com/haproxytech/kubernetes-ingress/deploy/tests/e2e"
)

func (suite *RateLimitingSuite) Test_Rate_Limiting() {
	for testName, tc := range map[string]struct {
		limitPeriodinSeconds int
		limitRequests        int
		customStatusCode     int
	}{
		"5req5s":     {5, 5, 403},
		"100req10s":  {10, 100, 403},
		"customcode": {5, 1, 429},
	} {
		suite.Run(testName, func() {
			suite.tmplData.Host = testName + ".test"
			suite.tmplData.IngAnnotations = []struct{ Key, Value string }{
				{"rate-limit-period", fmt.Sprintf("%ds", tc.limitPeriodinSeconds)},
				{"rate-limit-requests", fmt.Sprintf("%d", tc.limitRequests)},
				{"rate-limit-status-code", fmt.Sprintf("%d", tc.customStatusCode)},
			}
			suite.Require().NoError(suite.test.DeployYamlTemplate("config/ingress.yaml.tmpl", suite.test.GetNS(), suite.tmplData))
			suite.Require().Eventually(func() bool {
				var counter, responseCode int
				suite.client.Host = suite.tmplData.Host
				for responseCode != tc.customStatusCode {
					res, cls, err := suite.client.Do()
					if err != nil {
						suite.FailNow(err.Error())
					}
					defer cls()
					if res.StatusCode == 200 {
						counter++
					}
					responseCode = res.StatusCode
				}
				return counter == tc.limitRequests
			}, e2e.WaitDuration, e2e.TickDuration)
		})
	}
}
