# Falcon Operator

## About FalconOperator Custom Resource (CR)
Falcon Operator introduces the FalconOperator Custom Resource (CR) to the cluster. The resource is meant to install, configure, and uninstall any of the Falcon CRDs within a single manifest - FalconAdmission, FalconContainer, FalconImageAnalyzer, and FalconNodeSensor.

### FalconOperator CR Configuration using CrowdStrike API Keys
To start the FalconOperator installation using CrowdStrike API Keys to allow the operator to determine your Falcon Customer ID (CID) as well as pull down the CrowdStrike FalconAdmission, FalconContainer, FalconImageAnalyzer, or FalconImageAnalyzer image, please create the following FalconAdmission resource to your cluster.

> [!IMPORTANT]
> You will need to provide CrowdStrike API Keys and CrowdStrike cloud region for the installation. It is recommended to establish new API credentials for the installation at https://falcon.crowdstrike.com/support/api-clients-and-keys, required permissions are:
> * For FalconImageAnalyzer:
>   * Falcon Container CLI: **Write**
>   * Falcon Container Image: **Read/Write**
>   * Falcon Images Download: **Read**
> * For FalconAdmission, FalconContainer, or FalconNodeSensor:
>   * Falcon Images Download: **Read**
>   * Sensor Download: **Read**



### FalconOperator Reference Manual

#### Falcon API Settings
| Spec                       | Description                                                                                              |
| :------------------------- | :------------------------------------------------------------------------------------------------------- |
| falcon_api.client_id       | CrowdStrike API Client ID                                                                                |
| falcon_api.client_secret   | CrowdStrike API Client Secret                                                                            |
| falcon_api.cloud_region    | CrowdStrike cloud region (allowed values: us-1, us-2, eu-1, us-gov-1)                                    |
| falcon_api.cid             | (optional) CrowdStrike Falcon CID API override                                                           |

#### FalconOperator Registry Settings
| Spec                                 | Description                                                                                              |
| :----------------------------------- | :------------------------------------------------------------------------------------------------------- |
| registry.type                        | Registry to mirror Falcon Container (allowed values: acr, ecr, crowdstrike, gcr, openshift). Default: crowdstrike                                                                                     |
| registry.tls.insecure_skip_verify    | (optional) Skip TLS check when pushing container image to target registry (only for demoing purposes on self-signed openshift clusters)                                                               |
| registry.tls.caCertificate           | (optional) A string containing an optionally base64-encoded Certificate Authority Chain for self-signed TLS Registry Certificates                                                                     |
| registry.tls.caCertificateConfigMap  | (optional) The name of a ConfigMap containing CA Certificate Authority Chains under keys ending in ".tls"  for self-signed TLS Registry Certificates (ignored when registry.tls.caCertificate is set) |
| registry.acr_name                    | (optional) Name of ACR for the container image push. Only applicable to Azure cloud. (`registry.type="acr"`)

> [!IMPORTANT]
> Registry settings are only currently supported by FalconAdmission, FalconContainer, and FalconImageAnalyzer. Registry configurations for FalconNodeSensor are configured in `falconNodeSensor.node`

