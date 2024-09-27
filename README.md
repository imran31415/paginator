## Objective:

Implement pagination for a backend using Go, protobuf, grpc and mysql and xo library.

In this blog we will do the following:

1. Implement a db schema 
2. Integrate the DB schema with Golang
3. Leverage Xo library to create a custom template for pagination
4. Write a unit test to validate our custom template functions correctly
5. Integrate our db logic into a GRPC server and demonstrate the pagination


## Step 1: Implement db schema 
To implement pagination we fist need a table of something to paginate through.   So first we will create a SQL schema for the table we want to query, which we will simply call "resources".  

```sql
CREATE TABLE `resources` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `uuid` VARCHAR(100) NOT NULL UNIQUE,
    `name` VARCHAR(100) NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;
```

The following fields should be able to be paged through: 

- `id`: This wil be the db Id and paged numerically 
- `created_at`: Record creation date, paged by Date order
- `updated_at`: Record update date, paged by Date order
- `name`: user inputted field, should be able to be paged alphabetically

Now that we have the schema, we can apply it locally to our locally running mysql instance.  See the folder `db-init` for a sample implementation of applying the DB schema locally. 


## Step 2: Integrate with database using Go.   

To integrate with the database, there are many options in Golang.  In this case we will use the `xo` [library](https://github.com/xo/xo_).  Xo is a great library as it allows you to generate Go code to interact with your sql table by simply pointing the `xo` library to the schema.   

Since we have generated the table already we can now point Xo to our mysql table and it will generate the go code for us.  

```
mkdir models
xo schema 'mysql://root:Password123!@localhost/platform?parseTime=true' -o models 
```

Now we can see in `models/resource.xo.go` the xo library has generated the basic CRUD sql statements for us.   


## Step 3: Augment the xo library so that we can generate paginated based queries.   

It would be nice if the Xo library would just generate paginated queries for us based on our schema in the same way it does for the basic Insert/Select/Update/Delete statements.   Lets do this. 

1. First dump Xo templates to a template folder so that we can modify them:
```
mkdir templates
xo dump ./templates
```

Now we can run the xo command as above but apply --src ./templates and xo will generate the go code based on the local templates"
```
xo schema 'mysql://root:Password123!@localhost/platform?parseTime=true' -o models --src templates/
```

This should generate the same exact models/... code as before since we haven't modified the templates yet.   Now we can modify the templates to add our pagination logic:

1. Add helper function to `db.xo.go.tpl so its possible to paginate either Ascending or descending order:

```
// condition returns the appropriate SQL comparison operator based on the `order` parameter.
func condition(order string) string {
	if order == "ASC" {
		return ">"
	}
	return "<"
}
```

2. Update the schema.xo.go.tpl to add the template function to generate the pagination query.  Here is the template function to add:

```go


{{- $t := .Data -}}
// {{ $t.GoName }}KeysetPage retrieves a page of [{{ $t.GoName }}] records using keyset pagination with dynamic filtering.
//
// The keyset pagination retrieves results after or before a specific value (`key`)
// for a given column (`column`) with a limit (`limit`) and order (`ASC` or `DESC`).
//
// If `order` is `ASC`, it retrieves records where the value of `column` is greater than `key`.
// If `order` is `DESC`, it retrieves records where the value of `column` is less than `key`.
//
// Filters are dynamically provided via a `filters` map, where keys are column names and values are either single values or slices for `IN` clauses.
func {{ $t.GoName }}KeysetPage(ctx context.Context, db DB, column string, key interface{}, limit int, order string, filters map[string]interface{}) ([]*{{ $t.GoName }}, *{{ $t.GoName }}, error) {
    if order != "ASC" && order != "DESC" {
        return nil, nil, fmt.Errorf("invalid order: %s", order)
    }

    // Start building the query
    query := fmt.Sprintf(
        `SELECT * FROM {{ $t.SQLName }} 
         WHERE %s %s ?`, 
        column, condition(order),  // Ensure this is returning a valid operator
    )

    // Arguments for the query
    args := []interface{}{key}

    // Dynamically add filters from the `filters` map to the query
    for field, value := range filters {
        switch v := value.(type) {
        case []int:
            if len(v) > 0 {
                placeholders := make([]string, len(v))
                for i := range v {
                    placeholders[i] = "?"
                    args = append(args, v[i])
                }
                query += fmt.Sprintf(" AND %s IN (%s)", field, strings.Join(placeholders, ", "))
            }
        case []string:
            if len(v) > 0 {
                placeholders := make([]string, len(v))
                for i := range v {
                    placeholders[i] = "?"
                    args = append(args, v[i])
                }
                query += fmt.Sprintf(" AND %s IN (%s)", field, strings.Join(placeholders, ", "))
            }
        default:
            query += fmt.Sprintf(" AND %s = ?", field)
            args = append(args, value)
        }
    }

    // Finalize the query with the order and limit
    query += fmt.Sprintf(" ORDER BY %s %s LIMIT ?", column, order)
    args = append(args, limit)

    // Log the final query for debugging purposes
    log.Printf("Executing query: %s with args: %v", query, args)

    // Execute the query
    rows, err := db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, nil, logerror(err)
    }
    defer rows.Close()

    var results []*{{ $t.GoName }}
    var lastItem *{{ $t.GoName }} // Variable to store the last item

    for rows.Next() {
        {{ short $t.GoName }} := {{ $t.GoName }}{
        {{- if $t.PrimaryKeys }}
            _exists: true,
        {{ end -}}
        }
        if err := rows.Scan(
            {{ range $t.Fields -}}
            &{{ short $t.GoName }}.{{ .GoName }},
            {{- end }}
        ); err != nil {
            return nil, nil, logerror(err)
        }
        results = append(results, &{{ short $t.GoName }})
    }

    // Check for errors during row iteration.
    if err := rows.Err(); err != nil {
        return nil, nil, logerror(err)
    }

    // If we have results, set the lastItem to the last element in results.
    if len(results) > 0 {
        lastItem = results[len(results)-1]
    }

    return results, lastItem, nil
}

