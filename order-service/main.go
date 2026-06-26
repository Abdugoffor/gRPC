package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"order-service/order-service/productpb"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient(
		"localhost:50031",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatal("Product servicega ulanishda xatolik:", err)
	}

	defer conn.Close()

	productClient := productpb.NewProductServiceClient(conn)

	router := httprouter.New()
	router.GET("/products/:id", ProductByID(productClient))

	fmt.Println("Order service HTTP API started on port :8000")

	if err := http.ListenAndServe(":8000", router); err != nil {
		log.Fatal(err)
	}
}

func ProductByID(productClient productpb.ProductServiceClient) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		productID, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
		if err != nil {
			http.Error(w, "product id noto'g'ri", http.StatusBadRequest)
			return
		}

		log.Printf("[HTTP request] GET /products/%d", productID)
		log.Printf("[gRPC request] GetProduct id=%d -> product-service:50031", productID)

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		product, err := productClient.GetProduct(ctx, &productpb.GetProductRequest{
			Id: productID,
		})
		if err != nil {
			log.Printf("[gRPC response] error id=%d error=%v", productID, err)
			http.Error(w, "product topilmadi", http.StatusNotFound)
			return
		}

		log.Printf("[gRPC response] id=%d name=%q price=%.2f stock=%d", product.Id, product.Name, product.Price, product.Stock)

		response := ProductResponse{
			ID:          product.Id,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Stock:       product.Stock,
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("[HTTP response] JSON id=%d name=%q", response.ID, response.Name)
	}
}

type ProductResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int64   `json:"stock"`
}
