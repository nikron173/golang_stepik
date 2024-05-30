package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные
func StartMyMicroservice(ctx context.Context, listenerAddr string, ACLData string) error {

	adminService := &AdminService{
		mu:    sync.RWMutex{},
		wg:    sync.WaitGroup{},
		Logs:  sync.Map{},
		Stats: sync.Map{},
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
		grpc.UnaryInterceptor(adminService.ACLUnary),
		grpc.StreamInterceptor(adminService.ACLStream),
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
					server.GracefulStop()
					listener.Close()
					return ctx.Err()
				}
			}

		}
	}()

	return nil
}

type AdminService struct {
	mu    sync.RWMutex
	wg    sync.WaitGroup
	Logs  sync.Map
	Stats sync.Map
	ACL   map[string][]string
	UnimplementedAdminServer
}

func (as *AdminService) CheckACL(md metadata.MD, methodCall string) (bool, string, error) {
	consumerArr := md.Get("consumer")
	if len(consumerArr) == 0 {
		return false, "", grpc.Errorf(codes.Unauthenticated, "Unauthenticated")
	}

	consumer := consumerArr[0]
	methods, ok := as.ACL[consumer]
	if !ok {
		return false, "", grpc.Errorf(codes.Unauthenticated, "Unauthenticated")
	}

	cont := false
	for _, method := range methods {
		if strings.EqualFold(method, methodCall) {
			cont = true
			break
		}
		sep := strings.Split(method, "/")
		if len(sep) == 3 && sep[2] == "*" && strings.Split(methodCall, "/")[1] == sep[1] {
			cont = true
			break
		}
	}

	if !cont {
		return false, "", grpc.Errorf(codes.Unauthenticated, "Unauthenticated")
	}

	return true, consumer, nil
}

func (as *AdminService) ACLUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	p, _ := peer.FromContext(ctx)

	ok, consumer, err := as.CheckACL(md, info.FullMethod)

	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	event := &Event{
		Timestamp: time.Now().Unix(),
		Method:    info.FullMethod,
		Consumer:  consumer,
		Host:      p.Addr.String(),
	}

	as.wg.Add(2)
	go func() {
		defer as.wg.Done()
		as.Logs.Range(func(key, value interface{}) bool {
			eventResp, _ := value.(EventResponce)
			eventResp.Event <- *event
			return true
		})
	}()

	go func() {
		defer as.wg.Done()
		as.Stats.Range(func(key, value interface{}) bool {
			stat, _ := value.(Stat)
			stat.Timestamp = time.Now().Unix()
			stat.ByConsumer[consumer]++
			stat.ByMethod[info.FullMethod]++
			return true
		})
	}()
	as.wg.Wait()
	reply, err := handler(ctx, req)

	return reply, err
}

func (as *AdminService) ACLStream(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	md, _ := metadata.FromIncomingContext(ss.Context())
	p, _ := peer.FromContext(ss.Context())

	ok, consumer, err := as.CheckACL(md, info.FullMethod)

	if err != nil {
		return err
	}

	if !ok {
		return status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	event := &Event{
		Timestamp: time.Now().Unix(),
		Method:    info.FullMethod,
		Consumer:  consumer,
		Host:      p.Addr.String(),
	}

	as.mu.Lock()
	as.wg.Add(2)
	go func() {
		defer as.wg.Done()
		as.Logs.Range(func(key, value interface{}) bool {
			eventResp, _ := value.(EventResponce)
			eventResp.Event <- *event
			return true
		})
	}()

	go func() {
		defer as.wg.Done()
		as.Stats.Range(func(key, value interface{}) bool {
			stat, _ := value.(Stat)
			stat.Timestamp = time.Now().Unix()
			stat.ByConsumer[consumer]++
			stat.ByMethod[info.FullMethod]++
			return true
		})
	}()
	as.wg.Wait()
	as.mu.Unlock()
	if err := handler(srv, ss); err != nil {
		return err
	}

	return nil
}

func (as *AdminService) Logging(nothing *Nothing, in Admin_LoggingServer) error {
	ctx := in.Context()
	md, _ := metadata.FromIncomingContext(ctx)
	consumer := md.Get("consumer")[0]

	_, ok := as.Logs.Load(consumer)
	if !ok {
		as.Logs.Store(consumer, EventResponce{
			Event: make(chan Event),
		})
	}

	eventAny, _ := as.Logs.Load(consumer)
	eventResponce, _ := eventAny.(EventResponce)

	for {
		select {
		case <-ctx.Done():
			close(eventResponce.Event)
			as.Logs.Delete(consumer)
			return nil
		case event := <-eventResponce.Event:
			as.mu.Lock()
			err := in.Send(&event)
			as.mu.Unlock()
			if err != nil {
				log.Printf("Error: %s\n", err)
				as.mu.Lock()
				close(eventResponce.Event)
				as.Logs.Delete(consumer)
				as.mu.Unlock()
				return err
			}
		default:

		}
	}
}

func (as *AdminService) Statistics(interval *StatInterval, in Admin_StatisticsServer) error {
	tiker := time.NewTicker(time.Duration(interval.IntervalSeconds) * time.Second)

	ctx := in.Context()
	md, _ := metadata.FromIncomingContext(ctx)
	consumer := md.Get("consumer")[0]

	_, ok := as.Stats.Load(consumer)
	if !ok {
		as.Stats.Store(consumer, Stat{
			ByMethod:   make(map[string]uint64),
			ByConsumer: make(map[string]uint64),
		})
	}

	for {
		select {
		case <-ctx.Done():
			as.Stats.Delete(consumer)
			tiker.Stop()
			return nil
		case <-tiker.C:
			as.mu.Lock()
			statAny, _ := as.Stats.Load(consumer)
			stat, _ := statAny.(Stat)
			as.Stats.Store(consumer, Stat{
				ByMethod:   make(map[string]uint64),
				ByConsumer: make(map[string]uint64),
			})
			err := in.Send(&stat)
			as.mu.Unlock()
			if err != nil {
				log.Printf("Error: %s\n", err)
				as.Stats.Delete(consumer)
				tiker.Stop()
				return err
			}
		default:

		}
	}
}

type BizService struct {
	UnimplementedBizServer
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

type EventResponce struct {
	Event chan Event
}