{{ end }}

```

This code allows the caller to specify a column, filters, limit and sort order to be used for pagination. 


Now lets re-run the xo command as before and see the generated pagination function:

```
xo schema 'mysql://root:Password123!@localhost/platform?parseTime=true' -o models --src templates/
```

You should see the following code in `models/resource.xo.go`:

```go

// ResourceKeysetPage retrieves a page of [Resource] records using keyset pagination with dynamic filtering.
//
// The keyset pagination retrieves results after or before a specific value (`key`)
// for a given column (`column`) with a limit (`limit`) and order (`ASC` or `DESC`).
//
// If `order` is `ASC`, it retrieves records where the value of `column` is greater than `key`.
// If `order` is `DESC`, it retrieves records where the value of `column` is less than `key`.
//
// Filters are dynamically provided via a `filters` map, where keys are column names and values are either single values or slices for `IN` clauses.
func ResourceKeysetPage(ctx context.Context, db DB, column string, key interface{}, limit int, order string, filters map[string]interface{}) ([]*Resource, error) {
	if order != "ASC" && order != "DESC" {
		return nil, fmt.Errorf("invalid order: %s", order)
	}

	// Start building the query
	query := fmt.Sprintf(
		`SELECT * FROM resources 
         WHERE %s %s ?`,
		column, condition(order),
	)

	// Arguments for the query
	args := []interface{}{key}

	// Dynamically add filters from the `filters` map to the query
	for field, value := range filters {
		switch v := value.(type) {
		case []int:
			if len(v) > 0 {
				placeholders := make([]string, len(v))
				for i := range v {
					placeholders[i] = "?"
					args = append(args, v[i])
				}
				query += fmt.Sprintf(" AND %s IN (%s)", field, strings.Join(placeholders, ", "))
			}
		case []string:
			if len(v) > 0 {
				placeholders := make([]string, len(v))
				for i := range v {
					placeholders[i] = "?"
					args = append(args, v[i])
				}
				query += fmt.Sprintf(" AND %s IN (%s)", field, strings.Join(placeholders, ", "))
			}
		default:
			query += fmt.Sprintf(" AND %s = ?", field)
			args = append(args, value)
		}
	}

	// Finalize the query with the order and limit
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT ?", column, order)
	args = append(args, limit)

	// Log the final query for debugging purposes
	log.Printf("Executing query: %s with args: %v", query, args)

	// Execute the query
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()

	var results []*Resource
	for rows.Next() {
		r := Resource{
			_exists: true,
		}
		if err := rows.Scan(
			&r.ID, &r.UUID, &r.Name, &r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, logerror(err)
		}
		results = append(results, &r)
	}

	if err := rows.Err(); err != nil {
		return nil, logerror(err)
	}

	return results, nil
}
```
Now we have the requisite go code to paginate throug the `resources` table based on any column.  In addition, we are also able to specify filters, limit, and sort order. 



## Step 4: Test it out

Lets write a unit test that validates the function generated above works correctly.  To do this we can add a `resources_test.go` into our models folder.  This test will contain 2 main tests:

1. Test the `ResourceKeysetPage` function that it correctly queries based on the input column, filters, limit and sort order. 
2. Test the pagination logic so that callers can paginate through results.    


Please see the full `models/resources_test.go` file for the full test code.  Here are the key lines in this test that validate the pagination logic works as expected:


```go
// First Page: Get the first 2 resources ordered by created_at ASC
	firstPage, nextKey err := ResourceKeysetPage(ctx, db, "created_at", parseTime("2024-09-25T09:55:00Z"), 2, "ASC", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to get first page: %v", err)
	}

	// Use the last object's CreatedAt value as the key for the next page
	lastCreatedAt := nextKey.CreatedAt

	// Second Page: Get the next 2 resources ordered by created_at ASC
	secondPage, err := ResourceKeysetPage(ctx, db, "created_at", lastCreatedAt, 2, "ASC", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to get second page: %v", err)
	}
