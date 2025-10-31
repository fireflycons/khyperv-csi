# Container Storage Interface for Hyper-V

INCOMPLETE: Work in progress

| Feature                 | State                 |
|-------------------------|-----------------------|
| Windows REST Service    | :white_check_mark:    |
| Kubernetes CSI Workload | :construction_worker: |
| Helm Chart              | :x:                   |

This provides a CSI for mounting Hyper-V virtual hard disks as persistent volumes for Kubernetes clusters running on a single Hyper-V server. This includes the Hyper-V service running on your workstation/laptop if you have it. You do not require a dedicated Windows Server.

It comprises of the following components

* A Windows service that is installed on the Hyper-V server. This provides a REST interface to the disk subsystem that is utilised by the in-cluster provider workload. This is comprised of a Go wrapper round a PowerShell module that does the interaction with the disk subsystem.
* An in-cluster CSI provisioner which calls the REST service to provision and mount VHD disks.

## Security

* This should not be considered a production ready solution. It is intended for use by dev/test clusters running on a single Hyper-V server.
* The REST API may be served over HTTP, or HTTPS with either a self-signed or a provided certificate. If you choose self-signed, the service installer will generate this for you.
* The REST API is secured by a simple UUID API key, which is generated and displayed on the console when you install the Windows service, and should be passed to the Helm chart that installs the cluster components.

## Requirements

* A Windows machine running Hyper-V server on which you have a cluster running on Linux VMs
* The Linux VMs must have at least one SCSI adapter attached. This is where the PVs will be mounted.
* The Linux VMs must support the Hyper-V Data Exchange (KVP) service. If you've used AWS, this provides the same sort of information as the AWS metadata service `169.254.169.254`. You can verify this by checking the following at each node VM's terminal:
    * Directory `/var/lib/hyperv` exists.
    * This directory contains one or more files with names `.kvp_pool_X` where `X` is a number.

    Most modern distros support this out of the box and is provided by Linux Integration Services for Hyper-V (LIS). See [here](https://learn.microsoft.com/en-gb/windows-server/virtualization/hyper-v/Supported-Linux-and-FreeBSD-virtual-machines-for-Hyper-V-on-Windows).

    If your worker nodes do not have this, then you have the following options:

    * Rebuild the cluster completely on compatible VMs.
    * Deploy new nodes that do support the KVP service, migrate your workloads there and delete the old nodes.
    * See if it is possible to install the missing services for your kernel version and distro. Your favourite AI LLM can probably help with this.

## Installation

### 1. Install the REST Service

1. Extract `khypervprovisioner.exe` to a directory on your Hyper-V server
1. If you are going to use provided certificates, copy the certificate and key to the same directory.
1. Open a Windows PowerShell prompt in the directory where you copied the above files.
1. Run `.\kypervprovider install -h` to see available installation options.
1. Choose an installation method
    1. No SSL
        1. Execute `.\kypervprovider install`, optionally providing a port number if you don't want the default. API will be served over unsecured HTTP.
    1. Self signed certificates
        1. Execute `.\kypervprovider install --ssl`, optionally providing a port number if you don't want the default.
        1. Enter certificate information when prompted. Generated CA and server cert and key files will be output to the current directory.
    1. Provided certificate
        1. Execute `.\kypervprovider install --cert <path> --key <path>`, specifying the paths to the server cert and key files and optionally providing a port number if you don't want the default.

Unless you have specified a directory in which to store VHD files with `--directory`, when the service first starts, it will examine all local disks and pick the one with the most free space on which to store persistent volume disks. It will create a directory `Kubernetes Persistent Volumes` at the root of this disk. Volume provisioning will fail when the free space on the disk where the volume store has been created drops below 5GB. If specifying your own directory, it is advisable to pick one that is not the same directory being used by Hyper-V to store other VHDs, such as those created when you provisioned your cluster.

When the installation completes it will print the API key and the URL of the REST endpoint both of which will be required when installing the in-cluster provisioner.

You can verify the operation of the service by browsing its Swagger UI. Take the endpoint URL printed by the installation and paste to your browser.