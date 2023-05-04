package command

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// tag::command_options[]
const (
	// BuildbotCommand is the command that triggers the buildbot workflow in a
	// GitHub comment.
	BuildbotCommand = "/buildbot"

	// CommandOptionMandatory is the boolean option to make a check run
	// mandatory.
	CommandOptionMandatory = "mandatory"

	// CommandOptionBuilder is the option that can be used multiple times in a
	// command comment. The resulting builders will be a case-sensitive, sorted
	// list of builders with no duplicates.
	CommandOptionBuilder = "builder"

	// CommandOptionForce is the boolean option to enforce a new build even if
	// one is already present.
	CommandOptionForce = "force"
)

// end::command_options[]

// TODO(kwk): Maybe we can have a ResolveBuilderNames() method in order to have
// a more highlevel selection process for builder names. For example, if
// somebody narrows down a command to run on linux operating systems (os=linux)
// with an aarch64 architecture (arch=aarch64), we could resolve this to a list
// of builder names available.

// tag::command[]
// A Command represents all information about a /buildbot command
type Command struct {
	// When true, the command has to pass for the PR in order pass gating
	// (default: true).
	IsMandatory bool
	// Case-sensitive, sorted list of builders without duplicates to run build on.
	// TODO(kwk): Maybe we can default to something reasonable here?
	BuilderNames []string
	// The user's GitHub login that issued the /buildbot comment
	CommentAuthor string
	// When true, we'll try to run the build even if the PR has already been
	// tested at this stage (default: false).
	Force bool
}

// end::command[]

// New returns a Command object with defaults applied
func New() *Command {
	return &Command{
		IsMandatory:   true,
		BuilderNames:  []string{},
		CommentAuthor: "",
		Force:         false,
	}
}

// From creates a new command object from the given string.
func FromString(s string) (*Command, error) {
	if !StringIsCommand(s) {
		return nil, fmt.Errorf("string is no valid command: %s", s)
	}
	args, err := parseIntoMap(s)
	if err != nil {
		return nil, fmt.Errorf("failed to parse string into command map: %s", s)
	}

	// Get a command with proper defaults
	cmd := New()

	if mandatory, ok := args[CommandOptionMandatory]; ok {
		cmd.IsMandatory = valueIsTrue(mandatory)
	}
	if force, ok := args[CommandOptionForce]; ok {
		cmd.Force = valueIsTrue(force)
	}
	if builderNames, ok := args[CommandOptionBuilder]; ok {
		builderNamesArr, ok := builderNames.([]string)
		if !ok {
			return nil, fmt.Errorf("invalud builder names: %+v", builderNames)
		}
		cmd.BuilderNames = builderNamesArr
	}

	return cmd, nil
}

// ToGithubCheckNameString returns a string representation to be used as a GitHub check
// run name.
func (c Command) ToGithubCheckNameString() string {
	return fmt.Sprintf("@%s %s %s=%t %s=%t %s=%s", c.CommentAuthor, BuildbotCommand, CommandOptionMandatory, c.IsMandatory, CommandOptionForce, c.Force, CommandOptionBuilder, c.BuilderNames)
}

// ToTryBotPropertyArray returns a string array with properties set to be passed
// along to a "buildbot try" command.
func (c Command) ToTryBotPropertyArray() []string {
	return []string{
		fmt.Sprintf("--property=command_is_mandatory=%t", c.IsMandatory),
		fmt.Sprintf("--property=command_force=%t", c.Force),
		fmt.Sprintf("--property=command_builders=%s", strings.Join(c.BuilderNames, ";")),
	}
}

// tag::string_is_command[]
// StringIsCommand returns true if the given string is a valid /buildbot command.
func StringIsCommand(s string) bool {
	// force=yes|no : Can be used to allow for PRs to be build even when
	// they are closed or when a check run for the exact same SHA has been
	// run already.
	return regexp.MustCompile(buildRegexPattern()).MatchString(s)
}

// end::string_is_command[]

// TODO(kwk): implement fromTryBotPropertyArray() *Command

// toMap converts the command into a map
func (c Command) toMap() map[string]interface{} {
	return map[string]interface{}{
		CommandOptionMandatory: c.IsMandatory,
		CommandOptionBuilder:   c.BuilderNames,
		CommandOptionForce:     c.Force,
	}
}

