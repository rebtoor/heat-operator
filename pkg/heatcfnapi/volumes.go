package heatcfnapi

import (
	"github.com/openstack-k8s-operators/heat-operator/pkg/heat"
	corev1 "k8s.io/api/core/v1"
)

// getVolumes -
func getVolumes(parentName string, name string) []corev1.Volume {
	var config0640AccessMode int32 = 0640

	volumes := []corev1.Volume{
		{
			Name: "config-data-custom",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					DefaultMode: &config0640AccessMode,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: name + "-config-data",
					},
				},
			},
		},
	}

	return append(heat.GetVolumes(parentName), volumes...)
}

// getInitVolumeMounts - heat Cfn API init task VolumeMounts
func getInitVolumeMounts() []corev1.VolumeMount {

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "config-data-custom",
			MountPath: "/var/lib/config-data/custom",
			ReadOnly:  true,
		},
	}

	return append(heat.GetInitVolumeMounts(), volumeMounts...)
}

// getVolumeMounts - heat Cfn API VolumeMounts
func getVolumeMounts() []corev1.VolumeMount {

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "config-data-merged",
			MountPath: "/var/lib/kolla/config_files/config.json",
			SubPath:   "heat-cfnapi-config.json",
			ReadOnly:  true,
		},
	}

	return append(heat.GetVolumeMounts(), volumeMounts...)
}
