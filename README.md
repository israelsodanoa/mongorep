# Go MongoDB Generic Repository

A lightweight, high-performance implementation of the **Repository Pattern** for MongoDB using **Go Generics**. This library abstracts the official MongoDB driver (v2) to reduce boilerplate code while maintaining type safety and flexibility.

## 🚀 Features

* **Type Safe:** Powered by Go Generics (`[T any]`) for seamless integration with any struct.
* **Automatic Collection Naming:** Automatically derives collection names from struct names using reflection.
* **Built-in Pagination:** Includes a `GetAllSkipTake` method that returns both the data slice and the total document count.
* **Advanced Aggregation:** Support for complex pipelines with a built-in recursive parser to convert MongoDB Binary UUIDs into `google/uuid` objects.
* **Upsert Support:** The `Replace` method comes with `Upsert` enabled by default.
* **Clean API:** Simplified CRUD operations (Insert, Update, Replace, Delete, Count).

## 📦 Installation

```bash
go get github.com/israelsodanoa/mongorep
```

### Dependencies
* `go.mongodb.org/mongo-driver/v2`
* `github.com/google/uuid`

## 🛠️ Usage

### 1. Define your Entity
The repository uses the struct name (converted to lowercase) as the collection name.

```go
type Product struct {
    ID    string  `bson:"_id,omitempty"`
    Name  string  `bson:"name"`
    Price float64 `bson:"price"`
}
```

### 2. Initialize the Repository
```go
func main() {
    // ... setup your mongo.Database connection ...
    db := client.Database("your_database")

    // Initialize a generic repository for the Product struct
    productRepo := mongorep.NewMongoDbRepository[Product](db)
}
```

### 3. Basic Operations

#### Insert and Retrieve
```go
ctx := context.Background()

// Insert One
newProduct := &Product{Name: "Laptop", Price: 1200.00}
productRepo.Insert(ctx, newProduct)

// Get First match
filter := map[string]any{"name": "Laptop"}
p := productRepo.GetFirst(ctx, filter)
```

#### Pagination
The `Pagination[T]` struct provides an easy way to handle front-end data tables.
```go
// Skip 0, Take 10
result := productRepo.GetAllSkipTake(ctx, map[string]any{}, 0, 10)

fmt.Printf("Total Records: %d\n", result.Count)
for _, p := range result.Data {
    fmt.Println(p.Name)
}
```

## 📋 API Reference

| Method | Description |
| :--- | :--- |
| `GetAll` | Returns all documents matching the filter. |
| `GetAllSkipTake` | Returns a `Pagination[T]` object with data and total count. |
| `Count` | Returns the number of documents for a given filter. |
| `GetFirst` | Returns the first document found or `nil` if not found. |
| `Insert` / `InsertAll` | Persists one or multiple entities. |
| `Replace` | Replaces a document (Upsert: true). |
| `Update` | Updates specific fields using the `$set` operator. |
| `DeleteAll` | Removes all documents matching the filter. |
| `Aggregate` | Executes an aggregation pipeline and auto-parses UUIDs. |

## 🔍 Technical Details

### Collection Naming Convention
By default, `NewMongoDbRepository` uses `strings.ToLower(reflect.TypeOf(r).Name())`. 
Example: A struct named `UserSession` will map to the `usersession` collection.

### UUID Handling in Aggregations
The `Aggregate` method includes a recursive utility that detects `bson.Binary` fields. If they represent valid UUIDs, it automatically converts them to `uuid.UUID` types. This is particularly useful when working with BI dashboards or external APIs that expect standard UUID formats rather than Hex/Base64 strings.

## ⚖️ License
This project is licensed under the MIT License.

---

Does this structure work for you, or would you like to include a "Contribution" section as well?