```

# Success:
Now our pagination implementation is complete and we have a unit test to validate the functionality.   

---

## (Bonus) Step 5: Add a new table and get the paging logic for free.   

Now that we have made the effort to create a pagination template using Xo, we can add a new table/column and demonstrate the workflow for adding pagination logic:

1. Update schema to add a new table, `animal_rankings`.  There will be a field `animal_rank` which we will want to page on:

```sql
CREATE TABLE `animal_rankings` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `rank`  INT NOT NULL unique,
    `name` VARCHAR(100) NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL
) ENGINE=InnoDB;
```

2. Re-run the xo command to generate the db/pagination code in models/ for the new table:

`xo schema 'mysql://root:Password123!@localhost/platform?parseTime=true' -o models --src templates/`

3. Verify the `models/animalranking.xo.go` has pagination function generated by the template: 

```go
func AnimalRankingKeysetPage(ctx context.Context, db DB, column string, key interface{}, limit int, order string, filters map[string]interface{}) ([]*AnimalRanking, error) {
    ...
}
```

4. We can now add a similar test as before to validate everything is working correctly for the animal_ranging pagination:

Add test file `animalrankings_test.go` with the following test: (See models/animalranking.xo.go in package for full test code)
```go
func TestAnimalRankingsPagination(t *testing.T) {
	// Initialize the in-memory database with the animal_rankings table and data
	db, err := initAnimalRankingsTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// First Page: Get the first 2 animal rankings ordered by rank ASC
	firstPage, err := AnimalRankingKeysetPage(ctx, db, "rank", 0, 2, "ASC", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to get first page: %v", err)
	}

	expectedFirstPage := []*AnimalRanking{
		{ID: 1, Rank: 1, Name: "Lion", CreatedAt: parseTime("2024-09-25T10:00:00Z")},
		{ID: 2, Rank: 2, Name: "Tiger", CreatedAt: parseTime("2024-09-25T10:05:00Z")},
	}

	if !equalAnimalRankingSlices(firstPage, expectedFirstPage) {
		t.Errorf("Expected first page: %+v, got: %+v", printAnimalRankings(expectedFirstPage), printAnimalRankings(firstPage))
	}

	// Use the last object's rank value as the key for the next page
	lastRank := firstPage[len(firstPage)-1].Rank

	// Second Page: Get the next 2 animal rankings ordered by rank ASC
	secondPage, err := AnimalRankingKeysetPage(ctx, db, "rank", lastRank, 2, "ASC", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to get second page: %v", err)
	}

	expectedSecondPage := []*AnimalRanking{
		{ID: 3, Rank: 3, Name: "Elephant", CreatedAt: parseTime("2024-09-25T10:10:00Z")},
		{ID: 4, Rank: 4, Name: "Leopard", CreatedAt: parseTime("2024-09-25T10:15:00Z")},
	}

	if !equalAnimalRankingSlices(secondPage, expectedSecondPage) {
		t.Errorf("Expected second page: %+v, got: %+v", printAnimalRankings(expectedSecondPage), printAnimalRankings(secondPage))
	}

	// Use the last object's rank value as the key for the next page
	lastRank = secondPage[len(secondPage)-1].Rank

	// Third Page: Get the remaining animal rankings ordered by rank ASC
	thirdPage, err := AnimalRankingKeysetPage(ctx, db, "rank", lastRank, 2, "ASC", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to get third page: %v", err)
	}

	expectedThirdPage := []*AnimalRanking{
		{ID: 5, Rank: 5, Name: "Wolf", CreatedAt: parseTime("2024-09-25T10:20:00Z")},
	}

	if !equalAnimalRankingSlices(thirdPage, expectedThirdPage) {
		t.Errorf("Expected third page: %+v, got: %+v", printAnimalRankings(expectedThirdPage), printAnimalRankings(thirdPage))
	}
}
```


## (Bonus) Step 6: integrate with server 

- Now that we have functioning pagination logic in Go, we can demonstrate how to integrate this into our backend server.   

1. Create a protobuf spec that describes the interface for our new tables and the pagination logic:

```proto
syntax = "proto3";

