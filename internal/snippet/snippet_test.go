package snippet_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/tullo/snptx/internal/snippet"
	"github.com/tullo/snptx/internal/tests"
)

// TestSnippet validates the full set of CRUD operations on Snippet values.
func TestSnippet(t *testing.T) {

	// skip the test if the -short flag is provided
	if testing.Short() {
		t.Skip("database: skipping integration test")
	}

	db, teardown := tests.NewUnit(t)
	defer teardown()

	s := snippet.New(db)

	t.Log("Given the need to work with Snippet records.")
	{
		t.Log("\tWhen handling a single Snippet.")
		{
			ctx := tests.Context()
			now := time.Now()
			now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
			//oneDayLater := now.AddDate(0, 0, 1)
			//oneMonthLater := now.AddDate(0, 1, 0)
			oneYearLater := now.AddDate(1, 0, 0)

			ns := snippet.NewSnippet{
				Title:       "Foo",
				Content:     "Foo bar baz",
				DateExpires: oneYearLater,
			}

			spt, err := s.Create(ctx, ns, now)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create snippet : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create snippet.", tests.Success)

			latestS, err := s.Latest(ctx)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to list latest snippets: %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to list latest snippets.", tests.Success)
			if len(latestS) < 1 {
				t.Log("\t\tGot:", len(latestS))
				t.Log("\t\tExp:", 1)
				t.Fatalf("\t%s\tShould get back at least one snippet.\n", tests.Failed)
			}

			savedS, err := s.Retrieve(ctx, spt.ID)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retrieve snippet by ID: %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retrieve snippet by ID.", tests.Success)
			if diff := cmp.Diff(spt, savedS); diff != "" {
				t.Fatalf("\t%s\tShould get back the same snippet. Diff:\n%s", tests.Failed, diff)
			}

			addMonthDay := oneYearLater.AddDate(0, 1, 0)
			upd := snippet.UpdateSnippet{
				Title:       tests.StringPointer("Some Day"),
				Content:     tests.StringPointer("Some Day ..."),
				DateExpires: &addMonthDay,
			}
			if err := s.Update(ctx, spt.ID, upd, now); err != nil {
				t.Fatalf("\t%s\tShould be able to update snippet : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to update snippet.", tests.Success)

			savedS, err = s.Retrieve(ctx, spt.ID)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retrieve snippet by ID: %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retrieve snippet by ID.", tests.Success)

			if savedS.Title != *upd.Title {
				t.Errorf("\t%s\tShould be able to see updates to Title.", tests.Failed)
				t.Log("\t\tGot:", savedS.Title)
				t.Log("\t\tExp:", *upd.Title)
			} else {
				t.Logf("\t%s\tShould be able to see updates to Title.", tests.Success)
			}

			if savedS.Content != *upd.Content {
				t.Errorf("\t%s\tShould be able to see updates to Content.", tests.Failed)
				t.Log("\t\tGot:", savedS.Content)
				t.Log("\t\tExp:", *upd.Content)
			} else {
				t.Logf("\t%s\tShould be able to see updates to Content.", tests.Success)
			}

			if savedS.DateExpires.UTC() != upd.DateExpires.UTC() {
				t.Errorf("\t%s\tShould be able to see updates to DateExpires.", tests.Failed)
				t.Log("\t\tGot:", savedS.DateExpires.UTC())
				t.Log("\t\tExp:", upd.DateExpires.UTC())
			} else {
				t.Logf("\t%s\tShould be able to see updates to DateExpires.", tests.Success)
			}

			if err := s.Delete(ctx, spt.ID); err != nil {
				t.Fatalf("\t%s\tShould be able to delete snippet : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to delete snippet.", tests.Success)

			_, err = s.Retrieve(ctx, spt.ID)
			if errors.Cause(err) != snippet.ErrNotFound {
				t.Fatalf("\t%s\tShould NOT be able to retrieve snippet : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould NOT be able to retrieve snippet.", tests.Success)
		}
	}
}