// valueIsTrue returns true if a given value as is a string and its lowercase
// representation is one of: "true", "t", "yes", "y", "1".
func valueIsTrue(i interface{}) bool {
	s, ok := i.(string)
	if !ok {
		return false
	}
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "t" || s == "yes" || s == "y" || s == "1"
}

// valueIsFalse returns true if a given value is a string and its in lowecase
// representation is one of: "false", "f", "no", "n", "0".
func valueIsFalse(i interface{}) bool {
	s, ok := i.(string)
	if !ok {
		return false
	}
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "false" || s == "f" || s == "no" || s == "n" || s == "0"
}

// tag::command_regex[]
// buildRegexPattern returns the regex pattern to match a string against a
// /buildbot command
func buildRegexPattern() string {
	tfOptions := `(yes|no|true|false|f|t|y|n|0|1)`
	mandatoryOption := fmt.Sprintf(`%s=%s`, CommandOptionMandatory, tfOptions)
	forceOption := fmt.Sprintf(`%s=%s`, CommandOptionForce, tfOptions)
	builderOption := fmt.Sprintf(`%s=(\w+)`, CommandOptionBuilder)
	return fmt.Sprintf(`^%s(\s+|%s|%s|%s)*$`, BuildbotCommand, mandatoryOption, forceOption, builderOption)
}

// end::command_regex[]

// stripPrefix returns anything after the first appearance of "/buildbot" in a
// command string with trimmed spaces
func stripPrefix(s string) string {
	_, after, found := strings.Cut(s, BuildbotCommand)
	if found {
		after = strings.TrimSpace(after)
		return after
	}
	return s
}

// parseIntoMap returns a map with case insensitive options that can override or
// extend each other The options are parsed left-to-right. For overwritable
// options the last option given wins. For example:
//
//	/buildbot mandatory=yes builder=foo mandatory=no builder=bar
//
// Will return map{mandatory:no, builder:[]string{"foo", "bar"}}
func parseIntoMap(s string) (map[string]interface{}, error) {
	if !StringIsCommand(s) {
		return nil, fmt.Errorf("string is no valid command: %s", s)
	}

	arguments := map[string]interface{}{}
	argString := stripPrefix(s)
	kvList := strings.Split(argString, " ")
	if len(kvList) == 1 && kvList[0] == "" {
		return arguments, nil
	}

	for _, kvStr := range kvList {
		kv := strings.SplitN(kvStr, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid comment: %s", s)
		}
		key := strings.ToLower(kv[0])
		value := kv[1]
		switch key {
		case "builder":
			if err := makeCaseInsensitiveStringList(arguments, key, value); err != nil {
				return nil, err
			}
			if builderNames, ok := arguments[key].([]string); ok {
				arguments[key] = removeDuplicatesInPlace(builderNames)
			}
		default:
			arguments[key] = value
		}
	}
	return arguments, nil
}

// See https://hackernoon.com/how-to-remove-duplicates-in-go-slices
func removeDuplicatesInPlace(elements []string) []string {
	// if there are 0 or 1 items we return the slice itself.
	if len(elements) < 2 {
		return elements
	}

	// make the slice ascending sorted.
	sort.SliceStable(elements, func(i, j int) bool { return elements[i] < elements[j] })

	uniqPointer := 0

	for i := 1; i < len(elements); i++ {
		// compare a current item with the item under the unique pointer.
		// if they are not the same, write the item next to the right of the unique pointer.
		if elements[uniqPointer] != elements[i] {
			uniqPointer++
			elements[uniqPointer] = elements[i]
		}
	}

	return elements[:uniqPointer+1]
}

func makeCaseInsensitiveStringList(m map[string]interface{}, key string, value interface{}) error {
	s := fmt.Sprintf("%v", value)
	if _, ok := m[key]; ok {
		elements, ok := m[key].([]string)
		if !ok {
			return fmt.Errorf("list of '%s' is not a string slice: %s", key, s)
		}
		elements = append(elements, s)
		// sort.Strings(elements)
		m[key] = elements
	} else {
		m[key] = []string{s}
	}
	return nil
}