option go_package = "backend/proto;proto";
package backend;

// Message representing a single Resource record.
message Resource {
  int32 id = 1;
  string uuid = 2;
  string name = 3;
  string created_at = 4;
  string updated_at = 5;
}

// Message representing a single AnimalRanking record.
message AnimalRanking {
  int32 id = 1;
  int32 rank = 2;
  string name = 3;
  string created_at = 4;
  string updated_at = 5;
}

// Enum for specifying the sort order (ASC or DESC).
enum SortOrder {
  ASC = 0; // Ascending order
  DESC = 1; // Descending order
}

// Enum for specifying the columns to sort in the resources table.
enum ResourceSortColumn {
  RESOURCE_CREATED_AT = 0; // Created at column for Resource
  RESOURCE_NAME = 1; // Name column for Resource
}

// Enum for specifying the columns to sort in the animal_rankings table.
enum AnimalRankingSortColumn {
  ANIMAL_RANK = 0; // Rank column for AnimalRanking
  ANIMAL_NAME = 1; // Name column for AnimalRanking
}

// Request message for listing resources with pagination.
message ListResourcesRequest {
  string key = 1; // Pagination key (e.g., created_at or name value).
  int32 limit = 2; // Number of records to retrieve.
  SortOrder order = 3; // Enum specifying ASC or DESC.
  ResourceSortColumn sort_column = 4; // Enum specifying the column to sort by.
  map<string, string> filters = 5; // Optional filters as key-value pairs.
}

// Response message containing a list of resources.
message ListResourcesResponse {
  repeated Resource resources = 1; // List of resources.
  string next_key = 2; // Next key to use for pagination.
}

// Request message for listing animal rankings with pagination.
message ListAnimalRankingsRequest {
  int32 key = 1; // Pagination key (e.g., rank value).
  int32 limit = 2; // Number of records to retrieve.
  SortOrder order = 3; // Enum specifying ASC or DESC.
  AnimalRankingSortColumn sort_column = 4; // Enum specifying the column to sort by.
  map<string, string> filters = 5; // Optional filters as key-value pairs.
}

// Response message containing a list of animal rankings.
message ListAnimalRankingsResponse {
  repeated AnimalRanking animal_rankings = 1; // List of animal rankings.
  int32 next_key = 2; // Next key to use for pagination.
}

// Service for managing resources.
service ResourceService {
  // ListResources RPC for listing resources with pagination.
  rpc ListResources (ListResourcesRequest) returns (ListResourcesResponse);
}

// Service for managing animal rankings.
service AnimalRankingService {
  // ListAnimalRankings RPC for listing animal rankings with pagination.
  rpc ListAnimalRankings (ListAnimalRankingsRequest) returns (ListAnimalRankingsResponse);
}

```


2. Now we can install the dependencies and generate the golang stubs:

Dependencies:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
```

Command to generate PB files into /proto path:
```bash
mkdir proto
export PATH=$PATH:$(go env GOPATH)/bin
protoc --go_out=./proto --go_opt=paths=source_relative --go-grpc_out=./proto --go-grpc_opt=paths=source_relative backend.proto
```

