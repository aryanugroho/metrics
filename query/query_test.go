// Copyright 2015 Square Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package query

// show all metrics
// show tags WHERE predicate

import (
	"github.com/square/metrics/assert"
	"testing"
)

// these queries should successfully parse,
// with a corresponding command.
var inputs = []string{
	// describes
	"describe all",
	"describe x",
	"describe cpu_usage",
	"describe cpu_usage where key = 'value'",
	"describe cpu_usage where key = 'value\\''",
	"describe cpu_usage where key != 'value'",
	"describe cpu_usage where (key = 'value')",
	"describe cpu_usage where not (key = 'value')",
	"describe cpu_usage where not key = 'value'",
	"describe cpu_usage where (key = 'value' and key = 'value')",
	"describe cpu_usage where key = 'value' or key = 'value'",
	"describe cpu_usage where key in ('value', 'value')",
	"describe cpu_usage where key matches 'abc'",
	"describe nodes.cpu.usage where datacenter='sjc1b' and type='idle' and host matches 'fwd'",
}

var parseOnly = []string{
	// selects - trying out arithmetic
	"select x from 0 to 0",
	"select x-y-z from 0 to 0",
	"select (x)-(y)-(z) from 0 to 0",
	"select 0 from 0 to 0",
	"select x, y from 0 to 0",
	"select 1 + 2 * 3 + 4 from 0 to 0",
	"select x * (y + 123), z from 0 to 0",
	// testing escaping
	"select `x` from 0 to 0",
	// selects - timestamps
	"select x * (y + 123), z from '2014-01-01' to '2014-01-02'",
	"select x * (y + 123), z from 0 to 10000",
	// selects - aggregate functions
	"select scalar.max(x) from 0 to 0",
	"select scalar.max(x) from 0 to 0",
	"select aggregate.max(x, y) from 0 to 0",
	"select aggregate.max(x group by foo) + 3 from 0 to 0",
	// selects - where clause
	"select x where y = 'z' from 0 to 0",
	// selects - per-identifier where clause
	"select x + z[y = 'z'] from 0 to 0",
	"select x[y = 'z'] from 0 to 0",
	// selects - complicated queries
	"select aggregate.max(x[y = 'z'] group by foo) from 0 to 0",
	"select cpu.user + cpu.kernel where host = 'apa3.sjc2b' from 0 to 0",
}

// these queries should fail with a syntax error.
var syntaxErrorQuery = []string{
	"select (",
	"select )",
	"describe (",
	"describe invalid_regex where key matches 'ab['",
}

func TestParse_success(t *testing.T) {
	for _, row := range inputs {
		if err := checkSyntaxError(t, row); err != nil {
			t.Errorf("[%s] failed to parse: %s", row, err.Error())
		}
	}
	for _, row := range parseOnly {
		if err := checkSyntaxError(t, row); err != nil {
			t.Errorf("[%s] failed to parse: %s", row, err.Error())
		}
	}
}

func TestParse_syntaxError(t *testing.T) {
	for _, row := range syntaxErrorQuery {
		_, err := Parse(row)
		if err == nil {
			t.Errorf("[%s] should have failed to parse", row)
		} else if _, ok := err.(SyntaxErrors); !ok {
			t.Logf("[%s] Expected SyntaxErrors, got: %s", row, err.Error())
		}
	}
}

func TestCompile(t *testing.T) {
	a := assert.New(t)
	for _, row := range inputs {
		p := Parser{Buffer: row}
		p.Init()
		a.CheckError(p.Parse())
		p.Execute()
		testParserResult(t, p)
	}
}

// Helper functions
// ================

func checkSyntaxError(t *testing.T, input string) error {
	p := Parser{Buffer: input}
	p.Init()
	return p.Parse()
}

func testParserResult(t *testing.T, p Parser) {
	a := assert.New(t)
	a.EqInt(len(p.nodeStack), 0)
	a.EqInt(len(p.errors), 0)
}
