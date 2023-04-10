package helper

import (
	"context"
	"fmt"

	v1 "k8s-op-demo/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// 生成 POD name
func GetRedisPodNames(redisConfig *v1.Redis) []string {
	podNames := make([]string, redisConfig.Spec.Replicas)
	fmt.Printf("%+v", redisConfig)
	for i := 0; i < redisConfig.Spec.Replicas; i++ {
		podNames[i] = fmt.Sprintf("%s-%d", redisConfig.Name, i)
	}

	fmt.Println("PodNames: ", podNames)
	return podNames
}

//  判断 redis  pod 是否能获取
func IsExistPod(podName string, redis *v1.Redis, client client.Client) bool {
	err := client.Get(context.Background(), types.NamespacedName{
		Namespace: redis.Namespace,
		Name:      podName,
	},
		&corev1.Pod{},
	)

	if err != nil {
		return false
	}
	return true
}

// 是否存在于finalizers，finalizers 是人为删除动作添加的，
// 只要finalizers有值则删除无法顺利进行，直到finalizers为空；
func IsExistInFinalizers(podName string, redis *v1.Redis) bool {
	for _, fPodName := range redis.Finalizers {
		if podName == fPodName {
			return true

		}
	}
	return false
}

func CreateRedis(client client.Client, redisConfig *v1.Redis, podName string, schema *runtime.Scheme) (string, error) {
	if IsExistPod(podName, redisConfig, client) {
		return "", nil
	}
	// 建立 POD 对象
	newPod := &corev1.Pod{}
	newPod.Name = podName
	newPod.Namespace = redisConfig.Namespace
	newPod.Spec.Containers = []corev1.Container{
		{
			Name:            podName,
			Image:           "redis:5-alpine",
			ImagePullPolicy: corev1.PullIfNotPresent,
			Ports: []corev1.ContainerPort{
				{
					ContainerPort: int32(redisConfig.Spec.Port),
				},
			},
		},
	}

	// set owner reference，使用ControllerManager为我们管理 POD
	// 这个就和ReplicateSet是一个道理
	err := controllerutil.SetControllerReference(redisConfig, newPod, schema)
	if err != nil {
		return "", err
	}
	// 创建 POD
	err = client.Create(context.Background(), newPod)
	return podName, err
}
