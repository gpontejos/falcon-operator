package falcon

import (
	"context"
	"reflect"

	falconv1alpha1 "github.com/crowdstrike/falcon-operator/api/falcon/v1alpha1"
	"github.com/crowdstrike/falcon-operator/internal/controller/assets"
	k8sutils "github.com/crowdstrike/falcon-operator/internal/controller/common"
	"github.com/crowdstrike/falcon-operator/pkg/common"
	"github.com/crowdstrike/falcon-operator/pkg/k8s_utils"
	"github.com/crowdstrike/falcon-operator/pkg/node"
	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-lib/proxy"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	clog "sigs.k8s.io/controller-runtime/pkg/log"
)

// handleNamespace creates and updates the namespace
func (r *FalconNodeSensorReconciler) handleNamespace(ctx context.Context, req ctrl.Request, nodesensor *falconv1alpha1.FalconNodeSensor, logger logr.Logger) (bool, error) {
	ns := corev1.Namespace{}
	err := common.GetNamespacedObject(ctx, r.Client, r.Reader, types.NamespacedName{Name: nodesensor.Spec.InstallNamespace}, &ns)
	if errors.IsNotFound(err) {
		ns = corev1.Namespace{
			TypeMeta: metav1.TypeMeta{
				APIVersion: corev1.SchemeGroupVersion.String(),
				Kind:       "Namespace",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: nodesensor.Spec.InstallNamespace,
			},
		}

		err = k8sutils.Create(r.Client, r.Scheme, ctx, req, logger, nodesensor, &nodesensor.Status, &ns)
		if err != nil {
			return false, err
		}

		return true, nil
	} else if err != nil {
		logger.Error(err, "Failed to get FalconNodeSensor Namespace")
		return false, err
	}

	return false, nil
}

// handlePriorityClass creates and updates the priority class
func (r *FalconNodeSensorReconciler) handlePriorityClass(ctx context.Context, req ctrl.Request, nodesensor *falconv1alpha1.FalconNodeSensor, logger logr.Logger) error {
	existingPC := &schedulingv1.PriorityClass{}
	pcName := nodesensor.Spec.Node.PriorityClass.Name
	update := false

	if pcName == "" && nodesensor.Spec.Node.GKE.Enabled == nil && nodesensor.Spec.Node.PriorityClass.Deploy == nil {
		return nil
	} else if pcName != "" && nodesensor.Spec.Node.PriorityClass.Deploy == nil &&
		(nodesensor.Spec.Node.GKE.Enabled != nil && *nodesensor.Spec.Node.GKE.Enabled) {
		//logger.Info("Skipping PriorityClass creation on GKE AutoPilot because an existing priority class name was provided")
		return nil
	} else if pcName != "" && (nodesensor.Spec.Node.PriorityClass.Deploy == nil || !*nodesensor.Spec.Node.PriorityClass.Deploy) {
		//logger.Info("Skipping PriorityClass creation because an existing priority class name was provided")
		return nil
	}

	if pcName == "" {
		pcName = nodesensor.Name + "-priorityclass"
		nodesensor.Spec.Node.PriorityClass.Name = pcName
	}

	pc := assets.PriorityClass(pcName, nodesensor.Spec.Node.PriorityClass.Value)

	err := common.GetNamespacedObject(ctx, r.Client, r.Reader, types.NamespacedName{Name: pcName, Namespace: nodesensor.Spec.InstallNamespace}, existingPC)
	if errors.IsNotFound(err) {
		err = k8sutils.Create(r.Client, r.Scheme, ctx, req, logger, nodesensor, &nodesensor.Status, pc)
		if err != nil {
			return err
		}
		logger.Info("Creating FalconNodeSensor PriorityClass")

		return nil
	} else if err != nil {
		logger.Error(err, "Failed to get FalconNodeSensor PriorityClass")
		return err
	}

	if nodesensor.Spec.Node.PriorityClass.Value != nil && existingPC.Value != *nodesensor.Spec.Node.PriorityClass.Value {
		update = true
	}

	if nodesensor.Spec.Node.PriorityClass.Name != "" && existingPC.Name != nodesensor.Spec.Node.PriorityClass.Name {
		update = true
	}

	if update {
		err = r.Delete(ctx, existingPC)
		if err != nil {
			return err
		}

		err = k8sutils.Create(r.Client, r.Scheme, ctx, req, logger, nodesensor, &nodesensor.Status, pc)
		if err != nil {
			return err
		}
		logger.Info("Updating FalconNodeSensor PriorityClass")
	}

	return nil
}

