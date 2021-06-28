package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/longhorn/longhorn-engine/pkg/replica/client"
	"github.com/longhorn/longhorn-engine/proto/ptypes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %v <controller-ip:grpc-port> <optional time between delete and recreate>\n", os.Args[0])
		return
	}

	conn, err := grpc.Dial(os.Args[1], grpc.WithInsecure())
	if err != nil {
		fmt.Printf("Cannot connect to %v: %v\n", os.Args[1], err)
		return
	}
	defer conn.Close()

	cs := ptypes.NewControllerServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), client.GRPCServiceCommonTimeout)
	defer cancel()

	replicaList, err := cs.ReplicaList(ctx, &empty.Empty{})

	if err != nil {
		fmt.Printf("replica list error: %v\n", err)
		return
	}

	// Delete the first replica and immediately recreate it.
	if len(replicaList.Replicas) > 0 {
		addr := replicaList.Replicas[0].Address

		fmt.Printf("Deleting replica %v\n", addr.Address)
		_, err := cs.ReplicaDelete(ctx, addr)
		if err != nil {
			fmt.Printf("Error deleting replica: %v\n", err)
			return
		}

		req := &ptypes.ControllerReplicaCreateRequest{
			Address:          addr.Address,
			SnapshotRequired: true,
			Mode:             ptypes.ReplicaMode_RW,
		}

		if len(os.Args) > 2 {
			seconds, _ := strconv.Atoi(os.Args[2])
			time.Sleep((time.Duration(seconds)))
		}

		fmt.Printf("Recreating replica %v\n", addr.Address)
		_, er := cs.ControllerReplicaCreate(ctx, req)

		if er != nil {
			fmt.Printf("%v\n", er)
		}

	}
}
