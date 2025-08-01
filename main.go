package main

import (
	"log"
	"os"
	"strings"
	"time"
	"bytes"
	"io"
	"mime"
    "regexp"

	"autojobtracker/parser"
	"autojobtracker/notion"
    "autojobtracker/models"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/joho/godotenv"
)

type RawEmail struct {
	Subject string
	Body    string
    Email   string
	Date    time.Time
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	email := os.Getenv("GMAIL_USER")
	password := os.Getenv("GMAIL_APP_PASSWORD")
	notionToken := os.Getenv("NOTION_TOKEN")
	notionDB := os.Getenv("NOTION_DB_ID")
    openaiKey := os.Getenv("OPENAI_API_KEY")
    parser.Init(openaiKey)  
	notion.Init(notionToken, notionDB)

	c, err := client.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Logout()

	if err := c.Login(email, password); err != nil {
		log.Fatal(err)
	}

	_, err = c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}

	criteria := imap.NewSearchCriteria()
	criteria.Since = time.Now().AddDate(0, 0, -7)
	ids, err := c.Search(criteria)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Found %d messages.\n", len(ids))
	if len(ids) == 0 {
		return
	}
    
    // Limit to first 10 messages
    if len(ids) > 20 {
        //ids = ids[:10]
    }

	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)
	section := &imap.BodySectionName{}
	msgChan := make(chan *imap.Message)
	rawChan := make(chan RawEmail)
	jobChan := make(chan *models.Job)

	// Fetcher goroutine
	go func() {
		if err := c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, section.FetchItem()}, msgChan); err != nil {
			log.Fatal(err)
		}
		//close(msgChan)
	}()

	// Parser goroutine
	go func() {
		for msg := range msgChan {
			if msg.Envelope == nil {
				continue
			}
			subject := strings.ToLower(msg.Envelope.Subject)

			if isJobEmail(subject) {
				body := getBodyText(msg)
                email := getSenderEmail(msg)
				rawChan <- RawEmail{
					Subject: subject,
					Body:    body,
                    Email:  email,
					Date:    msg.Envelope.Date,
				}
			}
		}
		close(rawChan)
	}()

	// Parsing worker
	go func() {
		for raw := range rawChan {
			job := parser.ParseEmail(raw.Subject, raw.Body, raw.Email, raw.Date)
			log.Printf("🧠 Parsed: %+v\n", job)
			if job.Company == "" && job.Position == "" {
				// notion.FlagUnparsed(raw.Subject, raw.Body)
				continue
			}
			jobChan <- job
		}
		close(jobChan)
	}()

	// Notion writer (main thread can block here)
	for job := range jobChan {
		notion.UpdateOrCreate(job)
	}
}

func isJobEmail(subject string) bool {
	return strings.Contains(subject, "applied") ||
		strings.Contains(subject, "application") ||
		strings.Contains(subject, "thanks for applying") ||
		strings.Contains(subject, "thanks from") ||
		strings.Contains(subject, "follow-up") ||
		strings.Contains(subject, "update") ||
		strings.Contains(subject, "recruiting") ||
		strings.Contains(subject, "thank you for applying")
}

func getSenderEmail(msg *imap.Message) string {
	if msg == nil || msg.Envelope == nil || len(msg.Envelope.From) == 0 {
		log.Println("⚠️ No sender info in email")
		return ""
	}

	from := msg.Envelope.From[0] // typically the sender
	email := from.MailboxName + "@" + from.HostName
	log.Printf("📬 Sender email: %s", email)
	return email
}

func getBodyText(msg *imap.Message) string {
	if msg == nil {
		return ""
	}

	section := &imap.BodySectionName{}
	r := msg.GetBody(section)
	if r == nil {
		log.Println("No message body found")
		return ""
	}

	mr, err := mail.CreateReader(r)
	if err != nil {
		log.Println("CreateReader error:", err)
		return ""
	}

	var htmlBody string

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			log.Println("Reached end of email parts without finding text/plain")
			break
		}
		if err != nil {
			log.Println("NextPart error:", err)
			break
		}

		contentType := p.Header.Get("Content-Type")
		if contentType == "" {
			log.Println("⚠️ Missing Content-Type header in email part")
			continue
		}

		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			log.Println("Failed to parse media type:", err)
			continue
		}

		log.Printf("🔍 Found part with mediaType: %s", mediaType)

		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(p.Body)
		if err != nil {
			log.Println("ReadFrom error:", err)
			continue
		}

		if strings.HasPrefix(mediaType, "text/plain") {
			body := buf.String()
			log.Printf("✅ Extracted plain text body (length: %d)", len(body))
			return body
		}

		if strings.HasPrefix(mediaType, "text/html") {
			htmlBody = buf.String()
		}
	}

	if htmlBody != "" {
		log.Println("⚠️ No text/plain found, using HTML fallback")
		return stripHTMLTags(htmlBody)
	}

	log.Println("❌ No text/plain or usable html body found")
	return ""
}

func stripHTMLTags(html string) string {
	re := regexp.MustCompile("<[^>]*>")
	text := re.ReplaceAllString(html, "")
	// Optionally decode &nbsp; etc.
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	return strings.TrimSpace(text)
}