// handleConfigMaps creates and updates the node sensor configmap
func (r *FalconNodeSensorReconciler) handleConfigMaps(ctx context.Context, req ctrl.Request, config *node.ConfigCache, nodesensor *falconv1alpha1.FalconNodeSensor, logger logr.Logger) (*corev1.ConfigMap, bool, error) {
	var updated bool
	cmName := nodesensor.Name + "-config"
	confCm := &corev1.ConfigMap{}
	configmap := assets.SensorConfigMap(cmName, nodesensor.Spec.InstallNamespace, common.FalconKernelSensor, config.SensorEnvVars())

	err := common.GetNamespacedObject(ctx, r.Client, r.Reader, types.NamespacedName{Name: cmName, Namespace: nodesensor.Spec.InstallNamespace}, confCm)
	if errors.IsNotFound(err) {
		// does not exist, create
		err = k8sutils.Create(r.Client, r.Scheme, ctx, req, logger, nodesensor, &nodesensor.Status, configmap)
		if err != nil {
			if errors.IsAlreadyExists(err) {
				// We have got NotFound error during the Get(), but then we have got AlreadyExists error from Create(). Client cache is invalid.
				_ = k8sutils.Update(r.Client, ctx, req, logger, nodesensor, &nodesensor.Status, configmap)
				return configmap, updated, nil
			} else {
				return nil, updated, err

			}
		}

		logger.Info("Creating FalconNodeSensor Configmap")
		return nil, updated, nil
	} else if err != nil {
		logger.Error(err, "error getting Configmap")
		return nil, updated, err
	}

	if !reflect.DeepEqual(confCm.Data, configmap.Data) {
		err = k8sutils.Update(r.Client, ctx, req, logger, nodesensor, &nodesensor.Status, configmap)
		if err != nil {
			logger.Error(err, "Failed to update Configmap", "Configmap.Namespace", nodesensor.Spec.InstallNamespace, "Configmap.Name", cmName)
			return nil, updated, err
		}

		updated = true
	}

	return confCm, updated, nil
}

// handleCrowdStrikeSecrets creates and updates the image pull secrets for the nodesensor
func (r *FalconNodeSensorReconciler) handleCrowdStrikeSecrets(ctx context.Context, req ctrl.Request, config *node.ConfigCache, nodesensor *falconv1alpha1.FalconNodeSensor, logger logr.Logger) error {
	if !config.UsingCrowdStrikeRegistry() {
		return nil
	}
	secret := corev1.Secret{}

	err := common.GetNamespacedObject(ctx, r.Client, r.Reader, types.NamespacedName{Name: common.FalconPullSecretName, Namespace: nodesensor.Spec.InstallNamespace}, &secret)
	if err == nil || !errors.IsNotFound(err) {
		return err
	}

	pulltoken, err := config.GetPullToken(ctx)
	if err != nil {
		return err
	}

	secretData := map[string][]byte{corev1.DockerConfigJsonKey: common.CleanDecodedBase64(pulltoken)}
	secret = *assets.Secret(common.FalconPullSecretName, nodesensor.Spec.InstallNamespace, common.FalconKernelSensor, secretData, corev1.SecretTypeDockerConfigJson)
	err = k8sutils.Create(r.Client, r.Scheme, ctx, req, logger, nodesensor, &nodesensor.Status, &secret)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
	} else {
		logger.Info("Created a new Pull Secret", "Secret.Namespace", nodesensor.Spec.InstallNamespace, "Secret.Name", common.FalconPullSecretName)
	}
	return nil
}

