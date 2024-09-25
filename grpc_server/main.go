package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"backend/models"
	pb "backend/proto" // Update this import to your generated protobuf package path.

	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver
)

// DB connection (global variable for simplicity, can be improved with better struct management).
var db *sql.DB

// ResourceServiceServer is the server implementation for ResourceService.
type ResourceServiceServer struct {
	pb.UnimplementedResourceServiceServer
}

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

// AnimalRankingServiceServer is the server implementation for AnimalRankingService.
type AnimalRankingServiceServer struct {
	pb.UnimplementedAnimalRankingServiceServer
}

// ListAnimalRankings implements the ListAnimalRankings RPC.
func (s *AnimalRankingServiceServer) ListAnimalRankings(ctx context.Context, req *pb.ListAnimalRankingsRequest) (*pb.ListAnimalRankingsResponse, error) {
	// Implement your pagination logic here using req parameters.
	column := map[pb.AnimalRankingSortColumn]string{
		pb.AnimalRankingSortColumn_ANIMAL_RANK: "rank",
		pb.AnimalRankingSortColumn_ANIMAL_NAME: "name",
	}[req.SortColumn]

	// Fetch animal rankings using pagination logic.
	rankings, nextKey, err := models.AnimalRankingKeysetPage(ctx, db, column, int(req.Key), int(req.Limit), req.Order.String(), convertStringMapToInterfaceMap(req.GetFilters()))
	if err != nil {
		return nil, err
	}

	// Convert animal rankings to protobuf format.
	pbRankings := []*pb.AnimalRanking{}
	for _, r := range rankings {
		pbRankings = append(pbRankings, &pb.AnimalRanking{
			Id:        int32(r.ID),
			Rank:      int32(r.Rank),
			Name:      r.Name,
			CreatedAt: r.CreatedAt.String(),
			UpdatedAt: r.UpdatedAt.String(),
		})
	}

	if len(pbRankings) == 0 {
		return &pb.ListAnimalRankingsResponse{
			AnimalRankings: pbRankings,
		}, nil
	}

	return &pb.ListAnimalRankingsResponse{
		AnimalRankings: pbRankings,
		NextKey:        int32(nextKey.Rank),
	}, nil
}

// Main function to start the gRPC server.
func main() {
	// Set up the database connection using environment variables.
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "example")
	dbHost := getEnv("DB_HOST", "db")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "platform")

	// Construct the data source name (DSN) for the MySQL connection.
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbPort, dbName)

	// Open the database connection.
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	// Create a new gRPC server.
	grpcServer := grpc.NewServer()

	// Register the ResourceServiceServer and AnimalRankingServiceServer.
	pb.RegisterResourceServiceServer(grpcServer, &ResourceServiceServer{})
	pb.RegisterAnimalRankingServiceServer(grpcServer, &AnimalRankingServiceServer{})

	// Register the gRPC health check service.
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	// Set the health status to SERVING for the gRPC services.
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable reflection for grpcurl and other tools.
	reflection.Register(grpcServer)

	// Start listening for gRPC requests.
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen on port 50051: %v", err)
	}

	log.Println("gRPC server is listening on port 50051...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}

// getEnv retrieves the value of the environment variable named by the key.
// It returns the value, or the provided default if the variable is not set.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Convert map[string]string to map[string]interface{}
func convertStringMapToInterfaceMap(input map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range input {
		result[key] = value
	}
	return result
}
