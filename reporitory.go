package mongorep

import (
	"context"
	"iter"
	"log"

	"reflect"
	"strings"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type (
	Pagination[T any] struct {
		Data  []T
		Count int64
	}
	Repository[T any] interface {
		GetAll(ctx context.Context,
			filter map[string]any) []T
		GetIter(ctx context.Context,
			filter map[string]any) iter.Seq[T]
		GetAllSkipTake(ctx context.Context,
			filter map[string]any,
			skip int64,
			take int64,
			opts ...*options.FindOptionsBuilder) *Pagination[T]
		Count(ctx context.Context,
			filter map[string]any) int64
		GetFirst(ctx context.Context,
			filter map[string]any) *T
		Insert(ctx context.Context,
			entity *T)
		InsertAll(ctx context.Context,
			entities []T)
		Replace(ctx context.Context,
			filter map[string]any,
			entity *T)
		Update(ctx context.Context,
			filter map[string]any,
			fields map[string]any)
		DeleteAll(ctx context.Context,
			filter map[string]any)
		Aggregate(
			ctx context.Context,
			pipeline []map[string]any) []map[string]any
	}
	MongoDbRepository[T any] struct {
		collection *mongo.Collection
	}
)

func NewMongoDbRepository[T any](
	db *mongo.Database) Repository[T] {

	var r T
	coll := db.Collection(strings.ToLower(reflect.TypeOf(r).Name()))
	return &MongoDbRepository[T]{
		collection: coll,
	}
}

func (r *MongoDbRepository[T]) GetAll(
	ctx context.Context,
	filter map[string]any) []T {
	result := make([]T, 0)
	for n := range r.GetIter(ctx, filter) {
		result = append(result, n)
	}

	return result
}

func (r *MongoDbRepository[T]) GetIter(
	ctx context.Context,
	filter map[string]any) iter.Seq[T] {
	cur, err := r.collection.Find(ctx, filter)
	if err != nil {
		panic(err)
	}

	return func(yield func(T) bool) {
		for cur.Next(ctx) {
			var el T
			if err := cur.Decode(&el); err != nil {
				panic(err)
			}
			if !yield(el) {
				return
			}
		}
	}
}

func (r *MongoDbRepository[T]) GetAllSkipTake(
	ctx context.Context,
	filter map[string]any,
	skip int64,
	take int64,
	opts ...*options.FindOptionsBuilder) *Pagination[T] {
	ct, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		panic(err)
	}

	op := options.Find()
	op.SetSkip(skip)
	if take > 0 {
		op.SetLimit(take)
	}

	for _, o := range opts {
		op.Opts = append(op.Opts, o.List()...)
	}
	cur, err := r.collection.Find(ctx, filter, op)

	if err != nil {
		panic(err)
	}
	result := make([]T, 0)
	for cur.Next(ctx) {
		var el T
		err = cur.Decode(&el)
		if err != nil {
			panic(err)
		}
		result = append(result, el)
	}

	return &Pagination[T]{
		Data:  result,
		Count: ct,
	}
}

func (r *MongoDbRepository[T]) Count(ctx context.Context,
	filter map[string]any) int64 {
	ct, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		panic(err)
	}
	return ct
}

func (r *MongoDbRepository[T]) GetFirst(
	ctx context.Context,
	filter map[string]any) *T {
	var el T
	err := r.collection.FindOne(ctx, filter).Decode(&el)
	if err == mongo.ErrNoDocuments {
		return nil
	}

	if err != nil {
		panic(err)
	}

	return &el
}

func (r *MongoDbRepository[T]) Insert(
	ctx context.Context,
	entity *T) {
	_, err := r.collection.InsertOne(ctx, entity)
	if err != nil {
		panic(err)
	}
}

func (r *MongoDbRepository[T]) InsertAll(
	ctx context.Context,
	entities []T) {
	_, err := r.collection.InsertMany(ctx, entities)
	if err != nil {
		panic(err)
	}
}

func (r *MongoDbRepository[T]) Replace(
	ctx context.Context,
	filter map[string]any,
	entity *T) {
	_, err := r.collection.ReplaceOne(ctx, filter, entity, options.Replace().SetUpsert(true))
	if err != nil {
		panic(err)
	}
}

func (r *MongoDbRepository[T]) Update(
	ctx context.Context,
	filter map[string]any,
	fields map[string]any) {
	_, err := r.collection.UpdateOne(ctx, filter, map[string]any{
		"$set": fields,
	}, nil)
	if err != nil {
		panic(err)
	}
}

func (r *MongoDbRepository[T]) DeleteAll(
	ctx context.Context,
	filter map[string]any) {
	_, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		panic(err)
	}
}

func (r *MongoDbRepository[T]) Aggregate(
	ctx context.Context,
	pipeline []map[string]any) []map[string]any {
	cur, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		panic(err)
	}
	result := []map[string]any{}
	for cur.Next(ctx) {
		var el map[string]any
		err = cur.Decode(&el)
		if err != nil {
			panic(err)
		}
		result = append(result, parseMongoUUIDs(el))
	}

	return result
}

var (
	MapType = reflect.TypeFor[map[string]any]()
	BinType = reflect.TypeFor[bson.Binary]()
)

func parseMongoUUIDs(parentMap map[string]any) map[string]any {
	for key, value := range parentMap {
		if reflect.TypeOf(value) == MapType {
			parentMap[key] = parseMongoUUIDs(value.(map[string]any))
			continue
		}
		if reflect.TypeOf(value) == BinType {
			if id, err := uuid.FromBytes(value.(bson.Binary).Data); err == nil {
				parentMap[key] = id
			} else {
				log.Print(err.Error())
			}
		}
	}
	return parentMap
}
