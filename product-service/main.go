package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"product-service/product-service/productpb"

	"github.com/julienschmidt/httprouter"
	"google.golang.org/grpc"
)

type Product struct {
	ID          int64
	Name        string
	Description string
	Price       float64
	Stock       int64
}

var products = map[int64]Product{
	1: {
		ID:          1,
		Name:        "iPhone 15",
		Description: "Apple iPhone 15 128GB",
		Price:       12500000,
		Stock:       10,
	},
	2: {
		ID:          2,
		Name:        "Samsung S25",
		Description: "Samsung Galaxy S25",
		Price:       11000000,
		Stock:       15,
	},
	3: {
		ID:          3,
		Name:        "MacBook Air M3",
		Description: "Apple MacBook Air M3",
		Price:       18500000,
		Stock:       5,
	},
	4: {
		ID:          4,
		Name:        "Dell XPS 13",
		Description: "Dell XPS 13 Laptop",
		Price:       15000000,
		Stock:       8,
	},
	5: {
		ID:          5,
		Name:        "Sony WH-1000XM5",
		Description: "Sony Noise Cancelling Headphones",
		Price:       3500000,
		Stock:       20,
	},
	6: {
		ID:          6,
		Name:        "Apple Watch Series 9",
		Description: "Apple Watch Series 9 Smartwatch",
		Price:       4500000,
		Stock:       12,
	},
}

type ProductServer struct {
	productpb.UnimplementedProductServiceServer
}

func (s *ProductServer) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.ProductResponse, error) {
	log.Printf("[gRPC request] GetProduct id=%d", req.Id)

	product, ok := products[req.Id]
	if !ok {
		log.Printf("[gRPC response] product not found id=%d", req.Id)
		return nil, errors.New("product not found")
	}

	response := &productpb.ProductResponse{
		Id:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
	}

	log.Printf("[gRPC response] id=%d name=%q price=%.2f stock=%d", response.Id, response.Name, response.Price, response.Stock)

	return response, nil
}

func main() {
	listener, err := net.Listen("tcp", ":50031")
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()

	productpb.RegisterProductServiceServer(server, &ProductServer{})

	fmt.Println("Product service started on port :50031")

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	// if err := server.Serve(listener); err != nil {
	// 	log.Fatal(err)
	// }

	router := httprouter.New()
	router.GET("/products", ProductList)

	fmt.Println("HTTP API started on port :7000")

	if err := http.ListenAndServe(":7000", router); err != nil {
		log.Fatal(err)
	}
}

func ProductList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println("[HTTP request] GET /products")

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(products); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[HTTP response] products count=%d", len(products))
}
