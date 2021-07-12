// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package tensorflow provides a Kubernetes controller for a TFJob resource.
package tensorflow

import (
	"context"
	"fmt"
	"github.com/kubeflow/tf-operator/pkg/common/util"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/controller.v1/common"
	tflogger "github.com/kubeflow/common/pkg/util"
	"github.com/kubeflow/tf-operator/cmd/tf-operator.v1/app/options"
	tfv1 "github.com/kubeflow/tf-operator/pkg/apis/tensorflow/v1"
	tfjobclientset "github.com/kubeflow/tf-operator/pkg/client/clientset/versioned"
	tfjobscheme "github.com/kubeflow/tf-operator/pkg/client/clientset/versioned/scheme"
	tfjobinformers "github.com/kubeflow/tf-operator/pkg/client/informers/externalversions"
	tfjobinformersv1 "github.com/kubeflow/tf-operator/pkg/client/informers/externalversions/tensorflow/v1"
	tfjoblisters "github.com/kubeflow/tf-operator/pkg/client/listers/tensorflow/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/apimachinery/pkg/runtime/schema"
	volcanoclient "volcano.sh/apis/pkg/client/clientset/versioned"
)

const (
	controllerName = "tf-operator"

	// labels for pods and servers.
	tfReplicaTypeLabel  = "replica-type"
	tfReplicaIndexLabel = "replica-index"
	labelGroupName      = "group-name"
	// Deprecated label for backwards compatibility. Has to be removed
	labelTFJobName = "tf-job-name"
	// volcanoTaskSpecKey task spec key used in pod annotation when EnableGangScheduling is true
	volcanoTaskSpecKey = "volcano.sh/task-spec"
)

var (
	// KeyFunc is the short name to DeletionHandlingMetaNamespaceKeyFunc.
	// IndexerInformer uses a delta queue, therefore for deletes we have to use this
	// key function but it should be just fine for non delete events.
	KeyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc

	tfJobsDeletedCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tf_operator_jobs_deleted_total",
			Help: "Counts number of TF jobs deleted",
		},
		[]string{"job_namespace"},
	)
)

// TFController is the type for TFJob Controller, which manages
// the lifecycle of TFJobs.
type TFController struct {
	common.JobController

	// tfJobClientSet is a clientset for CRD TFJob.
	tfJobClientSet tfjobclientset.Interface

	// To allow injection of sync functions for testing.
	syncHandler func(string) (bool, error)

	// tfJobInformer is a temporary field for unstructured informer support.
	tfJobInformer cache.SharedIndexInformer

	// Listers for TFJob, Pod and Service
	// tfJobLister can list/get tfjobs from the shared informer's store.
	tfJobLister tfjoblisters.TFJobLister

	// tfJobInformerSynced returns true if the tfjob store has been synced at least once.
	tfJobInformerSynced cache.InformerSynced
}

