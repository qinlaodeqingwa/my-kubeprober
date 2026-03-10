package controller

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	// 保持你原有的 import
	kubeproberv1 "kubeprober/api/v1"
)

// ProbeReconciler reconciles a Probe object
type ProbeReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=kubeprober.erda.cloud,resources=probes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubeprober.erda.cloud,resources=probes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubeprober.erda.cloud,resources=probes/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// 新增：允许发送 Events 的权限
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *ProbeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// 1. 获取 Probe
	probe := &kubeproberv1.Probe{}
	err := r.Get(ctx, req.NamespacedName, probe)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. 定义 Pod 名字
	podName := probe.Name + "-runner"

	// 3. 检查 Pod 是否存在
	foundPod := &corev1.Pod{}
	err = r.Get(ctx, types.NamespacedName{Name: podName, Namespace: probe.Namespace}, foundPod)

	// 4. 如果 Pod 不存在 -> 创建它
	if err != nil && errors.IsNotFound(err) {
		desiredPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      podName,
				Namespace: probe.Namespace,
				Labels: map[string]string{
					"app":   "kubeprober",
					"probe": probe.Name,
				},
			},
			Spec: probe.Spec.Template.Spec,
		}

		if err := ctrl.SetControllerReference(probe, desiredPod, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("启动新一轮探测任务", "pod", desiredPod.Name)
		if err := r.Create(ctx, desiredPod); err != nil {
			return ctrl.Result{}, err
		}

		// 修改点 1：创建成功后，发个 Event
		// 这样用户 describe probe 的时候就能看到 "Created pod xxx"
		r.Recorder.Event(probe, corev1.EventTypeNormal, "Created", fmt.Sprintf("Created pod %s", desiredPod.Name))

		// 初始化状态
		now := metav1.Now()
		probe.Status.LastRun = &now
		probe.Status.Phase = "Pending"
		if err := r.Status().Update(ctx, probe); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// ====================================================
	// 5. Pod 已存在，处理状态
	// ====================================================

	podPhase := string(foundPod.Status.Phase)

	// 同步 Phase 状态
	if probe.Status.Phase != podPhase {
		probe.Status.Phase = podPhase
		if err := r.Status().Update(ctx, probe); err != nil {
			return ctrl.Result{}, err
		}
	}

	if podPhase == "Running" || podPhase == "Pending" {
		return ctrl.Result{}, nil
	}

	// 6. 处理结果 (Succeeded / Failed)
	if podPhase == "Succeeded" || podPhase == "Failed" {
		result := "Succeeded"
		message := "Probe executed successfully"

		if podPhase == "Failed" {
			result = "Failed"
			message = "Pod failed"
			if len(foundPod.Status.ContainerStatuses) > 0 {
				term := foundPod.Status.ContainerStatuses[0].State.Terminated
				if term != nil {
					message = fmt.Sprintf("ExitCode: %d, Reason: %s", term.ExitCode, term.Reason)
				}
			}
			//修改点 2：如果失败了，发个 Warning Event
			// 这样 Prometheus 或者 K8s 事件监控就能抓到了
			r.Recorder.Event(probe, corev1.EventTypeWarning, "ProbeFailed", message)
		}

		// 更新状态
		if probe.Status.LastResult != result || probe.Status.LastMessage != message {
			probe.Status.LastResult = result
			probe.Status.LastMessage = message
			probe.Status.TotalRuns++
			log.Info("探测完成", "Result", result, "Message", message)
			if err := r.Status().Update(ctx, probe); err != nil {
				return ctrl.Result{}, err
			}
		}

		// 计算下一次运行时间
		interval := probe.Spec.Policy.RunInterval
		if interval == 0 {
			interval = 60
		}

		var lastRuntime time.Time
		if probe.Status.LastRun != nil {
			lastRuntime = probe.Status.LastRun.Time
		} else {
			lastRuntime = foundPod.CreationTimestamp.Time
		}

		nextRuntime := lastRuntime.Add(time.Duration(interval) * time.Second)
		now := time.Now()

		if now.Before(nextRuntime) {
			waitTime := nextRuntime.Sub(now)
			return ctrl.Result{RequeueAfter: waitTime}, nil
		}

		// 🔥🔥🔥 修改点 3：清理旧 Pod
		log.Info("间隔时间已到，清理旧Pod", "Pod", podName)

		// 发个清理事件（可选，方便调试）
		r.Recorder.Event(probe, corev1.EventTypeNormal, "Cleanup", "Deleting old pod to restart")

		// ⚠️ 修复之前的 Bug：这里之前是 errr := ... return err
		if err := r.Delete(ctx, foundPod); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *ProbeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubeproberv1.Probe{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
