package main

import (
	"context"
	"log"
	"net"
	db "precios_provider/database"
	pb "precios_provider/monedas"
	"time"

	"go.elastic.co/apm/module/apmgrpc/v2"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedPreciosMonedasServer
}

func (s *server) Query(ctx context.Context, in *pb.MonedaRequest) (*pb.MonedaResponse, error) {
	log.Printf("Recived request to get value of currency: %v ", in)

	start, err := time.Parse("2006-01-02", in.GetFechaInicio())
	if err != nil {
		return nil, err
	}

	end, err := time.Parse("2006-01-02", in.GetFechaTermino())
	if err != nil {
		return nil, err
	}

	valores, err := db.GetValores(in.Moneda, start, end)
	if err != nil {
		return nil, err
	}
	log.Printf("Found %v", valores)
	valores_out := make([]*pb.ValorDia, len(valores))

	for num, val := range valores {
		valores_out[num] = &pb.ValorDia{
			Fecha: val.Fecha,
			Valor: val.Valor,
		}
	}

	return &pb.MonedaResponse{Moneda: in.Moneda, Valores: valores_out}, nil
}

func (s *server) Update(ctx context.Context, in *pb.UpdateMonedaRequest) (*pb.UpdateMonedaResponse, error) {
	valores_input := in.GetValores()
	valores_modificados := make([]db.ValorMoneda, len(valores_input))

	for num, val := range valores_input {
		valores_modificados[num] = db.ValorMoneda{
			Moneda: in.GetMoneda(),
			Fecha:  val.GetFecha(),
			Valor:  val.GetValor(),
		}
	}

	err := db.AddRowsMonedas(valores_modificados)
	if err.Error() == "entrada duplicada" {
		return &pb.UpdateMonedaResponse{Moneda: in.GetMoneda(), Status: 1.0}, nil
	} else if err != nil {
		return nil, err
	}

	return &pb.UpdateMonedaResponse{Moneda: in.GetMoneda(), Status: 0.0}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed setting up listener: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(apmgrpc.NewUnaryServerInterceptor(apmgrpc.WithRecovery())),
	)

	pb.RegisterPreciosMonedasServer(s, &server{})
	log.Printf("[+] Server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("[!] Failed to serve: %v", err)
	}
}
