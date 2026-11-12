package pagination_test

import (
	"net/http/httptest"
	"testing"

	"mapup-geofence/internal/pagination"
)

func TestPagination_Unpaginated(t *testing.T) {
	req := httptest.NewRequest("GET", "/geofences", nil)
	params := pagination.Parse(req)

	if params.Enabled {
		t.Fatalf("expected pagination to be disabled when neither page nor limit specified")
	}

	items := []string{"a", "b", "c", "d", "e"}
	sliced, meta := pagination.PaginateSlice(items, params)

	if len(sliced) != 5 {
		t.Fatalf("expected all 5 items returned, got %d", len(sliced))
	}
	if meta.TotalCount != 5 {
		t.Fatalf("expected totalCount 5, got %d", meta.TotalCount)
	}
}

func TestPagination_Paginated(t *testing.T) {
	req := httptest.NewRequest("GET", "/geofences?page=2&limit=2", nil)
	params := pagination.Parse(req)

	if !params.Enabled {
		t.Fatalf("expected pagination to be enabled")
	}
	if params.Page != 2 || params.Limit != 2 || params.Offset != 2 {
		t.Fatalf("unexpected params: %+v", params)
	}

	items := []string{"item1", "item2", "item3", "item4", "item5"}
	sliced, meta := pagination.PaginateSlice(items, params)

	if len(sliced) != 2 {
		t.Fatalf("expected 2 items on page 2, got %d", len(sliced))
	}
	if sliced[0] != "item3" || sliced[1] != "item4" {
		t.Fatalf("expected item3, item4; got %+v", sliced)
	}
	if meta.TotalCount != 5 {
		t.Errorf("expected total count 5, got %d", meta.TotalCount)
	}
	if meta.TotalPages != 3 {
		t.Errorf("expected 3 total pages, got %d", meta.TotalPages)
	}
}
