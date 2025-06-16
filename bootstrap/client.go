package bootstrap

import (
	"context"
	"github.com/stupidrun/mon/api/proto"
	"github.com/stupidrun/mon/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

type Client struct {
	conn          *grpc.ClientConn
	name          string
	monitorClient proto.MonitoringServiceClient
}

func NewClient(addr string, clientName string) (*Client, error) {
	clientConn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, err
	}
	client := proto.NewMonitoringServiceClient(clientConn)
	return &Client{
		conn:          clientConn,
		monitorClient: client,
		name:          clientName,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) PushMetrics(ctx context.Context, metrics []*proto.Metric) (*proto.PushMetricsResponse, error) {
	req := &proto.PushMetricsRequest{
		Metrics: metrics,
	}
	return c.monitorClient.PushMetrics(ctx, req)
}

func StartPeriodicMetricsCollection(ctx context.Context, client *Client, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("开始每 %v 收集一次系统指标", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("指标收集已停止")
			return
		case <-ticker.C:
			metric, err := utils.GetCurrentMetrics()
			if err != nil {
				log.Printf("获取系统指标失败: %v", err)
				continue
			}

			metric.Name = client.name
			_, err = client.PushMetrics(ctx, []*proto.Metric{metric})
			if err != nil {
				log.Printf("推送指标失败: %v", err)
				continue
			}

			log.Printf("已推送系统指标: Name=%s, CPU=%.2f%%, 内存=%.2fMB, 网络上行=%.2fKB/s, 网络下行=%.2fKB/s",
				client.name, metric.CpuUsage, metric.MemoryUsage, metric.NetworkOut, metric.NetworkIn)
		}
	}
}