func updateDaemonSetContainerProxy(ds *appsv1.DaemonSet, logger logr.Logger) bool {
	updated := false
	if len(proxy.ReadProxyVarsFromEnv()) > 0 {
		for i, container := range ds.Spec.Template.Spec.Containers {
			newContainerEnv := common.AppendUniqueEnvVars(container.Env, proxy.ReadProxyVarsFromEnv())
			updatedContainerEnv := common.UpdateEnvVars(container.Env, proxy.ReadProxyVarsFromEnv())
			if !equality.Semantic.DeepEqual(ds.Spec.Template.Spec.Containers[i].Env, newContainerEnv) {
				ds.Spec.Template.Spec.Containers[i].Env = newContainerEnv
				updated = true
			}
			if !equality.Semantic.DeepEqual(ds.Spec.Template.Spec.Containers[i].Env, updatedContainerEnv) {
				ds.Spec.Template.Spec.Containers[i].Env = updatedContainerEnv
				updated = true
			}
			if updated {
				logger.Info("Updating FalconNodeSensor DaemonSet Proxy Settings")
			}
		}
	}

	return updated
}

// If an update is needed, this will update the tolerations from the given DaemonSet
func (r *FalconNodeSensorReconciler) updateDaemonSetTolerations(ctx context.Context, req ctrl.Request, ds *appsv1.DaemonSet, nodesensor *falconv1alpha1.FalconNodeSensor, logger logr.Logger) (bool, error) {
	tolerations := &ds.Spec.Template.Spec.Tolerations
	origTolerations := nodesensor.Spec.Node.Tolerations
	tolerationsUpdate := !equality.Semantic.DeepEqual(*tolerations, *origTolerations)
	if tolerationsUpdate {
		logger.Info("Updating FalconNodeSensor DaemonSet Tolerations")
		mergedTolerations := k8s_utils.MergeTolerations(*tolerations, *origTolerations)
		*tolerations = mergedTolerations
		nodesensor.Spec.Node.Tolerations = &mergedTolerations

		// Double check if we need this.
		err := k8sutils.Update(r.Client, ctx, req, logger, nodesensor, &nodesensor.Status, ds)
		if err != nil {
			logger.Error(err, "Failed to update FalconNodeSensor Tolerations")
			return false, err
		}
	}
	return tolerationsUpdate, nil
}

// If an update is needed, this will update the affinity from the given DaemonSet
func updateDaemonSetAffinity(ds *appsv1.DaemonSet, nodesensor *falconv1alpha1.FalconNodeSensor, logger logr.Logger) bool {
	nodeAffinity := ds.Spec.Template.Spec.Affinity
	origNodeAffinity := corev1.Affinity{NodeAffinity: &nodesensor.Spec.Node.NodeAffinity}
	affinityUpdate := !equality.Semantic.DeepEqual(nodeAffinity.NodeAffinity, origNodeAffinity.NodeAffinity)
	if affinityUpdate {
		logger.Info("Updating FalconNodeSensor DaemonSet NodeAffinity")
		*nodeAffinity = origNodeAffinity
	}
	return affinityUpdate
}

// If an update is needed, this will update the containervolumes from the given DaemonSet
func updateDaemonSetContainerVolumes(ds, origDS *appsv1.DaemonSet, logger logr.Logger) bool {
	containerVolumeMounts := &ds.Spec.Template.Spec.Containers[0].VolumeMounts
	containerVolumeMountsUpdates := !equality.Semantic.DeepEqual(*containerVolumeMounts, origDS.Spec.Template.Spec.Containers[0].VolumeMounts)
	if containerVolumeMountsUpdates {
		logger.Info("Updating FalconNodeSensor DaemonSet Container volumeMounts")
		*containerVolumeMounts = origDS.Spec.Template.Spec.Containers[0].VolumeMounts
	}

	containerVolumeMounts = &ds.Spec.Template.Spec.InitContainers[0].VolumeMounts
	containerVolumeMountsUpdates = !equality.Semantic.DeepEqual(*containerVolumeMounts, origDS.Spec.Template.Spec.InitContainers[0].VolumeMounts)
	if containerVolumeMountsUpdates {
		logger.Info("Updating FalconNodeSensor DaemonSet InitContainer volumeMounts")
		*containerVolumeMounts = origDS.Spec.Template.Spec.InitContainers[0].VolumeMounts
	}

	return containerVolumeMountsUpdates
}

