package notify

import (
	"github.com/nlopes/slack"
)

func (n *Notify) sendMsg(preText, text string) error {
	attachment := slack.Attachment{
		Pretext: preText,
		Text:    text,
		// Uncomment the following part to send a field too
		/*
			Fields: []slack.AttachmentField{
				slack.AttachmentField{
					Title: "a",
					Value: "no",
				},
			},
		*/
	}

	var err error
	if n.notifyID != "" {
		notify := "@" + n.notifyID
		_, _, err = n.opr.Slack.GetAPI().PostMessage(n.channel, slack.MsgOptionText(notify, true),
			slack.MsgOptionAttachments(attachment))
	} else {
		_, _, err = n.opr.Slack.GetAPI().PostMessage(n.channel, slack.MsgOptionAttachments(attachment))
	}
	return err
	// channelID, timestamp, err := n.opr.Slack.GetAPI().PostMessage(n.channel, notify, params)
	// if err != nil {
	// 	log.Errorf("Send message to %s failed with error: %v", toChan, err)
	// 	return errors.Trace(err)
	// }
	// log.Debug("Send message to %s succ ChanID: %s, Time: %s", toChan, channelID, timestamp)
	// return nil
}
