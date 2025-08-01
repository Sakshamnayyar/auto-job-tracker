package notion

import (
	"context"
	"log"
	"strings"
	"time"

	"autojobtracker/models"
	"github.com/jomei/notionapi"
)

const (
	ColCompany      = "Company"
	ColStage        = "Stage"
	ColReferral     = "Referral?"
	ColJobURL       = "JobURL"
	ColApplyDate    = "Apply date"
	ColResponseDate = "Response date"
	ColPosition     = "Position"
)


var (
	notionToken string
	databaseID  string
	client      *notionapi.Client
	Unparseable []models.Job
)

func Init(token, db string) {
	notionToken = token
	databaseID = db
	client = notionapi.NewClient(notionapi.Token(notionToken))
}

func findMatchingPage(job *models.Job) (*notionapi.Page, error) {
	resp, err := client.Database.Query(context.Background(), notionapi.DatabaseID(databaseID), &notionapi.DatabaseQueryRequest{})
	if err != nil {
		return nil, err
	}
	for _, page := range resp.Results {
		company := getTitle(&page, ColCompany)
		position := getText(&page, ColPosition)
		if strings.EqualFold(company, job.Company) && strings.EqualFold(position, job.Position) {
			return &page, nil
		}
	}
	return nil, nil
}

func getTitle(page *notionapi.Page, field string) string {
	prop := page.Properties[field]
	if title, ok := prop.(*notionapi.TitleProperty); ok && len(title.Title) > 0 {
		return title.Title[0].Text.Content
	}
	return ""
}

func getText(page *notionapi.Page, field string) string {
	prop := page.Properties[field]
	if richText, ok := prop.(*notionapi.RichTextProperty); ok && len(richText.RichText) > 0 {
		return richText.RichText[0].Text.Content
	}
	return ""
}

func UpdateOrCreate(job *models.Job) {
	page, err := findMatchingPage(job)
	if err != nil {
		log.Println("⛳ Using DB ID:", databaseID)
		log.Println("❌ Notion lookup error:", err)
		Unparseable = append(Unparseable, *job)
		return
	}

	props := notionapi.Properties{
		ColStage: &notionapi.StatusProperty{
			Status: mapStageToNotionStatus(job.Stage),
		},
		ColReferral: &notionapi.SelectProperty{
			Select: notionapi.Option{Name: referralOption(job.Referral)},
		},
	}
	if job.JobURL != "" {
		props[ColJobURL] = &notionapi.URLProperty{
			URL: job.JobURL,
		}
	}
	

	if !job.ApplyDate.IsZero() {
		props[ColApplyDate] = &notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: toNotionDatePtr(job.ApplyDate),
			},
		}
	}

	if job.ResponseDate != nil {
		props[ColResponseDate] = &notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: toNotionDatePtr(*job.ResponseDate),
				//Start: notionapi.Date(job.ResponseDate.Format(time.RFC3339)),
			},
		}
	}

	if page != nil {
		log.Println("✏️ Updating existing entry:", job.Company, job.Position)
		_, err := client.Page.Update(context.Background(), notionapi.PageID(page.ID), &notionapi.PageUpdateRequest{
			Properties: props,
		})
		if err != nil {
			log.Println("❌ Update failed:", err)
			Unparseable = append(Unparseable, *job)
		}
		return
	}

	// Create new entry
	log.Println("➕ Creating new entry:", job.Company, job.Position)
	props[ColCompany] = &notionapi.TitleProperty{
		Title: []notionapi.RichText{{
			Text: &notionapi.Text{Content: job.Company},
		}},
	}
	props[ColPosition] = &notionapi.RichTextProperty{
		RichText: []notionapi.RichText{{
			Text: &notionapi.Text{Content: job.Position},
		}},
	}

	_, err = client.Page.Create(context.Background(), &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{DatabaseID: notionapi.DatabaseID(databaseID)},
		Properties: props,
	})
	if err != nil {
		log.Println("❌ Create failed:", err)
		Unparseable = append(Unparseable, *job)
	}
}

func referralOption(ref bool) string {
	if ref {
		return "Referred!"
	}
	return "No"
}

func mapStageToNotionStatus(stage string) notionapi.Option {
	switch stage {
	case "Applied":
		return notionapi.Option{Name: "Applied"}
	case "Interview":
		return notionapi.Option{Name: "Interview"}
	case "Rejected":
		return notionapi.Option{Name: "Rejected"}
	default:
		log.Printf("⚠️ Unknown stage '%s', defaulting to 'Applied'", stage)
		return notionapi.Option{Name: "Applied"}
	}
}


func toNotionDatePtr(t time.Time) *notionapi.Date {
	d := notionapi.Date(t) // time.Time → Date
	return &d
}