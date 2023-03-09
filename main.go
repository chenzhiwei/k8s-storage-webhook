/*
Copyright 2023 zhiwei@youya.org.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"github.com/chenzhiwei/k8s-storage-webhook/pkg/validator"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	//+kubebuilder:scaffold:imports
)

func init() {
	logf.SetLogger(zap.New())
	//+kubebuilder:scaffold:scheme
}

func main() {
	log := logf.Log.WithName("storage-webhook")

	// Setup a Manager
	log.Info("setting up manager")
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{
		Port: 8443,
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Setup Webhook
	log.Info("setting up webhook")
	if err := builder.WebhookManagedBy(mgr).
		For(&corev1.PersistentVolumeClaim{}).
		WithValidator(&validator.PersistentVolumeClaimValidator{Client: mgr.GetClient()}).
		Complete(); err != nil {
		log.Error(err, "unable to create webhook", "webhook", "PersistentVolumeClaim")
		os.Exit(1)
	}

	log.Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")
		os.Exit(1)
	}
}