You should see 2 files generated in the `proto` directory: `backend_grpc.pb.go` and `backend.pb.go`. 

3. Now we can implement the GRPC server which will leverage our pagination logic:

Here is the relevant part of the GRPC implementation showing how the ListResources pagination is done:

```go

// ListResources implements the ListResources RPC.
func (s *ResourceServiceServer) ListResources(ctx context.Context, req *pb.ListResourcesRequest) (*pb.ListResourcesResponse, error) {
	// Implement your pagination logic here using req parameters.
	column := map[pb.ResourceSortColumn]string{
		pb.ResourceSortColumn_RESOURCE_CREATED_AT: "created_at",
		pb.ResourceSortColumn_RESOURCE_NAME:       "name",
	}[req.SortColumn]

	// Fetch resources using pagination logic.
	resources, k, err := models.ResourceKeysetPage(ctx, db, column, req.Key, int(req.Limit), req.Order.String(), convertStringMapToInterfaceMap(req.GetFilters()))
	if err != nil {
		return nil, err
	}

	// Convert resources to protobuf format.
	pbResources := []*pb.Resource{}
	for _, r := range resources {
		pbResources = append(pbResources, &pb.Resource{
			Id:        int32(r.ID),
			Uuid:      r.UUID,
			Name:      r.Name,
			CreatedAt: r.CreatedAt.String(),
			UpdatedAt: r.UpdatedAt.String(),
		})
	}
	if len(pbResources) == 0 {
		return &pb.ListResourcesResponse{
			Resources: pbResources,
		}, nil
	}
	switch req.SortColumn {
	case pb.ResourceSortColumn_RESOURCE_CREATED_AT:
		return &pb.ListResourcesResponse{
			Resources: pbResources,
			NextKey:   k.CreatedAt.String(),
		}, nil
	case pb.ResourceSortColumn_RESOURCE_NAME:
		return &pb.ListResourcesResponse{
			Resources: pbResources,
			NextKey:   k.Name,
		}, nil
	default:
		return nil, fmt.Errorf("invalid page key")
	}

}
```

## Testing it out:

We can test the whole project out now by creating a Docker-compose file and run all our relevant components at once.  Here is the docker compose file:
```docker
services:
  grpc_server:
    build:
      context: .
      dockerfile: Dockerfile.grpc
    healthcheck:
      test: ["CMD", "grpc_health_probe", "-addr=:50051"]
      interval: 10s
      timeout: 5s
      retries: 5
    ports:
      - "50051:50051"
      - "9092:9092"   # Expose the Prometheus metrics port

    environment:
      - MYSQL_USER=root
      - MYSQL_PASSWORD=example
      - MYSQL_HOST=db
      - MYSQL_PORT=3306
      - MYSQL_DBNAME=platform
      - MYSQL_DATABASE=platform  
    depends_on:
      db:
        condition: service_healthy 
    networks:
      - platform
  db:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=example
      - MYSQL_PASSWORD=example
      - MYSQL_PORT=3306
      - MYSQL_DBNAME=platform
      - MYSQL_DATABASE=platform  
    volumes:
      - ./db/schema.sql:/docker-entrypoint-initdb.d/schema.sql
    ports:
      - "3306:3306"
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      retries: 5
      start_period: 30s
    networks:
      - platform
  grpcui:
    image: fullstorydev/grpcui:latest
    ports:
      - "8081:8080"  # Expose grpcui on port 8081
    depends_on:
      grpc_server:
        condition: service_healthy
    networks:
      - platform
    command: ["-plaintext", "grpc_server:50051"]  # Disable TLS and provide the gRPC server host and port
networks:
  platform:

```

We also add some test data to our schema so we can demostrate paging:

