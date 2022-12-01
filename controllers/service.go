package controllers

import (
	corev1 "k8s.io/api/core/v1"
	redisv1 "operator-redis/api/v1"
)

func MutateService(redisApp *redisv1.RedisCluster, service *corev1.Service) {

	service.Spec = corev1.ServiceSpec{
		Type: corev1.ServiceType(redisApp.Spec.ServiceType),
		Selector: map[string]string{
			"redisCluster": redisApp.Name,
		},
		Ports: redisApp.Spec.Ports,
	}

}
