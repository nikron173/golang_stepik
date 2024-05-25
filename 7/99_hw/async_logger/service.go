package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные
func StartMyMicroservice(ctx context.Context, listenerAddr string, ACLData string) error {

	stat := &Stat{
		ByMethod:   make(map[string]uint64),
		ByConsumer: make(map[string]uint64),
	}

	adminService := &AdminService{
		mu:   sync.RWMutex{},
		Stat: *stat,
	}
	bizService := new(BizService)

	if err := json.Unmarshal([]byte(ACLData), &adminService.ACL); err != nil {
		return fmt.Errorf("Error json parse ACLData")
	}

	listener, err := net.Listen("tcp", listenerAddr)
	if err != nil {
		return fmt.Errorf("Error up listener with addr: %s", listenerAddr)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(adminService.CheckACL),
	)

	RegisterAdminServer(server, adminService)
	RegisterBizServer(server, bizService)

	go func() error {
		if err := server.Serve(listener); err != nil {
			return fmt.Errorf("Error start server: %s", err)
		}
		return nil
	}()

	go func() error {
		for {
			select {
			case <-ctx.Done():
				{
					server.Stop()
					listener.Close()
					return ctx.Err()
				}
			}

		}
	}()

	return nil
}

type AdminService struct {
	mu sync.RWMutex
	Stat
	Event
	ACL map[string][]string
}

func (as *AdminService) CheckACL(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	p, _ := peer.FromContext(ctx)

	consumerArr := md.Get("consumer")
	if len(consumerArr) == 0 {
		return nil, grpc.Errorf(codes.Unauthenticated, "Unauthenticated")
	}

	consumer := consumerArr[0]
	// fmt.Printf("Consumer: %#v\n", consumer)
	// fmt.Println(as.ACL)

	methods, ok := as.ACL[consumer]
	if !ok {
		return nil, grpc.Errorf(codes.Unauthenticated, "Unauthenticated")
	}

	cont := false
	for _, method := range methods {
		if strings.EqualFold(method, info.FullMethod) {
			cont = true
			break
		}

		sep := strings.Split(method, "/")
		if len(sep) == 3 && sep[2] == "*" && strings.Split(info.FullMethod, "/")[1] == sep[1] {
			cont = true
			break
		}
	}

	if !cont {
		return nil, grpc.Errorf(codes.Unauthenticated, "Unauthenticated")
	}

	as.mu.Lock()
	as.Stat.ByConsumer[consumer]++
	as.Stat.ByMethod[info.FullMethod]++
	as.Stat.Timestamp = time.Now().Unix()
	as.Event.Consumer = consumer
	as.Event.Method = info.FullMethod
	as.Event.Timestamp = time.Now().Unix()
	as.Event.Host = p.Addr.String()
	as.mu.Unlock()
	// fmt.Println(as.Stat)
	// fmt.Println(as.Event)
	reply, err := handler(ctx, req)

	return reply, err
}

func (as *AdminService) CheckACLStream(ctx context.Context, methodCall string) error {
	md, _ := metadata.FromIncomingContext(ctx)
	consumerArr := md.Get("consumer")
	if len(consumerArr) == 0 {
		return grpc.Errorf(codes.Unauthenticated, "Unauthenticated")
	}

	consumer := consumerArr[0]
	// fmt.Printf("Consumer: %#v\n", consumer)
	// fmt.Println(as.ACL)

	methods, ok := as.ACL[consumer]
	if !ok {
		return grpc.Errorf(codes.Unauthenticated, "Unauthenticated")
	}

	cont := false
	for _, method := range methods {
		if strings.EqualFold(method, methodCall) {
			cont = true
			break
		}
		sep := strings.Split(method, "/")
		if len(sep) == 3 && sep[2] == "*" && methodCall == sep[1] {
			cont = true
			break
		}
	}

	if !cont {
		return grpc.Errorf(codes.Unauthenticated, "Unauthenticated")
	}

	return nil
}

func (as *AdminService) Logging(nothing *Nothing, in Admin_LoggingServer) error {
	if err := as.CheckACLStream(in.Context(), Admin_Logging_FullMethodName); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-in.Context().Done():
				return
			default:
				as.mu.Lock()
				event := &Event{
					Host:      as.Event.Host,
					Method:    as.Event.Method,
					Consumer:  as.Event.Consumer,
					Timestamp: as.Event.Timestamp,
				}
				as.mu.Unlock()
				fmt.Printf("Event: %#v\n", event)
				in.Send(event)
			}
		}
	}()
	return nil
}

func (as *AdminService) Statistics(interval *StatInterval, in Admin_StatisticsServer) error {
	if err := as.CheckACLStream(in.Context(), Admin_Statistics_FullMethodName); err != nil {
		return err
	}
	go func() {
		tiker := time.NewTicker(time.Duration(interval.IntervalSeconds) * time.Second)
		for _ = range tiker.C {
			select {
			case <-in.Context().Done():
				return
			default:
				as.mu.Lock()
				in.Send(&as.Stat)
				as.mu.Unlock()
			}
		}
	}()
	return nil
}

type BizService struct {
}

func (bs *BizService) Check(context.Context, *Nothing) (*Nothing, error) {
	return &Nothing{
		Dummy: true,
	}, nil
}

func (bs *BizService) Add(context.Context, *Nothing) (*Nothing, error) {
	return &Nothing{
		Dummy: true,
	}, nil
}
func (bs *BizService) Test(context.Context, *Nothing) (*Nothing, error) {
	return &Nothing{
		Dummy: true,
	}, nil
}
