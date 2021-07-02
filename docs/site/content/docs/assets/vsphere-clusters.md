## Create vSphere Clusters

This section describes setting up management and workload/guest clusters for
vSphere.

1. Download the machine image that matches the version of the Kubernetes you plan on deploying (1.20.1 is default).

    At this time, we cannot guarantee the plugin versions that will be used for cluster management.
    While using the kickstart UI to bootstrap your cluster, you may be asked to add an `OVA` to your vSphere environment. The following links are to the most recent OVAs at the time of writing this Getting Started guide.

    * [1.20.4
      OVA](http://build-squid.eng.vmware.com/build/mts/release/bora-17800251/publish/lin64/tkg_release/node/ova-photon-3-v1.20.4+vmware.1-tkg.0-2326554155028348692/photon-3-kube-v1.20.4+vmware.1-tkg.0-2326554155028348692.ova)
    * [1.19.8
      OVA](http://build-squid.eng.vmware.com/build/mts/release/bora-17759077/publish/lin64/tkg_release/node/ova-photon-3-v1.19.8+vmware.1-tkg.0-15338136437231643652/photon-3-kube-v1.19.8+vmware.1-tkg.0-15338136437231643652.ova)

    Note: To access the OVAs, you must have a VMware Customer Connect account. Complete the following steps to register a new account and access the OVAs:

    a. If you don't already have an account, register a new account on [VMware Customer Connect](http://my.vmware.com/).

    b. Log in to VMware Customer Connect, click on Products and Accounts > All Products.

    c. Search for "tanzu kubernetes grid", and from the search results, select "Tanzu Kubernetes Grid > Product Binaries > Tanzu Kubernetes Grid" to access the OVAs.
    <!--If you're asked for another `OVA` version by the kickstart UI, you can
    download the OVA that corresponds to the rc version (e.g. 1,2,3,etc) at the [TKG
    daily builds confluence page](https://confluence.eng.vmware.com/pages/viewpage.action?spaceKey=TKG&title=TKG+Release+Daily+Build#TKGReleaseDailyBuild-TKG1.3.0RC.3(March/09/2021)).-->

    Note: Validation work so far has focused on the Photon based images.

1. In vCenter, right-click on your datacenter and deploy the OVF template.

1. After importing, right-click and covert to a template.

1. Initialize the Tanzu kickstart UI.

    ```sh
    tanzu management-cluster create --ui
    ```

1. Go through the installation process for vSphere. With the following
   considerations:

   * Set all instance profile to large or larger.
     * In our testing, we found resource constraints caused bootstrapping
     issues. Choosing a large profile or more will give a better chance for
     successful bootstrapping.
   * Set your control plane IP
     * The control plane IP is a virtual IP that fronts the Kubernetes API
     server. You **must** set an IP that is routable and won't be taken by
     another system (e.g. DHCP).
   * Disable OIDC configuration.

    > Until we have more TCE documentation, you can find the full TKG docs
    > [here](https://docs.vmware.com/en/VMware-Tanzu-Kubernetes-Grid/1.2/vmware-tanzu-kubernetes-grid-12/GUID-mgmt-clusters-deploy-management-clusters.html).
    > We will have more complete `tanzu` cluster bootstrapping documentation available here in the near future.

1. Validate the management cluster started successfully.

    ```sh
    tanzu management-cluster get
    ```

1. Create a cluster name that will be used throughout this Getting Started guide. This instance of `MGMT_CLUSTER_NAME` should be set to whatever value is returned by `tanzu management-cluster get` above.

    ```sh
    export MGMT_CLUSTER_NAME="<INSERT_MGMT_CLUSTER_NAME_HERE>"
    export GUEST_CLUSTER_NAME="<INSERT_GUEST_CLUSTER_NAME_HERE>"
    ```

1. Capture the management cluster's kubeconfig.

    ```sh
    tanzu management-cluster kubeconfig get ${MGMT_CLUSTER_NAME} --admin

    Credentials of workload cluster 'mtce' have been saved
    You can now access the cluster by running 'kubectl config use-context mtce-admin@mtce'
    ```

    > Note the context name `${MGMT_CLUSTER_NAME}-admin@${MGMT_CLUSTER_NAME}`, you'll use the above command in
    > future steps.

1. Set your kubectl context to the management cluster.

    ```sh
    kubectl config use-context ${MGMT_CLUSTER_NAME}-admin@${MGMT_CLUSTER_NAME}
    ```

1. Validate you can access the management cluster's API server.

    ```sh
    kubectl get nodes

    NAME         STATUS   ROLES                  AGE    VERSION
    10-0-1-133   Ready    <none>                 123m   v1.20.1+vmware.2
    10-0-1-76    Ready    control-plane,master   125m   v1.20.1+vmware.2
    ```

1. Next you will create a guest cluster. First, setup a guest cluster config file.

    ```sh
    cp  ~/.tanzu/tkg/clusterconfigs/xw6nt8jduy.yaml ~/.tanzu/tkg/clusterconfigs/guest1.yaml
    ```

   > This step duplicates the configuration file that was created when you deployed your management cluster. This duplicated file will be used as the configuration file for your guest cluster. You can edit the parameters in this new  file as required. For an example of a guest cluster template, see  [vSphere Guest Cluster Template](../vsphere-wl-template).

   [](ignored)

   > In the next three steps you will edit the parameters in this new file (`guest1`) and then use the file to deploy a guest cluster.

   [](ignored)


1. In the new guest cluster file (`~/.tanzu/tkg/clusterconfigs/guest1.yaml`), edit the CLUSTER_NAME parameter to assign a name to your guest cluster. For example,

   ```yaml
   CLUSTER_CIDR: 100.96.0.0/11
   CLUSTER_NAME: my-guest-cluster
   CLUSTER_PLAN: dev
   ```
   #### Note
   * If you did not specify a name for your management cluster, the installer generated a unique name, in this case, you must manually add the CLUSTER_NAME parameter and assign a guest cluster name.
   * If you specified a name for your management cluster, the CLUSTER_NAME parameter is present and needs to be changed to the new guest cluster name.

1. In the guest cluster file (`~/.tanzu/tkg/clusterconfigs/guest1.yaml`), edit the VSPHERE_CONTROL_PLANE_ENDPOINT parameter to apply a viable IP.

   > This will be the API Server IP for your guest cluster. You must choose an IP that is routable and not used elsewhere in your network, e.g., out of your DHCP range.

   [](ignored)

   > The other parameters in ``guest1.yaml`` are likely fine as-is. However, you can change
   > them as required. Reference an example configuration template here:  [vSphere Guest Cluster Template](../vsphere-wl-template).

   > Validation is performed on the file prior to applying it, so the `tanzu` command will return a message if something necessary is omitted.

1. Create your guest cluster.

    ```sh
    tanzu cluster create ${GUEST_CLUSTER_NAME} --file ${HOME}/.tanzu/tkg/clusterconfigs/guest1.yaml
    ```

1. Validate the cluster starts successfully.

    ```sh
    tanzu cluster list
    ```

1. Capture the guest cluster's kubeconfig.

    ```sh
    tanzu cluster kubeconfig get ${GUEST_CLUSTER_NAME} --admin
    ```

1. Set your `kubectl` context accordingly.

    ```sh
    kubectl config use-context ${GUEST_CLUSTER_NAME}-admin@${GUEST_CLUSTER_NAME}
    ```

1. Verify you can see pods in the cluster.

    ```sh
    kubectl get pods --all-namespaces

    NAMESPACE     NAME                                                    READY   STATUS    RESTARTS   AGE
    kube-system   antrea-agent-9d4db                                      2/2     Running   0          3m42s
    kube-system   antrea-agent-vkgt4                                      2/2     Running   1          5m48s
    kube-system   antrea-controller-5d594c5cc7-vn5gt                      1/1     Running   0          5m49s
    kube-system   coredns-5d6f7c958-hs6vr                                 1/1     Running   0          5m49s
    kube-system   coredns-5d6f7c958-xf6cl                                 1/1     Running   0          5m49s
    kube-system   etcd-tce-guest-control-plane-b2wsf                      1/1     Running   0          5m56s
    kube-system   kube-apiserver-tce-guest-control-plane-b2wsf            1/1     Running   0          5m56s
    kube-system   kube-controller-manager-tce-guest-control-plane-b2wsf   1/1     Running   0          5m56s
    kube-system   kube-proxy-9825q                                        1/1     Running   0          5m48s
    kube-system   kube-proxy-wfktm                                        1/1     Running   0          3m42s
    kube-system   kube-scheduler-tce-guest-control-plane-b2wsf            1/1     Running   0          5m56s
    kube-system   kube-vip-tce-guest-control-plane-b2wsf                  1/1     Running   0          5m56s
    kube-system   vsphere-cloud-controller-manager-nwrg4                  1/1     Running   2          5m48s
    kube-system   vsphere-csi-controller-5b6f54ccc5-trgm4                 5/5     Running   0          5m49s
    kube-system   vsphere-csi-node-drnkz                                  3/3     Running   0          5m48s
    kube-system   vsphere-csi-node-flszf                                  3/3     Running   0          3m42s
    ```