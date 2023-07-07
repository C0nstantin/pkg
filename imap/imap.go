package imap

import (
	"errors"
	"fmt"
	"time"

	"github.com/emersion/go-imap"
	quota "github.com/emersion/go-imap-quota"
	sortthread "github.com/emersion/go-imap-sortthread"
	uidplus "github.com/emersion/go-imap-uidplus"
	"github.com/emersion/go-imap/client"
)

var (
	ErrMessageNotFound = errors.New("server didn't return message")
	ErrSystemFolder    = fmt.Errorf("system folder not found")
)

type Client struct {
	imap *client.Client
}

func Login(email, password, host string) (*Client, error) {
	clt, err := client.DialTLS(host, nil)
	if err != nil {
		return nil, err
	}
	err = clt.Login(email, password)
	if err != nil {
		return nil, err
	}
	c := Client{imap: clt}

	return &c, nil
}

func (client *Client) Logout() error {
	return client.imap.Logout()
}

func (client *Client) Client() *client.Client {
	return client.imap
}

func (client *Client) SelectMailbox(mboxName string) (*imap.MailboxStatus, error) {
	mboxName = imap.CanonicalMailboxName(mboxName)
	mbox, err := client.imap.Select(mboxName, false)
	if err != nil {
		return nil, err // errors.New("failed to select mailbox")
	}
	return mbox, nil
}

func (client *Client) FetchMessage(uid uint) (*imap.Message, error) {
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uint32(uid))

	var section imap.BodySectionName
	fetch := []imap.FetchItem{
		section.FetchItem(),
		imap.FetchEnvelope,
		imap.FetchUid,
		imap.FetchBodyStructure,
		imap.FetchFlags,
		"BODY.PEEK[HEADER.FIELDS (X-Crypted)]",
	}
	ch := make(chan *imap.Message, 1)

	if err := client.imap.UidFetch(seqSet, fetch, ch); err != nil {
		return nil, fmt.Errorf("failed to fetch message error: %w ", err)
	}

	msg := <-ch
	if msg == nil {
		return nil, ErrMessageNotFound
	}

	return msg, nil
}

func (client *Client) FetchByUids(uids []uint32) ([]*imap.Message, error) {
	var seqSet imap.SeqSet
	seqSet.AddNum(uids...)

	fetch := []imap.FetchItem{
		imap.FetchBodyStructure,
		imap.FetchFlags,
		imap.FetchEnvelope,
		imap.FetchUid,
		"BODY.PEEK[HEADER.FIELDS (X-Crypted)]",
	}

	ch := make(chan *imap.Message, len(uids))
	errChan := make(chan error)
	go func() {
		if err := client.imap.UidFetch(&seqSet, fetch, ch); err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	messages := make([]*imap.Message, 0)
	for msg := range ch {
		messages = append(messages, msg)
	}

	if err := <-errChan; err != nil {
		return nil, errors.New("error fetch messages")
	}

	return messages, nil
}

func (client *Client) FetchThreads(criteria *imap.SearchCriteria, mboxName string, page uint, limit uint) ([]*imap.Message, map[uint][]*imap.Message, uint, uint, error) {
	sc := sortthread.NewThreadClient(client.imap)

	ok, err := sc.SupportThread()
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("error check support thread > %w ", err)
	} else if !ok {
		return nil, nil, 0, 0, fmt.Errorf("server doesn't support THREAD")
	}

	threadAlgorithm := "ORDEREDSUBJECT" // REFS REFERENCES ORDEREDSUBJECT
	results, err := sc.UidThread(sortthread.ThreadAlgorithm(threadAlgorithm), criteria)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("error get uids > %w", err)
	}

	total := uint(len(results))
	if total == 0 {
		return []*imap.Message{}, nil, 0, 0, nil
	}

	var to, from uint
	if total <= (page-1)*limit {
		return []*imap.Message{}, nil, 0, 0, nil
	}
	to = total - (page-1)*limit

	if to <= limit+1 {
		from = 1
	} else {
		from = to - limit + 1
	}

	var uids []uint32
	childs := make(map[uint32][]uint32)
	threads := make(map[uint32]bool)

	for _, thread := range results {
		if thread.Children != nil {
			threadUids := make([]uint32, 0)
			threadUids = append(threadUids, thread.Id)
			threadUids = client.getThreadUids(thread.Children, threadUids)
			parentUID, childrenUids := threadUids[len(threadUids)-1], threadUids[:len(threadUids)-1]
			childs[parentUID] = childrenUids
			threads[parentUID] = true
		} else {
			threads[thread.Id] = true
		}
	}

	uidsSort, err := client.ListSortUids(criteria, mboxName)
	all := uint(len(uidsSort))

	if err != nil {
		return nil, nil, 0, 0, err
	}
	for _, uid := range uidsSort {
		if _, ok := threads[uid]; ok {
			uids = append(uids, uid)
		}
	}

	uids = uids[from-1 : to]

	childrens := make(map[uint][]*imap.Message)
	for _, p := range uids {
		if child, ok := childs[p]; ok {
			childrens[uint(p)], err = client.FetchByUids(child)
			if err != nil {
				return nil, nil, 0, 0, fmt.Errorf("error fetch messages %w", err)
			}
		}
	}

	unsort, err := client.FetchByUids(uids)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("error fetch messages %w", err)
	}

	var messages []*imap.Message
	for _, uid := range uids {
		for _, message := range unsort {
			if message.Uid == uid {
				messages = append(messages, message)
			}
		}
	}

	return messages, childrens, total, all, nil
}