// If an update is needed, this will update the volumes from the given DaemonSet
func updateDaemonSetVolumes(ds, origDS *appsv1.DaemonSet, logger logr.Logger) bool {
	volumeMounts := &ds.Spec.Template.Spec.Volumes
	volumeMountsUpdates := !equality.Semantic.DeepEqual(*volumeMounts, origDS.Spec.Template.Spec.Volumes)
	if volumeMountsUpdates {
		logger.Info("Updating FalconNodeSensor DaemonSet volumes")
		*volumeMounts = origDS.Spec.Template.Spec.Volumes
	}

	return volumeMountsUpdates
}

// If an update is needed, this will update the InitContainer image reference from the given DaemonSet
func updateDaemonSetImages(ds *appsv1.DaemonSet, origImg string, logger logr.Logger) bool {
	initImage := &ds.Spec.Template.Spec.InitContainers[0].Image
	imgUpdate := *initImage != origImg
	if imgUpdate {
		logger.Info("Updating FalconNodeSensor DaemonSet InitContainer image", "Original Image", origImg, "Current Image", initImage)
		*initImage = origImg
	}

	image := &ds.Spec.Template.Spec.Containers[0].Image
	imgUpdate = *image != origImg
	if imgUpdate {
		logger.Info("Updating FalconNodeSensor DaemonSet image", "Original Image", origImg, "Current Image", image)
		*image = origImg
	}

	return imgUpdate
}

// If an update is needed, this will update the resources from the given DaemonSet
func updateDaemonSetResources(ds, origDS *appsv1.DaemonSet, logger logr.Logger) bool {
	resources := &ds.Spec.Template.Spec.Containers[0].Resources
	resourcesUpdates := !equality.Semantic.DeepEqual(*resources, origDS.Spec.Template.Spec.Containers[0].Resources)
	if resourcesUpdates {
		logger.Info("Updating FalconNodeSensor DaemonSet resources")
		*resources = origDS.Spec.Template.Spec.Containers[0].Resources

	}

	return resourcesUpdates
}

func updateDaemonSetInitContainerResources(ds, origDS *appsv1.DaemonSet, logger logr.Logger) bool {
	resources := &ds.Spec.Template.Spec.InitContainers[0].Resources
	resourcesUpdates := !equality.Semantic.DeepEqual(*resources, origDS.Spec.Template.Spec.InitContainers[0].Resources)
	if resourcesUpdates {
		logger.Info("Updating FalconNodeSensor DaemonSet InitContainer resources")
		*resources = origDS.Spec.Template.Spec.InitContainers[0].Resources
	}

	return resourcesUpdates
}

// If an update is needed, this will update the priority class from the given DaemonSet
func updateDaemonSetPriorityClass(ds, origDS *appsv1.DaemonSet, logger logr.Logger) bool {
	priorityClass := &ds.Spec.Template.Spec.PriorityClassName
	priorityClassUpdates := *priorityClass != origDS.Spec.Template.Spec.PriorityClassName
	if priorityClassUpdates {
		logger.Info("Updating FalconNodeSensor DaemonSet priority class")
		*priorityClass = origDS.Spec.Template.Spec.PriorityClassName
	}

	return priorityClassUpdates
}

