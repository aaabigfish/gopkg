package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BaseCollection interface {
	SelectPage(ctx context.Context, filter interface{}, sort interface{}, skip, limit int64) (int64, []interface{}, error)
	SelectList(ctx context.Context, filter interface{}, sort interface{}) ([]interface{}, error)
	SelectListForExpect(ctx context.Context, filter interface{}, sort interface{}, expect interface{}, limit int64) (interface{}, error)
	SelectOne(ctx context.Context, filter interface{}) (interface{}, error)
	SelectCount(ctx context.Context, filter interface{}) (int64, error)
	UpdateOne(ctx context.Context, filter, update interface{}) (int64, error)
	UpsertOne(ctx context.Context, filter, update interface{}) (int64, error)
	UpdateMany(ctx context.Context, filter, update interface{}) (int64, error)
	Delete(ctx context.Context, filter interface{}) (int64, error)
	InsetOne(ctx context.Context, model interface{}) (interface{}, error)
	InsertMany(ctx context.Context, models []interface{}) ([]interface{}, error)
	InsertManyIgnore(ctx context.Context, models []interface{}) ([]interface{}, error)
	Aggregate(ctx context.Context, pipeline interface{}, result interface{}) error
	CreateIndexes(ctx context.Context, indexes []mongo.IndexModel) error
	GetCollection() *mongo.Collection
}

type BaseCollectionImpl struct {
	DbName     string
	ColName    string
	DataBase   *mongo.Database
	Collection *mongo.Collection
}

func (b *BaseCollectionImpl) SelectPage(ctx context.Context, filter interface{}, sort interface{}, skip, limit int64) (int64, []interface{}, error) {
	var err error

	resultCount, err := b.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, nil, err
	}

	opts := options.Find().SetSort(sort).SetSkip(skip).SetLimit(limit)
	finder, err := b.Collection.Find(ctx, filter, opts)
	if err != nil {
		return resultCount, nil, err
	}

	result := make([]interface{}, 0)
	if err := finder.All(ctx, &result); err != nil {
		return resultCount, nil, err
	}
	return resultCount, result, nil
}

func (b *BaseCollectionImpl) SelectList(ctx context.Context, filter interface{}, sort interface{}) ([]interface{}, error) {
	var err error

	opts := options.Find().SetSort(sort)
	finder, err := b.Collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, 0)
	if err := finder.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, err
}

func (b *BaseCollectionImpl) SelectListForExpect(ctx context.Context, filter interface{}, sort interface{}, expect interface{}, limit int64) (interface{}, error) {
	var err error
	opts := options.Find().SetSort(sort).SetLimit(limit)
	finder, err := b.Collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	if err := finder.All(ctx, &expect); err != nil {
		return nil, err
	}
	return expect, err
}

func (b *BaseCollectionImpl) SelectOne(ctx context.Context, filter interface{}) (interface{}, error) {
	result := new(interface{})
	err := b.Collection.FindOne(ctx, filter, options.FindOne()).Decode(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (b *BaseCollectionImpl) SelectCount(ctx context.Context, filter interface{}) (int64, error) {
	return b.Collection.CountDocuments(ctx, filter)
}

func (b *BaseCollectionImpl) UpdateOne(ctx context.Context, filter, update interface{}) (int64, error) {
	result, err := b.Collection.UpdateOne(ctx, filter, update, options.Update())
	if err != nil {
		return 0, err
	}
	if result.MatchedCount == 0 {
		return 0, fmt.Errorf("Update result: %s ", "document not found")
	}
	return result.MatchedCount, nil
}

func (b *BaseCollectionImpl) UpsertOne(ctx context.Context, filter, update interface{}) (int64, error) {
	result, err := b.Collection.UpdateOne(ctx, filter, bson.M{"$set": update}, options.Update().SetUpsert(true))
	if err != nil {
		return 0, err
	}

	return result.MatchedCount, nil
}

func (b *BaseCollectionImpl) UpdateMany(ctx context.Context, filter, update interface{}) (int64, error) {
	result, err := b.Collection.UpdateMany(ctx, filter, update, options.Update())
	if err != nil {
		return 0, err
	}
	if result.MatchedCount == 0 {
		return 0, fmt.Errorf("Update result: %s ", "document not found")
	}
	return result.MatchedCount, nil
}

func (b *BaseCollectionImpl) Delete(ctx context.Context, filter interface{}) (int64, error) {
	result, err := b.Collection.DeleteMany(ctx, filter, options.Delete())
	if err != nil {
		return 0, err
	}
	if result.DeletedCount == 0 {
		return 0, fmt.Errorf("DeleteOne result: %s ", "document not found")
	}
	return result.DeletedCount, nil
}

func (b *BaseCollectionImpl) InsetOne(ctx context.Context, model interface{}) (interface{}, error) {
	result, err := b.Collection.InsertOne(ctx, model, options.InsertOne())
	if err != nil {
		return nil, err
	}
	return result.InsertedID, err
}

func (b *BaseCollectionImpl) InsertMany(ctx context.Context, models []interface{}) ([]interface{}, error) {
	result, err := b.Collection.InsertMany(ctx, models, options.InsertMany())
	if err != nil {
		return nil, err
	}
	return result.InsertedIDs, err
}

func (b *BaseCollectionImpl) InsertManyIgnore(ctx context.Context, models []interface{}) ([]interface{}, error) {
	result, err := b.Collection.InsertMany(ctx, models, options.InsertMany().SetOrdered(false))
	if err != nil {
		return nil, err
	}
	return result.InsertedIDs, err
}

func (b *BaseCollectionImpl) Aggregate(ctx context.Context, pipeline interface{}, result interface{}) error {
	finder, err := b.Collection.Aggregate(ctx, pipeline, options.Aggregate())
	if err != nil {
		return err
	}
	if err := finder.All(ctx, &result); err != nil {
		return err
	}
	return nil
}

func (b *BaseCollectionImpl) CreateIndexes(ctx context.Context, indexes []mongo.IndexModel) error {
	_, err := b.Collection.Indexes().CreateMany(ctx, indexes, options.CreateIndexes())
	return err
}

func (b *BaseCollectionImpl) GetCollection() *mongo.Collection {
	return b.Collection
}
