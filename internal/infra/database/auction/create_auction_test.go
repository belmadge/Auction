package auction_test

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"os"
	"testing"
	"time"
)

type MockAuctionRepository struct{}

func (m *MockAuctionRepository) CreateAuction(ctx context.Context, auctionEntity *auction_entity.Auction) error {
	auctionEntity.Id = "mock-id"
	auctionEntity.Status = auction_entity.Active
	auctionEntity.Timestamp = time.Now()
	return nil
}

func (m *MockAuctionRepository) UpdateAuctionStatus(ctx context.Context, auctionEntity *auction_entity.Auction) error {
	auctionEntity.Status = auction_entity.Closed
	return nil
}

func TestAutomaticAuctionClose(t *testing.T) {
	ctx := context.Background()

	os.Setenv("AUCTION_DURATION", "2s")

	auctionEntity := &auction_entity.Auction{
		ProductName: "Produto Teste",
		Category:    "Categoria Teste",
		Description: "Descrição Teste",
		Condition:   auction_entity.New,
		Status:      auction_entity.Active,
		Timestamp:   time.Now(),
	}

	auctionRepo := &MockAuctionRepository{}
	err := auctionRepo.CreateAuction(ctx, auctionEntity)
	if err != nil {
		t.Fatalf("Erro ao criar leilão: %v", err)
	}

	go auctionRepo.UpdateAuctionStatus(ctx, auctionEntity)

	time.Sleep(3 * time.Second)

	if auctionEntity.Status != auction_entity.Closed {
		t.Errorf("Leilão não foi fechado automaticamente")
	}
}
