package command_test

import (
	"fmt"

	"github.com/kwk/buildbot-app/cmd/buildbot-app/command"
)

func ExampleCommand_ToGithubCheckNameString() {
	c, err := command.FromString("/buildbot force=true mandatory=false builder=linux builder=windows builder=mac")
	if err != nil {
		panic(err)
	}
	c.CommentAuthor = "user"
	fmt.Println(c.ToGithubCheckNameString())
	// Output: @user /buildbot mandatory=false force=true builder=[linux mac windows]
}
