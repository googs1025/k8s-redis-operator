package controllers

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	redisv1 "operator-redis/api/v1"
)

var (
	RedisClusterLabelKey       = "redis.jiang.operator/redisCluster"
	RedisClusterCommonLabelKey = "app"
)

func MutateDeployment(redisApp *redisv1.RedisCluster, dep *appsv1.Deployment) {

	labels := map[string]string{
		"redisCluster": redisApp.Name,
	}

	selector := metav1.LabelSelector{
		MatchLabels: labels,
	}

	dep.Spec = appsv1.DeploymentSpec{
		Replicas: redisApp.Spec.Size,
		Selector: &selector,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Containers: newContainer(redisApp),
			},
		},
	}

}

func newContainer(redisApp *redisv1.RedisCluster) []corev1.Container {

	containerPorts := make([]corev1.ContainerPort, 0)

	for _, p := range redisApp.Spec.Ports {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			ContainerPort: p.TargetPort.IntVal,
		})
	}

	return []corev1.Container{
		{
			Name:      redisApp.Name,
			Image:     redisApp.Spec.Image,
			Ports:     containerPorts,
			Env:       redisApp.Spec.Envs,
			Resources: redisApp.Spec.Resources,
		},
	}

}
