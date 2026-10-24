package check

import (
	"strings"
	"time"
)

// DateLayout is the only date format accepted in front-matter (YYYY-MM-DD).
const DateLayout = "2006-01-02"

// Front-matter date fields, in the order they must happen.
var dateFields = []string{"created", "approved_at", "closed_at"}

// parseDate parses a YYYY-MM-DD date and rejects impossible calendar dates.
func parseDate(s string) (time.Time, error) {
	return time.Parse(DateLayout, s)
}

// lifecycleRules says which lifecycle fields a document kind supports.
type lifecycleRules struct {
	hasStatus bool // documents without status (verification.md) never get approved_at
	canClose  bool // the constitution is never closed
}

// checkMeta validates the type and the date fields shared by every SDD
// document. It must run after the status has been validated, because the
// required dates depend on it.
func checkMeta(d *Doc, wantType string, rules lifecycleRules, r *Report) {
	if got := d.Fields["type"]; got != wantType {
		if got == "" {
			r.add(SevError, "EVA060", d.Path, "front-matter type is missing, expected %q", wantType)
		} else {
			r.add(SevError, "EVA060", d.Path, "front-matter type is %q, expected %q for this file", got, wantType)
		}
	}

	dates := map[string]time.Time{}
	for _, field := range dateFields {
		v := strings.TrimSpace(d.Fields[field])
		if v == "" {
			continue
		}
		t, err := parseDate(v)
		if err != nil {
			r.add(SevError, "EVA061", d.Path, "%s is %q, expected a real date as YYYY-MM-DD", field, v)
			continue
		}
		dates[field] = t
	}
	present := func(field string) bool { return strings.TrimSpace(d.Fields[field]) != "" }

	status := d.Status()
	if !present("created") && (!rules.hasStatus || status != "draft") {
		r.add(SevError, "EVA062", d.Path, "created is empty; set it to the day the document was written (YYYY-MM-DD)")
	}
	if rules.hasStatus {
		switch approved := d.Approved(); {
		case approved && !present("approved_at"):
			r.add(SevError, "EVA063", d.Path, "status is %s but approved_at is empty (the approver sets it)", status)
		case !approved && present("approved_at"):
			r.add(SevError, "EVA063", d.Path, "approved_at is set but status is %s; only an approved document has an approval date", status)
		}
	}
	switch done := status == "done"; {
	case !rules.canClose && present("closed_at"):
		r.add(SevError, "EVA064", d.Path, "closed_at must stay empty: this document is never closed")
	case rules.canClose && rules.hasStatus && done && !present("closed_at"):
		r.add(SevError, "EVA064", d.Path, "status is done but closed_at is empty (set it when closing)")
	case rules.canClose && rules.hasStatus && !done && present("closed_at"):
		r.add(SevError, "EVA064", d.Path, "closed_at is set but status is %s; only a done document is closed", status)
	}

	checkDateOrder(d, dates, r)
}

// checkDateOrder requires created ≤ approved_at ≤ closed_at for the dates present.
func checkDateOrder(d *Doc, dates map[string]time.Time, r *Report) {
	prevField := ""
	var prev time.Time
	for _, field := range dateFields {
		t, ok := dates[field]
		if !ok {
			continue
		}
		if prevField != "" && t.Before(prev) {
			r.add(SevError, "EVA065", d.Path, "%s (%s) is before %s (%s)", field, t.Format(DateLayout), prevField, prev.Format(DateLayout))
		}
		prevField, prev = field, t
	}
}
