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
var Headers = map[string][][]byte{
	"ASL2": {
		[]byte(`// Licensed to %s under one or more contributor`),
		[]byte(`// license agreements. See the NOTICE file distributed with`),
		[]byte(`// this work for additional information regarding copyright`),
		[]byte(`// ownership. %s licenses this file to you under`),
		[]byte(`// the Apache License, Version 2.0 (the "License"); you may`),
		[]byte(`// not use this file except in compliance with the License.`),
		[]byte(`// You may obtain a copy of the License at`),
		[]byte(`//`),
		[]byte(`//     http://www.apache.org/licenses/LICENSE-2.0`),
		[]byte(`//`),
		[]byte(`// Unless required by applicable law or agreed to in writing,`),
		[]byte(`// software distributed under the License is distributed on an`),
		[]byte(`// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY`),
		[]byte(`// KIND, either express or implied.  See the License for the`),
		[]byte(`// specific language governing permissions and limitations`),
		[]byte(`// under the License.`),
	},
	"ASL2-Short": {
		[]byte(`// Licensed to %s under one or more agreements.`),
		[]byte(`// %s licenses this file to you under the Apache 2.0 License.`),
		[]byte(`// See the LICENSE file in the project root for more information.`),
	},
	"Elastic": {
		[]byte(`// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one`),
		[]byte(`// or more contributor license agreements. Licensed under the Elastic License;`),
		[]byte(`// you may not use this file except in compliance with the Elastic License.`),
	},
	"Elasticv2": {
		[]byte(`// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one`),
		[]byte(`// or more contributor license agreements. Licensed under the Elastic License 2.0;`),
		[]byte(`// you may not use this file except in compliance with the Elastic License 2.0.`),
	},
	"Cloud": {
		[]byte(`// ELASTICSEARCH CONFIDENTIAL`),
		[]byte(`// __________________`),
		[]byte(`//`),
		[]byte(`//  Copyright Elasticsearch B.V. All rights reserved.`),
		[]byte(`//`),
		[]byte(`// NOTICE:  All information contained herein is, and remains`),
		[]byte(`// the property of Elasticsearch B.V. and its suppliers, if any.`),
		[]byte(`// The intellectual and technical concepts contained herein`),
		[]byte(`// are proprietary to Elasticsearch B.V. and its suppliers and`),
		[]byte(`// may be covered by U.S. and Foreign Patents, patents in`),
		[]byte(`// process, and are protected by trade secret or copyright`),
		[]byte(`// law.  Dissemination of this information or reproduction of`),
		[]byte(`// this material is strictly forbidden unless prior written`),
		[]byte(`// permission is obtained from Elasticsearch B.V.`),
	},
}
