package validation

import (
	"strings"

	"postsapi/internal/models"
)

// Errors maps a field name to a human-readable validation message.
type Errors map[string]string

// ValidatePostInput enforces the rules from the spec:
//   - title:    required, min 20 characters
//   - content:  required, min 200 characters
//   - category: required, min 3 characters
//   - status:   required, must be one of publish | draft | thrash
func ValidatePostInput(in models.PostInput) Errors {
	errs := Errors{}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		errs["title"] = "title is required"
	} else if len(title) < 20 {
		errs["title"] = "title must be at least 20 characters"
	}

	content := strings.TrimSpace(in.Content)
	if content == "" {
		errs["content"] = "content is required"
	} else if len(content) < 200 {
		errs["content"] = "content must be at least 200 characters"
	}

	category := strings.TrimSpace(in.Category)
	if category == "" {
		errs["category"] = "category is required"
	} else if len(category) < 3 {
		errs["category"] = "category must be at least 3 characters"
	}

	status := strings.TrimSpace(strings.ToLower(in.Status))
	if status == "" {
		errs["status"] = "status is required"
	} else if status != models.StatusPublish && status != models.StatusDraft && status != models.StatusThrash {
		errs["status"] = "status must be one of: publish, draft, thrash"
	}

	return errs
}
