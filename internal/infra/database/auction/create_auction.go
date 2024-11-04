package auction

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}
type AuctionRepository struct {
	Collection *mongo.Collection
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection: database.Collection("auctions"),
	}
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	go ar.monitorAuction(ctx, auctionEntity)

	return nil
}

func (ar *AuctionRepository) monitorAuction(ctx context.Context, auctionEntity *auction_entity.Auction) {
	duration, _ := getAuctionDuration()
	timer := time.NewTimer(duration)
	<-timer.C

	auctionEntity.Close()
	err := ar.UpdateAuctionStatus(ctx, auctionEntity)
	if err != nil {
		log.Printf("Erro ao fechar o leilão %v: %v", auctionEntity.Id, err)
	} else {
		log.Printf("Leilão %v fechado automaticamente", auctionEntity.Id)
	}
}

func getAuctionDuration() (time.Duration, error) {
	durationStr := os.Getenv("AUCTION_DURATION")
	if durationStr == "" {
		durationStr = "1m"
	}
	return time.ParseDuration(durationStr)
}

func (ar *AuctionRepository) UpdateAuctionStatus(ctx context.Context, auctionEntity *auction_entity.Auction) error {
	auctionEntity.Status = auction_entity.Closed

	filter := bson.M{"_id": auctionEntity.Id}
	update := bson.M{"$set": bson.M{"status": auction_entity.Closed}}

	_, err := ar.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error("Erro ao atualizar o status do leilão", err)
		return err
	}

	return nil
}
