package utils

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// FakeAiSpendingCapsAPI is an in-process stand-in for Core's labs AI spending caps endpoints. It keeps the
// defaults and the overrides, answers each entry the way Core does (UPDATED, UNCHANGED, or FAILED for
// the IDs in FailIds), pages the list, and records the order and size of the requests.
type FakeAiSpendingCapsAPI struct {
	mu sync.Mutex
	// GateOff answers every request 404, as Core does for an organization the flag does not cover.
	GateOff bool
	// FailIds are the subjects whose entry fails.
	FailIds map[string]bool
	// UnknownIds are subjects that are not in the organization: a request naming one answers 404
	// with the subject in the message and changes nothing, as Core does.
	UnknownIds map[string]bool
	// TotalCountPadding is added to the list's totalCount, to simulate a page that went missing.
	TotalCountPadding int

	Defaults map[string]int            // "USER" or "SERVICEKEY" to the default
	Caps     map[string]map[string]int // "USER" or "SERVICEKEY" to subject ID to override
	// calls records each request as "set-caps:3", "remove-caps:1", "set-defaults", "remove-defaults",
	// "get-defaults" or "list:USER".
	calls []string
}

func NewFakeAiSpendingCapsAPI() *FakeAiSpendingCapsAPI {
	return &FakeAiSpendingCapsAPI{
		FailIds:    map[string]bool{},
		UnknownIds: map[string]bool{},
		Defaults:   map[string]int{},
		Caps:       map[string]map[string]int{"USER": {}, "SERVICEKEY": {}},
	}
}

func (f *FakeAiSpendingCapsAPI) Seed(subjectType string, caps map[string]int) {
	for id, value := range caps {
		f.Caps[subjectType][id] = value
	}
}

func (f *FakeAiSpendingCapsAPI) CallsMade() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string{}, f.calls...)
}

func (f *FakeAiSpendingCapsAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	if f.GateOff {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "Path " + r.URL.Path + " not found", "statusCode": 404, "requestId": "test"})
		return
	}
	isDefaults := strings.HasSuffix(r.URL.Path, "/ai-spending-caps/defaults")
	if !isDefaults && !strings.HasSuffix(r.URL.Path, "/ai-spending-caps") {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	body, _ := io.ReadAll(r.Body)

	switch {
	case isDefaults && r.Method == http.MethodGet:
		f.calls = append(f.calls, "get-defaults")
		_ = json.NewEncoder(w).Encode(f.defaultsBody())
	case isDefaults && r.Method == http.MethodPost:
		f.calls = append(f.calls, "set-defaults")
		var req struct {
			UserSpendingCap     *int `json:"userSpendingCap"`
			ApiTokenSpendingCap *int `json:"apiTokenSpendingCap"`
		}
		_ = json.Unmarshal(body, &req)
		if req.UserSpendingCap != nil {
			f.Defaults["USER"] = *req.UserSpendingCap
		}
		if req.ApiTokenSpendingCap != nil {
			f.Defaults["SERVICEKEY"] = *req.ApiTokenSpendingCap
		}
		_ = json.NewEncoder(w).Encode(f.defaultsBody())
	case isDefaults && r.Method == http.MethodDelete:
		f.calls = append(f.calls, "remove-defaults")
		var req struct {
			SubjectTypes []string `json:"subjectTypes"`
		}
		_ = json.Unmarshal(body, &req)
		for _, subjectType := range req.SubjectTypes {
			delete(f.Defaults, subjectType)
		}
		_ = json.NewEncoder(w).Encode(f.defaultsBody())
	case r.Method == http.MethodPost:
		var req struct {
			SpendingCaps []struct {
				SubjectId   string `json:"subjectId"`
				SubjectType string `json:"subjectType"`
				SpendingCap int    `json:"spendingCap"`
			} `json:"spendingCaps"`
		}
		_ = json.Unmarshal(body, &req)
		f.calls = append(f.calls, "set-caps:"+strconv.Itoa(len(req.SpendingCaps)))
		for _, entry := range req.SpendingCaps {
			if f.UnknownIds[entry.SubjectId] {
				f.subjectNotFound(w, entry.SubjectType, entry.SubjectId)
				return
			}
		}
		results := []map[string]any{}
		for _, entry := range req.SpendingCaps {
			result := map[string]any{"subjectId": entry.SubjectId, "subjectType": entry.SubjectType}
			current, has := f.Caps[entry.SubjectType][entry.SubjectId]
			switch {
			case f.FailIds[entry.SubjectId]:
				result["status"] = "FAILED"
				if has {
					result["spendingCap"] = current
				}
			case has && current == entry.SpendingCap:
				result["status"], result["spendingCap"] = "UNCHANGED", current
			default:
				f.Caps[entry.SubjectType][entry.SubjectId] = entry.SpendingCap
				result["status"], result["spendingCap"] = "UPDATED", entry.SpendingCap
			}
			results = append(results, result)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"spendingCaps": results})
	case r.Method == http.MethodDelete:
		var req struct {
			Subjects []struct {
				SubjectId   string `json:"subjectId"`
				SubjectType string `json:"subjectType"`
			} `json:"subjects"`
		}
		_ = json.Unmarshal(body, &req)
		f.calls = append(f.calls, "remove-caps:"+strconv.Itoa(len(req.Subjects)))
		for _, entry := range req.Subjects {
			if f.UnknownIds[entry.SubjectId] {
				f.subjectNotFound(w, entry.SubjectType, entry.SubjectId)
				return
			}
		}
		results := []map[string]any{}
		for _, entry := range req.Subjects {
			result := map[string]any{"subjectId": entry.SubjectId, "subjectType": entry.SubjectType}
			switch _, has := f.Caps[entry.SubjectType][entry.SubjectId]; {
			case f.FailIds[entry.SubjectId]:
				result["status"] = "FAILED"
			case has:
				delete(f.Caps[entry.SubjectType], entry.SubjectId)
				result["status"] = "UPDATED"
			default:
				result["status"] = "UNCHANGED"
			}
			results = append(results, result)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"spendingCaps": results})
	case r.Method == http.MethodGet:
		subjectType := r.URL.Query().Get("subjectType")
		f.calls = append(f.calls, "list:"+subjectType)
		if r.URL.Query().Get("hasOverride") != "true" {
			http.Error(w, "the resource reads only subjects with an override", http.StatusBadRequest)
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		ids := make([]string, 0, len(f.Caps[subjectType]))
		for id := range f.Caps[subjectType] {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		page := []map[string]any{}
		for i := offset; i < len(ids) && i < offset+limit; i++ {
			page = append(page, map[string]any{
				"subjectId": ids[i], "subjectType": subjectType, "name": "name " + ids[i],
				"spendingCap": f.Caps[subjectType][ids[i]], "isCapReached": false,
			})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"aiSpendingCaps": page, "totalCount": len(ids) + f.TotalCountPadding, "offset": offset, "limit": limit,
			"periodStartAt": "2026-11-01T00:00:00Z", "periodEndAt": "2026-12-01T00:00:00Z",
			"isUsageAvailable": false, "isUsageEstimated": false,
		})
	}
}

func (f *FakeAiSpendingCapsAPI) subjectNotFound(w http.ResponseWriter, subjectType, id string) {
	kind := "user"
	if subjectType == "SERVICEKEY" {
		kind = "api token"
	}
	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(map[string]any{"message": kind + " with id " + id + " not found", "statusCode": 404, "requestId": "test"})
}

func (f *FakeAiSpendingCapsAPI) defaultsBody() map[string]any {
	body := map[string]any{}
	if value, ok := f.Defaults["USER"]; ok {
		body["userSpendingCap"] = value
	}
	if value, ok := f.Defaults["SERVICEKEY"]; ok {
		body["apiTokenSpendingCap"] = value
	}
	return body
}