// If an update is needed, this will update the capabilities from the given DaemonSet
func updateDaemonSetCapabilities(ds, origDS *appsv1.DaemonSet, logger logr.Logger) bool {
	capabilities := &ds.Spec.Template.Spec.Containers[0].SecurityContext.Capabilities
	capabilitiesUpdates := !equality.Semantic.DeepEqual(*capabilities, origDS.Spec.Template.Spec.Containers[0].SecurityContext.Capabilities)
	if capabilitiesUpdates {
		logger.Info("Updating FalconNodeSensor DaemonSet Container capabilities")
		*capabilities = origDS.Spec.Template.Spec.Containers[0].SecurityContext.Capabilities
	}

	capabilities = &ds.Spec.Template.Spec.InitContainers[0].SecurityContext.Capabilities
	capabilitiesUpdates = !equality.Semantic.DeepEqual(*capabilities, origDS.Spec.Template.Spec.InitContainers[0].SecurityContext.Capabilities)
	if capabilitiesUpdates {
		logger.Info("Updating FalconNodeSensor DaemonSet InitContainer capabilities")
		*capabilities = origDS.Spec.Template.Spec.InitContainers[0].SecurityContext.Capabilities
	}

	return capabilitiesUpdates
}

// If an update is needed, this will update the init args from the given DaemonSet
func updateDaemonSetInitArgs(ds, origDS *appsv1.DaemonSet, logger logr.Logger) bool {
	initArgs := &ds.Spec.Template.Spec.InitContainers[0].Args
	initArgsUpdates := !equality.Semantic.DeepEqual(*initArgs, origDS.Spec.Template.Spec.InitContainers[0].Args)
	if initArgsUpdates {
		logger.Info("Updating FalconNodeSensor DaemonSet init args")
		*initArgs = origDS.Spec.Template.Spec.InitContainers[0].Args
	}

	return initArgsUpdates
}

// handlePermissions creates and updates the service account, role and role binding
func (r *FalconNodeSensorReconciler) handlePermissions(ctx context.Context, req ctrl.Request, nodesensor *falconv1alpha1.FalconNodeSensor, logger logr.Logger) (bool, error) {
	created, err := r.handleServiceAccount(ctx, req, nodesensor, logger)
	if created || err != nil {
		return created, err
	}

	return r.handleClusterRoleBinding(ctx, req, nodesensor, logger)
}

// handleRoleBinding creates and updates RoleBinding
func (r *FalconNodeSensorReconciler) handleClusterRoleBinding(ctx context.Context, req ctrl.Request, nodesensor *falconv1alpha1.FalconNodeSensor, logger logr.Logger) (bool, error) {
	binding := rbacv1.ClusterRoleBinding{}

	err := common.GetNamespacedObject(ctx, r.Client, r.Reader, types.NamespacedName{Name: common.NodeClusterRoleBindingName}, &binding)
	if errors.IsNotFound(err) {
		binding = rbacv1.ClusterRoleBinding{
			TypeMeta: metav1.TypeMeta{
				APIVersion: rbacv1.SchemeGroupVersion.String(),
				Kind:       "ClusterRoleBinding",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:   common.NodeClusterRoleBindingName,
				Labels: common.CRLabels("clusterrolebinding", common.NodeClusterRoleBindingName, common.FalconKernelSensor),
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "falcon-operator-node-sensor-role",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      common.NodeServiceAccountName,
					Namespace: nodesensor.Spec.InstallNamespace,
				},
			},
		}

		logger.Info("Creating FalconNodeSensor ClusterRoleBinding")
		err = k8sutils.Create(r.Client, r.Scheme, ctx, req, logger, nodesensor, &nodesensor.Status, &binding)
		if err != nil && !errors.IsAlreadyExists(err) {
			logger.Error(err, "Failed to create new ClusterRoleBinding", "ClusteRoleBinding.Name", common.NodeClusterRoleBindingName)
			return false, err
		} else {
			logger.Info("ClusterRoleBinding already exists. Ignoring error.")
		}

		return true, nil
	} else if err != nil {
		logger.Error(err, "Failed to get FalconNodeSensor ClusterRoleBinding")
		return false, err
	}

	return false, nil
}

