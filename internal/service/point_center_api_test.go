package service

import (
	"context"
	"testing"
	"time"

	"github.com/dabowin/sola/internal/api"
	"github.com/dabowin/sola/internal/model"
	"github.com/google/uuid"
)

func TestPointCenterExchangeCodeAmountFilterAndSummary(t *testing.T) {
	st := newServiceTestStore(t)
	createPointCenterTables(t, st.DB)
	now := time.Now()
	rows := []model.ExchangeCode{
		{BaseModel: model.BaseModel{ID: uuid.New(), CreatedAt: now, UpdatedAt: now}, Code: "A-10", Amount: 10, Status: "available"},
		{BaseModel: model.BaseModel{ID: uuid.New(), CreatedAt: now, UpdatedAt: now}, Code: "B-10", Amount: 10, Status: "assigned"},
		{BaseModel: model.BaseModel{ID: uuid.New(), CreatedAt: now, UpdatedAt: now}, Code: "C-50", Amount: 50, Status: "used"},
	}
	for _, row := range rows {
		if err := st.DB.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}

	service := &pointCenterAPIService{service: NewPointCenterService(st, "12345:test")}
	amount := 10
	items, err := service.ListExchangeCodes(context.Background(), api.ExchangeCodeListQuery{Amount: &amount})
	if err != nil || len(items) != 2 {
		t.Fatalf("amount filter returned %d items, err=%v", len(items), err)
	}
	for _, item := range items {
		if item.Amount != 10 {
			t.Fatalf("amount filter returned %+v", item)
		}
	}

	summary, err := service.SummarizeExchangeCodes(context.Background())
	if err != nil || len(summary) != 2 {
		t.Fatalf("summary = %+v, err=%v", summary, err)
	}
	if summary[0].Amount != 10 || summary[0].Available != 1 || summary[0].Assigned != 1 || summary[0].Total != 2 {
		t.Fatalf("10 summary = %+v", summary[0])
	}
	if summary[1].Amount != 50 || summary[1].Used != 1 || summary[1].Total != 1 {
		t.Fatalf("50 summary = %+v", summary[1])
	}
}

func TestPointCenterExchangeCodeBatchOperations(t *testing.T) {
	st := newServiceTestStore(t)
	createPointCenterTables(t, st.DB)
	now := time.Now()
	available := model.ExchangeCode{BaseModel: model.BaseModel{ID: uuid.New(), CreatedAt: now, UpdatedAt: now}, Code: "BATCH-A", Amount: 10, Status: "available", RedeemURL: "https://old.example"}
	assigned := model.ExchangeCode{BaseModel: model.BaseModel{ID: uuid.New(), CreatedAt: now, UpdatedAt: now}, Code: "BATCH-B", Amount: 10, Status: "assigned", RedeemURL: "https://old.example"}
	if err := st.DB.Create(&available).Error; err != nil {
		t.Fatal(err)
	}
	if err := st.DB.Create(&assigned).Error; err != nil {
		t.Fatal(err)
	}
	service := &pointCenterAPIService{service: NewPointCenterService(st, "12345:test")}
	url := "https://new.example/redeem"
	result, err := service.BatchUpdateExchangeCodes(context.Background(), api.ExchangeCodeBatchUpdateRequest{IDs: []string{available.ID.String(), assigned.ID.String()}, RedeemURL: &url})
	if err != nil || result.Updated != 1 || result.Skipped != 1 {
		t.Fatalf("batch update = %+v, err=%v", result, err)
	}
	result, err = service.BatchDeleteExchangeCodes(context.Background(), api.ExchangeCodeBatchDeleteRequest{IDs: []string{available.ID.String(), assigned.ID.String()}})
	if err != nil || result.Deleted != 1 || result.Skipped != 1 {
		t.Fatalf("batch delete = %+v, err=%v", result, err)
	}
}
