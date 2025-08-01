# Auto Job Tracker 📩 → 📊

A fully automated system to **parse job application emails** from Gmail and **update your Notion job tracker** — using GPT and Go.
Runs daily via GitHub Actions, or manually from your terminal.

---

## ⚙️ What It Does

* Connects to your Gmail inbox via IMAP
* Reads recent job-related emails
* Uses OpenAI (GPT-4o) to extract:

  * Company
  * Position
  * Stage (Applied, Interview, Rejected)
  * Referral status
  * Job URL
* Pushes structured results to your Notion database
* Logs unparseable emails for review

---

## 🚀 Quick Start

You can either **fork + run via GitHub Actions**, or **run locally via CLI**.

---

## ♻️ Option 1: Use with GitHub Actions (Recommended)

### ✅ 1. Fork this repo

### ✅ 2. Set up GitHub Secrets

Go to your repo → **Settings → Secrets → Actions** → Add the following:

| Name                 | Value                                  |
| -------------------- | -------------------------------------- |
| `GMAIL_USER`         | Your Gmail address                     |
| `GMAIL_APP_PASSWORD` | Gmail App Password (not your password) |
| `NOTION_TOKEN`       | Integration token from Notion          |
| `NOTION_DB_ID`       | Your Notion database ID                |
| `OPENAI_API_KEY`     | Your OpenAI API key (`sk-...`)         |

> ✅ Your Gmail must have IMAP enabled, and you must use an App Password.

### ✅ 3. That's it — it runs daily at 12 AM EST

You can also trigger it manually from the **Actions** tab.

---

## 💻 Option 2: Run Locally

### ✅ 1. Clone the repo

```bash
git clone https://github.com/your-username/auto-job-tracker.git
cd auto-job-tracker
```

### ✅ 2. Create a `.env` file

```
GMAIL_USER=your@gmail.com
GMAIL_APP_PASSWORD=your_app_password
NOTION_TOKEN=your_notion_secret
NOTION_DB_ID=your_notion_database_id
OPENAI_API_KEY=your_openai_key
```

### ✅ 3. Run it

```bash
go run main.go
```

---

## 📒 Notion Database Format

Your Notion database must have these columns:

| Column Name     | Type   |
| --------------- | ------ |
| `Company`       | Title  |
| `Position`      | Text   |
| `Stage`         | Status |
| `Referral?`     | Select |
| `JobURL`        | URL    |
| `Apply date`    | Date   |
| `Response date` | Date   |

---

## 🧠 Technologies Used

* Go 🐹
* OpenAI GPT-4o
* Notion API
* IMAP (via Gmail)
* GitHub Actions (for scheduling)

---

## 🤠 Why This Exists

Tired of tracking job apps manually?
This automates your grind so you can focus on prepping, not updating spreadsheets.

---

## ❓ FAQ

### Can I modify the prompt?

Yes — `prompt.txt` controls what’s sent to GPT.

### Is my data secure?

Yes. Nothing is logged or stored beyond what's pushed to Notion.

---

## 🧱 Contribute

1. Fork it
2. Open a PR
3. Drop issues if something breaks

---