// handleServiceAccount creates and updates the service account and grants necessary permissions to it
func (r *FalconNodeSensorReconciler) handleServiceAccount(ctx context.Context, req ctrl.Request, nodesensor *falconv1alpha1.FalconNodeSensor, logger logr.Logger) (bool, error) {
	sa := corev1.ServiceAccount{}

	err := common.GetNamespacedObject(ctx, r.Client, r.Reader, types.NamespacedName{Name: common.NodeServiceAccountName, Namespace: nodesensor.Spec.InstallNamespace}, &sa)
	if errors.IsNotFound(err) {
		sa = corev1.ServiceAccount{
			TypeMeta: metav1.TypeMeta{
				APIVersion: corev1.SchemeGroupVersion.String(),
				Kind:       "ServiceAccount",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: nodesensor.Spec.InstallNamespace,
				Name:      common.NodeServiceAccountName,
				Labels:    common.CRLabels("serviceaccount", common.NodeServiceAccountName, common.FalconKernelSensor),
			},
		}

		logger.Info("Creating FalconNodeSensor ServiceAccount")
		err = k8sutils.Create(r.Client, r.Scheme, ctx, req, logger, nodesensor, &nodesensor.Status, &sa)
		if !errors.IsAlreadyExists(err) {
			logger.Error(err, "Failed to create new ServiceAccount", "Namespace.Name", nodesensor.Spec.InstallNamespace, "ServiceAccount.Name", common.NodeServiceAccountName)
			return false, err
		}

		return true, nil
	} else if err != nil {
		logger.Error(err, "Failed to get FalconNodeSensor ServiceAccount")
		return false, err
	}

	return false, nil
}

// handleServiceAccount creates and updates the service account and grants necessary permissions to it
func (r *FalconNodeSensorReconciler) handleSAAnnotations(ctx context.Context, req ctrl.Request, nodesensor *falconv1alpha1.FalconNodeSensor, logger logr.Logger) error {
	sa := corev1.ServiceAccount{}
	saAnnotations := nodesensor.Spec.Node.ServiceAccount.Annotations

	err := common.GetNamespacedObject(ctx, r.Client, r.Reader, types.NamespacedName{Name: common.NodeServiceAccountName, Namespace: nodesensor.Spec.InstallNamespace}, &sa)
	if errors.IsNotFound(err) {
		logger.Error(err, "Could not get FalconNodeSensor ServiceAccount")
		return err
	}

	// If there are no existing annotations, go ahead and create a map
	if sa.Annotations == nil {
		sa.Annotations = make(map[string]string)
	}

	// Add the CR configured annotations to the service account
	for key, value := range saAnnotations {
		sa.Annotations[key] = value
	}

	err = k8sutils.Update(r.Client, ctx, req, logger, nodesensor, &nodesensor.Status, &sa)
	if err != nil {
		return err
	}
	logger.Info("Updating FalconNodeSensor ServiceAccount Annotations", "Annotations", saAnnotations)

	return nil
}

