//go:build windows

package powershell

func (s *PowershellTestSuite) TestErrUnsupportedDataType() {

	err1 := &ErrUnsupportedDataType{Message: "foo"}

	s.Require().ErrorIs(err1, &ErrUnsupportedDataType{})
}

func (s *PowershellTestSuite) TestBuildCmdlet() {

	tests := []struct {
		cmdlet      string
		args        map[string]any
		expectOneOf []string // Map to allow for unordered args
	}{
		{
			cmdlet: "Get-ChildItem",
			args: map[string]any{
				"Recurse": nil,
				"Path":    "C:\\Windows",
			},
			expectOneOf: []string{
				"Get-ChildItem -Recurse -Path \"C:\\Windows\"",
				"Get-ChildItem -Path \"C:\\Windows\" -Recurse",
			},
		},
		{
			cmdlet: "Get-Foo",
			args: map[string]any{
				"-MultiString": []string{
					"a",
					"b",
					"c",
				},
			},
			expectOneOf: []string{
				"Get-Foo -MultiString \"a\",\"b\",\"c\"",
			},
		},
		{
			cmdlet: "Get-Foo",
			args: map[string]any{
				"-Int": 1,
			},
			expectOneOf: []string{
				"Get-Foo -Int 1",
			},
		},
		{
			cmdlet: "Get-Foo",
			args: map[string]any{
				"-MultiInt": []int{
					1,
					2,
					3,
				},
			},
			expectOneOf: []string{
				"Get-Foo -MultiInt 1,2,3",
			},
		},
		{
			cmdlet: "Get-Foo",
			args: map[string]any{
				"-MultiInt64": []int64{
					1,
					2,
					3,
				},
			},
			expectOneOf: []string{
				"Get-Foo -MultiInt64 1,2,3",
			},
		},
		{
			cmdlet: "Get-Foo",
			args: map[string]any{
				"-F64": 1.2,
			},
			expectOneOf: []string{
				"Get-Foo -F64 1.2",
			},
		},
		{
			cmdlet: "Get-Foo",
			args: map[string]any{
				"-Bool": true,
			},
			expectOneOf: []string{
				"Get-Foo -Bool $true",
			},
		},
		{
			cmdlet: "Get-Foo",
			args: map[string]any{
				"MultiFloat": []float32{
					1.2,
					2,
					3,
				},
			},
			expectOneOf: []string{
				"Get-Foo -MultiFloat 1.2,2,3",
			},
		},
	}

	for _, test := range tests {
		s.Run(test.expectOneOf[0], func() {
			c, err := buildCmdlet(test.cmdlet, test.args)
			s.Require().NoError(err)
			s.Require().Contains(test.expectOneOf, c)
		})
	}
}