func (client *Client) getThreadUids(threads []*sortthread.Thread, uids []uint32) []uint32 {
	for _, thread := range threads {
		uids = append(uids, thread.Id)
		if thread.Children != nil {
			uids = client.getThreadUids(thread.Children, uids)
		}
	}

	return uids
}

func (client *Client) ListUids(searchCriteria *imap.SearchCriteria, mboxName string) ([]uint32, error) {
	mbox, err := client.SelectMailbox(mboxName)
	if err != nil {
		return nil, fmt.Errorf("select mailbox error > %w ", err)
	}

	total := uint(mbox.Messages)
	if total == 0 {
		return []uint32{}, nil
	}

	return client.imap.UidSearch(searchCriteria)
}

func (client *Client) ListSortUids(searchCriteria *imap.SearchCriteria, mboxName string) ([]uint32, error) {
	mbox, err := client.SelectMailbox(mboxName)
	if err != nil {
		return nil, fmt.Errorf("select mailbox error > %w ", err)
	}

	total := uint(mbox.Messages)
	if total == 0 {
		return []uint32{}, nil
	}

	// Create a new SORT client
	sc := sortthread.NewSortClient(client.imap)

	// Check the server supports the extension
	ok, err := sc.SupportSort()
	if err != nil {
		return nil, fmt.Errorf("error check support sort > %w ", err)
	} else if !ok {
		return nil, fmt.Errorf("server doesn't support SORT")
	}

	sortCriteria := []sortthread.SortCriterion{
		{Field: sortthread.SortDate},
	}

	searchCriteria.SeqNum = new(imap.SeqSet)
	searchCriteria.SeqNum.AddRange(1, uint32(total))

	uids, err := sc.UidSort(sortCriteria, searchCriteria)
	if err != nil {
		return nil, fmt.Errorf("error fetch uids SORT")
	}

	return uids, nil

}

func (client *Client) FetchMessages(criteria *imap.SearchCriteria, mboxName string, page uint, limit uint) ([]*imap.Message, uint, error) {
	mbox, err := client.SelectMailbox(mboxName)
	if err != nil {
		return nil, 0, fmt.Errorf("select mailbox error > %w ", err)
	}
	total := uint(mbox.Messages)
	if total == 0 {
		return []*imap.Message{}, 0, nil
	}

	uids, err := client.ListSortUids(criteria, mboxName)
	if err != nil {
		return nil, 0, err
	}
	total = uint(len(uids))
	var to, from uint
	if total <= (page-1)*limit {
		return []*imap.Message{}, 0, nil
	}
	to = total - (page-1)*limit

	if to <= limit+1 {
		from = 1
	} else {
		from = to - limit + 1
	}

	uids = uids[from-1 : to]

	if len(uids) == 0 {
		return []*imap.Message{}, 0, nil
	}

	unsort, err := client.FetchByUids(uids)
	if err != nil {
		return nil, 0, fmt.Errorf("error fetch messages %w", err)
	}

	var messages []*imap.Message

	for _, uid := range uids {
		for _, message := range unsort {
			if message.Uid == uid {
				messages = append(messages, message)
			}
		}
	}

	return messages, total, nil
}