// finalizeDaemonset deletes the Daemonset running the Falcon Sensor and then runs a Daemonset to cleanup the /opt/CrowdStrike directory
func (r *FalconNodeSensorReconciler) finalizeDaemonset(ctx context.Context, req ctrl.Request, image string, serviceAccount string, nodesensor *falconv1alpha1.FalconNodeSensor, logger logr.Logger) error {
	dsCleanupName := nodesensor.Name + "-cleanup"
	daemonset := &appsv1.DaemonSet{}
	pods := corev1.PodList{}
	dsList := &appsv1.DaemonSetList{}
	var nodeCount int32 = 0

	// Get a list of DS and return the DS within the correct NS
	listOptions := &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{common.FalconComponentKey: common.FalconKernelSensor}),
		Namespace:     nodesensor.Spec.InstallNamespace,
	}

	if err := r.List(ctx, dsList, listOptions); err != nil {
		if err = r.Reader.List(ctx, dsList, listOptions); err != nil {
			return err
		}
	}

	// Delete the Daemonset containing the sensor
	if err := r.Delete(ctx,
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodesensor.Name, Namespace: nodesensor.Spec.InstallNamespace,
			},
		}); err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Failed to cleanup Falcon sensor DaemonSet pods")
		return err
	}

	// Check if the cleanup DS is created. If not, create it.
	err := common.GetNamespacedObject(ctx, r.Client, r.Reader, types.NamespacedName{Name: dsCleanupName, Namespace: nodesensor.Spec.InstallNamespace}, daemonset)
	if errors.IsNotFound(err) {
		// Define a new DS for cleanup
		ds := assets.RemoveNodeDirDaemonset(dsCleanupName, image, serviceAccount, nodesensor)

		// Create the cleanup DS
		err = k8sutils.Create(r.Client, r.Scheme, ctx, req, logger, nodesensor, &nodesensor.Status, ds)
		if err != nil {
			logger.Error(err, "Failed to delete node directory with cleanup DaemonSet", "Path", common.FalconHostInstallDir)
			return err
		}

		// Start inifite loop to check that all pods have either completed or are running in the DS
		for {
			// List all pods with the "cleanup" label in the appropriate NS
			cleanupListOptions := &client.ListOptions{
				LabelSelector: labels.SelectorFromSet(labels.Set{common.FalconInstanceNameKey: "cleanup"}),
				Namespace:     nodesensor.Spec.InstallNamespace,
			}
			if err := r.List(ctx, &pods, cleanupListOptions); err != nil {
				if err = r.Reader.List(ctx, &pods, cleanupListOptions); err != nil {
					return err
				}
			}

			// Reset completedCount each loop, to ensure we don't count the same node(s) multiple times
			var completedCount int32 = 0
			// Reset the nodeCount to the desired number of pods to be scheduled for cleanup each loop, in case the cluster has scaled down
			for _, dSet := range dsList.Items {
				nodeCount = dSet.Status.DesiredNumberScheduled
				logger.Info("Setting DaemonSet node count", "Number of nodes", nodeCount)
			}

			// When the pods have a status of completed or running, increment the count.
			// The reason running is an acceptable value is because the pods should be running the sleep command and have already cleaned up /opt/CrowdStrike
			for _, pod := range pods.Items {
				if pod.Status.Phase == "Completed" || pod.Status.Phase == "Running" || pod.Status.Phase == "CrashLoopBackOff" {
					completedCount++
				}
			}

			// Break out of the infinite loop for cleanup when the completed or running DS count reaches the desired node count
			if completedCount == nodeCount {
				logger.Info("Clean up pods should be done. Continuing deleting.")
				break
			} else if completedCount < nodeCount && completedCount > 0 {
				logger.Info("Waiting for cleanup pods to complete. Retrying....", "Number of pods still processing task", completedCount)
			}

			err = common.GetNamespacedObject(ctx, r.Client, r.Reader, types.NamespacedName{Name: dsCleanupName, Namespace: nodesensor.Spec.InstallNamespace}, daemonset)
			if errors.IsNotFound(err) {
				logger.Info("Clean-up daemonset has been removed")
				break
			}
		}

		// The cleanup DS should be completed so delete the cleanup DS
		if err := r.Delete(ctx,
			&appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name: dsCleanupName, Namespace: nodesensor.Spec.InstallNamespace,
				},
			}); err != nil && !errors.IsNotFound(err) {
			logger.Error(err, "Failed to cleanup Falcon sensor DaemonSet pods")
			return err
		}

		// If we have gotten here, the cleanup should be successful
		logger.Info("Successfully deleted node directory", "Path", common.FalconDataDir)
	} else if err != nil {
		logger.Error(err, "error getting the cleanup DaemonSet")
		return err
	}

	logger.Info("Successfully finalized daemonset")
	return nil
}

func (r *FalconNodeSensorReconciler) reconcileObjectWithName(ctx context.Context, name types.NamespacedName) error {
	obj := &falconv1alpha1.FalconNodeSensor{}
	err := r.Get(ctx, name, obj)
	if err != nil {
		return err
	}

	clog.FromContext(ctx).Info("reconciling FalconNodeSensor object", "namespace", obj.Namespace, "name", obj.Name)
	r.reconcileObject(obj)
	return nil
}

func shouldTrackSensorVersions(obj *falconv1alpha1.FalconNodeSensor) bool {
	return obj.Spec.FalconAPI != nil && obj.Spec.Node.Advanced.IsAutoUpdating()
}
