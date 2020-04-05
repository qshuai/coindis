package utils

import (
	"github.com/qshuai/coindis/pb"
	"google.golang.org/grpc"
)

func WalletClient(host string, port string) (pb.APIClient, error) {
	conn, err := grpc.Dial(host+":"+port, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}

	return  pb.NewAPIClient(conn), nil
}
