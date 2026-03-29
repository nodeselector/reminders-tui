package remindctl

import (
	"testing"
	"time"
)

func TestParseReminders(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name: "valid reminders",
			input: `[
				{
					"id": "ABC123",
					"title": "Buy milk",
					"dueDate": "2026-03-29T06:00:00Z",
					"isCompleted": false,
					"listID": "LIST1",
					"listName": "Groceries",
					"priority": "high",
					"notes": "whole milk"
				},
				{
					"id": "DEF456",
					"title": "Call mom",
					"isCompleted": false,
					"listID": "LIST2",
					"listName": "Personal",
					"priority": "none"
				}
			]`,
			want: 2,
		},
		{
			name:  "empty array",
			input: `[]`,
			want:  0,
		},
		{
			name:    "invalid json",
			input:   `not json`,
			wantErr: true,
		},
		{
			name: "completed reminder",
			input: `[{
				"id": "XYZ",
				"title": "Done task",
				"isCompleted": true,
				"completionDate": "2026-03-28T10:00:00Z",
				"listID": "L1",
				"listName": "Personal",
				"priority": "none"
			}]`,
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseReminders([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if len(got) != tt.want {
				t.Errorf("got %d reminders, want %d", len(got), tt.want)
			}
		})
	}
}

func TestParseRemindersFields(t *testing.T) {
	input := `[{
		"id": "ABC-123",
		"title": "Test task",
		"dueDate": "2026-04-01T06:00:00Z",
		"isCompleted": false,
		"listID": "L1",
		"listName": "Work",
		"priority": "high",
		"notes": "important"
	}]`

	reminders, err := ParseReminders([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := reminders[0]

	if r.ID != "ABC-123" {
		t.Errorf("ID = %q, want %q", r.ID, "ABC-123")
	}
	if r.Title != "Test task" {
		t.Errorf("Title = %q, want %q", r.Title, "Test task")
	}
	if r.DueDate != "2026-04-01T06:00:00Z" {
		t.Errorf("DueDate = %q, want %q", r.DueDate, "2026-04-01T06:00:00Z")
	}
	if r.IsCompleted {
		t.Errorf("IsCompleted = true, want false")
	}
	if r.ListName != "Work" {
		t.Errorf("ListName = %q, want %q", r.ListName, "Work")
	}
	if r.Priority != "high" {
		t.Errorf("Priority = %q, want %q", r.Priority, "high")
	}
	if r.Notes != "important" {
		t.Errorf("Notes = %q, want %q", r.Notes, "important")
	}
}

func TestParseLists(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name: "valid lists",
			input: `[
				{"id": "L1", "title": "Personal", "reminderCount": 15, "overdueCount": 5},
				{"id": "L2", "title": "Work", "reminderCount": 3, "overdueCount": 0}
			]`,
			want: 2,
		},
		{
			name:  "empty",
			input: `[]`,
			want:  0,
		},
		{
			name:    "bad json",
			input:   `{broken`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLists([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if len(got) != tt.want {
				t.Errorf("got %d lists, want %d", len(got), tt.want)
			}
		})
	}
}

func TestReminderIsOverdue(t *testing.T) {
	now := time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		r    Reminder
		want bool
	}{
		{
			name: "overdue",
			r:    Reminder{DueDate: "2026-03-28T06:00:00Z", IsCompleted: false},
			want: true,
		},
		{
			name: "due today",
			r:    Reminder{DueDate: "2026-03-29T06:00:00Z", IsCompleted: false},
			want: false,
		},
		{
			name: "future",
			r:    Reminder{DueDate: "2026-04-01T06:00:00Z", IsCompleted: false},
			want: false,
		},
		{
			name: "overdue but completed",
			r:    Reminder{DueDate: "2026-03-28T06:00:00Z", IsCompleted: true},
			want: false,
		},
		{
			name: "no due date",
			r:    Reminder{IsCompleted: false},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.IsOverdue(now)
			if got != tt.want {
				t.Errorf("IsOverdue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReminderIsDueToday(t *testing.T) {
	now := time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		r    Reminder
		want bool
	}{
		{
			name: "due today",
			r:    Reminder{DueDate: "2026-03-29T06:00:00Z", IsCompleted: false},
			want: true,
		},
		{
			name: "yesterday",
			r:    Reminder{DueDate: "2026-03-28T06:00:00Z", IsCompleted: false},
			want: false,
		},
		{
			name: "tomorrow",
			r:    Reminder{DueDate: "2026-03-30T06:00:00Z", IsCompleted: false},
			want: false,
		},
		{
			name: "today but completed",
			r:    Reminder{DueDate: "2026-03-29T06:00:00Z", IsCompleted: true},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.IsDueToday(now)
			if got != tt.want {
				t.Errorf("IsDueToday() = %v, want %v", got, tt.want)
			}
		})
	}
}