func (client *Client) FetchUids(mboxName string) ([]uint, error) {
	mbox, err := client.SelectMailbox(mboxName)
	if err != nil {
		return nil, fmt.Errorf("select mailbox error > %w ", err)
	}

	uids := make([]uint, 0)
	total := uint(mbox.Messages)
	if total == 0 {
		return uids, nil
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddRange(uint32(1), uint32(total))

	fetch := []imap.FetchItem{
		imap.FetchUid,
	}

	ch := make(chan *imap.Message, total+1)

	errChan := make(chan error)

	go func() {
		if err := client.imap.Fetch(seqSet, fetch, ch); err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()
	if err := <-errChan; err != nil {
		return nil, fmt.Errorf("error fetch messages %w", err)
	}

	for msg := range ch {
		uids = append(uids, uint(msg.Uid))
	}

	return uids, nil
}

func (client *Client) Expunge() error {
	if err := client.imap.Expunge(nil); err != nil {
		return errors.New("failed to delete messages")
	}

	return nil
}

func (client *Client) Append(mboxName string, message imap.Literal) (validity uint32, uid uint32, err error) {
	mbox, err := client.SelectMailbox(mboxName)
	if err != nil {
		err = fmt.Errorf("select mailbox error > %w ", err)
		return
	}

	flags := []string{imap.SeenFlag}
	if mboxName == "INBOX.Drafts" {
		flags = append(flags, imap.DraftFlag)
	}

	plus := uidplus.NewClient(client.imap)
	return plus.Append(mbox.Name, flags, time.Now(), message)
}

func (client *Client) Move(uids []uint, mboxName string) error {
	seqSet := new(imap.SeqSet)
	var seq []uint32
	for _, uid := range uids {
		seq = append(seq, uint32(uid))
	}

	seqSet.AddNum(seq...)
	return client.imap.UidMove(seqSet, mboxName)
}

func (client *Client) AddFlags(uids []uint, flags []interface{}) error {
	seqSet := new(imap.SeqSet)
	var seq []uint32
	for _, uid := range uids {
		seq = append(seq, uint32(uid))
	}

	seqSet.AddNum(seq...)
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	return client.imap.UidStore(seqSet, item, flags, nil)
}

func (client *Client) RemoveFlags(uids []uint, flags []interface{}) error {
	seqSet := new(imap.SeqSet)
	var seq []uint32
	for _, uid := range uids {
		seq = append(seq, uint32(uid))
	}

	seqSet.AddNum(seq...)
	item := imap.FormatFlagsOp(imap.RemoveFlags, true)
	return client.imap.UidStore(seqSet, item, flags, nil)
}

func (client *Client) MarkMessageAnswered(uids []uint) error {
	return client.AddFlags(uids, []interface{}{imap.AnsweredFlag})
}
func (client *Client) MarkMessageSeen(uids []uint) error {
	return client.AddFlags(uids, []interface{}{imap.SeenFlag})
}
func (client *Client) MarkMessageUnseen(uids []uint) error {
	return client.RemoveFlags(uids, []interface{}{imap.SeenFlag})
}
func (client *Client) MarkMessageFlagget(uids []uint) error {
	return client.AddFlags(uids, []interface{}{imap.FlaggedFlag})
}
func (client *Client) MarkMessageUnflagget(uids []uint) error {
	return client.RemoveFlags(uids, []interface{}{imap.FlaggedFlag})
}
func (client *Client) MarkMessageLabeled(uids []uint, label string) error {
	return client.AddFlags(uids, []interface{}{label})
}
func (client *Client) MarkMessageUnlabeled(uids []uint, label string) error {
	return client.RemoveFlags(uids, []interface{}{label})
}

func (client *Client) FolderUnreadCount(name string) (count uint, err error) {
	items := []imap.StatusItem{imap.StatusUnseen}
	mb, err := client.imap.Status(name, items)
	if err != nil {
		return
	}
	count = uint(mb.Unseen)
	return
}

func (client *Client) ListMailboxes() (mailboxes []*imap.MailboxInfo, err error) {
	ch := make(chan *imap.MailboxInfo, 20)
	done := make(chan error, 1)
	defer close(done)

	go func() {
		done <- client.imap.List("", "*", ch)
	}()

	for mbox := range ch {
		mailboxes = append(mailboxes, mbox)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("list mailboxes error > %w", err)
	}

	return
}

func (client *Client) Quota() (size uint, limit uint, err error) {
	qc := quota.NewClient(client.imap)
	quotas, err := qc.GetQuotaRoot("INBOX")
	if err != nil {
		err = fmt.Errorf("select mailbox error > %w ", err)
		return
	}

	for _, quotaI := range quotas {
		for name, usage := range quotaI.Resources {
			if name == "STORAGE" {
				size = uint(usage[0])
				limit = uint(usage[1])
			}
		}
	}

	return
}
