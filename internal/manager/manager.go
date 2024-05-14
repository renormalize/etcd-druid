// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gardener/etcd-druid/internal/client/kubernetes"
	"github.com/gardener/etcd-druid/internal/controller/compaction"
	"github.com/gardener/etcd-druid/internal/controller/etcd"
	"github.com/gardener/etcd-druid/internal/controller/etcdcopybackupstask"
	"github.com/gardener/etcd-druid/internal/controller/secret"
	"github.com/gardener/etcd-druid/internal/webhook/sentinel"
	coordinationv1 "k8s.io/api/coordination/v1"
	coordinationv1beta1 "k8s.io/api/coordination/v1beta1"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	eventsv1beta1 "k8s.io/api/events/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	defaultTimeout = time.Minute
)

// InitializeManager creates a controller manager and adds all the controllers and webhooks to the controller-manager using the passed in Config.
func InitializeManager(config *Config) (ctrl.Manager, error) {
	var (
		err error
		mgr ctrl.Manager
	)
	config.populateControllersFeatureGates()
	if mgr, err = createManager(config); err != nil {
		return nil, err
	}
	if err = registerControllers(mgr, config); err != nil {
		return nil, err
	}
	if err = registerWebhooks(mgr, config); err != nil {
		return nil, err
	}
	return mgr, nil
}

func createManager(config *Config) (ctrl.Manager, error) {
	// TODO: this can be removed once we have an improved informer, see https://github.com/gardener/etcd-druid/issues/215
	// list of objects which should not be cached.
	uncachedObjects := []client.Object{
		&corev1.Event{},
		&eventsv1beta1.Event{},
		&eventsv1.Event{},
	}

	if config.DisableLeaseCache {
		uncachedObjects = append(uncachedObjects, &coordinationv1.Lease{}, &coordinationv1beta1.Lease{})
	}

	// TODO: remove this check once `--metrics-addr` flag is removed, and directly compute the address:port when setting managerOptions.Metrics.BindAddress
	if !strings.Contains(config.Server.Metrics.BindAddress, ":") {
		config.Server.Metrics.BindAddress = net.JoinHostPort(config.Server.Metrics.BindAddress, strconv.Itoa(config.Server.Metrics.Port))
	}

	return ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Client: client.Options{
			Cache: &client.CacheOptions{
				DisableFor: uncachedObjects,
			},
		},
		Scheme: kubernetes.Scheme,
		Metrics: metricsserver.Options{
			BindAddress: config.Server.Metrics.BindAddress,
		},
		WebhookServer: webhook.NewServer(webhook.Options{
			Host:    config.Server.Webhook.BindAddress,
			Port:    config.Server.Webhook.Port,
			CertDir: config.Server.Webhook.TLSConfig.ServerCertDir,
		}),
		LeaderElection:             config.LeaderElection.Enabled,
		LeaderElectionID:           config.LeaderElection.ID,
		LeaderElectionResourceLock: config.LeaderElection.ResourceLock,
	})
}

func registerControllers(mgr ctrl.Manager, config *Config) error {
	var err error

	// Add etcd reconciler to the manager
	etcdReconciler, err := etcd.NewReconciler(mgr, config.EtcdControllerConfig)
	if err != nil {
		return err
	}
	if err = etcdReconciler.RegisterWithManager(mgr); err != nil {
		return err
	}

	// Add compaction reconciler to the manager if the CLI flag enable-backup-compaction is true.
	if config.CompactionControllerConfig.EnableBackupCompaction {
		compactionReconciler, err := compaction.NewReconciler(mgr, config.CompactionControllerConfig)
		if err != nil {
			return err
		}
		if err = compactionReconciler.RegisterWithManager(mgr); err != nil {
			return err
		}
	}

	// Add etcd-copy-backups-task reconciler to the manager
	etcdCopyBackupsTaskReconciler, err := etcdcopybackupstask.NewReconciler(mgr, config.EtcdCopyBackupsTaskControllerConfig)
	if err != nil {
		return err
	}
	if err = etcdCopyBackupsTaskReconciler.RegisterWithManager(mgr); err != nil {
		return err
	}

	// Add secret reconciler to the manager
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return secret.NewReconciler(
		mgr,
		config.SecretControllerConfig,
	).RegisterWithManager(ctx, mgr)
}

func registerWebhooks(mgr ctrl.Manager, config *Config) error {
	// Add sentinel webhook to the manager
	if config.SentinelWebhookConfig.Enabled {
		sentinelWebhook, err := sentinel.NewHandler(
			mgr,
			config.SentinelWebhookConfig,
		)
		if err != nil {
			return err
		}
		return sentinelWebhook.RegisterWithManager(mgr)
	}
	return nil
}