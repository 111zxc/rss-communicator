package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

type IMAPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Mailbox  string
}

type IMAPMailbox struct {
	cfg IMAPConfig
}

func NewIMAPMailbox(cfg IMAPConfig) *IMAPMailbox {
	return &IMAPMailbox{cfg: cfg}
}

func (m *IMAPMailbox) FetchUnread(ctx context.Context) ([]IncomingMessage, error) {
	c, err := m.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Logout()

	if _, err := c.Select(m.cfg.Mailbox, false); err != nil {
		return nil, err
	}

	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	uids, err := c.UidSearch(criteria)
	if err != nil {
		return nil, err
	}
	if len(uids) == 0 {
		return nil, nil
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)

	items := []imap.FetchItem{imap.FetchUid, imap.FetchEnvelope}
	ch := make(chan *imap.Message, len(uids))
	done := make(chan error, 1)
	go func() {
		done <- c.UidFetch(seqset, items, ch)
	}()

	out := make([]IncomingMessage, 0, len(uids))
	for msg := range ch {
		if msg == nil || msg.Envelope == nil || len(msg.Envelope.From) == 0 {
			continue
		}
		addr := msg.Envelope.From[0]
		out = append(out, IncomingMessage{
			UID:     msg.Uid,
			From:    addr.MailboxName + "@" + addr.HostName,
			Subject: msg.Envelope.Subject,
		})
	}
	if err := <-done; err != nil {
		return nil, err
	}
	return out, nil
}

func (m *IMAPMailbox) MarkSeen(ctx context.Context, uids []uint32) error {
	if len(uids) == 0 {
		return nil
	}

	c, err := m.connect(ctx)
	if err != nil {
		return err
	}
	defer c.Logout()

	if _, err := c.Select(m.cfg.Mailbox, false); err != nil {
		return err
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.SeenFlag}
	return c.UidStore(seqset, item, flags, nil)
}

func (m *IMAPMailbox) connect(_ context.Context) (*client.Client, error) {
	addr := net.JoinHostPort(m.cfg.Host, fmt.Sprintf("%d", m.cfg.Port))
	c, err := client.DialTLS(addr, &tls.Config{ServerName: m.cfg.Host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return nil, err
	}
	if err := c.Login(m.cfg.Username, m.cfg.Password); err != nil {
		c.Logout()
		return nil, err
	}
	return c, nil
}
