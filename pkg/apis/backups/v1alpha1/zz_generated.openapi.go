// +build !ignore_autogenerated

// Code generated by openapi-gen. DO NOT EDIT.

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1.VolumeBackup":               schema_pkg_apis_backups_v1alpha1_VolumeBackup(ref),
		"github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1.VolumeBackupProvider":       schema_pkg_apis_backups_v1alpha1_VolumeBackupProvider(ref),
		"github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1.VolumeBackupProviderSpec":   schema_pkg_apis_backups_v1alpha1_VolumeBackupProviderSpec(ref),
		"github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1.VolumeBackupProviderStatus": schema_pkg_apis_backups_v1alpha1_VolumeBackupProviderStatus(ref),
		"github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1.VolumeBackupSpec":           schema_pkg_apis_backups_v1alpha1_VolumeBackupSpec(ref),
		"github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1.VolumeBackupStatus":         schema_pkg_apis_backups_v1alpha1_VolumeBackupStatus(ref),
	}
}

func schema_pkg_apis_backups_v1alpha1_VolumeBackup(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "VolumeBackup is the Schema for the volumebackups API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1.VolumeBackupSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1.VolumeBackupStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1.VolumeBackupSpec", "github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1.VolumeBackupStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_backups_v1alpha1_VolumeBackupProvider(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "VolumeBackupProvider is the Schema for the volumebackupproviders API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1.VolumeBackupProviderSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1.VolumeBackupProviderStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1.VolumeBackupProviderSpec", "github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1.VolumeBackupProviderStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_backups_v1alpha1_VolumeBackupProviderSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "VolumeBackupProviderSpec defines the desired state of VolumeBackupProvider",
				Properties:  map[string]spec.Schema{},
			},
		},
		Dependencies: []string{},
	}
}

func schema_pkg_apis_backups_v1alpha1_VolumeBackupProviderStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "VolumeBackupProviderStatus defines the observed state of VolumeBackupProvider",
				Properties:  map[string]spec.Schema{},
			},
		},
		Dependencies: []string{},
	}
}

func schema_pkg_apis_backups_v1alpha1_VolumeBackupSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "VolumeBackupSpec defines the desired state of VolumeBackup",
				Properties: map[string]spec.Schema{
					"applicationName": {
						SchemaProps: spec.SchemaProps{
							Description: "INSERT ADDITIONAL SPEC FIELDS - desired state of cluster Important: Run \"operator-sdk generate k8s\" to regenerate code after modifying this file Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"volumeName": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"containerName": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
				Required: []string{"applicationName", "volumeName", "containerName"},
			},
		},
		Dependencies: []string{},
	}
}

func schema_pkg_apis_backups_v1alpha1_VolumeBackupStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "VolumeBackupStatus defines the observed state of VolumeBackup",
				Properties: map[string]spec.Schema{
					"volumeBackupCondition": {
						SchemaProps: spec.SchemaProps{
							Description: "INSERT ADDITIONAL STATUS FIELD - define observed state of cluster Important: Run \"operator-sdk generate k8s\" to regenerate code after modifying this file Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html",
							Ref:         ref("github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1.VolumeBackupCondition"),
						},
					},
				},
				Required: []string{"volumeBackupCondition"},
			},
		},
		Dependencies: []string{
			"github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1.VolumeBackupCondition"},
	}
}