#### FalconOperator Configuration Settings
The additional configurations for `falconNodeSensor`, `imageAnalyzer`, `falconContainer`, `falconAdmission` are all mapped to the Spec for each of the CRDs. 
- For falconAdmission, see: [falconadmission-reference-manual](https://github.com/CrowdStrike/falcon-operator/blob/main/docs/resources/admission/README.md#falconadmission-reference-manual)
- For imageAnalyzer, see: [falconimageanalyzer-reference-manual](https://github.com/CrowdStrike/falcon-operator/blob/main/docs/resources/imageanalyzer/README.md#falconimageanalyzer-reference-manual)
- For falconContainer, see: [falconcontainer-reference-manual](https://github.com/CrowdStrike/falcon-operator/blob/main/docs/resources/container/README.md#falconcontainer-reference-manual)
- For falconNodeSensor, see: [falconnodesensor-reference-manual](https://github.com/CrowdStrike/falcon-operator/blob/main/docs/resources/node/README.md#falconnodesensor-reference-manual)
> [!NOTE]
> Any required values for a level within the Spec will be enforced, e.g., if `falconAdmission.falcon_api.client_id` is assigned, then `falconAdmission.falcon_api.client_secret` will also be required.

| Spec                       | Description                                                                                                                                    |
| :------------------------- | :--------------------------------------------------------------------------------------------------------------------------------------------- |
| deployImageAnalyzer        | (Optional) Boolean to deploy the Image Analyzer. Default: True                                                                                 |
| deployAdmissionController  | (Optional) Boolean to deploy the Admission Controller. Default: True                                                                           |
| deployNodeSensor           | (Optional) Boolean to deploy Falcon Node Sensor. Default True                                                                                  |
| deployFalconContainer      | (Optional) Boolean to deploy Falcon Container. Default: False                                                                                  |
| falconNodeSensor           | (Optional) Additional configurations that map to FalconNodeSensorSpec. All values within the custom resource spec can be overridden here.      | 
| imageAnalyzer              | (Optional) Additional configurations that map to FalconImageAnalyzerSpec. All values within the custom resource spec can be overridden here.   | 
| falconContainer            | (Optional) Additional configurations that map to FalconContainerSpec. All values within the custom resource spec can be overridden here.       |
| falconAdmission            | (Optional) Additional configurations that map to FalconAdmissionConfigSpec. All values within the custom resource spec can be overridden here. |


#### Deployment Manifest Examples
Example of a minimal deployment:

```yaml
apiVersion: falcon.crowdstrike.com/v1alpha1
kind: FalconOperator
metadata:
  name: falcon-operator
spec:
  falcon_api:
    client_id: PLEASE_FILL_IN
    client_secret: PLEASE_FILL_IN
    cloud_region: <cloud region>
```

Example of a deployment containing custom configurations for each resource:

```yaml
apiVersion: falcon.crowdstrike.com/v1alpha1
kind: FalconOperator
metadata:
 name: falcon-operator
spec:
 falcon_api:
   client_id: PLEASE_FILL_IN
   client_secret: PLEASE_FILL_IN
   cloud_region: PLEASE_FILL_IN
 deployAdmissionController: true
 deployNodeSensor: true
 deployImageAnalyzer: true
 deployContainerSensor: true
 falconNodeSensor:
   installNamespace: falcon-node
   node:
     imagePullPolicy: IfNotPresent
   falcon:
     trace: warn
 falconAdmission:
   installNamespace: falcon-kac
   admissionConfig:
     resources:
       limits:
         memory: "1Gi"
     deployWatcher: false
   falcon:
     trace: warn
 falconImageAnalyzer:
   installNamespace: falcon-iar
   imageAnalyzerConfig:
     imagePullPolicy: IfNotPresent
     resources:
       limits:
         memory: "1Gi"
 falconContainerSensor:
   installNamespace: falcon-sidecar
   falcon:
     trace: warn
   injector:
     replicas: 3
     resources:
       limits:
         memory: "1Gi"
```
> [!NOTE]  
> Multiple restarts for the Falcon Admission Controller Pods may occur when deploying alongside other resources. Falcon KAC is designed to ignore namespaces managed by CrowdStrike, so as new resources are added, such as Falcon Container or Falcon Node Sensor, the KAC pod will redeploy to ignore the new namespaces.

### Install Steps
To install Falcon Admission Controller, run the following command to install the FalconAdmission CR:
```sh
oc create -f https://raw.githubusercontent.com/crowdstrike/falcon-operator/main/config/samples/falcon_v1alpha1_falconoperator.yaml --edit=true
```

### Uninstall Steps
To uninstall FalconOperator simply remove the FalconOperator resource. The operator will uninstall the FalconOperator resource and any CRs deployed from the cluster.

```sh
oc delete falconoperator --all
``` 

### Sensor upgrades

To upgrade the sensor version, simply add and/or update the `version` field in the FalconOperator resource and apply the change. Alternatively if the `image` field was used instead of using the Falcon API credentials, add and/or update the `image` field in the FalconAdmission resource and apply the change. The operator will detect the change and perform the upgrade.

### Troubleshooting

- You can get more insight by viewing the FalconOperator CRD in full detail by running the following command:

  ```sh
  oc get falconoperator -o yaml
  ```

- To review the logs of Falcon Operator:
  ```sh
  oc -n falcon-operator logs -f deploy/falcon-operator-controller-manager -c manager
  ```

### Additional Documentation
End-to-end guide(s) to install falcon-operator together with FalconAdmission resource.
 - [Deployment Guide for OpenShift](../../README.md)



