package pkg

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	fruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

const Name = "NetworkTraffic"

var _ = framework.ScorePlugin(&NetworkTraffic{})

type NetworkTraffic struct {
	prometheus *PrometheusHandle
	// FrameworkHandle 提供插件可以使用的数据和一些工具。
	// 它在插件初始化时传递给 plugin 工厂类。
	// plugin 必须存储和使用这个handle来调用framework函数。
	handle framework.FrameworkHandle
}

// FitArgs holds the args that are used to configure the plugin.
type NetworkTrafficArgs struct {
	IP         string `json:"ip"`
	DeviceName string `json:"deviceName"`
	TimeRange  int    `json:"timeRange"`
}

// New initializes a new plugin and returns it.
func New(plArgs runtime.Object, h framework.FrameworkHandle) (framework.Plugin, error) {
	args := &NetworkTrafficArgs{}
	if err := fruntime.DecodeInto(plArgs, args); err != nil {
		return nil, err
	}

	klog.Infof("[NetworkTraffic] args received. Device: %s; TimeRange: %d, Address: %s", args.DeviceName, args.TimeRange, args.IP)

	return &NetworkTraffic{
		handle:     h,
		prometheus: NewProme(args.IP, args.DeviceName, time.Minute*time.Duration(args.TimeRange)),
	}, nil
}

// Name returns name of the plugin. It is used in logs, etc.
func (n *NetworkTraffic) Name() string {
	return Name
}

// 如果返回framework.ScoreExtensions 就需要实现framework.ScoreExtensions
func (n *NetworkTraffic) ScoreExtensions() framework.ScoreExtensions {
	return n
}

// NormalizeScore与ScoreExtensions是固定格式
func (n *NetworkTraffic) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, scores framework.NodeScoreList) *framework.Status {
	var higherScore int64
	for _, node := range scores {
		if higherScore < node.Score {
			higherScore = node.Score
		}
	}
	// 计算公式为，满分 - (当前带宽 / 最高最高带宽 * 100)
	// 公式的计算结果为，带宽占用越大的机器，分数越低
	for i, node := range scores {
		scores[i].Score = framework.MaxNodeScore - (node.Score * 100 / higherScore)
		klog.Infof("[NetworkTraffic] Nodes final score: %v", scores)
	}

	klog.Infof("[NetworkTraffic] Nodes final score: %v", scores)
	return nil
}

func (n *NetworkTraffic) Score(ctx context.Context, state *framework.CycleState, p *corev1.Pod, nodeName string) (int64, *framework.Status) {
	nodeBandwidth, err := n.prometheus.GetGauge(nodeName)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("error getting node bandwidth measure: %s", err))
	}
	bandWidth := int64(nodeBandwidth.Value)
	klog.Infof("[NetworkTraffic] node '%s' bandwidth: %s", nodeName, bandWidth)
	return bandWidth, nil
}
