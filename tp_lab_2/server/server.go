package Server

import (
	"context"

	logtp "example.com/tpmod/Api/v1"

	"google.golang.org/grpc"
)

type Parametros struct {
	Registro CommitLog
}

var _ logtp.LogServer = (*grpcServer)(nil)

type grpcServer struct {
	logtp.UnimplementedLogServer
	Parametros *Parametros
}

func NewGRPCServer(param *Parametros) (*grpc.Server, error) {
	// Crear un nuevo servidor gRPC
	gsrv := grpc.NewServer()

	// Crear el servidor de log
	srv, err := newgrpcServer(param)
	if err != nil {
		return nil, err
	}

	// Registrar el servidor de log
	logtp.RegisterLogServer(gsrv, srv)
	return gsrv, nil
}

func newgrpcServer(param *Parametros) (*grpcServer, error) {
	srv := &grpcServer{
		Parametros: param,
	}
	return srv, nil
}

func (s *grpcServer) Produce(ctx context.Context, req *logtp.ProduceRequest) (*logtp.ProduceResponse, error) {
	offset, err := s.Parametros.Registro.Append(req.Record)
	if err != nil {
		return nil, err
	}
	return &logtp.ProduceResponse{Offset: offset}, nil
}

func (s *grpcServer) Consume(ctx context.Context, req *logtp.ConsumeRequest) (*logtp.ConsumeResponse, error) {
	record, err := s.Parametros.Registro.Read(req.Offset)
	if err != nil {
		return nil, err
	}
	return &logtp.ConsumeResponse{Record: record}, nil
}

func (s *grpcServer) ProduceStream(stream logtp.Log_ProduceStreamServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == grpc.ErrClientConnClosing {
				return nil // El cliente ha cerrado la conexión
			}
			return err
		}
		res, err := s.Produce(stream.Context(), req)
		if err != nil {
			return err
		}
		if err := stream.Send(res); err != nil {
			return err
		}
	}
}

func (s *grpcServer) ConsumeStream(req *logtp.ConsumeRequest, stream logtp.Log_ConsumeStreamServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return nil // Salir si el contexto se cierra
		default:
			res, err := s.Consume(stream.Context(), req)
			if err != nil {
				return err
			}
			if err := stream.Send(res); err != nil {
				return err
			}
			req.Offset++ // Incrementa el offset para el siguiente registro
		}
	}
}

type CommitLog interface {
	Append(*logtp.Record) (uint64, error)
	Read(uint64) (*logtp.Record, error)
}
