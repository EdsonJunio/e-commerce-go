package domain

import "testing"

func TestCategoryUpdateStateClearsParent(t *testing.T) {
	parentID := 2
	category := Category{ParentID: &parentID}
	category.UpdateState(CategoryChanges{ClearParent: true})
	if category.ParentID != nil {
		t.Fatalf("explicit null parent did not clear relationship: %d", *category.ParentID)
	}
}

func TestCategoryUpdateStateAssignsDifferentParent(t *testing.T) {
	oldParentID := 2
	newParentID := 3
	category := Category{ParentID: &oldParentID}
	category.UpdateState(CategoryChanges{ParentID: &newParentID})
	if category.ParentID == nil || *category.ParentID != newParentID {
		t.Fatalf("parent = %v, want %d", category.ParentID, newParentID)
	}
}