```sql

-- Insert sample data into `resources` table
INSERT INTO `resources` (`uuid`, `name`, `created_at`, `updated_at`) VALUES
('uuid-1', 'Resource 1', '2024-09-25 10:00:00', '2024-09-25 10:00:00'),
('uuid-2', 'Resource 2', '2024-09-25 10:05:00', '2024-09-25 10:05:00'),
('uuid-3', 'Resource 3', '2024-09-25 10:10:00', '2024-09-25 10:10:00'),
('uuid-4', 'Resource 4', '2024-09-25 10:15:00', '2024-09-25 10:15:00'),
('uuid-5', 'Resource 5', '2024-09-25 10:20:00', '2024-09-25 10:20:00'),
('uuid-6', 'Resource 6', '2024-09-25 10:25:00', '2024-09-25 10:25:00'),
('uuid-7', 'Resource 7', '2024-09-25 10:30:00', '2024-09-25 10:30:00'),
('uuid-8', 'Resource 8', '2024-09-25 10:35:00', '2024-09-25 10:35:00'),
('uuid-9', 'Resource 9', '2024-09-25 10:40:00', '2024-09-25 10:40:00'),
('uuid-10', 'Resource 10', '2024-09-25 10:45:00', '2024-09-25 10:45:00'),
('uuid-11', 'Resource 11', '2024-09-25 10:50:00', '2024-09-25 10:50:00'),
('uuid-12', 'Resource 12', '2024-09-25 10:55:00', '2024-09-25 10:55:00'),
('uuid-13', 'Resource 13', '2024-09-25 11:00:00', '2024-09-25 11:00:00'),
('uuid-14', 'Resource 14', '2024-09-25 11:05:00', '2024-09-25 11:05:00'),
('uuid-15', 'Resource 15', '2024-09-25 11:10:00', '2024-09-25 11:10:00'),
('uuid-16', 'Resource 16', '2024-09-25 11:15:00', '2024-09-25 11:15:00'),
('uuid-17', 'Resource 17', '2024-09-25 11:20:00', '2024-09-25 11:20:00'),
('uuid-18', 'Resource 18', '2024-09-25 11:25:00', '2024-09-25 11:25:00'),
('uuid-19', 'Resource 19', '2024-09-25 11:30:00', '2024-09-25 11:30:00'),
('uuid-20', 'Resource 20', '2024-09-25 11:35:00', '2024-09-25 11:35:00');

-- Insert sample data into `animal_rankings` table
INSERT INTO `animal_rankings` (`rank`, `name`, `created_at`, `updated_at`) VALUES
(1, 'Lion', '2024-09-25 10:00:00', '2024-09-25 10:00:00'),
(2, 'Tiger', '2024-09-25 10:05:00', '2024-09-25 10:05:00'),
(3, 'Elephant', '2024-09-25 10:10:00', '2024-09-25 10:10:00'),
(4, 'Leopard', '2024-09-25 10:15:00', '2024-09-25 10:15:00'),
(5, 'Wolf', '2024-09-25 10:20:00', '2024-09-25 10:20:00'),
(6, 'Fox', '2024-09-25 10:25:00', '2024-09-25 10:25:00'),
(7, 'Bear', '2024-09-25 10:30:00', '2024-09-25 10:30:00'),
(8, 'Giraffe', '2024-09-25 10:35:00', '2024-09-25 10:35:00'),
(9, 'Zebra', '2024-09-25 10:40:00', '2024-09-25 10:40:00'),
(10, 'Rhinoceros', '2024-09-25 10:45:00', '2024-09-25 10:45:00'),
(11, 'Hippopotamus', '2024-09-25 10:50:00', '2024-09-25 10:50:00'),
(12, 'Cheetah', '2024-09-25 10:55:00', '2024-09-25 10:55:00'),
(13, 'Jaguar', '2024-09-25 11:00:00', '2024-09-25 11:00:00'),
(14, 'Panda', '2024-09-25 11:05:00', '2024-09-25 11:05:00'),
(15, 'Kangaroo', '2024-09-25 11:10:00', '2024-09-25 11:10:00'),
(16, 'Koala', '2024-09-25 11:15:00', '2024-09-25 11:15:00'),
(17, 'Penguin', '2024-09-25 11:20:00', '2024-09-25 11:20:00'),
(18, 'Ostrich', '2024-09-25 11:25:00', '2024-09-25 11:25:00'),
(19, 'Eagle', '2024-09-25 11:30:00', '2024-09-25 11:30:00'),
(20, 'Falcon', '2024-09-25 11:35:00', '2024-09-25 11:35:00');
```


Now we can run `docker-compose up --build` and we should see the following components spin up:
 1. Db
 2. GRPC server



Now we can run the following commands against the GRPC server and obtain the paginated results. 

We will specify the key for the second item `2024-09-25 10:05:00` which should return items 3-7 and the next key of 
`2024-09-25 10:30:00 +0000 UTC`, which should be the 7th item

