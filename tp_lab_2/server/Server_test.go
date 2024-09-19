package Server

import (
	"context"
	"net"
	"os"
	"testing"

	api "example.com/tpmod/Api/v1"
	Log "example.com/tpmod/Log"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func TestServer(t *testing.T) {
	scenarios := map[string]func(t *testing.T, client api.LogClient, param *Parametros){
		"produce/consume a message to/from the log succeeds": testProduceConsume,
		"produce/consume stream succeeds":                    testProduceConsumeStream,
		"consume past log boundary fails":                    testConsumePastBoundary,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			client, param, teardown := setupTest(t, nil)
			defer teardown()
			fn(t, client, param)
		})
	}
}

func setupTest(t *testing.T, fn func(*Parametros)) (
	client api.LogClient,
	param *Parametros,
	teardown func(),
) {
	t.Helper()

	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	clientOptions := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	cc, err := grpc.Dial(l.Addr().String(), clientOptions...)
	require.NoError(t, err)

	dir, err := os.MkdirTemp("", "server-test")
	require.NoError(t, err)

	clog, err := Log.NewLog(dir, Log.Config{})
	require.NoError(t, err)

	param = &Parametros{Registro: clog}
	if fn != nil {
		fn(param)
	}

	server, err := newgrpcServer(param)
	require.NoError(t, err)

	gsrv := grpc.NewServer()
	api.RegisterLogServer(gsrv, server)

	go func() {
		require.NoError(t, gsrv.Serve(l))
	}()

	client = api.NewLogClient(cc)

	return client, param, func() {
		gsrv.Stop()
		cc.Close()
		l.Close()
		os.RemoveAll(dir) // Asegúrate de eliminar el directorio temporal
	}
}

func testProduceConsume(t *testing.T, client api.LogClient, param *Parametros) {
	ctx := context.Background()

	want := &api.Record{Value: []byte("hello world")}
	produce, err := client.Produce(ctx, &api.ProduceRequest{Record: want})
	require.NoError(t, err)

	consume, err := client.Consume(ctx, &api.ConsumeRequest{Offset: produce.Offset})
	require.NoError(t, err)
	require.Equal(t, want.Value, consume.Record.Value)
	require.Equal(t, produce.Offset, consume.Record.Offset)
}

func testConsumePastBoundary(t *testing.T, client api.LogClient, param *Parametros) {
	ctx := context.Background()

	produce, err := client.Produce(ctx, &api.ProduceRequest{Record: &api.Record{Value: []byte("hello world")}})
	require.NoError(t, err)

	consume, err := client.Consume(ctx, &api.ConsumeRequest{Offset: produce.Offset + 1})
	require.Error(t, err)
	require.Nil(t, consume)

	got := status.Code(err)
	want := status.Code(api.ErrOffsetOutOfRange{}.GRPCStatus().Err())
	require.Equal(t, want, got)
}

func testProduceConsumeStream(t *testing.T, client api.LogClient, param *Parametros) {
	ctx := context.Background()

	records := []*api.Record{
		{Value: []byte("first message")},
		{Value: []byte("second message")},
	}

	stream, err := client.ProduceStream(ctx)
	require.NoError(t, err)

	for offset, record := range records {
		err = stream.Send(&api.ProduceRequest{Record: record})
		require.NoError(t, err)

		res, err := stream.Recv()
		require.NoError(t, err)
		require.Equal(t, uint64(offset), res.Offset)
	}

	consumeStream, err := client.ConsumeStream(ctx, &api.ConsumeRequest{Offset: 0})
	require.NoError(t, err)

	for i, expectedRecord := range records {
		res, err := consumeStream.Recv()
		require.NoError(t, err)
		require.Equal(t, expectedRecord.Value, res.Record.Value)
		require.Equal(t, uint64(i), res.Record.Offset)
	}
}
