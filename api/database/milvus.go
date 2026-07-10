package database

import (
	"context"
	"fmt"
	"learnlang-api/config"
	"log"
	"net"
	"strings"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

var MilvusClient *milvusclient.Client

func ConnectMilvus(cfg *config.Config) error {
	address := milvusAddress(cfg.Milvus)
	if address == "" {
		return fmt.Errorf("milvus address is required")
	}

	client, err := milvusclient.New(context.Background(), &milvusclient.ClientConfig{
		Address: address,
	})
	if err != nil {
		return err
	}

	MilvusClient = client
	log.Println("Milvus connected successfully")
	return nil
}

func milvusAddress(cfg config.MilvusConfig) string {
	host := strings.TrimSpace(cfg.Host)
	port := strings.TrimSpace(cfg.Port)
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "19530"
	}
	if strings.Contains(host, ":") {
		return host
	}
	return net.JoinHostPort(host, port)
}