```bash
grpcurl -plaintext -d '{
  "key": "2024-09-25 10:05:00",
  "limit": 5,
  "order": "ASC",
  "sortColumn": "RESOURCE_CREATED_AT",
  "filters": {}
}' grpc_server:50051 backend.ResourceService.ListResources
```

This returns:

```
{
  "resources": [
    {
      "id": 3,
      "uuid": "uuid-3",
      "name": "Resource 3",
      "created_at": "2024-09-25 10:10:00 +0000 UTC",
      "updated_at": "2024-09-25 10:10:00 +0000 UTC"
    },
    {
      "id": 4,
      "uuid": "uuid-4",
      "name": "Resource 4",
      "created_at": "2024-09-25 10:15:00 +0000 UTC",
      "updated_at": "2024-09-25 10:15:00 +0000 UTC"
    },
    {
      "id": 5,
      "uuid": "uuid-5",
      "name": "Resource 5",
      "created_at": "2024-09-25 10:20:00 +0000 UTC",
      "updated_at": "2024-09-25 10:20:00 +0000 UTC"
    },
    {
      "id": 6,
      "uuid": "uuid-6",
      "name": "Resource 6",
      "created_at": "2024-09-25 10:25:00 +0000 UTC",
      "updated_at": "2024-09-25 10:25:00 +0000 UTC"
    },
    {
      "id": 7,
      "uuid": "uuid-7",
      "name": "Resource 7",
      "created_at": "2024-09-25 10:30:00 +0000 UTC",
      "updated_at": "2024-09-25 10:30:00 +0000 UTC"
    }
  ],
  "next_key": "2024-09-25 10:30:00 +0000 UTC"
}
```

Now we should be able to do the same query but specify the next_key from the response of `2024-09-25 10:30:00 +0000 UTC`

This gives us items 8-12 as expected:

```
{
  "resources": [
    {
      "id": 8,
      "uuid": "uuid-8",
      "name": "Resource 8",
      "created_at": "2024-09-25 10:35:00 +0000 UTC",
      "updated_at": "2024-09-25 10:35:00 +0000 UTC"
    },
    {
      "id": 9,
      "uuid": "uuid-9",
      "name": "Resource 9",
      "created_at": "2024-09-25 10:40:00 +0000 UTC",
      "updated_at": "2024-09-25 10:40:00 +0000 UTC"
    },
    {
      "id": 10,
      "uuid": "uuid-10",
      "name": "Resource 10",
      "created_at": "2024-09-25 10:45:00 +0000 UTC",
      "updated_at": "2024-09-25 10:45:00 +0000 UTC"
    },
    {
      "id": 11,
      "uuid": "uuid-11",
      "name": "Resource 11",
      "created_at": "2024-09-25 10:50:00 +0000 UTC",
      "updated_at": "2024-09-25 10:50:00 +0000 UTC"
    },
    {
      "id": 12,
      "uuid": "uuid-12",
      "name": "Resource 12",
      "created_at": "2024-09-25 10:55:00 +0000 UTC",
      "updated_at": "2024-09-25 10:55:00 +0000 UTC"
    }
  ],
  "next_key": "2024-09-25 10:55:00 +0000 UTC"
}
```

In addition, we can leverage our filter as well to ensure we get paginated filtered results:

```bash
grpcurl -plaintext -d '{
  "key": "2024-09-25 10:00:00 +0000 UTC",
  "limit": 5,
  "order": "ASC",
  "sortColumn": "RESOURCE_CREATED_AT",
  "filters": {
    "name": "Resource 5"
  }
}' grpc_server:50051 backend.ResourceService.ListResources
```

This correctly returns only 1 resource (Resource 5)
```json
{
  "resources": [
    {
      "id": 5,
      "uuid": "uuid-5",
      "name": "Resource 5",
      "created_at": "2024-09-25 10:20:00 +0000 UTC",
      "updated_at": "2024-09-25 10:20:00 +0000 UTC"
    }
  ],
  "next_key": "2024-09-25 10:20:00 +0000 UTC"
}
```


# Success:
We have now successfully completed all these tasks:

1. Implement a db schema 
2. Integrate the DB schema with Golang
3. Leverage Xo library to create a custom template for pagination
4. Write a unit test to validate our custom template functions correctly
5. Integrate our db logic into a GRPC server and demonstrate the pagination





