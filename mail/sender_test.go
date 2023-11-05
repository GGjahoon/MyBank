package mail

import (
	"github.com/GGjahoon/MySimpleBank/util"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSendEmail(t *testing.T) {
	config, err := util.LoadConfig("..")
	require.NoError(t, err)
	sender := NewOutLookSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	subject := "a test email"
	content := `
	<h1>HelloWord</h1>
	<p>this is a massage from <a href="http://987009146@qq.com"> Ja Hoon</a></p>
	`
	to := []string{"987009146@qq.com"}
	attachFiles := []string{"..README.md"}

	sender.SendEmail(subject, content, to, nil, nil, attachFiles)
	require.NoError(t, err)
}
