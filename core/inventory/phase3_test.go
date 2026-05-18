package inventory

import (
	"database/sql"
	"os"
	"testing"

	"github.com/Fernandogf021207/IPNventario/core/database"
	"github.com/Fernandogf021207/IPNventario/core/models"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	dbFile := "./test_ipnventario.db"
	_ = os.Remove(dbFile)

	db, err := database.Open(dbFile)
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}

	schemaBytes, err := os.ReadFile("../../schema.sql")
	if err != nil {
		db.Close()
		os.Remove(dbFile)
		t.Fatalf("Failed to read schema.sql: %v", err)
	}

	err = database.Initialize(db, string(schemaBytes))
	if err != nil {
		db.Close()
		os.Remove(dbFile)
		t.Fatalf("Failed to initialize test DB: %v", err)
	}

	cleanup := func() {
		db.Close()
		_ = os.Remove(dbFile)
	}
	return db, cleanup
}

func TestCategoryCRUD(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db)

	// Create
	id, err := repo.CreateCategory("Test Category", "Test Description")
	if err != nil {
		t.Fatalf("CreateCategory failed: %v", err)
	}

	// Read
	cat, err := repo.GetCategoryByID(id)
	if err != nil {
		t.Fatalf("GetCategoryByID failed: %v", err)
	}
	if cat.Name != "Test Category" {
		t.Errorf("Expected category name 'Test Category', got '%s'", cat.Name)
	}

	// Update
	err = repo.UpdateCategory(id, "Updated Category", "Updated Description")
	if err != nil {
		t.Fatalf("UpdateCategory failed: %v", err)
	}

	cat, _ = repo.GetCategoryByID(id)
	if cat.Name != "Updated Category" {
		t.Errorf("Expected updated category name 'Updated Category', got '%s'", cat.Name)
	}

	// List
	cats, err := repo.ListCategories()
	if err != nil {
		t.Fatalf("ListCategories failed: %v", err)
	}
	found := false
	for _, c := range cats {
		if c.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Created category not found in ListCategories")
	}

	// Delete
	err = repo.DeleteCategory(id)
	if err != nil {
		t.Fatalf("DeleteCategory failed: %v", err)
	}
}

func TestItemCRUDAndStockRB2RB5(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db)

	catID, _ := repo.CreateCategory("Test Cat", "")

	item := &models.Item{
		SKU:               "TOOL-001",
		Name:              "Test Tool",
		ItemType:          "tool",
		CategoryID:        sql.NullInt64{Int64: catID, Valid: true},
		Stock:             10.0,
		MinStock:          2.0,
		Unit:              "pza",
		MaintenanceStatus: "ok",
		Location:          sql.NullString{String: "Rack A", Valid: true},
		ModuleData:        "{}",
	}

	// Create
	itemID, err := repo.CreateItem(item)
	if err != nil {
		t.Fatalf("CreateItem failed: %v", err)
	}

	// Read
	fetched, err := repo.GetItemByID(itemID)
	if err != nil {
		t.Fatalf("GetItemByID failed: %v", err)
	}
	if fetched.Stock != 10.0 {
		t.Errorf("Expected stock 10.0, got %f", fetched.Stock)
	}

	// RB2 & RB5: Adjust Stock - Positive
	err = repo.AdjustStock(itemID, 5.0, "adjustment", 1, nil, nil, nil, nil, "Adding 5 more")
	if err != nil {
		t.Fatalf("AdjustStock failed: %v", err)
	}

	fetched, _ = repo.GetItemByID(itemID)
	if fetched.Stock != 15.0 {
		t.Errorf("Expected stock 15.0, got %f", fetched.Stock)
	}

	// Check transaction registered (RB5)
	txs, err := repo.GetTransactionsByItemID(itemID)
	if err != nil {
		t.Fatalf("GetTransactionsByItemID failed: %v", err)
	}
	if len(txs) != 1 {
		t.Errorf("Expected 1 transaction record, got %d", len(txs))
	} else {
		if txs[0].Quantity != 5.0 || txs[0].StockAfter != 15.0 || txs[0].Type != "adjustment" {
			t.Errorf("Unexpected transaction content: %+v", txs[0])
		}
	}

	// RB2: Prevent negative stock
	err = repo.AdjustStock(itemID, -20.0, "adjustment", 1, nil, nil, nil, nil, "Subtracting too much")
	if err == nil {
		t.Errorf("Expected error when stock goes negative, got nil")
	}

	// Check stock remains 15.0
	fetched, _ = repo.GetItemByID(itemID)
	if fetched.Stock != 15.0 {
		t.Errorf("Expected stock to remain 15.0, got %f", fetched.Stock)
	}
}

func TestItemAvailabilityRB8(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db)

	okItem := &models.Item{
		SKU:               "OK-001",
		Name:              "Ok Item",
		ItemType:          "tool",
		Stock:             1.0,
		MaintenanceStatus: "ok",
	}
	criticalItem := &models.Item{
		SKU:               "CRIT-001",
		Name:              "Critical Item",
		ItemType:          "tool",
		Stock:             1.0,
		MaintenanceStatus: "critical",
	}

	_, _ = repo.CreateItem(okItem)
	_, _ = repo.CreateItem(criticalItem)

	// List all active
	itemsAll, err := repo.ListItems("", "", "", true, false)
	if err != nil {
		t.Fatalf("ListItems failed: %v", err)
	}
	if len(itemsAll) < 2 {
		t.Errorf("Expected at least 2 items, got %d", len(itemsAll))
	}

	// List available only (RB8)
	itemsAvail, err := repo.ListItems("", "", "", true, true)
	if err != nil {
		t.Fatalf("ListItems failed: %v", err)
	}

	for _, item := range itemsAvail {
		if item.MaintenanceStatus == "critical" || item.MaintenanceStatus == "out_of_service" {
			t.Errorf("Found item with status '%s' in available items list", item.MaintenanceStatus)
		}
	}
}
