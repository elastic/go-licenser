// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package licensing

// Headers is the map of supported licenses
var Headers = map[string][]string{
	"ASL2": {
		`// Licensed to %s under one or more contributor`,
		`// license agreements. See the NOTICE file distributed with`,
		`// this work for additional information regarding copyright`,
		`// ownership. %s licenses this file to you under`,
		`// the Apache License, Version 2.0 (the "License"); you may`,
		`// not use this file except in compliance with the License.`,
		`// You may obtain a copy of the License at`,
		`//`,
		`//     http://www.apache.org/licenses/LICENSE-2.0`,
		`//`,
		`// Unless required by applicable law or agreed to in writing,`,
		`// software distributed under the License is distributed on an`,
		`// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY`,
		`// KIND, either express or implied.  See the License for the`,
		`// specific language governing permissions and limitations`,
		`// under the License.`,
	},
	"ASL2-Short": {
		`// Licensed to %s under one or more agreements.`,
		`// %s licenses this file to you under the Apache 2.0 License.`,
		`// See the LICENSE file in the project root for more information.`,
	},
	"Elastic": {
		`// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one`,
		`// or more contributor license agreements. Licensed under the Elastic License;`,
		`// you may not use this file except in compliance with the Elastic License.`,
	},
	"Elasticv2": {
		`// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one`,
		`// or more contributor license agreements. Licensed under the Elastic License 2.0;`,
		`// you may not use this file except in compliance with the Elastic License 2.0.`,
	},
	"Cloud": {
		`// ELASTICSEARCH CONFIDENTIAL`,
		`// __________________`,
		`//`,
		`//  Copyright Elasticsearch B.V. All rights reserved.`,
		`//`,
		`// NOTICE:  All information contained herein is, and remains`,
		`// the property of Elasticsearch B.V. and its suppliers, if any.`,
		`// The intellectual and technical concepts contained herein`,
		`// are proprietary to Elasticsearch B.V. and its suppliers and`,
		`// may be covered by U.S. and Foreign Patents, patents in`,
		`// process, and are protected by trade secret or copyright`,
		`// law.  Dissemination of this information or reproduction of`,
		`// this material is strictly forbidden unless prior written`,
		`// permission is obtained from Elasticsearch B.V.`,
	},
}
