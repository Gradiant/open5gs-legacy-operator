# Open5GS Kubernetes Operator

## What is it?

The **Open5GS Kubernetes Operator** is a custom Kubernetes operator designed to automate the deployment, configuration, and lifecycle management of Open5GS and its subscribers in a declarative way. It uses two CRDs (Custom Resource Definitions): one for managing Open5GS deployments and another for managing Open5GS users, allowing efficient and automated core and subscriber management.

## Features

### Open5GS Deployment and Reconfiguration

Using the Helm plugin of the Operator SDK, the Open5GS Operator deploys the Open5GS Helm charts developed by Gradiant, enabling the reconfiguration of these deployments as needed. With the 2.2.2 version of the charts, it is possible to modify the configuration of individual Open5GS functions. The controller detects changes and only deletes and restarts the pod of the specific function that needs updating, without affecting the entire core. This is done through a mechanism that generates a hash from the `ConfigMap` contents, allowing the controller to apply updates efficiently while minimizing service downtime.

### Multi-Namespace Support

The operator handles multiple Open5GS deployments across different Kubernetes namespaces, ensuring resource isolation. It can also manage several Open5GS deployments within the same namespace, allowing independent management of each Open5GS instance.

### Open5GS Users Management

The operator provides full management of Open5GS subscribers, including configuration of network slices and the target Open5GS deployment to which they should be assigned. It distinguishes between **Managed Users** and **Unmanaged Users**:

- **Managed Users**: These are users whose IMSI is defined in a CR (Custom Resource). The operator controls their configuration, and any discrepancy between the actual state and the desired state in the CR will be detected as drift and corrected automatically, ensuring the configuration always aligns with the declarative source of truth.
  
- **Unmanaged Users**: These users are not controlled by the operator and are created externally (e.g., via scripts that directly modify the database or the Open5GS WebUI). Unmanaged users will not be altered by the operator, allowing compatibility with external tools and temporary deployments that don't need strict management by the operator.

## Prerequisites

- **Kubernetes**: A running Kubernetes cluster.
- **Helm** (optional): Only if you want to install the operator using the Helm chart. If you prefer not to use Helm, there is an alternative method to install it without Helm.

## Development Requirements

- **Kubernetes**: A running Kubernetes cluster between 1.25 and 1.28.
- **Operator SDK**: OperatorSDK 1.34.1 version
- **Go**: go 1.21.7 version

## How to Install

### Option 1: Installation using Helm (recommended)
   ```bash
  helm install open5gs-operator oci://registry-1.docker.io/gradiant/open5gs-legacy-operator-chart --version 0.1.0
   ```

#### Uninstall with Helm

Delete all the Open5GS and Open5GSUser resources and run:

```bash
   helm uninstall open5gs-operator 
   ```

### Option 2: Installation without Helm

If you prefer not to use Helm, you can apply the Kubernetes manifests directly or use the Makefile to install de CRD and deploy de operator.

```bash
   make deploy IMG=gradiant/open5gs-legacy-operator:0.1.0
   ```

#### Uninstall without Helm

Delete all the Open5GS and Open5GSUser resources and run:

```bash
   make undeploy
   ```

### Option 3: Run locally (outside the cluster)

This option is only recommended for development

```bash
   make install run
   ```

## How to Use

### Create an Open5GS Deployment

1. Create a deployment configuration file for Open5GS. Here’s a basic example:

    ```yaml
    apiVersion: net.gradiant.org/v1
    kind: Open5GS
    metadata:
      name: open5gs-1
    spec:
      hss:
        enabled: false
      mme:
        enabled: false
      pcrf:
        enabled: false
      smf:
        config:
          pcrf:
            enabled: false
      sgwc:
        enabled: false
      sgwu:
        enabled: false
      amf:
        config:
          guamiList:
            - plmn_id:
                mcc: "999"
                mnc: "70"
              amf_id:
                region: 2
                set: 1
          taiList:
            - plmn_id:
                mcc: "999"
                mnc: "70"
              tac: [1]
          plmnList:
            - plmn_id:
                mcc: "999"
                mnc: "70"
              s_nssai:
                - sst: 1
                  sd: "0x111111"
      nssf:
        config:
          nsiList:
            - uri: ""
              sst: 1
              sd: "0x111111"
      webui:
        enabled: false
      populate:
        enabled: false
    ```

    These values are the ones defined in the Gradiant Open5GS Helm Charts.

2. Apply the deployment file:

   ```bash
   kubectl apply -f open5gs-deployment.yaml
   ```

### Create Open5GS Users

1. Create a configuration file for the users you want to add. Here’s an example:

    ```yaml
    apiVersion: net.gradiant.org/v1
    kind: Open5GSUser
    metadata:
      name: open5gsuser-1
    spec:
      imsi: "999700000000001"
      key: "465B5CE8B199B49FAA5F0A2EE238A6BC"
      opc: "E8ED289DEBA952E4283B54E88E6183CA"
      sd: "111111"
      sst: "1"
      apn: "internet"
      open5gs: 
        name: "open5gs-1"
        namespace: "default"
    ```

    The apn, sst, and sd fields are optional. If they are not provided in the configuration, default values will be used by the system.

2. Apply the user configuration:

   ```bash
   kubectl apply -f open5gsuser-1.yaml
   ```
