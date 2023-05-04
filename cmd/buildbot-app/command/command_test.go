package command

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want *Command
	}{
		{"default", &Command{
			IsMandatory:   true,
			BuilderNames:  []string{},
			CommentAuthor: "",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Command
		wantErr bool
	}{
		{
			"t1",
			"/buildbot",
			New(),
			false,
		},
		{
			"t1",
			"/buildbot force=true",
			func() *Command {
				c := New()
				c.Force = true
				return c
			}(),
			false,
		},
		{
			"t2",
			"/buildbot force=false",
			func() *Command {
				c := New()
				c.Force = false
				return c
			}(),
			false,
		},
		{
			"t3",
			"/buildbot mandatory=true",
			func() *Command {
				c := New()
				c.IsMandatory = true
				return c
			}(),
			false,
		},
		{
			"t4",
			"/buildbot mandatory=false",
			func() *Command {
				c := New()
				c.IsMandatory = false
				return c
			}(),
			false,
		},
		{
			"t5",
			"/buildbot mandatory=true force=true",
			func() *Command {
				c := New()
				c.IsMandatory = true
				c.Force = true
				return c
			}(),
			false,
		},
		{
			"overwrite",
			"/buildbot mandatory=true force=false mandatory=false force=true",
			func() *Command {
				c := New()
				c.IsMandatory = false
				c.Force = true
				return c
			}(),
			false,
		},
		{
			"t6",
			"/buildbot mandatory=false force=false",
			func() *Command {
				c := New()
				c.IsMandatory = false
				c.Force = false
				return c
			}(),
			false,
		},
		{
			"t7",
			"/buildbot mandatory=true force=false",
			func() *Command {
				c := New()
				c.IsMandatory = true
				c.Force = false
				return c
			}(),
			false,
		},
		{
			"t8",
			"/buildbot mandatory=false force=true",
			func() *Command {
				c := New()
				c.IsMandatory = false
				c.Force = true
				return c
			}(),
			false,
		},
		{
			"t9",
			"/buildbot builder=hello",
			func() *Command {
				c := New()
				c.BuilderNames = []string{"hello"}
				return c
			}(),
			false,
		},
		{
			"t10",
			"/buildbot builder=hello builder=world",
			func() *Command {
				c := New()
				c.BuilderNames = []string{"hello", "world"}
				return c
			}(),
			false,
		},
		{
			"t11",
			"/buildbot force=true builder=hello mandatory=false builder=world",
			func() *Command {
				c := New()
				c.BuilderNames = []string{"hello", "world"}
				c.Force = true
				c.IsMandatory = false
				return c
			}(),
			false,
		},
		{
			"t12",
			"/buildbot builder=world builder=hello builder=world",
			func() *Command {
				c := New()
				// NOTE: "world" only appears once and list is sorted
				c.BuilderNames = []string{"hello", "world"}
				return c
			}(),
			false,
		},
		{
			"t13",
			"/buildbot builder=world builder=hello builder=World",
			func() *Command {
				c := New()
				// NOTE: "world" and "World" are listed because the build names are case-sensitive
				c.BuilderNames = []string{"World", "hello", "world"}
				return c
			}(),
			false,
		},
		{
			"t14",
			"/buildbot builder=",
			nil, // not important because we expect an error
			true,
		},
		{
			"t15",
			"/buildbot force=",
			nil, // not important because we expect an error
			true,
		},
		{
			"t16",
			"/buildbot mandatory=",
			nil, // not important because we expect an error
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStripPrefix(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput string
	}{
		{"t1", "/buildbot foo=bar arg1=arg2 ", "foo=bar arg1=arg2"},
		{"t2", "@kwk's /buildbot foo=bar arg1=arg2 ", "foo=bar arg1=arg2"},
		{"t3", "@kwk's /buildbot", ""},
		{"t4", "/buildbot", ""},
		{"t5", "buildbot", "buildbot"},
		{"t6", "foobar", "foobar"},
		{"t7", "foo=bar", "foo=bar"},
		{"t8", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripPrefix(tt.input); got != tt.expectedOutput {
				t.Errorf("StripPrefix() = %v, want %v", got, tt.expectedOutput)
			}
		})
	}
}

func TestStringIsCommand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"t1", "/buildbot mandatory=yes force=true builder=foo builder=bar ", true},
		{"t2", "/buildbot mandatory=no force=false builder=foo builder=bar ", true},
		{"t3", "@kwk's /buildbot mandatory=yes ", false},
		{"t4", "@kwk's /buildbot", false},
		{"t5", "/buildbot", true},
		{"t6", "buildbot", false},
		{"t7", "foobar", false},
		{"t8", "foo=bar", false},
		{"t9", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringIsCommand(tt.input); got != tt.want {
				t.Errorf("StringIsCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommand_ToMap(t *testing.T) {
	type fields struct {
		IsMandatory   bool
		BuilderNames  []string
		CommentAuthor string
		Force         bool
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]interface{}
	}{
		{
			"t1",
			fields{
				IsMandatory:   true,
				BuilderNames:  []string{"foo", "bar"},
				CommentAuthor: "johndoe",
				Force:         false,
			},
			map[string]interface{}{
				CommandOptionBuilder:   []string{"foo", "bar"},
				CommandOptionMandatory: true,
				CommandOptionForce:     false,
			},
		},
		{
			"t2",
			fields{
				IsMandatory:   false,
				BuilderNames:  []string{"hello", "world"},
				CommentAuthor: "janedoe",
				Force:         true,
			},
			map[string]interface{}{
				CommandOptionBuilder:   []string{"hello", "world"},
				CommandOptionMandatory: false,
				CommandOptionForce:     true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Command{
				IsMandatory:   tt.fields.IsMandatory,
				BuilderNames:  tt.fields.BuilderNames,
				CommentAuthor: tt.fields.CommentAuthor,
				Force:         tt.fields.Force,
			}
			if got := c.toMap(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Command.ToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildRegexPattern(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"default", `^/buildbot(\s+|mandatory=(yes|no|true|false|f|t|y|n|0|1)|force=(yes|no|true|false|f|t|y|n|0|1)|builder=(\w+))*$`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildRegexPattern(); got != tt.want {
				t.Errorf("BuildRegexPattern() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValueIsTrue(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
		want bool
	}{
		{`"true"`, "true", true},
		{`"t"`, "t", true},
		{`"yes"`, "yes", true},
		{`"y"`, "y", true},
		{`"1"`, "1", true},

		{`""`, "", false},

		{`0`, 0, false},
		{`1`, 1, false},

		{`"false"`, "false", false},
		{`"f"`, "f", false},
		{`"no"`, "no", false},
		{`"n"`, "n", false},
		{`"0"`, "0", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := valueIsTrue(tt.args); got != tt.want {
				t.Errorf("ValueIsTrue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValueIsFalse(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
		want bool
	}{
		{`"false"`, "false", true},
		{`"f"`, "f", true},
		{`"no"`, "no", true},
		{`"n"`, "n", true},
		{`"0"`, "0", true},

		{`""`, "", false},

		{`0`, 0, false},
		{`1`, 1, false},

		{`"true"`, "true", false},
		{`"t"`, "t", false},
		{`"yes"`, "yes", false},
		{`"y"`, "y", false},
		{`"1"`, "1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := valueIsFalse(tt.args); got != tt.want {
				t.Errorf("ValueIsTrue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommand_ToGithubCheckNameString(t *testing.T) {
	type fields struct {
		IsMandatory   bool
		BuilderNames  []string
		CommentAuthor string
		Force         bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"t1",
			fields{
				IsMandatory:   true,
				BuilderNames:  []string{"foo", "bar"},
				CommentAuthor: "johndoe",
				Force:         false,
			},
			"@johndoe /buildbot mandatory=true force=false builder=[foo bar]",
		},
		{
			"t2",
			fields{
				IsMandatory:   false,
				BuilderNames:  []string{"hello", "world"},
				CommentAuthor: "janedoe",
				Force:         true,
			},
			"@janedoe /buildbot mandatory=false force=true builder=[hello world]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Command{
				IsMandatory:   tt.fields.IsMandatory,
				BuilderNames:  tt.fields.BuilderNames,
				CommentAuthor: tt.fields.CommentAuthor,
				Force:         tt.fields.Force,
			}
			if got := c.ToGithubCheckNameString(); got != tt.want {
				t.Errorf("Command.ToGithubCheckNameString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseIntoMap(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]interface{}
		wantErr bool
	}{
		{
			"t1",
			"/buildbot",
			map[string]interface{}{},
			false,
		},
		{
			"t2",
			"/buildbot force=true",
			map[string]interface{}{
				CommandOptionForce: "true",
			},
			false,
		},
		{
			"t3",
			"/buildbot force=false",
			map[string]interface{}{
				CommandOptionForce: "false",
			},
			false,
		},
		{
			"t4",
			"/buildbot mandatory=true",
			map[string]interface{}{
				CommandOptionMandatory: "true",
			},
			false,
		},
		{
			"t5",
			"/buildbot mandatory=false",
			map[string]interface{}{
				CommandOptionMandatory: "false",
			},
			false,
		},
		{
			"t6",
			"/buildbot builder=hello builder=world",
			map[string]interface{}{
				CommandOptionBuilder: []string{"hello", "world"},
			},
			false,
		},
		{
			"t7",
			"/buildbot builder=hello force=true builder=world mandatory=false",
			map[string]interface{}{
				CommandOptionMandatory: "false",
				CommandOptionForce:     "true",
				CommandOptionBuilder:   []string{"hello", "world"},
			},
			false,
		},
		{
			"t8",
			"/buildbot builder=hello force=false builder=world mandatory=true",
			map[string]interface{}{
				CommandOptionMandatory: "true",
				CommandOptionForce:     "false",
				CommandOptionBuilder:   []string{"hello", "world"},
			},
			false,
		},
		{
			"t9",
			"/buildbot foo=bar",
			nil, // output not relevant because of error
			true,
		},
		{
			"t10",
			"/buildbot builder=",
			nil, // output not relevant because of error
			true,
		},
		{
			"t11",
			"/buildbot force=",
			nil, // output not relevant because of error
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIntoMap(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseIntoMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseIntoMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
