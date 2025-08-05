# Auto Job Tracker 📩 → 📊

A fully automated system to **parse job application emails** from Gmail and **update your Notion job tracker** — using GPT or Gemini and Go.  
Runs daily via GitHub Actions, or manually from your terminal.

---

## ⚙️ What It Does

* Connects to your Gmail inbox via IMAP
* Reads recent job-related emails
* Uses OpenAI **(GPT-4o)** or **Gemini Pro** to extract:

  * Company
  * Position
  * Stage (Applied, Interview, Rejected)
  * Referral status
  * Job URL
* Pushes structured results to your Notion database
* Logs unparseable emails privately for manual review

---

## 🚀 Quick Start

You can either **fork + run via GitHub Actions**, or **run locally via CLI**.

---

## ♻️ Option 1: Use with GitHub Actions (Recommended)

### ✅ 1. Fork this repo

### ✅ 2. Set up GitHub Secrets

Go to your repo → **Settings → Secrets → Actions** → Add the following:

| Name                 | Value                                   |
| -------------------- | --------------------------------------- |
| `GMAIL_USER`         | Your Gmail address                      |
| `GMAIL_APP_PASSWORD` | Gmail App Password (not your password)  |
| `NOTION_TOKEN`       | Integration token from Notion           |
| `NOTION_DB_ID`       | Your Notion database ID                 |
| `OPENAI_API_KEY`     | Your OpenAI API key (`sk-...`)          |
| `GEMINI_API_KEY`     | Your Gemini API Key *(optional)*        |
| `USE_GEMINI`         | `true` to use Gemini, `false` for GPT   |

> 🧠 Gemini is free to use — no billing required. Just add `GEMINI_API_KEY` and set `USE_GEMINI=true` to switch.

> ✅ Your Gmail must have IMAP enabled, and you must use an App Password.

### ✅ 3. Done — runs daily at 12 AM EST

Or trigger it manually from the **Actions** tab.

---

### ⚠️ Important: Make Your Repo Private (Highly Recommended)

By default, forks are **public**. This means:

- Your GitHub Actions logs can expose sensitive metadata  
- If not handled properly, files like `unparsed_emails.csv` could leak job info

To keep your data safe:

> 🔒 **Go to your repo → Settings → Change visibility → Make private**

This ensures your job applications, emails, and logs stay confidential.

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
GEMINI_API_KEY=your_gemini_key
USE_GEMINI=true
```

### ✅ 3. Run it

```bash
go run main.go
```

---

### 🔹 How do I switch between GPT and Gemini?

Use this in your `.env` file or GitHub Secrets:

```env
USE_GEMINI=true      # to use Gemini  
USE_GEMINI=false     # to use GPT (default)
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

## 🛡️ CSV Logging of Failed Emails

Any job-related email that fails parsing or cannot be added to Notion is automatically saved in a CSV file (`unparsed_emails.csv`) for review.

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