// NewTFController returns a new TFJob controller.
func NewTFController(
	// This variable is for unstructured informer.
	tfJobInformer tfjobinformersv1.TFJobInformer,
	kubeClientSet kubeclientset.Interface,
	volcanoClientSet volcanoclient.Interface,
	tfJobClientSet tfjobclientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	// This field is not used now but we keep it since it will be used
	// after we support CRD validation.
	tfJobInformerFactory tfjobinformers.SharedInformerFactory,
	option options.ServerOption) *TFController {

	err := tfjobscheme.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("Failed to add tfjob scheme: %v", err)
	}

	log.Info("Creating TFJob controller")
	// Create new TFController.
	tc := &TFController{
		tfJobClientSet: tfJobClientSet,
	}

	// Create base controller
	log.Info("Creating Job controller")

	jc := common.NewJobController(tc, metav1.Duration{Duration: 15 * time.Second},
		option.EnableGangScheduling, kubeClientSet, volcanoClientSet, kubeInformerFactory, "tfjobs")

	// Set sync handler.
	tc.syncHandler = tc.syncTFJob

	// TODO(ChanYiLin): these are originally for testing, but with using common library,
	// we can not replcae the function. Also need to update or remove some tests

	// tc.updateStatusHandler = tc.UpdateJobStatusInApiServer
	// set delete handler.
	// tc.deleteTFJobHandler = tc.DeleteJob

	// Set up an event handler for when tfjob resources change.
	tfJobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    tc.addTFJob,
		UpdateFunc: tc.updateTFJob,
		// This will enter the sync loop and no-op,
		// because the tfjob has been deleted from the store.
		DeleteFunc: tc.enqueueTFJob,
	})

	tc.tfJobInformer = tfJobInformer.Informer()
	tc.tfJobLister = tfJobInformer.Lister()
	tc.tfJobInformerSynced = tfJobInformer.Informer().HasSynced

	// Create pod informer.
	podInformer := kubeInformerFactory.Core().V1().Pods()

	// Set up an event handler for when pod resources change
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    jc.AddPod,
		UpdateFunc: jc.UpdatePod,
		DeleteFunc: jc.DeletePod,
	})

	// tc.PodLister = podInformer.Lister()
	// tc.PodInformerSynced = podInformer.Informer().HasSynced
	jc.PodLister = podInformer.Lister()
	jc.PodInformerSynced = podInformer.Informer().HasSynced

	// Create service informer.
	serviceInformer := kubeInformerFactory.Core().V1().Services()

	// Set up an event handler for when service resources change.
	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    jc.AddService,
		UpdateFunc: jc.UpdateService,
		DeleteFunc: jc.DeleteService,
	})

	// tc.ServiceLister = serviceInformer.Lister()
	// tc.ServiceInformerSynced = serviceInformer.Informer().HasSynced
	jc.ServiceLister = serviceInformer.Lister()
	jc.ServiceInformerSynced = serviceInformer.Informer().HasSynced

	tc.JobController = jc

	return tc
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (tc *TFController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer tc.WorkQueue.ShutDown()

	// Start the informer factories to begin populating the informer caches.
	log.Info("Starting TFJob controller")

	// Wait for the caches to be synced before starting workers.
	log.Info("Waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(stopCh, tc.tfJobInformerSynced,
		tc.PodInformerSynced, tc.ServiceInformerSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	log.Infof("Starting %v workers", threadiness)
	// Launch workers to process TFJob resources.
	for i := 0; i < threadiness; i++ {
		go wait.Until(tc.runWorker, time.Second, stopCh)
	}

	log.Info("Started workers")
	<-stopCh
	log.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (tc *TFController) runWorker() {
	for tc.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (tc *TFController) processNextWorkItem() bool {
	obj, quit := tc.WorkQueue.Get()
	if quit {
		return false
	}
	defer tc.WorkQueue.Done(obj)

	var key string
	var ok bool
	if key, ok = obj.(string); !ok {
		// As the item in the workqueue is actually invalid, we call
		// Forget here else we'd go into a loop of attempting to
		// process a work item that is invalid.
		tc.WorkQueue.Forget(obj)
		utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
		return true
	}
	logger := tflogger.LoggerForKey(key)

	tfJob, err := tc.getTFJobFromKey(key)
	if err != nil {
		if err == errNotExists {
			logger.Infof("TFJob has been deleted: %v", key)
			namespace, _, keyerr := cache.SplitMetaNamespaceKey(key)
			if keyerr == nil && len(namespace) != 0 {
				tfJobsDeletedCount.WithLabelValues(namespace).Inc()
			} else {
				logger.Errorf("Invalid TFJob key %s: Namespace is missing %v", key, keyerr)
			}
			return true
		}

		// Log the failure to conditions.
		logger.Errorf("Failed to get TFJob from key %s: %v", key, err)
		if err == errFailedMarshal {
			errMsg := fmt.Sprintf("Failed to unmarshal the object to TFJob object: %v", err)
			tflogger.LoggerForJob(tfJob).Warn(errMsg)
			tc.Recorder.Event(tfJob, v1.EventTypeWarning, failedMarshalTFJobReason, errMsg)
		}

		return true
	}

	// Sync TFJob to match the actual state to this desired state.
	forget, err := tc.syncHandler(key)
	if err == nil {
		if forget {
			tc.WorkQueue.Forget(key)
		}
		return true
	}

	utilruntime.HandleError(fmt.Errorf("error syncing tfjob: %v", err))
	tc.WorkQueue.AddRateLimited(key)

	return true
}

func (tc *TFController) enqueueTFJob(tfjob interface{}) {
	key, err := KeyFunc(tfjob)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for tfjob object %#v: %v", tfjob, err))
		return
	}

	// TODO: we may need add backoff here
	tc.WorkQueue.Add(key)
}

// syncTFJob syncs the tfjob with the given key if it has had its expectations fulfilled, meaning
// it did not expect to see any more of its pods/services created or deleted.
// This function is not meant to be invoked concurrently with the same key.
func (tc *TFController) syncTFJob(key string) (bool, error) {
	startTime := time.Now()
	logger := tflogger.LoggerForKey(key)
	defer func() {
		logger.Infof("Finished syncing tfjob %q (%v)", key, time.Since(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return false, err
	}
	if len(namespace) == 0 || len(name) == 0 {
		return false, fmt.Errorf("invalid tfjob key %q: either namespace or name is missing", key)
	}

	sharedTFJob, err := tc.getTFJobFromName(namespace, name)
	if err != nil {
		if err == errNotExists {
			logger.Infof("TFJob has been deleted: %v", key)
			tfJobsDeletedCount.WithLabelValues(namespace).Inc()
			return true, nil
		}
		return false, err
	}

	tfjob := sharedTFJob.DeepCopy()

	// Sync tfjob every time if EnableDynamicWorker is true
	jobKey, err := common.KeyFunc(tfjob)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get jobKey for job object %#v: %v", tfjob, err))
	}

	replicaTypes := util.GetReplicaTypes(tfjob.Spec.TFReplicaSpecs)
	tfjobNeedsSync := tfjob.Spec.EnableDynamicWorker || util.SatisfiedExpectations(tc.Expectations, jobKey, replicaTypes)

	// Set default for the new tfjob.
	scheme.Scheme.Default(tfjob)

	var reconcileTFJobsErr error
	if tfjobNeedsSync && tfjob.DeletionTimestamp == nil {
		reconcileTFJobsErr = tc.ReconcileJobs(tfjob, tfjob.Spec.TFReplicaSpecs, tfjob.Status, &tfjob.Spec.RunPolicy)
	}

	if reconcileTFJobsErr != nil {
		return false, reconcileTFJobsErr
	}

	return true, err
}

func (tc *TFController) GetJobFromInformerCache(namespace, name string) (metav1.Object, error) {
	return tc.getTFJobFromName(namespace, name)
}

func (tc *TFController) GetJobFromAPIClient(namespace, name string) (metav1.Object, error) {
	return tc.tfJobClientSet.KubeflowV1().TFJobs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func (tc *TFController) GetAPIGroupVersionKind() schema.GroupVersionKind {
	return tfv1.GroupVersion.WithKind(tfv1.Kind)
}

func (tc *TFController) GetAPIGroupVersion() schema.GroupVersion {
	return tfv1.GroupVersion
}

func (tc *TFController) GetGroupNameLabelKey() string {
	return labelGroupName
}

// Deprecated function for backwards compatibility. Has to be removed later
func (tc *TFController) GetJobNameLabelKey() string {
	return labelTFJobName
}

func (tc *TFController) GetGroupNameLabelValue() string {
	return tfv1.GroupVersion.Group
}

func (tc *TFController) GetReplicaTypeLabelKey() string {
	return tfReplicaTypeLabel
}

func (tc *TFController) GetReplicaIndexLabelKey() string {
	return tfReplicaIndexLabel
}

func (tc *TFController) ControllerName() string {
	return controllerName
}

func (tc *TFController) GetDefaultContainerName() string {
	return tfv1.DefaultContainerName
}

func (tc *TFController) GetDefaultContainerPortName() string {
	return tfv1.DefaultPortName
}

func (tc *TFController) IsMasterRole(replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec, rtype commonv1.ReplicaType, index int) bool {

	if ContainChieforMasterSpec(replicas) {
		return rtype == tfv1.TFReplicaTypeChief || rtype == tfv1.TFReplicaTypeMaster
	}
	// else check if it is worker with index 0
	return rtype == tfv1.TFReplicaTypeWorker && index == 0
}
