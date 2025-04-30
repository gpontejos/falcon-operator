package falcon

import (
	"context"

	falconv1alpha1 "github.com/crowdstrike/falcon-operator/api/falcon/v1alpha1"
	"github.com/crowdstrike/falcon-operator/internal/controller/assets"
	k8sutils "github.com/crowdstrike/falcon-operator/internal/controller/common"
	"github.com/crowdstrike/falcon-operator/internal/controller/common/sensorversion"
	"github.com/crowdstrike/falcon-operator/pkg/common"
	"github.com/crowdstrike/falcon-operator/pkg/k8s_utils"
	"github.com/crowdstrike/falcon-operator/pkg/node"
	"github.com/crowdstrike/falcon-operator/version"
	"github.com/crowdstrike/gofalcon/falcon"
	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-lib/proxy"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	clog "sigs.k8s.io/controller-runtime/pkg/log"
)

// FalconNodeSensorReconciler reconciles a FalconNodeSensor object
type FalconNodeSensorReconciler struct {
	client.Client
	Reader          client.Reader
	Log             logr.Logger
	Scheme          *runtime.Scheme
	reconcileObject func(client.Object)
	tracker         sensorversion.Tracker
}

// SetupWithManager sets up the controller with the Manager.
func (r *FalconNodeSensorReconciler) SetupWithManager(mgr ctrl.Manager, tracker sensorversion.Tracker) error {
	nodeSensorController, err := ctrl.NewControllerManagedBy(mgr).
		For(&falconv1alpha1.FalconNodeSensor{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.DaemonSet{}).
		Owns(&corev1.Secret{}).
		Build(r)
	if err != nil {
		return err
	}

	r.reconcileObject, err = k8sutils.NewReconcileTrigger(nodeSensorController)
	if err != nil {
		return err
	}

	r.tracker = tracker
	return nil
}

// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;delete;deletecollection

//+kubebuilder:rbac:groups=falcon.crowdstrike.com,resources=falconnodesensors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=falcon.crowdstrike.com,resources=falconnodesensors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=falcon.crowdstrike.com,resources=falconnodesensors/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;create;update
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;create;update
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterrolebindings,verbs=get;list;watch;create
//+kubebuilder:rbac:groups="security.openshift.io",resources=securitycontextconstraints,resourceNames=privileged,verbs=use
//+kubebuilder:rbac:groups="scheduling.k8s.io",resources=priorityclasses,verbs=get;list;watch;create;delete;update
//+kubebuilder:rbac:groups="",resources=pods;services;nodes;daemonsets;replicasets;deployments;jobs;ingresses;cronjobs;persistentvolumes,verbs=get;watch;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *FalconNodeSensorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := clog.FromContext(ctx)
	logger := log.WithValues("DaemonSet", req.NamespacedName)
	logger.Info("reconciling FalconNodeSensor")

	// Fetch the FalconNodeSensor instance.
	nodesensor := &falconv1alpha1.FalconNodeSensor{}

	// Step 1: Check if Sensor Exists

	err := r.Get(ctx, req.NamespacedName, nodesensor)
	if errors.IsNotFound(err) {
		r.tracker.StopTracking(req.NamespacedName)

		// Request object not found, could have been deleted after reconcile request.
		// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
		// Return and don't requeue
		log.Info("FalconNodeSensor resource not found. Ignoring since object must be deleted")
		return ctrl.Result{}, nil
	}

	validate, err := k8sutils.CheckRunningPodLabels(r.Reader, ctx, nodesensor.Spec.InstallNamespace, common.CRLabels("daemonset", nodesensor.Name, common.FalconKernelSensor))
	if err != nil {
		return ctrl.Result{}, err
	}
	if !validate {
		err = k8sutils.ConditionsUpdate(r.Client, ctx, req, log, nodesensor, &nodesensor.Status, metav1.Condition{
			Type:    falconv1alpha1.ConditionFailed,
			Status:  metav1.ConditionFalse,
			Reason:  falconv1alpha1.ReasonReqNotMet,
			Message: "FalconNodeSensor must not be installed in a namespace with other workloads running. Please change the namespace in the CR configuration.",
		})
		if err != nil {
			return ctrl.Result{}, err
		}
		logger.Error(nil, "FalconNodeSensor is attempting to install in a namespace with existing pods. Please update the CR configuration to a namespace that does not have workoads already running.")
		return ctrl.Result{}, nil
	}

	dsCondition := meta.FindStatusCondition(nodesensor.Status.Conditions, falconv1alpha1.ConditionSuccess)
	if dsCondition == nil {
		err = k8sutils.ConditionsUpdate(r.Client, ctx, req, log, nodesensor, &nodesensor.Status, metav1.Condition{
			Type:    falconv1alpha1.ConditionPending,
			Status:  metav1.ConditionTrue,
			Reason:  falconv1alpha1.ReasonReqNotMet,
			Message: "FalconNodeSensor progressing",
		})
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	}

	if nodesensor.Status.Version != version.Get() {
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			err := r.Get(ctx, req.NamespacedName, nodesensor)
			if err != nil {
				return err
			}

			nodesensor.Status.Version = version.Get()
			return r.Status().Update(ctx, nodesensor)
		})
		if err != nil {
			log.Error(err, "Failed to update FalconNodeSensor status for nodesensor.Status.Version")
			return ctrl.Result{}, err
		}
	}

	created, err := r.handleNamespace(ctx, req, nodesensor, logger)
	if err != nil {
		return ctrl.Result{}, err
	}
	if created {
		return ctrl.Result{Requeue: true}, nil
	}

	err = r.handlePriorityClass(ctx, req, nodesensor, logger)
	if err != nil {
		return ctrl.Result{}, err
	}

	serviceAccount := common.NodeServiceAccountName

	created, err = r.handlePermissions(ctx, req, nodesensor, logger)
	if err != nil {
		return ctrl.Result{}, err
	}
	if created {
		return ctrl.Result{Requeue: true}, nil
	}

	if nodesensor.Spec.Node.ServiceAccount.Annotations != nil {
		err = r.handleSAAnnotations(ctx, req, nodesensor, logger)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	config, err := node.NewConfigCache(ctx, logger, nodesensor)
	if err != nil {
		return ctrl.Result{}, err
	}

	if shouldTrackSensorVersions(nodesensor) {
		getSensorVersion := sensorversion.NewFalconCloudQuery(falcon.NodeSensor, nodesensor.Spec.FalconAPI.ApiConfig())
		r.tracker.Track(req.NamespacedName, getSensorVersion, r.reconcileObjectWithName, nodesensor.Spec.Node.Advanced.IsAutoUpdatingForced())
	} else {
		r.tracker.StopTracking(req.NamespacedName)
	}

	sensorConf, updated, err := r.handleConfigMaps(ctx, req, config, nodesensor, logger)
	if err != nil {
		err = k8sutils.ConditionsUpdate(r.Client, ctx, req, log, nodesensor, &nodesensor.Status, metav1.Condition{
			Type:    falconv1alpha1.ConditionFailed,
			Status:  metav1.ConditionFalse,
			Reason:  falconv1alpha1.ReasonInstallFailed,
			Message: "FalconNodeSensor ConfigMap failed to be installed",
		})
		if err != nil {
			return ctrl.Result{}, err
		}

		logger.Error(err, "error handling configmap")
		return ctrl.Result{}, nil
	}
	if sensorConf == nil {
		err = k8sutils.ConditionsUpdate(r.Client, ctx, req, log, nodesensor, &nodesensor.Status, metav1.Condition{
			Type:    falconv1alpha1.ConditionConfigMapReady,
			Status:  metav1.ConditionTrue,
			Reason:  falconv1alpha1.ReasonInstallSucceeded,
			Message: "FalconNodeSensor ConfigMap has been successfully created",
		})
		if err != nil {
			return ctrl.Result{}, err
		}

		// this just got created, so re-queue.
		logger.Info("Configmap was just created. Re-queuing")
		return ctrl.Result{Requeue: true}, nil
	}
	if updated {
		err = k8sutils.ConditionsUpdate(r.Client, ctx, req, log, nodesensor, &nodesensor.Status, metav1.Condition{
			Type:    falconv1alpha1.ConditionConfigMapReady,
			Status:  metav1.ConditionTrue,
			Reason:  falconv1alpha1.ReasonUpdateSucceeded,
			Message: "FalconNodeSensor ConfigMap has been successfully updated",
		})
		if err != nil {
			return ctrl.Result{}, err
		}

		logger.Info("Configmap was updated")
	}

	err = r.handleCrowdStrikeSecrets(ctx, req, config, nodesensor, logger)
	if err != nil {
		return ctrl.Result{}, err
	}

	image, err := config.GetImageURI(ctx, logger)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Check if the daemonset already exists, if not create a new one
	daemonset := &appsv1.DaemonSet{}

	err = common.GetNamespacedObject(ctx, r.Client, r.Reader, types.NamespacedName{Name: nodesensor.Name, Namespace: nodesensor.Spec.InstallNamespace}, daemonset)
	if errors.IsNotFound(err) {
		ds := assets.Daemonset(nodesensor.Name, image, serviceAccount, nodesensor)

		if len(proxy.ReadProxyVarsFromEnv()) > 0 {
			for i, container := range ds.Spec.Template.Spec.Containers {
				ds.Spec.Template.Spec.Containers[i].Env = append(container.Env, proxy.ReadProxyVarsFromEnv()...)
			}
		}

		_, err = r.updateDaemonSetTolerations(ctx, req, ds, nodesensor, logger)
		if err != nil {
			return ctrl.Result{}, err
		}

		err = k8sutils.Create(r.Client, r.Scheme, ctx, req, log, nodesensor, &nodesensor.Status, ds)
		if err != nil {
			return ctrl.Result{}, err
		}

		err = k8sutils.ConditionsUpdate(r.Client, ctx, req, log, nodesensor, &nodesensor.Status, metav1.Condition{
			Type:    falconv1alpha1.ConditionDaemonSetReady,
			Status:  metav1.ConditionTrue,
			Reason:  falconv1alpha1.ReasonInstallSucceeded,
			Message: "FalconNodeSensor DaemonSet has been successfully installed",
		})
		if err != nil {
			return ctrl.Result{}, err
		}

		logger.Info("Created a new DaemonSet", "DaemonSet.Namespace", ds.Namespace, "DaemonSet.Name", ds.Name)
		// Daemonset created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil

	} else if err != nil {
		return ctrl.Result{}, err
	} else {
		// Copy Daemonset for updates
		dsUpdate := daemonset.DeepCopy()
		dsTarget := assets.Daemonset(dsUpdate.Name, image, serviceAccount, nodesensor)

		// Objects to check for updates to re-spin pods
		imgUpdate := updateDaemonSetImages(dsUpdate, image, logger)
		affUpdate := updateDaemonSetAffinity(dsUpdate, nodesensor, logger)
		containerVolUpdate := updateDaemonSetContainerVolumes(dsUpdate, dsTarget, logger)
		volumeUpdates := updateDaemonSetVolumes(dsUpdate, dsTarget, logger)
		resources := updateDaemonSetResources(dsUpdate, dsTarget, logger)
		initResources := updateDaemonSetInitContainerResources(dsUpdate, dsTarget, logger)
		pc := updateDaemonSetPriorityClass(dsUpdate, dsTarget, logger)
		capabilities := updateDaemonSetCapabilities(dsUpdate, dsTarget, logger)
		initArgs := updateDaemonSetInitArgs(dsUpdate, dsTarget, logger)
		proxyUpdates := updateDaemonSetContainerProxy(dsUpdate, logger)
		tolsUpdate, err := r.updateDaemonSetTolerations(ctx, req, dsUpdate, nodesensor, logger)
		if err != nil {
			return ctrl.Result{}, err
		}

		// Update the daemonset and re-spin pods with changes
		if imgUpdate || tolsUpdate || affUpdate || containerVolUpdate || volumeUpdates || resources || pc || capabilities || initArgs || initResources || proxyUpdates || updated {
			err = k8sutils.Update(r.Client, ctx, req, log, nodesensor, &nodesensor.Status, dsUpdate)
			if err != nil {
				return ctrl.Result{}, err
			}

			err := k8s_utils.RestartDaemonSet(ctx, r.Client, dsUpdate)
			if err != nil {
				logger.Error(err, "Failed to restart pods after DaemonSet configuration changed.")
				return ctrl.Result{}, err
			}

			err = k8sutils.ConditionsUpdate(r.Client, ctx, req, log, nodesensor, &nodesensor.Status, metav1.Condition{
				Type:    falconv1alpha1.ConditionDaemonSetReady,
				Status:  metav1.ConditionTrue,
				Reason:  falconv1alpha1.ReasonUpdateSucceeded,
				Message: "FalconNodeSensor DaemonSet has been successfully updated",
			})
			if err != nil {
				return ctrl.Result{}, err
			}
			logger.Info("FalconNodeSensor DaemonSet configuration changed. Pods have been restarted.")
		}
	}

	imgVer := common.ImageVersion(image)
	if nodesensor.Status.Sensor != imgVer {
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			err := r.Get(ctx, req.NamespacedName, nodesensor)
			if err != nil {
				return err
			}

			nodesensor.Status.Sensor = imgVer
			return r.Status().Update(ctx, nodesensor)
		})
		if err != nil {
			log.Error(err, "Failed to update FalconNodeSensor status for nodesensor.Status.Sensor")
			return ctrl.Result{}, err
		}
	}

	err = k8sutils.ConditionsUpdate(r.Client, ctx, req, log, nodesensor, &nodesensor.Status, metav1.Condition{
		Type:    falconv1alpha1.ConditionSuccess,
		Status:  metav1.ConditionTrue,
		Reason:  falconv1alpha1.ReasonInstallSucceeded,
		Message: "FalconNodeSensor installation completed",
	})
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	// Check if the FalconNodeSensor instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isDSMarkedToBeDeleted := nodesensor.GetDeletionTimestamp() != nil
	if isDSMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(nodesensor, common.FalconFinalizer) {
			logger.Info("Successfully finalized daemonset")
			// Allows the cleanup to be disabled by disableCleanup option
			if !*nodesensor.Spec.Node.NodeCleanup {
				// Run finalization logic for common.FalconFinalizer. If the
				// finalization logic fails, don't remove the finalizer so
				// that we can retry during the next reconciliation.
				if err := r.finalizeDaemonset(ctx, req, image, serviceAccount, nodesensor, logger); err != nil {
					return ctrl.Result{}, err
				}
			} else {
				logger.Info("Skipping cleanup because it is disabled", "disableCleanup", *nodesensor.Spec.Node.NodeCleanup)
			}

			// Remove common.FalconFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(nodesensor, common.FalconFinalizer)
			err = k8sutils.Update(r.Client, ctx, req, log, nodesensor, &nodesensor.Status, nodesensor)
			if err != nil {
				return ctrl.Result{}, err
			}
			log.Info("Removing finalizer")

		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(nodesensor, common.FalconFinalizer) {
		controllerutil.AddFinalizer(nodesensor, common.FalconFinalizer)
		err = k8sutils.Update(r.Client, ctx, req, log, nodesensor, &nodesensor.Status, nodesensor)
		if err != nil {
			logger.Error(err, "Unable to update finalizer")
			return ctrl.Result{}, err
		}
		log.Info("Adding finalizer")

	}

	return ctrl.Result{}, nil
}
